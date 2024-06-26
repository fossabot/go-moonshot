package moonshot

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/northes/go-moonshot/enum"
	"github.com/northes/gox/httpx"
)

type chat struct {
	client *httpx.Client
}

func (c *Client) Chat() *chat {
	return &chat{
		client: c.newHTTPClient(),
	}
}

type ChatCompletionsRequest struct {
	Messages         []*ChatCompletionsMessage `json:"messages"`
	Model            enum.ChatCompletionsModelID
	MaxTokens        int64
	Temperature      float64
	TopP             float64
	N                int64
	PresencePenalty  float64
	FrequencyPenalty float64
	Stop             []string
	Stream           bool
}

type ChatCompletionsResponse struct {
	Id      string                            `json:"id"`
	Object  string                            `json:"object"`
	Created int                               `json:"created"`
	Model   string                            `json:"model"`
	Choices []*ChatCompletionsResponseChoices `json:"choices"`
	Usage   *ChatCompletionsResponseUsage     `json:"usage"`
}

type ChatCompletionsResponseChoices struct {
	Index int `json:"index"`

	Message *ChatCompletionsMessage `json:"message,omitempty"`
	Delta   *ChatCompletionsMessage `json:"delta,omitempty"`

	FinishReason enum.ChatCompletionsFinishReason `json:"finish_reason"`
}

type ChatCompletionsResponseUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type ChatCompletionsMessage struct {
	Role    enum.ChatCompletionsMessageRole `json:"role"`
	Content string                          `json:"content"`
}

func (c *chat) Completions(ctx context.Context, req *ChatCompletionsRequest) (*ChatCompletionsResponse, error) {
	const path = "/v1/chat/completions"
	req.Stream = false
	chatCompletionsResp := new(ChatCompletionsResponse)
	resp, err := c.client.AddPath(path).SetBody(req).Post()
	if err != nil {
		return chatCompletionsResp, err
	}
	if !resp.StatusOK() {
		return nil, StatusCodeToError(resp.Raw().StatusCode)
	}
	err = resp.Unmarshal(chatCompletionsResp)
	if err != nil {
		return nil, err
	}
	return chatCompletionsResp, nil
}

func (c *chat) CompletionsStream(ctx context.Context, req *ChatCompletionsRequest, respCh chan<- *ChatCompletionsResponse, done chan<- struct{}) error {
	const path = "/v1/chat/completions"

	if respCh == nil || done == nil {
		return errors.New("chat completions streaming requests must have a non-nil channel")
	}

	req.Stream = true

	resp, err := c.client.AddPath(path).SetBody(req).Post()
	if err != nil {
		return err
	}
	if !resp.StatusOK() {
		return StatusCodeToError(resp.Raw().StatusCode)
	}
	defer func() {
		_ = resp.Raw().Body.Close()
	}()

	reader := bufio.NewReader(resp.Raw().Body)
	for {
		line, err := reader.ReadBytes('\n')
		//fmt.Printf("next line: %v\n", string(line))
		if err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("error reading response body line: %w", err)
		}

		prefix := []byte("data: ")

		if !bytes.HasPrefix(line, prefix) {
			//fmt.Println("no hava prefix,continue")
			continue
		}

		line = bytes.TrimPrefix(bytes.TrimSpace(line), prefix)

		if string(line) == "[DONE]" {
			break
		}

		//fmt.Printf("trim prefix line: %v %v %v\n", string(line), "[DONE]", string(line) == "[DONE]")

		rr := ChatCompletionsResponse{}
		err = json.Unmarshal(line, &rr)
		if err != nil {
			return fmt.Errorf("error unmarshalling response body line: %w", err)
		}
		respCh <- &rr
	}

	done <- struct{}{}

	return nil
}

func (i *ChatCompletionsResponseChoices) IsFinishStop() bool {
	return i.FinishReason == enum.FinishReasonStop
}

func (i *ChatCompletionsResponseChoices) IsFinishLength() bool {
	return i.FinishReason == enum.FinishReasonLength
}

func (c *ChatCompletionsResponse) CanGetContent() bool {
	for _, choice := range c.Choices {
		if choice.FinishReason == enum.FinishReasonStop {
			return false
		}
		if choice.Message != nil {
			if len(choice.Message.Content) == 0 {
				return false
			}
		}
		if choice.Delta != nil {
			if len(choice.Delta.Content) == 0 {
				return false
			}
		}
	}
	return true
}
