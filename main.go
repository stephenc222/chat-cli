package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

/*
TODO:
Further work could include:
- simple menu for selecting an existing conversation (with a -l flag, else straight to new conversation?)
- Enhancing logging with structured logging libraries.
- Adding a context (context.Context) to handle cancellation and timeouts across HTTP requests.
- Writing unit tests for each method, mocking the Client interface to test different scenarios.
- Adding a more advanced configuration system for managing API keys and other settings, possibly using environment variables and .env files.
- Implementing a command-line interface (CLI) to improve user interaction, using libraries like tview, cobra or urfave/cli.
*/

// What should we do next?
// - Create a new assistant
// - Interact with an existing assistant
// - List existing assistants
// - Delete an existing assistant
// - Update an existing assistant
// - Get details about an existing assistant
// - Get details about an existing conversation
// - Get details about an existing message
// - Get details about an existing tool
// - Get details about an existing user
// - Get details about an existing user message

var instructions = `As a GPT Assistant, your primary role is to serve as a knowledgeable and efficient guide for users navigating the Unix Terminal environment. Your objective is to provide clear, detailed, and practical assistance that enhances the user's capabilities and efficiency in executing a wide range of tasks within the Unix Terminal. This includes offering help with command syntax, scripting, troubleshooting, system management, and optimizing workflows, all while ensuring a user-friendly experience suitable for both beginners and advanced users in the Unix environment.`

func main() {
	config, err := ReadConfig()
	if err != nil || config.APIKey == "" {
		config, err = PromptForConfig()
		if err != nil {
			fmt.Printf("Error obtaining API key: %v\n", err)
			os.Exit(1)
		}
	}

	openAI := NewOpenAI(config.APIKey, &http.Client{})
	if config.AssistantID == "" {
		assistantID, err := openAI.CreateAssistant(map[string]interface{}{
			"model":        "gpt-4-1106-preview",
			"name":         "Shell Assistant",
			"instructions": instructions,
			"tools":        []interface{}{map[string]string{"type": "code_interpreter"}},
		})
		if err != nil {
			fmt.Printf("Error creating assistant: %v\n", err)
			os.Exit(1)
		}
		config.AssistantID = assistantID

		// Update the config file with the new assistant ID
		configBytes, err := json.MarshalIndent(config, "", "  ")
		if err != nil {
			fmt.Printf("Error updating config with assistant ID: %v\n", err)
			os.Exit(1)
		}
		configFilePath, err := GetConfigFilePath()
		if err != nil {
			fmt.Printf("Error obtaining config file path: %v\n", err)
			os.Exit(1)
		}
		err = os.WriteFile(configFilePath, configBytes, 0644)
		if err != nil {
			fmt.Printf("Error writing updated config to file: %v\n", err)
			os.Exit(1)
		}
	}

	openAI.InteractWithAssistant(config.AssistantID)
}
