package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/briandowns/spinner"
)

// Client interface defines the HTTP client behavior
type Client interface {
	Do(req *http.Request) (*http.Response, error)
}

// OpenAI represents the client for OpenAI API
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

func (oai *OpenAI) CreateAssistant(assistantData map[string]interface{}) (string, error) {
	response, err := oai.makeRequest("POST", "/assistants", assistantData)
	if err != nil {
		return "", err
	}
	return response["id"].(string), nil
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

func (oai *OpenAI) InteractWithAssistant(assistantID string) {
	reader := bufio.NewReader(os.Stdin)
	threadID, err := oai.CreateThread()
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

		err = oai.SendMessage(threadID, userInput)
		if err != nil {
			fmt.Printf("Error sending message: %v\n", err)
			continue
		}

		s.Start()                                          // Start the spinner
		s.Suffix = " " + ElectricBlue + "Thinking" + Reset // Append text after the spinner

		runID, err := oai.CreateRun(threadID, assistantID)
		if err != nil {
			fmt.Printf("Error creating run: %v\n", err)
			continue
		}

		for runStatus := ""; runStatus != "completed"; {
			runStatus, err = oai.GetRunStatus(threadID, runID)
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

		messages, err := oai.GetMessages(threadID)
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
