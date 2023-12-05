package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/briandowns/spinner"
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

var instructions = `GPT Assistant, your role is to be an informative and effective assistant for users working within a Unix Terminal environment. Your goal is to deliver concise yet comprehensive guidance to enhance the user's proficiency and efficiency within the Unix terminal.`

type Config struct {
	APIKey      string `json:"api_key"`
	AssistantID string `json:"assistant_id"`
}

type Client interface {
	Do(req *http.Request) (*http.Response, error)
}

type OpenAI struct {
	apiKey  string
	client  Client
	baseURL string
}

func NewOpenAI(apiKey string, client Client) *OpenAI {
	return &OpenAI{
		apiKey:  apiKey,
		client:  client,
		baseURL: "https://api.openai.com/v1",
	}
}

func (oai *OpenAI) makeRequest(method, endpoint string, payload interface{}) (map[string]interface{}, error) {
	url := oai.baseURL + endpoint
	var body bytes.Buffer
	if payload != nil {
		if err := json.NewEncoder(&body).Encode(payload); err != nil {
			return nil, fmt.Errorf("error encoding payload: %w", err)
		}
	}

	req, err := http.NewRequest(method, url, &body)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+oai.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("OpenAI-Beta", "assistants=v1")

	resp, err := oai.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("error unmarshaling response: %w", err)
	}

	return result, nil
}

func (oai *OpenAI) CreateAssistant(assistantData map[string]interface{}) (string, error) {
	response, err := oai.makeRequest("POST", "/assistants", assistantData)
	if err != nil {
		return "", err
	}
	return response["id"].(string), nil
}

func (oai *OpenAI) CreateThread() (string, error) {
	response, err := oai.makeRequest("POST", "/threads", nil)
	if err != nil {
		return "", err
	}
	return response["id"].(string), nil
}

func (oai *OpenAI) SendMessage(threadID, message string) error {
	messageData := map[string]interface{}{
		"role":    "user",
		"content": message,
	}
	_, err := oai.makeRequest("POST", fmt.Sprintf("/threads/%s/messages", threadID), messageData)
	return err
}

func (oai *OpenAI) CreateRun(threadID, assistantID string) (string, error) {
	runData := map[string]string{"assistant_id": assistantID}
	response, err := oai.makeRequest("POST", fmt.Sprintf("/threads/%s/runs", threadID), runData)
	if err != nil {
		return "", err
	}
	return response["id"].(string), nil
}

func (oai *OpenAI) GetRunStatus(threadID, runID string) (string, error) {
	response, err := oai.makeRequest("GET", fmt.Sprintf("/threads/%s/runs/%s", threadID, runID), nil)
	if err != nil {
		return "", err
	}
	return response["status"].(string), nil
}

func (oai *OpenAI) GetMessages(threadID string) ([]interface{}, error) {
	response, err := oai.makeRequest("GET", fmt.Sprintf("/threads/%s/messages", threadID), nil)
	if err != nil {
		return nil, err
	}
	return response["data"].([]interface{}), nil
}

func interactWithAssistant(openAI *OpenAI, assistantID string) {
	reader := bufio.NewReader(os.Stdin)
	threadID, err := openAI.CreateThread()
	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond) // Build our new spinner
	s.Color(SpinnerBgCyan, SpinnerReset, SpinnerFgHiCyan)
	fmt.Println("Enter your message (type 'exit' to quit)")
	fmt.Print("> ")
	for {
		userInput, _ := reader.ReadString('\n')
		userInput = strings.TrimSpace(userInput)
		if userInput == "exit" {
			break
		}

		if err != nil {
			fmt.Printf("Error creating thread: %v\n", err)
			continue
		}

		err = openAI.SendMessage(threadID, userInput)
		if err != nil {
			fmt.Printf("Error sending message: %v\n", err)
			continue
		}

		s.Start()                                          // Start the spinner
		s.Suffix = " " + ElectricBlue + "Thinking" + Reset // Append text after the spinner

		runID, err := openAI.CreateRun(threadID, assistantID)
		if err != nil {
			fmt.Printf("Error creating run: %v\n", err)
			continue
		}

		for runStatus := ""; runStatus != "completed"; {
			runStatus, err = openAI.GetRunStatus(threadID, runID)
			if err != nil {
				fmt.Printf("Error getting run status: %v\n", err)
				break
			}
			if runStatus == "failed" {
				fmt.Println("Run failed.")
				break
			}
			time.Sleep(1 * time.Second)
		}

		s.Stop()

		messages, err := openAI.GetMessages(threadID)
		if err != nil {
			fmt.Printf("Error retrieving messages: %v\n", err)
			continue
		}
		for _, messageInterface := range messages {
			message := messageInterface.(map[string]interface{})
			if message["role"] == "assistant" {
				contents, ok := message["content"].([]interface{})
				if !ok {
					fmt.Println("Error asserting content format.")
					continue
				}
				for _, contentInterface := range contents {
					content, ok := contentInterface.(map[string]interface{})
					if !ok {
						fmt.Println("Error asserting content entry format.")
						continue
					}
					if textMap, ok := content["text"].(map[string]interface{}); ok {
						if value, ok := textMap["value"].(string); ok {
							fmt.Printf(LighterLightGray+"AI: %s\n"+Reset, value)
						}
					}
				}
				break
			}
		}
		fmt.Print("> ")
	}
}

func getConfigFilePath() (string, error) {
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

func promptForConfig() (*Config, error) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter your OpenAI API Key: ")
	apiKey, err := reader.ReadString('\n')
	if err != nil {
		return nil, err
	}

	apiKey = strings.TrimSpace(apiKey)
	config := &Config{APIKey: apiKey}

	configFilePath, err := getConfigFilePath()
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

func promptForAPIKey() (string, error) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter your OpenAI API Key: ")
	apiKey, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	apiKey = strings.TrimSpace(apiKey)
	configFilePath, err := getConfigFilePath()
	if err != nil {
		return "", err
	}

	err = os.WriteFile(configFilePath, []byte(apiKey), 0644)
	if err != nil {
		return "", err
	}

	return apiKey, nil
}

func readConfig() (*Config, error) {
	configFilePath, err := getConfigFilePath()
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

func main() {
	// apiKey := os.Getenv("OPENAI_API_KEY")
	// if apiKey == "" {
	// 	fmt.Println("The OPENAI_API_KEY environment variable is not set.")
	// 	os.Exit(1)
	// }
	config, err := readConfig()
	if err != nil || config.APIKey == "" {
		config, err = promptForConfig()
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
		configFilePath, err := getConfigFilePath()
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

	interactWithAssistant(openAI, config.AssistantID)
}
