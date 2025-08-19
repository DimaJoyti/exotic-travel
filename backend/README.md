# Exotic Travel Booking - LLM-Powered Backend

A comprehensive LLM-powered travel booking system built with Go, featuring intelligent travel planning, flight and hotel search, weather integration, and location services.

## 🚀 Features

### Core LLM Framework
- **Multi-Provider LLM Support**: OpenAI, Anthropic, and local providers
- **Advanced Chain Execution**: Sequential, parallel, conditional, and map-reduce workflows
- **Conversation Memory**: Persistent conversation context with automatic summarization
- **Prompt Templates**: Reusable, parameterized prompt templates for travel scenarios

### Travel Tools & APIs
- **Flight Search**: Comprehensive flight search with multiple filters
- **Hotel Search**: Hotel and accommodation search with amenities filtering
- **Weather Integration**: Real-time weather data and forecasts
- **Location Services**: Place search, geocoding, and reverse geocoding

### Intelligent Agents
- **Travel Agent**: AI-powered travel planning with workflow orchestration
- **Intent Recognition**: Automatic classification of travel requests
- **Personalized Recommendations**: Context-aware travel suggestions

### Workflow System (LangGraph-inspired)
- **Visual Workflow Designer**: Graph-based workflow creation
- **Node Types**: LLM, Tool, Decision, Parallel, Transform nodes
- **Conditional Routing**: Dynamic workflow paths based on conditions
- **Execution Monitoring**: Real-time workflow execution tracking

## 🏗️ Architecture

```
backend/
├── cmd/
│   └── llm-server/          # LLM-powered server entry point
├── internal/
│   ├── api/                 # HTTP API layer
│   │   ├── handlers/        # Request handlers
│   │   ├── middleware/      # HTTP middleware
│   │   └── server.go        # Server setup
│   ├── agents/              # AI agents
│   │   └── travel_agent.go  # Main travel planning agent
│   ├── llm/                 # LLM framework
│   │   ├── providers/       # LLM provider implementations
│   │   ├── chains/          # Chain execution engine
│   │   ├── memory/          # Conversation memory
│   │   ├── prompts/         # Prompt template system
│   │   └── manager.go       # LLM manager
│   ├── workflow/            # Workflow system
│   │   ├── types.go         # Core types and interfaces
│   │   ├── graph.go         # Workflow graph implementation
│   │   ├── executor.go      # Workflow execution engine
│   │   ├── nodes.go         # Node implementations
│   │   └── registry.go      # Workflow registry
│   └── tools/               # External tools
│       ├── flight_search.go # Flight search tool
│       ├── hotel_search.go  # Hotel search tool
│       ├── weather.go       # Weather tool
│       ├── location.go      # Location services tool
│       └── tool.go          # Base tool interface
└── go.mod                   # Go module definition
```

## 🛠️ Installation & Setup

### Prerequisites
- Go 1.22 or later
- OpenTelemetry (optional, for observability)

### Environment Variables
```bash
# LLM Provider API Keys (optional for development)
export OPENAI_API_KEY="your-openai-api-key"
export ANTHROPIC_API_KEY="your-anthropic-api-key"

# External API Keys (optional for development)
export AMADEUS_API_KEY="your-amadeus-api-key"
export WEATHERAPI_KEY="your-weatherapi-key"
export GOOGLE_PLACES_API_KEY="your-google-places-api-key"
```

### Build & Run
```bash
# Clone the repository
git clone https://github.com/exotic-travel-booking/backend.git
cd backend

# Install dependencies
go mod download

# Build the LLM server
go build -o bin/llm-server ./cmd/llm-server

# Run the server
./bin/llm-server -port 8081 -host 0.0.0.0
```

### Development Mode
```bash
# Run directly with Go
go run ./cmd/llm-server/main.go -port 8081
```

## 📡 API Endpoints

### Travel Planning
- `POST /api/v1/travel/plan` - Comprehensive trip planning
- `GET /api/v1/travel/flights/search` - Flight search
- `GET /api/v1/travel/hotels/search` - Hotel search
- `GET /api/v1/travel/weather` - Weather information
- `GET /api/v1/travel/locations/search` - Location search
- `GET /api/v1/travel/tools` - Available tools information

### System
- `GET /health` - Health check
- `GET /` - API information

## 🔧 Usage Examples

### Comprehensive Trip Planning
```bash
curl -X POST http://localhost:8081/api/v1/travel/plan \
  -H "Content-Type: application/json" \
  -d '{
    "query": "Plan a 5-day romantic trip to Paris for 2 people in March",
    "destination": "Paris",
    "start_date": "2024-03-15",
    "end_date": "2024-03-20",
    "travelers": 2,
    "budget": "$3000",
    "travel_style": "romantic",
    "interests": ["museums", "fine dining", "architecture"]
  }'
```

### Flight Search
```bash
curl "http://localhost:8081/api/v1/travel/flights/search?origin=JFK&destination=CDG&departure_date=2024-03-15&return_date=2024-03-20&adults=2&class=economy"
```

### Hotel Search
```bash
curl "http://localhost:8081/api/v1/travel/hotels/search?location=Paris&check_in_date=2024-03-15&check_out_date=2024-03-20&adults=2&rooms=1&star_rating=4"
```

### Weather Information
```bash
curl "http://localhost:8081/api/v1/travel/weather?location=Paris&days=5&units=metric"
```

## 🧠 LLM Framework Features

### Chain Types
- **Sequential**: Execute steps one after another
- **Parallel**: Execute multiple steps concurrently
- **Conditional**: Route based on conditions
- **Map-Reduce**: Parallel processing with aggregation

### Memory Types
- **In-Memory**: Fast, temporary storage
- **Conversation**: Session-based conversation context
- **Summary**: Automatic conversation summarization

### Prompt Templates
- **Travel Intent Extraction**: Extract structured travel requirements
- **Flight Search**: Generate flight search prompts
- **Hotel Search**: Generate hotel search prompts
- **Itinerary Planning**: Create detailed travel itineraries
- **Recommendations**: Generate personalized travel advice

## 🔄 Workflow System

### Node Types
- **LLM Node**: Call language models with prompts
- **Tool Node**: Execute external tools
- **Decision Node**: Route based on conditions
- **Parallel Node**: Execute multiple sub-nodes concurrently
- **Transform Node**: Transform data between steps

### Workflow Examples
```go
// Create a flight search workflow
builder := workflow.NewWorkflowBuilder("flight_search", "Flight Search", "Search and analyze flights")

// Add nodes
builder.AddToolNode("search_flights", "Search Flights", flightTool)
builder.AddLLMNode("analyze_flights", "Analyze Results", provider, promptTemplate)

// Add edges
builder.AddSimpleEdge("search_flights", "analyze_flights")

// Set start node and build
builder.SetStartNode("search_flights")
workflow, err := builder.Build()
```

## 🔍 Observability

The system includes comprehensive observability with OpenTelemetry:
- **Distributed Tracing**: Track requests across services
- **Metrics**: Monitor performance and usage
- **Structured Logging**: Detailed request/response logging
- **Error Tracking**: Comprehensive error reporting

## 🧪 Testing

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific package tests
go test ./internal/llm/...
go test ./internal/workflow/...
go test ./internal/tools/...
```

## 🚀 Deployment

### Docker
```dockerfile
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o llm-server ./cmd/llm-server

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/llm-server .
EXPOSE 8081
CMD ["./llm-server"]
```

### Kubernetes
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: exotic-travel-llm
spec:
  replicas: 3
  selector:
    matchLabels:
      app: exotic-travel-llm
  template:
    metadata:
      labels:
        app: exotic-travel-llm
    spec:
      containers:
      - name: llm-server
        image: exotic-travel/llm-backend:latest
        ports:
        - containerPort: 8081
        env:
        - name: OPENAI_API_KEY
          valueFrom:
            secretKeyRef:
              name: llm-secrets
              key: openai-api-key
```

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🔗 Related Projects

- [Frontend Application](../frontend) - React-based user interface
- [Mobile App](../mobile) - React Native mobile application
- [Infrastructure](../infrastructure) - Deployment and infrastructure code

## 📞 Support

For support and questions:
- Create an issue in this repository
- Contact the development team
- Check the documentation in the `/docs` directory
