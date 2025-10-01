# macOS AI Command Agent

An intelligent command agent for macOS that uses free AI services to understand natural language requests and execute safe system commands.

## Features

-  **AI-Powered Command Generation**: Uses free AI services to translate natural language into appropriate system commands
-  **Security-First Design**: Only executes whitelisted, safe commands (no sudo, rm -rf, etc.)
-  **Multiple AI Backends**: Supports Ollama, Hugging Face, and local AI services
-  **Interactive Mode**: Chat with the agent in real-time
-  **Command Timeouts**: Built-in timeout protection for long-running commands
-  **Detailed Output**: JSON responses with execution results, timing, and confidence levels
-  **Comprehensive Testing**: Full test suite with unit tests for all components

## Supported AI Services

### 1. Ollama (Default)
- **Local AI models** - runs entirely on your machine
- **No API keys required**
- **Default model**: `llama3.2`

### 2. Hugging Face
- **Cloud-based inference**
- **Requires API key**
- **Default model**: `microsoft/DialoGPT-medium`

### 3. Local AI Service
- **Custom endpoints**
- **Compatible with any OpenAI-compatible API**
- **Flexible model selection**

## Installation

### Prerequisites
- Go 1.21 or later
- macOS (designed specifically for macOS systems)

### Build from Source
```bash
git clone https://github.com/timkrans/mac-agent
cd mac-agent
go build -o mac-agent
```

## Configuration

Create a `.env` file in the project directory:

```env
# AI Service Configuration
FREE_AI_SERVICE=ollama  # Options: ollama, huggingface, local

# Ollama Configuration
OLLAMA_MODEL=llama3.2

# Hugging Face Configuration (if using HF)
HUGGINGFACE_API_KEY=your_api_key_here
HF_MODEL=microsoft/DialoGPT-medium

# Local AI Configuration (if using local service)
LOCAL_AI_URL=http://localhost:8080
LOCAL_AI_MODEL=your_model_name
```

## Usage

### Interactive Mode
```bash
./mac-agent
```
This starts an interactive chat session where you can ask the agent to perform tasks using natural language.

### Single Command Mode
```bash
./mac-agent "list all files in the current directory"
```

### Example Interactions

**User**: "Show me all running processes"
**Agent**: 
- Thoughts: "The user wants to see running processes, which I can do with the 'ps' command"
- Commands: `ps aux`
- Explanation: "I'll use the 'ps aux' command to show all running processes with detailed information"
- Confidence: 95%

**User**: "Check disk usage for my home directory"
**Agent**:
- Thoughts: "The user wants to check disk usage, I'll use 'du' to show directory sizes"
- Commands: `du -sh ~`
- Explanation: "I'll use 'du -sh ~' to show the total size of your home directory in human-readable format"
- Confidence: 90%

## Allowed Commands

The agent maintains a whitelist of safe commands:

**File Operations**: `ls`, `pwd`, `cat`, `head`, `tail`, `grep`, `find`, `file`, `stat`, `cp`, `mv`, `rm`, `mkdir`, `rmdir`, `chmod`, `chown`

**System Information**: `whoami`, `date`, `uptime`, `ps`, `top`, `df`, `du`, `which`, `whereis`, `system_profiler`, `sw_vers`, `defaults`, `launchctl`

**Network Tools**: `netstat`, `lsof`, `ifconfig`, `ping`, `nslookup`, `dig`, `curl`, `wget`

**Text Processing**: `wc`, `sort`, `uniq`, `echo`

## Security Features

- **Command Whitelist**: Only pre-approved safe commands are allowed
- **No Elevated Privileges**: No sudo or administrative commands
- **Dangerous Command Prevention**: Blocks commands like `rm -rf /`, `format`, `dd`
- **Timeout Protection**: Commands automatically timeout after 30 seconds (configurable)
- **Error Handling**: Graceful error handling with detailed error messages

## Testing

Run the test suite:
```bash
go test -v
```

The test suite covers:
- Agent initialization
- Command whitelist validation
- Command execution with various scenarios
- Timeout handling
- System information gathering
- Response structure validation

## API Response Format

The agent returns structured JSON responses:

```json
{
  "thoughts": "AI reasoning about the request",
  "commands": [
    {
      "command": "ls",
      "args": ["-la"],
      "timeout": 30
    }
  ],
  "explanation": "Human-readable explanation of actions",
  "confidence": 0.95,
  "results": [
    {
      "success": true,
      "output": "command output",
      "error": "",
      "exit_code": 0,
      "duration": "50ms",
      "timestamp": "2024-01-01T12:00:00Z"
    }
  ]
}
```

## Development

### Project Structure
```
├── main.go                 # Main application entry point
├── free_ai_integration.go  # AI service integrations
├── load_env.go            # Environment variable loading
├── main_test.go           # Test suite
├── go.mod                 # Go module definition
└── README.md             # This file
```

### Adding New AI Services
1. Implement the service interface in `free_ai_integration.go`
2. Add service type handling in the main function
3. Update environment variable configuration
4. Add tests for the new service

## Troubleshooting

### Ollama Issues
- Ensure Ollama is installed and running: `ollama serve`
- Check if the model is available: `ollama list`
- Pull the model if needed: `ollama pull llama3.2`

### Hugging Face Issues
- Verify your API key is valid
- Check rate limits and quotas
- Ensure the model is accessible with your account

### Command Execution Issues
- Verify the command is in the whitelist
- Check system permissions
- Review timeout settings for long-running commands

## Disclaimer

This tool is designed for educational and productivity purposes. Always review commands before execution and use responsibly. The author is not responsible for any damage caused by misuse of this tool.

## Future plans

Add openai and other option for models in this system.