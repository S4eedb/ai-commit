package main

import (
	"encoding/json"
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"os"
	"path/filepath"
)

type apiKeyAnswer struct {
	APIKey string `survey:"apikey"`
}

func PromptToken() (string, error) {
	answer := apiKeyAnswer{}
	err := survey.AskOne(&survey.Input{
		Message: "Paste your OpenAI apikey here:",
	}, &answer.APIKey, survey.WithValidator(survey.Required))

	if err != nil {
		fmt.Println("Aborted.")
		fmt.Println(err)
		return "", err
	}

	return answer.APIKey, nil
}

const (
	GlobalConfigPath      = "${HOME}/.ai_commit.json"
	DefaultModel          = "text-davinci-003"
	DefaultTemperature    = 0.7
	DefaultMaxTokens      = 200
	DefaultPromptTemplate = `
suggest 10 commit messages based on the following diff:
{{diff}}
commit messages should:
- follow conventional commits
- message format should be: <type>(scope): <description>

examples:
- fix(authentication): add password regex pattern
- feat(storage): add new test cases`
)

type Config struct {
	APIKey         string  `json:"apiKey,omitempty"`
	PromptTemplate string  `json:"promptTemplate,omitempty"`
	Model          string  `json:"model,omitempty"`
	Temperature    float32 `json:"temperature,omitempty"`
	MaxTokens      int     `json:"maxTokens,omitempty"`
}

func SetAPIKey(key string) error {
	globalConfigPath := os.ExpandEnv(GlobalConfigPath)
	if _, err := os.Stat(globalConfigPath); os.IsNotExist(err) {
		// Create the directory if it doesn't exist
		err := os.MkdirAll(filepath.Dir(globalConfigPath), 0755)
		if err != nil {
			return err
		}

		// Write the JSON data to the file
		jsonData, err := json.MarshalIndent(Config{
			APIKey:         key,
			PromptTemplate: DefaultPromptTemplate,
			Model:          DefaultModel,
			Temperature:    DefaultTemperature,
			MaxTokens:      DefaultMaxTokens,
		}, "", "  ")
		if err != nil {
			return err
		}
		err = os.WriteFile(globalConfigPath, jsonData, 0644)
		if err != nil {
			return err
		}
		return nil
	} else {
		// Read the file
		jsonData, err := os.ReadFile(globalConfigPath)
		if err != nil {
			return err
		}

		// Unmarshal the JSON data
		var config Config
		err = json.Unmarshal(jsonData, &config)
		if err != nil {
			return err
		}

		// Update the API key
		config.APIKey = key

		// Write the JSON data to the file
		err = writeJsonFile(globalConfigPath, config)
		if err != nil {
			return err
		}
		return nil

	}
}

func writeJsonFile(path string, data interface{}) error {
	// Create the directory if it doesn't exist
	err := os.MkdirAll(filepath.Dir(path), 0755)
	if err != nil {
		return err
	}

	// Write the JSON data to the file
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	err = os.WriteFile(path, jsonData, 0644)
	if err != nil {
		return err
	}

	return nil
}

func loadConfig() (*Config, error) {
	globalConfigPath := os.ExpandEnv(GlobalConfigPath)
	if _, err := os.Stat(globalConfigPath); os.IsNotExist(err) {
		// If the file doesn't exist, return the default config
		return &Config{
			APIKey:         "",
			PromptTemplate: DefaultPromptTemplate,
			Model:          DefaultModel,
			Temperature:    DefaultTemperature,
			MaxTokens:      DefaultMaxTokens,
		}, nil
	} else {
		// Read the file
		jsonData, err := os.ReadFile(globalConfigPath)
		if err != nil {
			return nil, err
		}

		// Unmarshal the JSON data
		var config Config
		err = json.Unmarshal(jsonData, &config)
		if err != nil {
			return nil, err
		}

		return &config, nil
	}
}
