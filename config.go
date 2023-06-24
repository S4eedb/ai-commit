package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/AlecAivazis/survey/v2"
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
	apiKey = getConfig("apiKey").(string)

	if apiKey == "" {
		var err error
		apiKey, err = promptToken()
		if err != nil {
			return "", err
		}

		err = setGlobalConfig("apiKey", apiKey)
		if err != nil {
			return "", err
		}
	}

	return apiKey, nil
}

const (
	GLOBAL_CONFIG_PATH          = "${HOME}/.commitgpt.json"
	LOCAL_CONFIG_PATH           = "./.commitgpt.json"
	GLOBAL_PROMPT_TEMPLATE_PATH = "${HOME}/.commitgpt-template"
	DEFAULT_MODEL               = "text-davinci-003"
	DEFAULT_TEMPERATURE         = 0.7
	DEFAULT_MAX_TOKENS          = 200
	DEFAULT_PROMPT_TEMPLATE     = `
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

func defaultConfig() Config {
	return Config{
		Model:       DEFAULT_MODEL,
		Temperature: DEFAULT_TEMPERATURE,
		MaxTokens:   DEFAULT_MAX_TOKENS,
	}
}

func loadConfig() Config {
	var config Config

	// Load global config
	globalConfigPath := os.ExpandEnv(GLOBAL_CONFIG_PATH)
	if _, err := os.Stat(globalConfigPath); err == nil {
		globalConfigData, err := ioutil.ReadFile(globalConfigPath)
		if err == nil {
			err = json.Unmarshal(globalConfigData, &config)
			if err != nil {
				panic(err)
			}
		}
	}

	// Load local config
	if _, err := os.Stat(LOCAL_CONFIG_PATH); err == nil {
		localConfigData, err := ioutil.ReadFile(LOCAL_CONFIG_PATH)
		if err == nil {
			var localConfig Config
			err = json.Unmarshal(localConfigData, &localConfig)
			if err != nil {
				panic(err)
			}
			mergeConfigs(&config, &localConfig)
		}
	}

	return config
}

func mergeConfigs(dest, src *Config) {
	if src.APIKey != "" {
		dest.APIKey = src.APIKey
	}
	if src.PromptTemplate != "" {
		dest.PromptTemplate = src.PromptTemplate
	}
	if src.Model != "" {
		dest.Model = src.Model
	}
	if src.Temperature != 0 {
		dest.Temperature = src.Temperature
	}
	if src.MaxTokens != 0 {
		dest.MaxTokens = src.MaxTokens
	}
}

func getConfig(key string) interface{} {
	config := loadConfig()
	values := map[string]interface{}{
		"apiKey":         config.APIKey,
		"promptTemplate": config.PromptTemplate,
		"model":          config.Model,
		"temperature":    config.Temperature,
		"maxTokens":      config.MaxTokens,
	}
	return values[key]
}

func setGlobalConfig(key string, value interface{}) error {
	globalConfigPath := os.ExpandEnv(GLOBAL_CONFIG_PATH)
	globalConfigData, err := ioutil.ReadFile(globalConfigPath)
	if err != nil {
		return err
	}

	var config Config
	err = json.Unmarshal(globalConfigData, &config)
	if err != nil {
		return err
	}

	switch key {
	case "apiKey":
		config.APIKey = value.(string)
	case "promptTemplate":
		config.PromptTemplate = value.(string)
	case "model":
		config.Model = value.(string)
	case "temperature":
		config.Temperature = value.(float32)
	case "maxTokens":
		config.MaxTokens = value.(int)
	default:
		return errors.New("invalid config key")
	}

	writeJsonFile(globalConfigPath, config)

	return nil
}

func loadPromptTemplate() (string, error) {

	// Load global prompt template
	globalPromptTemplatePath := os.ExpandEnv(GLOBAL_PROMPT_TEMPLATE_PATH)
	if _, err := os.Stat(globalPromptTemplatePath); err == nil {
		temp, err := ioutil.ReadFile(globalPromptTemplatePath)
		if err == nil && containsDiffPlaceholder(string(temp)) {
			return string(temp), nil
		}
	}

	// Return default prompt template
	return DEFAULT_PROMPT_TEMPLATE, nil
}

func writeJsonFile(path string, data interface{}) error {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	return ioutil.WriteFile(path, jsonData, 0644)
}

func containsDiffPlaceholder(t string) bool {
	return strings.Contains(t, "{{diff}}") || strings.Contains(t, "{{diff")
}

func ensureGlobal() {
	globalConfigPath := os.ExpandEnv(GLOBAL_CONFIG_PATH)
	if _, err := os.Stat(globalConfigPath); os.IsNotExist(err) {
		writeJsonFile(globalConfigPath, defaultConfig())
	}
}
