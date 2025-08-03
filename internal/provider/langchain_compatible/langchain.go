package langchaincompatible

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/dmitrii/llm-gateway/internal/types"
	"github.com/openai/openai-go"
	"github.com/tmc/langchaingo/llms"
)

type LangchainProvider struct {
	model llms.Model
}

func NewLangchainProvider(model llms.Model) *LangchainProvider {
	return &LangchainProvider{
		model: model,
	}
}

func openaiMsgToLangchainMsg(msg openai.ChatCompletionMessageParamUnion) (llms.MessageContent, error) {
	llmsMsg := llms.MessageContent{}
	if msg.OfAssistant != nil {
		llmsMsg.Role = llms.ChatMessageTypeAI
	} else if msg.OfUser != nil {
		llmsMsg.Role = llms.ChatMessageTypeHuman
	} else if msg.OfSystem != nil {
		llmsMsg.Role = llms.ChatMessageTypeSystem
	} else if msg.OfFunction != nil {
		llmsMsg.Role = llms.ChatMessageTypeFunction
	} else if msg.OfTool != nil {
		llmsMsg.Role = llms.ChatMessageTypeTool
	} else if msg.OfDeveloper != nil {
		llmsMsg.Role = llms.ChatMessageTypeGeneric
	} else {
		return llms.MessageContent{}, fmt.Errorf("unknown message type: %T", msg)
	}

	switch v := msg.GetContent().AsAny().(type) {
	case *string:
		llmsMsg.Parts = []llms.ContentPart{
			llms.TextPart(*v),
		}
	case *[]openai.ChatCompletionContentPartTextParam:
		for _, part := range *v {
			llmsMsg.Parts = append(llmsMsg.Parts, llms.TextPart(part.Text))
		}
	case *[]openai.ChatCompletionContentPartUnionParam:
		for _, part := range *v {
			// TODO: handle other types
			if part.OfFile != nil {
				llmsMsg.Parts = append(llmsMsg.Parts, llms.BinaryPart(
					// TODO: where do i get content type from?
					"",
					[]byte(part.OfFile.File.FileData.Value),
				))
			} else if part.OfText != nil {
				llmsMsg.Parts = append(llmsMsg.Parts, llms.TextPart(part.OfText.Text))
			} else {
				return llms.MessageContent{}, fmt.Errorf("unsupported content part type: %T", part)
			}
		}
	case *[]openai.ChatCompletionAssistantMessageParamContentArrayOfContentPartUnion:
		for _, part := range *v {
			if part.OfRefusal != nil {
				llmsMsg.Parts = append(llmsMsg.Parts, llms.TextPart(part.OfRefusal.Refusal))
			} else if part.OfText != nil {
				llmsMsg.Parts = append(llmsMsg.Parts, llms.TextPart(part.OfText.Text))
			} else {
				return llms.MessageContent{}, fmt.Errorf("unsupported content part type: %T", part)
			}
		}
	default:
		return llms.MessageContent{}, fmt.Errorf("unsupported message content type: %T", v)
	}

	return llmsMsg, nil
}

func openaiOptionsToLangchainOptions(opts openai.ChatCompletionNewParams) []llms.CallOption {
	options := []llms.CallOption{
		llms.WithModel(opts.Model),
	}

	if opts.FrequencyPenalty.Valid() {
		options = append(options, llms.WithFrequencyPenalty(opts.FrequencyPenalty.Value))
	}
	if opts.PresencePenalty.Valid() {
		options = append(options, llms.WithPresencePenalty(opts.PresencePenalty.Value))
	}
	if opts.MaxTokens.Valid() {
		options = append(options, llms.WithMaxTokens(int(opts.MaxTokens.Value)))
	}
	if opts.Temperature.Valid() {
		options = append(options, llms.WithTemperature(opts.Temperature.Value))
	}
	if opts.TopP.Valid() {
		options = append(options, llms.WithTopP(opts.TopP.Value))
	}
	if opts.N.Valid() {
		options = append(options, llms.WithN(int(opts.N.Value)))
	}
	if len(opts.Stop.OfStringArray) > 0 {
		options = append(options, llms.WithStopWords(opts.Stop.OfStringArray))
	}

	return options
}

func (p *LangchainProvider) ChatCompletion(ctx context.Context, req *openai.ChatCompletionNewParams) (*types.ChatCompletionResponse, error) {
	options := openaiOptionsToLangchainOptions(*req)

	messages := make([]llms.MessageContent, len(req.Messages))
	for i, msg := range req.Messages {
		llmsMsg, err := openaiMsgToLangchainMsg(msg)
		if err != nil {
			return nil, fmt.Errorf("failed to convert OpenAI message to Langchain message: %w", err)
		}
		messages[i] = llmsMsg
	}

	// Call the Langchain model
	langchainResp, err := p.model.GenerateContent(ctx, messages, options...)
	if err != nil {
		return nil, fmt.Errorf("failed to generate content: %w", err)
	}

	// convert the response to the types.ChatCompletionResponse format
	res := types.ChatCompletionResponse{
		Choices: make([]types.ChatCompletionChoice, len(langchainResp.Choices)),
	}
	for i, choice := range langchainResp.Choices {
		res.Choices[i] = types.ChatCompletionChoice{
			Index: i,
			Message: types.ChatMessage{
				Role:    "assistant",
				Content: choice.Content,
			},
			FinishReason: choice.StopReason,
		}
		if choice.GenerationInfo != nil {
			complTokens, ok := choice.GenerationInfo["CompletionTokens"].(int)
			if !ok {
				slog.Warn("invalid type for CompletionTokens", "type", fmt.Sprintf("%T", choice.GenerationInfo["CompletionTokens"]))
			} else {
				res.Usage.CompletionTokens = complTokens
			}

			promptTokens, ok := choice.GenerationInfo["PromptTokens"].(int)
			if !ok {
				slog.Warn("invalid type for PromptTokens", "type", fmt.Sprintf("%T", choice.GenerationInfo["PromptTokens"]))
			} else {
				res.Usage.PromptTokens = promptTokens
			}

			totalTokens, ok := choice.GenerationInfo["TotalTokens"].(int)
			if !ok {
				slog.Warn("invalid type for TotalTokens", "type", fmt.Sprintf("%T", choice.GenerationInfo["TotalTokens"]))
			} else {
				res.Usage.TotalTokens = totalTokens
			}
		}
	}
	return &res, nil
}
