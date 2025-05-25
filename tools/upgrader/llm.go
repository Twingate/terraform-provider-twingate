package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

const (
	llmApiKeyEnv   = "LLM_API_KEY"
	llmModel       = "deepseek-ai/DeepSeek-V3-0324"
	url            = "https://llm.chutes.ai/v1/chat/completions"
	guideURL       = "https://github.com/Twingate/terraform-provider-twingate/blob/main/docs/guides/migration-v2-to-v3-guide.md"
	promptTemplate = `
Please upgrade Terraform code snapshot from v2 to v3 folliwing the guide: %s

And please double check the following:
1. Modify only <b>access</b> block in resources <b>twingate_resource</b>
2. Use new blocks <b>access_group</b> and <b>access_service</b> accordingly
3. In case of multiple ids use dynamic block

'''
%s
'''
`
)

type Message struct {
	Role             string      `json:"role"`
	Content          string      `json:"content"`
	ReasoningContent interface{} `json:"reasoning_content"`
	ToolCalls        interface{} `json:"tool_calls"`
}

type Choice struct {
	Index        int         `json:"index"`
	Message      Message     `json:"message"`
	LogProbs     interface{} `json:"logprobs"`
	FinishReason string      `json:"finish_reason"`
	MatchedStop  int         `json:"matched_stop"`
}

type Usage struct {
	PromptTokens        int         `json:"prompt_tokens"`
	TotalTokens         int         `json:"total_tokens"`
	CompletionTokens    int         `json:"completion_tokens"`
	PromptTokensDetails interface{} `json:"prompt_tokens_details"`
}

type Response struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created int64    `json:"created"`
	Model   string   `json:"model"`
	Choices []Choice `json:"choices"`
	Usage   Usage    `json:"usage"`
}

func getAPIKey() string {
	apiKey := os.Getenv(llmApiKeyEnv)
	if apiKey == "" {
		panic(fmt.Sprintf("%s is not set", llmApiKeyEnv))
	}

	return apiKey
}

func callLLM(input string) string {
	apiKey := getAPIKey()

	jsonData, err := json.Marshal(map[string]interface{}{
		"model": llmModel,
		"messages": []interface{}{map[string]interface{}{
			"role":    "user",
			"content": fmt.Sprintf(promptTemplate, guideURL, input)},
		},
		"stream":      false,
		"max_tokens":  1024,
		"temperature": 0.7,
	})
	if err != nil {
		panic(fmt.Errorf("Failed to marshal request body: %w", err))
	}

	client := &http.Client{}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		panic(fmt.Errorf("Failed to create request: %w", err))
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		panic(fmt.Errorf("Failed to send request: %w", err))
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(fmt.Errorf("Failed to read response body: %w", err))
	}

	var response Response
	if err := json.Unmarshal(body, &response); err != nil {
		panic(fmt.Errorf("Failed to unmarshal response: %w", err))
	}

	var result string
	if content := response.Choices[0].Message.Content; content != "" {
		const startToken = "```terraform"
		const endToken = "```"

		if start := strings.Index(content, startToken); start != -1 {
			if end := strings.Index(content[start+len(startToken):], endToken); end != -1 {
				result = content[start+len(startToken) : start+len(startToken)+end]
			}
		}
	}

	result = strings.TrimSpace(result)
	if result == "" {
		panic("Failed to get result from LLM")
	}

	return result
}
