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
)

const SystemPrompt = `You are a helpful AI assistant that can execute commands on macOS systems. Your role is to:

1. Understand user requests and translate them into appropriate system commands
2. Only suggest safe, whitelisted commands
3. Provide clear explanations of what you're doing
4. Express confidence in your decisions

Available commands: ls, pwd, whoami, date, uptime, ps, top, df, du, find, grep, cat, head, tail, wc, sort, uniq, echo, mkdir, rmdir, cp, mv, rm, chmod, chown, file, stat, which, whereis, system_profiler, sw_vers, defaults, launchctl, netstat, lsof, ifconfig, ping, nslookup, dig, curl, wget

Safety rules:
- Never suggest dangerous commands like 'sudo', 'rm -rf /', 'format', or 'dd'
- Always use safe alternatives
- Explain what each command does
- Be specific about arguments and options

Respond in valid JSON format with thoughts, commands array, explanation, and confidence level.`

type FreeAIAgent struct {
	*Agent
	serviceType string
	baseURL     string
	apiKey      string
	model       string
}

type FreeAIResponse struct {
	Thoughts     string              `json:"thoughts"`
	Commands     []CommandRequest    `json:"commands"`
	Explanation  string              `json:"explanation"`
	Confidence   float64             `json:"confidence"`
	Results      []CommandResponse   `json:"results,omitempty"`
	Error        string              `json:"error,omitempty"`
}

func NewFreeAIAgent(serviceType, baseURL, apiKey, model string) *FreeAIAgent {
	return &FreeAIAgent{
		Agent:       NewAgent(),
		serviceType: serviceType,
		baseURL:     baseURL,
		apiKey:      apiKey,
		model:       model,
	}
}

func NewOllamaAgent(model string) *FreeAIAgent {
	return NewFreeAIAgent("ollama", "http://localhost:11434", "", model)
}

func NewHuggingFaceAgent(apiKey, model string) *FreeAIAgent {
	return NewFreeAIAgent("huggingface", "https://api-inference.huggingface.co", apiKey, model)
}

func NewLocalAgent(baseURL, model string) *FreeAIAgent {
	return NewFreeAIAgent("local", baseURL, "", model)
}

func (fa *FreeAIAgent) ProcessUserRequest(userMessage string) (*FreeAIResponse, error) {
	systemInfo := fa.GetSystemInfo()
	context := fmt.Sprintf("Current system: %s %s, macOS %s", 
		systemInfo["os"], systemInfo["arch"], systemInfo["macos_version"])

	prompt := fmt.Sprintf(`%s

Context: %s

User request: %s

Please respond in valid JSON format with the following structure:
{
  "thoughts": "Your reasoning about what the user wants",
  "commands": [
    {
      "command": "command_name",
      "args": ["arg1", "arg2"],
      "timeout": 30
    }
  ],
  "explanation": "Explain what you're going to do and why",
  "confidence": 0.95
}`, SystemPrompt, context, userMessage)

	var response *FreeAIResponse
	var err error

	switch fa.serviceType {
	case "ollama":
		response, err = fa.callOllama(prompt)
	case "huggingface":
		response, err = fa.callHuggingFace(prompt)
	case "local":
		response, err = fa.callLocalAPI(prompt)
	default:
		return nil, fmt.Errorf("unsupported service type: %s", fa.serviceType)
	}

	if err != nil {
		return nil, err
	}

	if len(response.Commands) > 0 {
		response.Results = make([]CommandResponse, 0, len(response.Commands))
		
		for _, cmd := range response.Commands {
			result := fa.ExecuteCommand(cmd)
			response.Results = append(response.Results, result)
		}
	}

	return response, nil
}

func (fa *FreeAIAgent) callOllama(prompt string) (*FreeAIResponse, error) {
	requestBody := map[string]interface{}{
		"model":  fa.model,
		"prompt": prompt,
		"stream": false,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(fa.baseURL+"/api/generate", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("Ollama API error: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var ollamaResp struct {
		Response string `json:"response"`
		Error    string `json:"error,omitempty"`
	}

	if err := json.Unmarshal(body, &ollamaResp); err != nil {
		return nil, err
	}

	if ollamaResp.Error != "" {
		return nil, fmt.Errorf("Ollama error: %s", ollamaResp.Error)
	}

	var aiResponse FreeAIResponse
	if err := json.Unmarshal([]byte(ollamaResp.Response), &aiResponse); err != nil {
		aiResponse = FreeAIResponse{
			Thoughts:    "Failed to parse AI response as JSON",
			Commands:    []CommandRequest{},
			Explanation: ollamaResp.Response,
			Confidence:  0.0,
			Error:       "Invalid JSON response from AI",
		}
	}

	return &aiResponse, nil
}

func (fa *FreeAIAgent) callHuggingFace(prompt string) (*FreeAIResponse, error) {
	requestBody := map[string]interface{}{
		"inputs": prompt,
		"parameters": map[string]interface{}{
			"max_new_tokens": 1000,
			"temperature":    0.1,
		},
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", fa.baseURL+"/models/"+fa.model, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+fa.apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Hugging Face API error: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var hfResp []struct {
		GeneratedText string `json:"generated_text"`
	}

	if err := json.Unmarshal(body, &hfResp); err != nil {
		return nil, err
	}

	if len(hfResp) == 0 {
		return nil, fmt.Errorf("no response from Hugging Face")
	}

	var aiResponse FreeAIResponse
	if err := json.Unmarshal([]byte(hfResp[0].GeneratedText), &aiResponse); err != nil {
		aiResponse = FreeAIResponse{
			Thoughts:    "Failed to parse AI response as JSON",
			Commands:    []CommandRequest{},
			Explanation: hfResp[0].GeneratedText,
			Confidence:  0.0,
			Error:       "Invalid JSON response from AI",
		}
	}

	return &aiResponse, nil
}

func (fa *FreeAIAgent) callLocalAPI(prompt string) (*FreeAIResponse, error) {
	requestBody := map[string]interface{}{
		"model":  fa.model,
		"prompt": prompt,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(fa.baseURL+"/generate", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("Local API error: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var localResp struct {
		Response string `json:"response"`
		Error    string `json:"error,omitempty"`
	}

	if err := json.Unmarshal(body, &localResp); err != nil {
		return nil, err
	}

	if localResp.Error != "" {
		return nil, fmt.Errorf("Local API error: %s", localResp.Error)
	}

	var aiResponse FreeAIResponse
	if err := json.Unmarshal([]byte(localResp.Response), &aiResponse); err != nil {
		aiResponse = FreeAIResponse{
			Thoughts:    "Failed to parse AI response as JSON",
			Commands:    []CommandRequest{},
			Explanation: localResp.Response,
			Confidence:  0.0,
			Error:       "Invalid JSON response from AI",
		}
	}

	return &aiResponse, nil
}

func (fa *FreeAIAgent) InteractiveFreeAIMode() {
	fmt.Printf("Free AI-Powered macOS Command Agent (%s)\n", fa.serviceType)
	fmt.Println("===============================================")
	fmt.Println("I can help you execute commands on your Mac safely and intelligently.")
	fmt.Println("Just tell me what you want to do, and I'll figure out the best commands to run.")
	fmt.Println("Type 'quit' to exit.")
	fmt.Println()

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("You: ")
		if !scanner.Scan() {
			break
		}

		userInput := strings.TrimSpace(scanner.Text())
		if userInput == "" {
			continue
		}

		if userInput == "quit" || userInput == "exit" {
			fmt.Println("Goodbye")
			break
		}

		fmt.Println("Thinking...")
		response, err := fa.ProcessUserRequest(userInput)
		
		if err != nil {
			fmt.Printf("Error: %s\n\n", err)
			continue
		}

		fmt.Printf("\nThoughts: %s\n", response.Thoughts)
		fmt.Printf("Explanation: %s\n", response.Explanation)
		fmt.Printf("Confidence: %.1f%%\n", response.Confidence*100)

		if len(response.Commands) > 0 {
			fmt.Printf("\nExecuting %d command(s):\n", len(response.Commands))
			
			for i, result := range response.Results {
				fmt.Printf("\n--- Command %d: %s %s ---\n", 
					i+1, response.Commands[i].Command, strings.Join(response.Commands[i].Args, " "))
				
				if result.Success {
					fmt.Printf("Success (%.2fms)\n", 
						parseDuration(result.Duration).Seconds()*1000)
					if result.Output != "" {
						fmt.Printf("Output:\n%s", result.Output)
					}
				} else {
					fmt.Printf("Failed (Exit: %d)\n", result.ExitCode)
					if result.Error != "" {
						fmt.Printf("Error: %s\n", result.Error)
					}
					if result.Output != "" {
						fmt.Printf("Output:\n%s", result.Output)
					}
				}
			}
		}

		fmt.Println("\n" + strings.Repeat("-", 50) + "\n")
	}
}

func parseDuration(durationStr string) time.Duration {
	duration, err := time.ParseDuration(durationStr)
	if err != nil {
		return 0
	}
	return duration
}

func (fa *FreeAIAgent) TestFreeAIConnection() error {
	_, err := fa.ProcessUserRequest("test")
	return err
} 