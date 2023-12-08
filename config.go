package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Config holds the configuration details
type Config struct {
	APIKey      string `json:"api_key"`
	AssistantID string `json:"assistant_id"`
}

func GetConfigFilePath() (string, error) {
	// If the executable path contains `/tmp/`, assume it's run with `go run`
	// loosely check if it was executed using "go run"
	isGoRun := strings.HasPrefix(os.Args[0], os.TempDir())
	if isGoRun {
		// Use the current working directory for development purposes
		return filepath.Join(".", "cli-chat-config.json"), nil
	}

	configFilePath := filepath.Join("/usr/local/etc", "cli-chat-config.json")
	return configFilePath, nil
}

func ReadConfig() (*Config, error) {
	configFilePath, err := GetConfigFilePath()
	if err != nil {
		return nil, err
	}

	configBytes, err := os.ReadFile(configFilePath)
	if err != nil {
		return nil, err
	}

	var config Config
	err = json.Unmarshal(configBytes, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

func PromptForAPIKey() (string, error) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter your OpenAI API Key: ")
	apiKey, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	apiKey = strings.TrimSpace(apiKey)
	configFilePath, err := GetConfigFilePath()
	if err != nil {
		return "", err
	}

	err = os.WriteFile(configFilePath, []byte(apiKey), 0644)
	if err != nil {
		return "", err
	}

	return apiKey, nil
}

func PromptForConfig() (*Config, error) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter your OpenAI API Key: ")
	apiKey, err := reader.ReadString('\n')
	if err != nil {
		return nil, err
	}

	apiKey = strings.TrimSpace(apiKey)
	config := &Config{APIKey: apiKey}

	configFilePath, err := GetConfigFilePath()
	if err != nil {
		return nil, err
	}

	configBytes, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return nil, err
	}

	err = os.WriteFile(configFilePath, configBytes, 0644)
	if err != nil {
		return nil, err
	}

	return config, nil
}
