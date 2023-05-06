package comms

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/sashabaranov/go-openai"
)

var openaiClient *openai.Client

type OpenaiMessage struct {
	Received bool
	Message  string
}

func GetOpenAIReply(messages ...*OpenaiMessage) string {
	if len(messages) == 0 {
		errMsg := "No messages provided to GetOpenAIReply"
		SendMessageToTelegram(errMsg)
		log.Fatalln(errMsg)
	}

	// Create an "empty" chat with initial system message
	openaiMessages := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: strings.TrimSpace(appConfig.OpenAI.Instructions),
		},
	}

	// Add given messages to the chat
	for _, m := range messages {
		messageRole := openai.ChatMessageRoleUser
		if !m.Received {
			messageRole = openai.ChatMessageRoleAssistant
		}

		openaiMessages = append(openaiMessages, openai.ChatCompletionMessage{
			Role:    messageRole,
			Content: strings.TrimSpace(m.Message),
		})
	}

	// Request AI reply
	resp, err := openaiClient.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model:       appConfig.OpenAI.Model,
			Temperature: appConfig.OpenAI.Temperature,
			Messages:    openaiMessages,
		},
	)
	if err != nil {
		errMsg := fmt.Sprintf("Failed to retrieve message from OpenAI (ChatCompletion error): %v\n", err)
		SendMessageToTelegram(errMsg)
		log.Fatalln(errMsg)
	}

	// Return AI reply
	return resp.Choices[0].Message.Content
}
