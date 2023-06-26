package main

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"github.com/briandowns/spinner"
	"github.com/sashabaranov/go-openai"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"
	"unicode"
)

type ChatGPTClient struct {
	client *openai.Client
}

func NewChatGPTClient(apiKey string) *ChatGPTClient {
	return &ChatGPTClient{
		client: openai.NewClient(apiKey),
	}
}

func (c *ChatGPTClient) GetAnswer(question string) (string, error) {
	req := openai.CompletionRequest{
		Model:       DefaultModel,
		MaxTokens:   DefaultMaxTokens,
		Prompt:      question,
		Temperature: DefaultTemperature,
	}
	fmt.Println("Sending request to OpenAI API...")
	// question
	fmt.Println("Question: ", question)

	resp, err := c.client.CreateCompletion(context.Background(), req)
	if err != nil {
		return "", err
	}

	if len(resp.Choices) == 0 {
		return "", errors.New("empty response from the API")
	}

	return resp.Choices[0].Text, nil
}

func run(diff string) error {
	const (
		maxDiffTokens    = 400
		maxPromptTokens  = 400
		customMessageOpt = "Enter a custom message"
	)

	// Trim the diff to the maximum number of tokens
	diffTokens := len(strings.Fields(diff))
	if diffTokens > maxDiffTokens {
		diff = strings.Join(strings.Fields(diff)[:maxPromptTokens], " ")
	}
	config, err := LoadGlobalConfig()
	if err != nil {
		fmt.Println("Failed to load config:", err)
		os.Exit(1)
	}

	apiKey := config.APIKey
	if apiKey == "" {
		var err error
		apiKey, err := PromptToken()
		if err != nil {
			fmt.Println("Failed to get API key:", err)
			os.Exit(1)
		}
		err = SetAPIKey(apiKey)
		if err != nil {
			fmt.Println("Failed to set API key:", err)
			os.Exit(1)
		}
	}

	api := NewChatGPTClient(apiKey)
	// load prompt from file
	// decpde  base 64 to string
	decode, err := base64.StdEncoding.DecodeString(DefaultPromptTemplate)
	prompt := strings.ReplaceAll(string(decode), "{{diff}}", "```\n"+diff+"\n```")

	for {
		choices, err := getMessages(api, prompt)
		if err != nil {
			return err
		}
		choices = append(choices, customMessageOpt)

		var message string
		prompt := &survey.Select{
			Message: "Pick a message",
			Options: choices,
		}
		if err = survey.AskOne(prompt, &message); err != nil {
			return err
		}

		if message == customMessageOpt {
			cmd := exec.Command("git", "commit")
			cmd.Stdout, cmd.Stdin, cmd.Stderr = os.Stdout, os.Stdin, os.Stderr
			if err = cmd.Run(); err != nil {
				return err
			}
			return nil
		}

		cmd := exec.Command("git", "commit", "-m", escapeCommitMessage(message))
		cmd.Stdout, cmd.Stdin, cmd.Stderr = os.Stdout, os.Stdin, os.Stderr
		if err = cmd.Run(); err != nil {
			return err
		}
		return nil
	}
}

func getMessages(api *ChatGPTClient, request string) ([]string, error) {
	s := spinner.New(spinner.CharSets[11], 100*time.Millisecond)
	s.Start()
	defer s.Stop()

	// Call the OpenAI API to get the response
	response, err := api.GetAnswer(request)
	if err != nil {
		return nil, err
	}
	fmt.Println("Response: ", response)

	// Parse the response into a list of commit messages
	messages, err := parseStringList(response)
	if err != nil {
		return nil, err
	}

	// Normalize and filter the commit messages
	var choices []string
	for _, message := range messages {
		message = normalizeMessage(message)
		if message != "" {
			choices = append(choices, message)
		}
	}

	return choices, nil
}

func parseStringList(message string) ([]string, error) {
	//commitMessages := make([]string, 0)

	lines := strings.Split(message, "\n")

	return lines, nil
}

func normalizeMessage(line string) string {
	// Trim whitespace and common prefixes
	line = strings.TrimSpace(strings.TrimPrefix(line, "- "))
	line = strings.TrimSpace(strings.TrimPrefix(line, "* "))
	line = strings.TrimSpace(strings.TrimPrefix(line, "+ "))
	line = strings.TrimSpace(strings.TrimPrefix(line, "> "))
	line = strings.TrimSpace(strings.TrimPrefix(line, "# "))
	line = strings.TrimSpace(strings.TrimPrefix(line, "~ "))
	line = strings.TrimSpace(strings.TrimPrefix(line, ": "))
	line = strings.TrimSpace(strings.TrimPrefix(line, "| "))
	line = strings.TrimSpace(strings.TrimPrefix(line, "• "))
	line = strings.TrimSpace(strings.TrimPrefix(line, "▸ "))

	// Trim common suffixes
	line = strings.TrimSuffix(line, "`")
	line = strings.TrimSuffix(line, "\"")
	line = strings.TrimSuffix(line, "'")
	line = strings.TrimSuffix(line, ":")

	// Remove any leading numbers and whitespace
	for i := 0; i < len(line); i++ {
		if !unicode.IsDigit(rune(line[i])) {
			line = strings.TrimSpace(line[i:])
			break
		}
	}

	return strings.TrimSpace(line)
}

func escapeCommitMessage(message string) string {
	return strings.ReplaceAll(message, "'", `''`)
}

func getDiff() string {
	diff, err := exec.Command("git", "diff", "--cached").Output()
	if err != nil {
		log.Fatal("Failed to run git diff --cached")
	}

	if len(diff) == 0 {
		fmt.Println("No changes to commit.")
		return ""
	}

	return string(diff)
}
