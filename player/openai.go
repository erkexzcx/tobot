package player

import (
	"context"
	"log"

	"github.com/sashabaranov/go-openai"
)

// Set this variable before using this package
var (
	OPENAI_API_KEY      string
	OPENAI_INSTRUCTIONS string
)

var openai_client *openai.Client

func getAIReply(msg string) string {
	resp, err := openai_client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT3Dot5Turbo,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: OPENAI_INSTRUCTIONS,
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: msg,
				},
			},
		},
	)

	if err != nil {
		log.Println("User msg:", msg)
		log.Printf("ChatCompletion error: %v\n", err)
		return ""
	}

	return resp.Choices[0].Message.Content
}

func init() {
	openai_client = openai.NewClient(OPENAI_API_KEY)
}
