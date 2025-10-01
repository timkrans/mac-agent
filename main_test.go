package main

import (
	"testing"
	"time"
)

func TestNewAgent(t *testing.T) {
	agent := NewAgent()
	
	if agent == nil {
		t.Fatal("Agent should not be nil")
	}
	
	if len(agent.allowedCommands) == 0 {
		t.Error("Allowed commands should not be empty")
	}
}

func TestIsCommandAllowed(t *testing.T) {
	agent := NewAgent()
	allowedCommands := []string{"ls", "pwd", "whoami", "date", "ps"}
	for _, cmd := range allowedCommands {
		if !agent.isCommandAllowed(cmd) {
			t.Errorf("Command '%s' should be allowed", cmd)
		}
	}
	disallowedCommands := []string{"sudo", "rm -rf /", "format", "dd"}
	for _, cmd := range disallowedCommands {
		if agent.isCommandAllowed(cmd) {
			t.Errorf("Command '%s' should not be allowed", cmd)
		}
	}
}

func TestExecuteCommand(t *testing.T) {
	agent := NewAgent()
	req := CommandRequest{Command: "pwd"}
	response := agent.ExecuteCommand(req)
	if !response.Success {
		t.Errorf("Expected successful execution, got error: %s", response.Error)
	}
	
	if response.ExitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", response.ExitCode)
	}
	
	if response.Output == "" {
		t.Error("Expected non-empty output")
	}
	
	req = CommandRequest{Command: "sudo"}
	response = agent.ExecuteCommand(req)
	
	if response.Success {
		t.Error("Expected failed execution for disallowed command")
	}
	
	if response.Error == "" {
		t.Error("Expected error message for disallowed command")
	}
}

func TestExecuteCommandWithTimeout(t *testing.T) {
	agent := NewAgent()
	req := CommandRequest{
		Command: "sleep",
		Args:    []string{"2"},
		Timeout: 1,
	}
	
	start := time.Now()
	response := agent.ExecuteCommand(req)
	duration := time.Since(start)

	if response.Success {
		t.Error("Expected timeout failure")
	}
	
	if duration > 2*time.Second {
		t.Errorf("Command should have timed out, took %v", duration)
	}
}

func TestGetSystemInfo(t *testing.T) {
	agent := NewAgent()
	info := agent.GetSystemInfo()
	
	requiredKeys := []string{"os", "arch", "version"}
	for _, key := range requiredKeys {
		if _, exists := info[key]; !exists {
			t.Errorf("System info should contain key: %s", key)
		}
	}
	
	if info["os"] != "darwin" {
		t.Errorf("Expected OS to be 'darwin', got %s", info["os"])
	}
}

func TestCommandResponseStructure(t *testing.T) {
	agent := NewAgent()
	req := CommandRequest{Command: "echo", Args: []string{"hello"}}
	response := agent.ExecuteCommand(req)
	
	if response.Timestamp == "" {
		t.Error("Response should have timestamp")
	}
	
	if response.Duration == "" {
		t.Error("Response should have duration")
	}
	
	if !response.Success {
		t.Errorf("Echo command should succeed, got error: %s", response.Error)
	}
	
	if response.Output != "hello\n" {
		t.Errorf("Expected output 'hello\\n', got '%s'", response.Output)
	}
} 