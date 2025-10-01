package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

type CommandRequest struct {
	Command string `json:"command"`
	Args    []string `json:"args,omitempty"`
	Timeout int     `json:"timeout,omitempty"`
}

type CommandResponse struct {
	Success   bool   `json:"success"`
	Output    string `json:"output"`
	Error     string `json:"error,omitempty"`
	ExitCode  int    `json:"exit_code"`
	Duration  string `json:"duration"`
	Timestamp string `json:"timestamp"`
}

type Agent struct {
	allowedCommands map[string]bool
}

func NewAgent() *Agent {
	allowedCommands := map[string]bool{
		"ls": true, "pwd": true, "whoami": true, "date": true, "uptime": true,
		"ps": true, "top": true, "df": true, "du": true, "find": true,
		"grep": true, "cat": true, "head": true, "tail": true, "wc": true,
		"sort": true, "uniq": true, "echo": true, "mkdir": true, "rmdir": true,
		"cp": true, "mv": true, "rm": true, "chmod": true, "chown": true,
		"file": true, "stat": true, "which": true, "whereis": true,
		"system_profiler": true, "sw_vers": true, "defaults": true,
		"launchctl": true, "netstat": true, "lsof": true, "ifconfig": true,
		"ping": true, "nslookup": true, "dig": true, "curl": true, "wget": true,
	}

	return &Agent{
		allowedCommands: allowedCommands,
	}
}

func (a *Agent) ExecuteCommand(req CommandRequest) CommandResponse {
	start := time.Now()
	
	if !a.isCommandAllowed(req.Command) {
		return CommandResponse{
			Success:   false,
			Output:    "",
			Error:     fmt.Sprintf("Command '%s' is not allowed for security reasons", req.Command),
			ExitCode:  -1,
			Duration:  time.Since(start).String(),
			Timestamp: time.Now().Format(time.RFC3339),
		}
	}

	timeout := 30 * time.Second
	if req.Timeout > 0 {
		timeout = time.Duration(req.Timeout) * time.Second
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	var cmd *exec.Cmd
	if len(req.Args) > 0 {
		cmd = exec.CommandContext(ctx, req.Command, req.Args...)
	} else {
		cmd = exec.CommandContext(ctx, req.Command)
	}

	output, err := cmd.CombinedOutput()
	
	response := CommandResponse{
		Success:   err == nil,
		Output:    string(output),
		ExitCode:  cmd.ProcessState.ExitCode(),
		Duration:  time.Since(start).String(),
		Timestamp: time.Now().Format(time.RFC3339),
	}

	if err != nil {
		response.Error = err.Error()
	}

	return response
}

func (a *Agent) GetSystemInfo() map[string]interface{} {
	info := map[string]interface{}{
		"os":      runtime.GOOS,
		"arch":    runtime.GOARCH,
		"version": runtime.Version(),
	}

	if runtime.GOOS == "darwin" {
		if output, err := exec.Command("sw_vers", "-productVersion").Output(); err == nil {
			info["macos_version"] = strings.TrimSpace(string(output))
		}
		if output, err := exec.Command("uname", "-m").Output(); err == nil {
			info["machine"] = strings.TrimSpace(string(output))
		}
	}

	return info
}

func (a *Agent) isCommandAllowed(command string) bool {
	return a.allowedCommands[command]
}

func runFreeAI() {
	if runtime.GOOS != "darwin" {
		log.Fatal("This agent is designed to run on macOS only")
	}

	serviceType := os.Getenv("FREE_AI_SERVICE")
	if serviceType == "" {
		serviceType = "ollama"
	}

	var freeAgent *FreeAIAgent

	switch serviceType {
	case "ollama":
		model := os.Getenv("OLLAMA_MODEL")
		if model == "" {
			model = "llama3.2"
		}
		freeAgent = NewOllamaAgent(model)
		
	case "huggingface":
		apiKey := os.Getenv("HUGGINGFACE_API_KEY")
		if apiKey == "" {
			log.Fatal("HUGGINGFACE_API_KEY environment variable is required for Hugging Face")
		}
		model := os.Getenv("HF_MODEL")
		if model == "" {
			model = "microsoft/DialoGPT-medium"
		}
		freeAgent = NewHuggingFaceAgent(apiKey, model)
		
	case "local":
		baseURL := os.Getenv("LOCAL_AI_URL")
		if baseURL == "" {
			log.Fatal("LOCAL_AI_URL environment variable is required for local AI")
		}
		model := os.Getenv("LOCAL_AI_MODEL")
		if model == "" {
			model = "default"
		}
		freeAgent = NewLocalAgent(baseURL, model)
		
	default:
		log.Fatalf("Unsupported free AI service: %s. Supported: ollama, huggingface, local", serviceType)
	}

	if len(os.Args) > 2 {
		userRequest := strings.Join(os.Args[2:], " ")
		fmt.Printf("Processing request with %s: %s\n", serviceType, userRequest)
		
		response, err := freeAgent.ProcessUserRequest(userRequest)
		if err != nil {
			log.Fatalf("Error: %s", err)
		}

		responseJSON, _ := json.MarshalIndent(response, "", "  ")
		fmt.Printf("Response: %s\n", string(responseJSON))
	} else {
		freeAgent.InteractiveFreeAIMode()
	}
}

func main() {
	_ = loadEnvFile(".env")
	runFreeAI()
} 