package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/joho/godotenv"
)

// see API-doc: https://platform.openai.com/docs/api-reference/models/list
// see secret: https://platform.openai.com/account/api-keys
type Model string

const (
	Turbo Model = "gpt-3.5-turbo"
)

func (m Model) String() string {
	return string(m)
}

type Role string

const (
	System    Role = "system"
	User      Role = "user"
	Assistant Role = "assistant"
)

// https://platform.openai.com/docs/guides/chat/instructing-chat-models
type Message struct {
	Role    `json:"role"`
	Content string `json:"content"`
}

type OpenAIRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Prompt      string    `json:"prompt,omitempty"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
	Temperature float64   `json:"temperature,omitempty"`
}

type OpenAIResponse struct {
	Index   int    `json:"index"`
	Object  string `json:"object"`
	Created int    `json:"created"`
	Choices []struct {
		// Text  string `json:"text"`
		Index int `json:"index"`
		// Logprobs struct {
		// } `json:"logprobs"`
		Message      `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

const Endpoint = "https://api.openai.com/v1/chat/completions"

func main() {

	godotenv.Load(".env")

	APIKey := os.Getenv("OPENAI_API_KEY")
	if APIKey == "" {
		fmt.Print("Enter your OpenAI API key: ")
		reader := bufio.NewReader(os.Stdin)
		APIKey, _ = reader.ReadString('\n')
		APIKey = strings.TrimSpace(APIKey)
	}

	// 環境変数から保存ファイル名を取得
	filename := os.Getenv("CONVERSATION_FILE")
	if filename == "" {
		fmt.Print("Enter the filename for the conversation (e.g., conversation.md): ")
		reader := bufio.NewReader(os.Stdin)
		filename, _ = reader.ReadString('\n')
		filename = strings.TrimSpace(filename)
	}

	path := strings.Split(filename, "/")
	if len(path) > 1 {
		filename = strings.Join(path[:len(path)-1], "/") + "/" + time.Now().Format("2006-01-02-15:04:05-") + path[len(path)-1]
	} else {
		filename = time.Now().Format("2006-01-02-15:04:05-") + filename
	}

	reader := bufio.NewReader(os.Stdin)

	var conversation []string

	model := Turbo
	me := "You"
	format := fmt.Sprintf("%%%ds", utf8.RuneCountInString(model.String())-utf8.RuneCountInString(me)+3)
	me = fmt.Sprintf(format, me)

	fmt.Print("\nWelcome to the OpenAI chatbot demo!\n\n")
	for {
		fmt.Print(me + ": ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		conversation = append(conversation, fmt.Sprintf(me+": %s\n", input))

		responseText, err := sendRequest(APIKey, model, "", strings.Join(conversation, ""))
		if err != nil {
			fmt.Println("Error:", err)
			continue
		}

		fmt.Printf("%s: %s\n", model, responseText)

		conversation = append(conversation, fmt.Sprintf("%s: %s\n", model, responseText))

		err = ioutil.WriteFile(filename, []byte(strings.Join(conversation, "")), 0644)
		if err != nil {
			fmt.Println("Error:", err)
		}
	}
}

func sendRequest(APIKey string, model Model, prompt, msg string) (string, error) {
	data, err := json.Marshal(OpenAIRequest{
		Model: string(Turbo),
		Messages: []Message{
			{
				Role:    User,
				Content: msg,
			},
		},
		Prompt:      prompt,
		Temperature: 0.2,
	})
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", Endpoint, bytes.NewBuffer(data))
	if err != nil {
		return "", err
	}

	// APIキーを追加
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", APIKey))
	req.Header.Add("Content-Type", "application/json")

	// r, err := httputil.DumpRequest(req, true)
	// if err != nil {
	// 	return "", err
	// }
	// fmt.Printf("DEBUG:Request:\n%s\n\n", string(r))

	// リクエストを送信
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// r, err := httputil.DumpResponse(resp, true)
	// if err != nil {
	// 	return "", err
	// }
	// fmt.Printf("DEBUG:Response:\n%s\n\n", string(r))

	body, _ := ioutil.ReadAll(resp.Body)
	var response OpenAIResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return "", err
	}

	// rally := len(response.Choices)

	if len(response.Choices) == 0 {
		return "", fmt.Errorf("no response")
	}

	if strings.Contains(response.Choices[0].Message.Content, ":") {
		response.Choices[0].Message.Content = strings.Split(response.Choices[0].Message.Content, ":")[1]
	}

	// 応答のテキストを返す
	return strings.TrimSpace(response.Choices[0].Message.Content), nil
}
