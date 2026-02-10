package llm

import (
	"context"
	"fmt"
	"strings"

	"github.com/sashabaranov/go-openai"
)

type Client struct {
	APIKey         string
	BaseURL        string
	ChatModel      string
	EmbeddingModel string
}

func NewClient(apiKey string) *Client {
	return &Client{
		APIKey:         apiKey,
		BaseURL:        "https://open.bigmodel.cn/api/paas/v4/",
		ChatModel:      "glm-4",
		EmbeddingModel: "embedding-3",
	}
}

func (c *Client) Call(ctx context.Context, prompt string) (string, error) {
	client := c.newOpenAIClient()
	resp, err := client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: c.ChatModel,
		Messages: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleUser, Content: prompt},
		},
	})
	if err != nil {
		return "", err
	}
	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("empty response")
	}
	return strings.TrimSpace(resp.Choices[0].Message.Content), nil
}

// CallStream returns a channel that streams the response chunks
func (c *Client) CallStream(ctx context.Context, prompt string) (<-chan string, <-chan error) {
	textChan := make(chan string)
	errChan := make(chan error, 1)

	go func() {
		defer close(textChan)
		defer close(errChan)

		client := c.newOpenAIClient()
		req := openai.ChatCompletionRequest{
			Model: c.ChatModel,
			Messages: []openai.ChatCompletionMessage{
				{Role: openai.ChatMessageRoleUser, Content: prompt},
			},
			Stream: true,
		}

		stream, err := client.CreateChatCompletionStream(ctx, req)
		if err != nil {
			errChan <- err
			return
		}
		defer stream.Close()

		for {
			response, err := stream.Recv()
			if err != nil {
				// io.EOF is normal end of stream, but for openai-go it returns io.EOF
				// We should check if it's the end.
				// Actually standard openai-go stream.Recv returns io.EOF when done.
				if err.Error() != "EOF" {
					errChan <- err
				}
				return
			}

			if len(response.Choices) > 0 {
				content := response.Choices[0].Delta.Content
				if content != "" {
					textChan <- content
				}
			}
		}
	}()

	return textChan, errChan
}

func (c *Client) Embed(ctx context.Context, text string) ([]float32, error) {
	client := c.newOpenAIClient()
	resp, err := client.CreateEmbeddings(ctx, openai.EmbeddingRequest{
		Input: []string{text},
		Model: openai.EmbeddingModel(c.EmbeddingModel),
	})
	if err != nil {
		return nil, err
	}
	if len(resp.Data) == 0 {
		return nil, fmt.Errorf("empty embedding")
	}
	return resp.Data[0].Embedding, nil
}

func (c *Client) newOpenAIClient() *openai.Client {
	config := openai.DefaultConfig(c.APIKey)
	if c.BaseURL != "" {
		config.BaseURL = c.BaseURL
	}
	return openai.NewClientWithConfig(config)
}
