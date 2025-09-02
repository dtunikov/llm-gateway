package langchaincompatible

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/dmitrii/llm-gateway/api"
	"github.com/dmitrii/llm-gateway/internal/errors"
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

func openaiMsgToLangchainMsg(msg *api.ChatMessage) (llms.MessageContent, error) {
	llmsMsg := llms.MessageContent{}
	switch msg.Role {
	case api.ChatMessageRoleUser:
		llmsMsg.Role = llms.ChatMessageTypeHuman
	case api.ChatMessageRoleAssistant:
		llmsMsg.Role = llms.ChatMessageTypeAI
	case api.ChatMessageRoleSystem:
		llmsMsg.Role = llms.ChatMessageTypeSystem
	case api.ChatMessageRoleFunction:
		llmsMsg.Role = llms.ChatMessageTypeFunction
	case api.ChatMessageRoleTool:
		llmsMsg.Role = llms.ChatMessageTypeTool
	default:
		return llms.MessageContent{}, errors.ErrNotFound.WithMessage(fmt.Sprintf("unknown chat message role: %s", msg.Role))
	}

	var contentParts []api.MessageContentPart
	contentString, err := msg.Content.AsChatMessageContent0()
	if err != nil {
		contentParts, err = msg.Content.AsChatMessageContent1()
		if err != nil {
			return llms.MessageContent{}, fmt.Errorf("failed to convert content: %w", err)
		}
	}
	if len(contentParts) > 0 {
		llmsMsg.Parts = make([]llms.ContentPart, len(contentParts))
		for i, part := range contentParts {
			if part.Text != nil {
				llmsMsg.Parts[i] = llms.TextPart(*part.Text)
			} else if part.ImageUrl != nil {
				llmsMsg.Parts[i] = llms.ImageURLPart(part.ImageUrl.Url)
			}
		}
		return llmsMsg, nil
	} else {
		llmsMsg.Parts = []llms.ContentPart{
			llms.TextPart(contentString),
		}
	}

	return llmsMsg, nil
}

func openaiOptionsToLangchainOptions(req *api.ChatCompletionRequest) ([]llms.CallOption, error) {
	options := []llms.CallOption{
		llms.WithModel(req.Model),
	}

	if req.FrequencyPenalty != nil {
		options = append(options, llms.WithFrequencyPenalty(float64(*req.FrequencyPenalty)))
	}
	if req.PresencePenalty != nil {
		options = append(options, llms.WithPresencePenalty(float64(*req.PresencePenalty)))
	}
	if req.MaxTokens != nil {
		options = append(options, llms.WithMaxTokens(int(*req.MaxTokens)))
	}
	if req.Temperature != nil {
		options = append(options, llms.WithTemperature(float64(*req.Temperature)))
	}
	if req.TopP != nil {
		options = append(options, llms.WithTopP(float64(*req.TopP)))
	}
	if req.N != nil {
		options = append(options, llms.WithN(int(*req.N)))
	}
	if req.Stop != nil {
		stopArr, err := req.Stop.AsChatCompletionRequestStop1()
		if err != nil {
			stopWord, err := req.Stop.AsChatCompletionRequestStop0()
			if err != nil {
				return nil, fmt.Errorf("failed to convert stop words: %w", err)
			}
			stopArr = []string{stopWord}
		}
		options = append(options, llms.WithStopWords(stopArr))

	}

	return options, nil
}

func (p *LangchainProvider) ChatCompletion(ctx context.Context, req *api.ChatCompletionRequest) (*api.ChatCompletionResponse, error) {
	options, err := openaiOptionsToLangchainOptions(req)
	if err != nil {
		return nil, fmt.Errorf("failed to convert OpenAI options to Langchain options: %w", err)
	}

	messages := make([]llms.MessageContent, len(req.Messages))
	for i, msg := range req.Messages {
		llmsMsg, err := openaiMsgToLangchainMsg(&msg)
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
	res := api.ChatCompletionResponse{
		Choices: make([]api.ChatCompletionChoice, len(langchainResp.Choices)),
	}
	for i, choice := range langchainResp.Choices {
		converted := api.ChatCompletionChoice{
			Index:        i,
			FinishReason: api.ChatCompletionChoiceFinishReason(choice.StopReason),
		}

		if choice.FuncCall != nil {
			converted.Message.Role = api.ChatMessageRoleFunction
			converted.Message.FunctionCall = &api.FunctionCall{
				Name:      choice.FuncCall.Name,
				Arguments: choice.FuncCall.Arguments,
			}
		} else if len(choice.ToolCalls) > 0 {
			converted.Message.Role = api.ChatMessageRoleTool
			calls := make([]api.ToolCall, len(choice.ToolCalls))
			for j, toolCall := range choice.ToolCalls {
				calls[j] = api.ToolCall{
					Id:   toolCall.ID,
					Type: api.ToolCallType(toolCall.Type),
					Function: api.FunctionCall{
						Name:      toolCall.FunctionCall.Name,
						Arguments: toolCall.FunctionCall.Arguments,
					},
				}
			}
			converted.Message.ToolCalls = &calls
		} else {
			converted.Message.Role = api.ChatMessageRoleAssistant
		}
		if len(choice.Content) > 0 {
			content := &api.ChatMessage_Content{}
			content.FromChatMessageContent0(choice.Content)
			converted.Message.Content = content
		}

		res.Choices[i] = converted
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
