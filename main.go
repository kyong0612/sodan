package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type OpenAIResponse struct {
	Choices []struct {
		Text string `json:"text"`
	} `json:"choices"`
}

const Endpoint = "https://api.openai.com/v1/engines/davinci-codex/completions"

func main() {

	godotenv.Load(".env")

	// 環境変数からAPIキーを取得
	APIKey := os.Getenv("OPENAI_API_KEY")
	if APIKey == "" {
		// 環境変数が設定されていない場合、ユーザーにAPIキーを入力させる
		fmt.Print("Enter your OpenAI API key: ")
		reader := bufio.NewReader(os.Stdin)
		APIKey, _ = reader.ReadString('\n')
		APIKey = strings.TrimSpace(APIKey)
	}

	// 環境変数から保存ファイル名を取得
	filename := os.Getenv("CONVERSATION_FILE")
	if filename == "" {
		// 環境変数が設定されていない場合、ユーザーにファイル名を入力させる
		fmt.Print("Enter the filename for the conversation (e.g., conversation.md): ")
		reader := bufio.NewReader(os.Stdin)
		filename, _ = reader.ReadString('\n')
		filename = strings.TrimSpace(filename)
	}

	// ユーザー入力を受け付けるためのリーダーを初期化
	reader := bufio.NewReader(os.Stdin)

	// 対話内容を保存するための変数
	var conversation []string

	for {
		// ユーザーからの入力を受け取る
		fmt.Print("You: ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		// 入力を対話内容に追加
		conversation = append(conversation, fmt.Sprintf("You: %s\n", input))

		// GPT-4に対話内容を送信し、応答を取得
		responseText, err := sendRequest(APIKey, strings.Join(conversation, ""))
		if err != nil {
			fmt.Println("Error:", err)
			continue
		}

		// GPT-4からの応答を表示
		fmt.Printf("GPT-4: %s\n", responseText)

		// 応答を対話内容に追加
		conversation = append(conversation, fmt.Sprintf("GPT-4: %s\n", responseText))

		// 対話内容をファイルに書き込む
		err = ioutil.WriteFile(filename, []byte(strings.Join(conversation, "")), 0644)
		if err != nil {
			fmt.Println("Error:", err)
		}
	}
}

func sendRequest(APIKey, prompt string) (string, error) {
	// リクエストデータを作成
	data := fmt.Sprintf(`{"prompt": "%s", "max_tokens": 100}`, prompt)
	req, _ := http.NewRequest("POST", Endpoint, strings.NewReader(data))

	// APIキーを追加
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", APIKey))
	req.Header.Add("Content-Type", "application/json")

	// リクエストを送信
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// 応答データを解析
	body, _ := ioutil.ReadAll(resp.Body)
	var response OpenAIResponse
	json.Unmarshal(body, &response)

	// 応答のテキストを返す
	return strings.TrimSpace(response.Choices[0].Text), nil
}
