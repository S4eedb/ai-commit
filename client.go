package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"
	"unicode"

	"github.com/AlecAivazis/survey/v2"
	"github.com/briandowns/spinner"
	"github.com/sashabaranov/go-openai"
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
	model := getConfig("model").(string)
	maxTokens := getConfig("maxTokens").(int)
	temperature := getConfig("temperature").(float32)

	req := openai.CompletionRequest{
		Model:       model,
		MaxTokens:   maxTokens,
		Prompt:      question,
		Temperature: temperature,
	}

	resp, err := c.client.CreateCompletion(context.Background(), req)
	if err != nil {
		return "", err
	}

	if len(resp.Choices) == 0 {
		return "", errors.New("empty response from the API")
	}

	return resp.Choices[0].Text, nil
}

const CUSTOM_MESSAGE_OPTION = "[write own message]..."

func run(diff string) error {
	const maxDiffTokens = 2000
	const maxPromptTokens = 700
	const customMessageOption = "Enter a custom message"

	// Trim the diff to the maximum number of tokens
	diffTokens := len(strings.Fields(diff))
	if diffTokens > maxDiffTokens {
		diff = strings.Join(strings.Fields(diff)[:maxPromptTokens], " ")
	}

	apiKey, err := getApiKey()
	if err != nil {
		return err
	}
	api := NewChatGPTClient(apiKey)

	prompt, err := loadPromptTemplate()
	if err != nil {
		return err
	}
	prompt = strings.ReplaceAll(prompt, "{{diff}}", "```\n"+diff+"\n```")

	for {
		choices, err := getMessages(api, prompt)
		if err != nil {
			return err
		}
		choices = append(choices, customMessageOption)

		var message string
		prompt := &survey.Select{
			Message: "Pick a message",
			Options: choices,
		}
		err = survey.AskOne(prompt, &message)
		if err != nil {
			return err
		}

		if message == customMessageOption {
			cmd := exec.Command("git", "commit")
			cmd.Stdout = os.Stdout
			cmd.Stdin = os.Stdin
			cmd.Stderr = os.Stderr
			err = cmd.Run()
			if err != nil {
				return err
			}
			return nil
		}

		cmd := exec.Command("git", "commit", "-m", escapeCommitMessage(message))
		cmd.Stdout = os.Stdout
		cmd.Stdin = os.Stdin
		cmd.Stderr = os.Stderr
		err = cmd.Run()
		if err != nil {
			return err
		}
		return nil
	}
}
func getMessages(api *ChatGPTClient, request string) ([]string, error) {
	spinner := spinner.New(spinner.CharSets[11], 100*time.Millisecond)
	spinner.Start()
	defer spinner.Stop()

	response, err := api.GetAnswer(request)
	if err != nil {
		return nil, err
	}

	messages, err := parseStringList(response)
	if err != nil {
		println(err)
	}

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
	var commitMessages []string

	lines := strings.Split(message, "\n")
	for _, line := range lines {
		// Ignore empty lines and lines that don't start with a number
		line = strings.TrimSpace(line)
		if line == "" || !unicode.IsDigit(rune(line[0])) {
			continue
		}

		// Add the commit message to the result
		commitMessages = append(commitMessages, line)
	}

	if len(commitMessages) == 0 {
		return nil, errors.New("no valid commit messages found")
	}

	return commitMessages, nil
}
func normalizeMessage(line string) string {
	line = strings.TrimSpace(line)
	prefixes := []string{"- ", "* ", "+ ", "> ", "# ", "~ ", ": ", "| ", "• ", "▸ "}
	for _, prefix := range prefixes {
		line = strings.TrimPrefix(line, prefix)
	}
	suffixes := []string{"`", `"`, `'`, `:`}
	for _, suffix := range suffixes {
		line = strings.TrimSuffix(line, suffix)
	}
	if line != "" {
		// Find the first non-digit character
		i := 0
		for i < len(line) && unicode.IsDigit(rune(line[i])) {
			i++
		}
		if i < len(line) {
			// Remove the number and any leading whitespace
			line = strings.TrimSpace(line[i:])
		}
	}

	line = strings.TrimSpace(line)
	return line
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
