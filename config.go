package main

import (
	"encoding/json"
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"io/ioutil"
	"os"
)

type apiKeyAnswer struct {
	APIKey string `survey:"apikey"`
}

func promptToken() (string, error) {
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

func getApiKey() (string, error) {
	var apiKey string
	if apiKey == "" {
		var err error
		apiKey, err = promptToken()
		if err != nil {
			return "", err
		}
		err = setAPIKey(apiKey)
		if err != nil {
			return "", err
		}
	}
	return apiKey, nil
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

func setAPIKey(key string) error {
	globalConfigPath := os.ExpandEnv(GlobalConfigPath)
	globalConfigData, err := ioutil.ReadFile(globalConfigPath)
	if err != nil {
		return err
	}
	var config Config
	err = json.Unmarshal(globalConfigData, &config)
	if err != nil {
		return err
	}
	config.APIKey = key
	writeJsonFile(globalConfigPath, config)
	return nil
}

func writeJsonFile(path string, data interface{}) error {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(path, jsonData, 0644)
}
