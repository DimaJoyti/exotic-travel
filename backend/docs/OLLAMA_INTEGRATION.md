# 🦙 Ollama Integration Guide

This guide explains how to set up and use Ollama for local LLM inference in the Exotic Travel Booking platform.

## Overview

Ollama integration provides:
- **Local LLM inference** without external API dependencies
- **Privacy-focused** AI processing on your own hardware
- **Cost-effective** solution for high-volume usage
- **Offline capability** for air-gapped environments
- **Multiple model support** with easy switching

## Prerequisites

### 1. Install Ollama

**macOS:**
```bash
brew install ollama
```

**Linux:**
```bash
curl -fsSL https://ollama.ai/install.sh | sh
```

**Windows:**
Download from [ollama.ai](https://ollama.ai/download)

### 2. Start Ollama Service

```bash
# Start Ollama service (runs on localhost:11434 by default)
ollama serve
```

### 3. Pull Recommended Models

```bash
# Small, fast model (1.3GB)
ollama pull llama3.2:1b

# Medium model with good performance (2.0GB)
ollama pull llama3.2:3b

# Full model for best quality (4.7GB)
ollama pull llama3.2

# Alternative models
ollama pull mistral        # 4.1GB - Good alternative to Llama
ollama pull phi3          # 2.3GB - Microsoft's efficient model
ollama pull gemma2        # 5.4GB - Google's Gemma model
```

## Configuration

### Environment Variables

```bash
# Optional: Custom Ollama URL (default: http://localhost:11434)
export OLLAMA_BASE_URL=http://localhost:11434

# Optional: Default model (default: llama3.2)
export OLLAMA_DEFAULT_MODEL=llama3.2:3b

# Optional: Request timeout (default: 60s)
export OLLAMA_TIMEOUT=60s
```

### Server Configuration

The Ollama provider is automatically configured in `internal/api/server.go`:

```go
// Ollama provider (for local inference)
ollamaConfig := &providers.LLMConfig{
    Provider: "ollama",
    Model:    "llama3.2", // Default model
    BaseURL:  "http://localhost:11434",
    Timeout:  60 * time.Second,
}
```

## Usage

### 1. Basic API Usage

**Health Check:**
```bash
curl http://localhost:8080/api/v1/ollama/health
```

**List Models:**
```bash
curl http://localhost:8080/api/v1/ollama/models
```

**Generate Response:**
```bash
curl -X POST http://localhost:8080/api/v1/ollama/generate \
  -H "Content-Type: application/json" \
  -d '{
    "model": "llama3.2",
    "prompt": "What are the best travel destinations in Japan?"
  }'
```

**Streaming Response:**
```bash
curl -X POST http://localhost:8080/api/v1/ollama/generate \
  -H "Content-Type: application/json" \
  -d '{
    "model": "llama3.2",
    "prompt": "Plan a 3-day Tokyo itinerary",
    "stream": true
  }'
```

### 2. Programmatic Usage

```go
package main

import (
    "context"
    "fmt"
    "github.com/exotic-travel-booking/backend/internal/services"
)

func main() {
    // Create Ollama service
    ollamaService := services.NewOllamaService("http://localhost:11434")
    
    ctx := context.Background()
    
    // Check health
    if err := ollamaService.CheckHealth(ctx); err != nil {
        panic(err)
    }
    
    // Generate response
    response, err := ollamaService.GenerateResponse(ctx, 
        "llama3.2", 
        "What are unique destinations in Southeast Asia?")
    if err != nil {
        panic(err)
    }
    
    fmt.Println(response)
}
```

### 3. LLM Provider Integration

```go
import "github.com/exotic-travel-booking/backend/internal/llm/providers"

// Create Ollama provider
config := &providers.LLMConfig{
    Provider: "ollama",
    Model:    "llama3.2",
    BaseURL:  "http://localhost:11434",
    Timeout:  60 * time.Second,
}

provider, err := providers.NewOllamaProvider(config)
if err != nil {
    panic(err)
}

// Use with LLM manager
llmManager.AddProvider("ollama", provider)
```

## Model Management

### Pull Models

```bash
# Pull specific model
curl -X POST http://localhost:8080/api/v1/ollama/models/pull \
  -H "Content-Type: application/json" \
  -d '{"model_name": "llama3.2:3b"}'
```

### Ensure Model Availability

```bash
# Ensure model is available (pulls if not present)
curl -X POST http://localhost:8080/api/v1/ollama/models/ensure \
  -H "Content-Type: application/json" \
  -d '{"model_name": "llama3.2"}'
```

### Check Model Status

```bash
curl http://localhost:8080/api/v1/ollama/models/status
```

## Travel-Specific Use Cases

### 1. Destination Recommendations

```go
prompt := `As a travel expert, recommend 5 unique destinations in %s. 
For each destination, provide:
- Name and location
- Best time to visit
- Key attractions
- Estimated budget per day
- Why it's special

Format as JSON.`

response, err := ollamaService.GenerateResponse(ctx, "llama3.2", 
    fmt.Sprintf(prompt, "Southeast Asia"))
```

### 2. Itinerary Planning

```go
prompt := `Create a detailed %d-day itinerary for %s including:
- Daily schedule with timing
- Must-see attractions
- Restaurant recommendations
- Transportation options
- Budget estimates
- Local tips

Keep it practical and realistic.`

response, err := ollamaService.GenerateResponse(ctx, "llama3.2",
    fmt.Sprintf(prompt, 5, "Tokyo, Japan"))
```

### 3. Cultural Insights

```go
prompt := `Provide cultural insights for travelers visiting %s:
- Local customs and etiquette
- Language basics
- Tipping practices
- Dress codes
- Common mistakes to avoid
- Cultural experiences not to miss`

response, err := ollamaService.GenerateResponse(ctx, "llama3.2",
    fmt.Sprintf(prompt, "Morocco"))
```

## Performance Optimization

### Model Selection Guidelines

| Model | Size | Speed | Quality | Use Case |
|-------|------|-------|---------|----------|
| llama3.2:1b | 1.3GB | Fast | Good | Quick responses, high volume |
| llama3.2:3b | 2.0GB | Medium | Better | Balanced performance |
| llama3.2 | 4.7GB | Slower | Best | Detailed planning, complex queries |
| mistral | 4.1GB | Medium | Excellent | Alternative to Llama |
| phi3 | 2.3GB | Fast | Good | Efficient, Microsoft-optimized |

### Hardware Requirements

**Minimum:**
- 8GB RAM
- 4 CPU cores
- 10GB disk space

**Recommended:**
- 16GB+ RAM
- 8+ CPU cores
- 50GB+ disk space
- GPU (optional, for faster inference)

### GPU Acceleration

If you have an NVIDIA GPU:

```bash
# Install CUDA support
ollama pull llama3.2

# Ollama automatically uses GPU if available
# Check GPU usage: nvidia-smi
```

## Troubleshooting

### Common Issues

**1. Ollama not responding:**
```bash
# Check if Ollama is running
curl http://localhost:11434/api/tags

# Restart Ollama
pkill ollama
ollama serve
```

**2. Model not found:**
```bash
# List available models
ollama list

# Pull missing model
ollama pull llama3.2
```

**3. Out of memory:**
```bash
# Use smaller model
ollama pull llama3.2:1b

# Or increase system memory
```

**4. Slow responses:**
- Use smaller models (1b or 3b variants)
- Enable GPU acceleration
- Increase system RAM
- Close other applications

### Monitoring

```bash
# Check Ollama logs
journalctl -u ollama -f

# Monitor resource usage
htop

# Check GPU usage (if available)
nvidia-smi
```

## Integration with Travel Agents

The Ollama provider integrates seamlessly with the travel agent system:

```go
// Travel agent using Ollama
agent := agents.NewTravelAgent(llmManager, toolRegistry)

request := &agents.TravelRequest{
    UserID:      "user123",
    SessionID:   "session456",
    Query:       "Plan a romantic getaway to Santorini",
    Destination: "Santorini, Greece",
    StartDate:   "2024-06-15",
    EndDate:     "2024-06-22",
    Travelers:   2,
    Budget:      5000,
}

response, err := agent.ProcessRequest(ctx, request)
```

## Next Steps

1. **Set up Ollama** following the installation guide
2. **Pull recommended models** for your use case
3. **Test the integration** using the demo script
4. **Configure your travel agents** to use Ollama
5. **Monitor performance** and adjust models as needed

For more advanced configurations and LangGraph integration, see the [LangGraph Guide](LANGGRAPH_INTEGRATION.md).
