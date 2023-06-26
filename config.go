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

func getApiKey() (string, error) {
	var apiKey string
	if apiKey == "" {
		var err error
		apiKey, err = PromptToken()
		if err != nil {
			return "", err
		}
		err = SetAPIKey(apiKey)
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
	DefaultPromptTemplate = `WW91IGFyZSBhbiBleHBlcnQgcHJvZ3JhbW1lciBzdW1tYXJpemluZyBhIGdpdCBkaWZmLgpSZW1pbmRlcnMgYWJvdXQgdGhlIGdpdCBkaWZmIGZvcm1hdDoKRm9yIGV2ZXJ5IGZpbGUsIHRoZXJlIGFyZSBhIGZldyBtZXRhZGF0YSBsaW5lcywgbGlrZSAoZm9yIGV4YW1wbGUpOgpgYGAKZGlmZiAtLWdpdCBhL2xpYi9pbmRleC5qcyBiL2xpYi9pbmRleC5qcwppbmRleCBhYWRmNjkxLi5iZmVmNjAzIDEwMDY0NAotLS0gYS9saWIvaW5kZXguanMKKysrIGIvbGliL2luZGV4LmpzCmBgYApUaGlzIG1lYW5zIHRoYXQgYGxpYi9pbmRleC5qc2Agd2FzIG1vZGlmaWVkIGluIHRoaXMgY29tbWl0LiBOb3RlIHRoYXQgdGhpcyBpcyBvbmx5IGFuIGV4YW1wbGUuClRoZW4gdGhlcmUgaXMgYSBzcGVjaWZpZXIgb2YgdGhlIGxpbmVzIHRoYXQgd2VyZSBtb2RpZmllZC4KQSBsaW5lIHN0YXJ0aW5nIHdpdGggYCtgIG1lYW5zIGl0IHdhcyBhZGRlZC4KQSBsaW5lIHRoYXQgc3RhcnRpbmcgd2l0aCBgLWAgbWVhbnMgdGhhdCBsaW5lIHdhcyBkZWxldGVkLgpBIGxpbmUgdGhhdCBzdGFydHMgd2l0aCBuZWl0aGVyIGArYCBub3IgYC1gIGlzIGNvZGUgZ2l2ZW4gZm9yIGNvbnRleHQgYW5kIGJldHRlciB1bmRlcnN0YW5kaW5nLgpJdCBpcyBub3QgcGFydCBvZiB0aGUgZGlmZi4KQWZ0ZXIgdGhlIGdpdCBkaWZmIG9mIHRoZSBmaXJzdCBmaWxlLCB0aGVyZSB3aWxsIGJlIGFuIGVtcHR5IGxpbmUsIGFuZCB0aGVuIHRoZSBnaXQgZGlmZiBvZiB0aGUgbmV4dCBmaWxlLgoKRG8gbm90IGluY2x1ZGUgdGhlIGZpbGUgbmFtZSBhcyBhbm90aGVyIHBhcnQgb2YgdGhlIGNvbW1lbnQuCkRvIG5vdCB1c2UgdGhlIGNoYXJhY3RlcnMgYFtgIG9yIGBdYCBpbiB0aGUgc3VtbWFyeS4KV3JpdGUgZXZlcnkgc3VtbWFyeSBjb21tZW50IGluIGEgbmV3IGxpbmUuCkNvbW1lbnRzIHNob3VsZCBiZSBpbiBhIGJ1bGxldCBwb2ludCBsaXN0LCBlYWNoIGxpbmUgc3RhcnRpbmcgd2l0aCBhIGAtYC4KVGhlIHN1bW1hcnkgc2hvdWxkIG5vdCBpbmNsdWRlIGNvbW1lbnRzIGNvcGllZCBmcm9tIHRoZSBjb2RlLgpUaGUgb3V0cHV0IHNob3VsZCBiZSBlYXNpbHkgcmVhZGFibGUuIFdoZW4gaW4gZG91YnQsIHdyaXRlIGZld2VyIGNvbW1lbnRzIGFuZCBub3QgbW9yZS4gRG8gbm90IG91dHB1dCBjb21tZW50cyB0aGF0CnNpbXBseSByZXBlYXQgdGhlIGNvbnRlbnRzIG9mIHRoZSBmaWxlLgpSZWFkYWJpbGl0eSBpcyB0b3AgcHJpb3JpdHkuIFdyaXRlIG9ubHkgdGhlIG1vc3QgaW1wb3J0YW50IGNvbW1lbnRzIGFib3V0IHRoZSBkaWZmLgoKRVhBTVBMRSBTVU1NQVJZIENPTU1FTlRTOgpgYGAKLSBSYWlzZSB0aGUgYW1vdW50IG9mIHJldHVybmVkIHJlY29yZGluZ3MgZnJvbSBgMTBgIHRvIGAxMDBgCi0gRml4IGEgdHlwbyBpbiB0aGUgZ2l0aHViIGFjdGlvbiBuYW1lCi0gTW92ZSB0aGUgYG9jdG9raXRgIGluaXRpYWxpemF0aW9uIHRvIGEgc2VwYXJhdGUgZmlsZQotIEFkZCBhbiBPcGVuQUkgQVBJIGZvciBjb21wbGV0aW9ucwotIExvd2VyIG51bWVyaWMgdG9sZXJhbmNlIGZvciB0ZXN0IGZpbGVzCi0gQWRkIDIgdGVzdHMgZm9yIHRoZSBpbmNsdXNpdmUgc3RyaW5nIHNwbGl0IGZ1bmN0aW9uCmBgYApNb3N0IGNvbW1pdHMgd2lsbCBoYXZlIGxlc3MgY29tbWVudHMgdGhhbiB0aGlzIGV4YW1wbGVzIGxpc3QuClRoZSBsYXN0IGNvbW1lbnQgZG9lcyBub3QgaW5jbHVkZSB0aGUgZmlsZSBuYW1lcywKYmVjYXVzZSB0aGVyZSB3ZXJlIG1vcmUgdGhhbiB0d28gcmVsZXZhbnQgZmlsZXMgaW4gdGhlIGh5cG90aGV0aWNhbCBjb21taXQuCkRvIG5vdCBpbmNsdWRlIHBhcnRzIG9mIHRoZSBleGFtcGxlIGluIHlvdXIgc3VtbWFyeS4KSXQgaXMgZ2l2ZW4gb25seSBhcyBhbiBleGFtcGxlIG9mIGFwcHJvcHJpYXRlIGNvbW1lbnRzLgoKClRIRSBHSVQgRElGRiBUTyBCRSBTVU1NQVJJWkVEOgpgYGAKe3tkaWZmfX0KYGBgCgogY29tbWl0IG1lc3NhZ2VzIHNob3VsZDoKICAtIGZvbGxvdyBjb252ZW50aW9uYWwgY29tbWl0cwogIC0gbWVzc2FnZSBmb3JtYXQgc2hvdWxkIGJlOiA8dHlwZT5bc2NvcGVdOiA8ZGVzY3JpcHRpb24+CmV4YW1wbGVzOgogIC0gZml4KGF1dGhlbnRpY2F0aW9uKTogYWRkIHBhc3N3b3JkIHJlZ2V4IHBhdHRlcm4KICAtIGZlYXQoc3RvcmFnZSk6IGFkZCBuZXcgdGVzdCBjYXNlcwpUSEUgU1VNTUFSWTo=`
)

type Config struct {
	APIKey         string  `json:"apiKey,omitempty"`
	PromptTemplate string  `json:"promptTemplate,omitempty"`
	Model          string  `json:"model,omitempty"`
	Temperature    float32 `json:"temperature,omitempty"`
	MaxTokens      int     `json:"maxTokens,omitempty"`
}

func LoadGlobalConfig() (*Config, error) {
	globalConfigPath := os.ExpandEnv(GlobalConfigPath)
	if _, err := os.Stat(globalConfigPath); os.IsNotExist(err) {
		return &Config{
			APIKey:         "",
			PromptTemplate: DefaultPromptTemplate,
			Model:          DefaultModel,
			Temperature:    DefaultTemperature,
			MaxTokens:      DefaultMaxTokens,
		}, nil
	} else if err != nil {
		return nil, err
	} else {
		jsonData, err := os.ReadFile(globalConfigPath)
		if err != nil {
			return nil, err
		}

		var config Config
		err = json.Unmarshal(jsonData, &config)
		if err != nil {
			return nil, err
		}

		if config.PromptTemplate == "" {
			config.PromptTemplate = DefaultPromptTemplate
		}
		if config.Model == "" {
			config.Model = DefaultModel
		}
		if config.Temperature == 0 {
			config.Temperature = DefaultTemperature
		}
		if config.MaxTokens == 0 {
			config.MaxTokens = DefaultMaxTokens
		}

		return &config, nil
	}
}

func SetAPIKey(key string) error {
	globalConfigPath := os.ExpandEnv(GlobalConfigPath)

	// Load the config file
	config, err := LoadGlobalConfig()
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
