package langchaincompatible

import (
	"context"
	"fmt"

	"github.com/dmitrii/llm-gateway/internal/types"
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

var rolesMapping = map[string]llms.ChatMessageType{
	"system":    llms.ChatMessageTypeSystem,
	"user":      llms.ChatMessageTypeHuman,
	"assistant": llms.ChatMessageTypeAI,
	"tool":      llms.ChatMessageTypeTool,
	"function":  llms.ChatMessageTypeFunction,
}

func (p *LangchainProvider) ChatCompletion(ctx context.Context, req *types.ChatCompletionRequest) (*types.ChatCompletionResponse, error) {
	options := []llms.CallOption{}
	options = append(options,
		llms.WithFrequencyPenalty(req.FrequencyPenalty),
		llms.WithPresencePenalty(req.PresencePenalty),
		llms.WithMaxTokens(req.MaxTokens),
		llms.WithTemperature(req.Temperature),
		llms.WithTopP(req.TopP),
		llms.WithN(req.N),
		llms.WithStopWords(req.Stop),
	)

	messages := make([]llms.MessageContent, len(req.Messages))
	for i, msg := range req.Messages {
		role, ok := rolesMapping[msg.Role]
		if !ok {
			return nil, fmt.Errorf("unknown role %q in message %d", msg.Role, i)
		}
		messages[i] = llms.MessageContent{
			Role: role,
			Parts: []llms.ContentPart{
				// TODO: think about other content types
				llms.TextPart(msg.Content),
			},
		}
	}

	// Call the Langchain model
	langchainResp, err := p.model.GenerateContent(ctx, messages, options...)
	if err != nil {
		return nil, err
	}

	// convert the response to the types.ChatCompletionResponse format
	choices := make([]types.ChatCompletionChoice, len(langchainResp.Choices))
	for i, choice := range langchainResp.Choices {
		choices[i] = types.ChatCompletionChoice{
			Index: i,
			Message: types.ChatMessage{
				Role:    "assistant",
				Content: choice.Content,
			},
			FinishReason: choice.StopReason,
		}
	}
	// TODO: handle usage stats if available + other fields like FunctionCall and ToolCalls

	return &types.ChatCompletionResponse{
		Choices: choices,
	}, nil
}
