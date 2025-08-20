package langchain

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// OutputParser defines the interface for parsing LLM outputs
type OutputParser interface {
	// Parse parses the raw output into structured data
	Parse(ctx context.Context, output string) (interface{}, error)
	
	// GetFormatInstructions returns instructions for the LLM on how to format output
	GetFormatInstructions() string
	
	// GetName returns the parser name
	GetName() string
	
	// Validate validates the parser configuration
	Validate() error
}

// BaseParser provides common functionality for all parsers
type BaseParser struct {
	Name   string `json:"name"`
	tracer trace.Tracer `json:"-"`
}

// NewBaseParser creates a new base parser
func NewBaseParser(name string) *BaseParser {
	return &BaseParser{
		Name:   name,
		tracer: otel.Tracer("langchain.parser"),
	}
}

// GetName returns the parser name
func (p *BaseParser) GetName() string {
	return p.Name
}

// Validate validates the base parser
func (p *BaseParser) Validate() error {
	if p.Name == "" {
		return fmt.Errorf("parser name cannot be empty")
	}
	return nil
}

// JSONParser parses JSON output from LLMs
type JSONParser struct {
	*BaseParser
	Schema map[string]interface{} `json:"schema"`
	Strict bool                   `json:"strict"`
}

// NewJSONParser creates a new JSON parser
func NewJSONParser(name string, schema map[string]interface{}, strict bool) *JSONParser {
	return &JSONParser{
		BaseParser: NewBaseParser(name),
		Schema:     schema,
		Strict:     strict,
	}
}

// Parse parses JSON output
func (p *JSONParser) Parse(ctx context.Context, output string) (interface{}, error) {
	ctx, span := p.tracer.Start(ctx, "json_parser.parse")
	defer span.End()

	span.SetAttributes(
		attribute.String("parser.name", p.Name),
		attribute.Bool("parser.strict", p.Strict),
	)

	// Clean the output - remove markdown code blocks if present
	cleaned := p.cleanJSONOutput(output)
	
	var result interface{}
	err := json.Unmarshal([]byte(cleaned), &result)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Validate against schema if strict mode is enabled
	if p.Strict && p.Schema != nil {
		if err := p.validateAgainstSchema(result); err != nil {
			span.RecordError(err)
			return nil, fmt.Errorf("schema validation failed: %w", err)
		}
	}

	return result, nil
}

// cleanJSONOutput removes markdown code blocks and extra whitespace
func (p *JSONParser) cleanJSONOutput(output string) string {
	// Remove markdown code blocks
	re := regexp.MustCompile("```(?:json)?\n?(.*?)\n?```")
	matches := re.FindStringSubmatch(output)
	if len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}
	
	// If no code blocks, return trimmed output
	return strings.TrimSpace(output)
}

// validateAgainstSchema performs basic schema validation
func (p *JSONParser) validateAgainstSchema(data interface{}) error {
	// This is a simplified schema validation
	// In a production system, you'd use a proper JSON schema validator
	dataMap, ok := data.(map[string]interface{})
	if !ok {
		return fmt.Errorf("expected object, got %T", data)
	}

	// Check required fields if specified in schema
	if required, exists := p.Schema["required"]; exists {
		if requiredFields, ok := required.([]interface{}); ok {
			for _, field := range requiredFields {
				if fieldName, ok := field.(string); ok {
					if _, exists := dataMap[fieldName]; !exists {
						return fmt.Errorf("required field '%s' is missing", fieldName)
					}
				}
			}
		}
	}

	return nil
}

// GetFormatInstructions returns JSON formatting instructions
func (p *JSONParser) GetFormatInstructions() string {
	instructions := "Please format your response as valid JSON."
	
	if p.Schema != nil {
		if schemaJSON, err := json.MarshalIndent(p.Schema, "", "  "); err == nil {
			instructions += fmt.Sprintf("\n\nExpected schema:\n```json\n%s\n```", string(schemaJSON))
		}
	}
	
	instructions += "\n\nEnsure the JSON is properly formatted and can be parsed."
	return instructions
}

// ListParser parses list output from LLMs
type ListParser struct {
	*BaseParser
	Separator   string `json:"separator"`
	Numbered    bool   `json:"numbered"`
	TrimSpaces  bool   `json:"trim_spaces"`
}

// NewListParser creates a new list parser
func NewListParser(name, separator string, numbered, trimSpaces bool) *ListParser {
	return &ListParser{
		BaseParser: NewBaseParser(name),
		Separator:  separator,
		Numbered:   numbered,
		TrimSpaces: trimSpaces,
	}
}

// Parse parses list output
func (p *ListParser) Parse(ctx context.Context, output string) (interface{}, error) {
	ctx, span := p.tracer.Start(ctx, "list_parser.parse")
	defer span.End()

	span.SetAttributes(
		attribute.String("parser.name", p.Name),
		attribute.String("parser.separator", p.Separator),
		attribute.Bool("parser.numbered", p.Numbered),
	)

	lines := strings.Split(output, p.Separator)
	result := make([]string, 0, len(lines))

	for _, line := range lines {
		if p.TrimSpaces {
			line = strings.TrimSpace(line)
		}
		
		if line == "" {
			continue
		}
		
		// Remove numbering if specified
		if p.Numbered {
			line = p.removeNumbering(line)
		}
		
		result = append(result, line)
	}

	span.SetAttributes(attribute.Int("parsed.items", len(result)))
	return result, nil
}

// removeNumbering removes common numbering patterns
func (p *ListParser) removeNumbering(line string) string {
	// Remove patterns like "1. ", "1) ", "- ", "* "
	patterns := []string{
		`^\d+\.\s*`,  // "1. "
		`^\d+\)\s*`,  // "1) "
		`^-\s*`,      // "- "
		`^\*\s*`,     // "* "
		`^•\s*`,      // "• "
	}
	
	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		if re.MatchString(line) {
			return re.ReplaceAllString(line, "")
		}
	}
	
	return line
}

// GetFormatInstructions returns list formatting instructions
func (p *ListParser) GetFormatInstructions() string {
	instructions := fmt.Sprintf("Please format your response as a list separated by '%s'.", p.Separator)
	
	if p.Numbered {
		instructions += " Use numbered items (1. item, 2. item, etc.)."
	} else {
		instructions += " Use bullet points or simple line breaks."
	}
	
	return instructions
}

// KeyValueParser parses key-value pairs from LLM output
type KeyValueParser struct {
	*BaseParser
	KeyValueSeparator string `json:"key_value_separator"`
	PairSeparator     string `json:"pair_separator"`
	TrimSpaces        bool   `json:"trim_spaces"`
}

// NewKeyValueParser creates a new key-value parser
func NewKeyValueParser(name, kvSeparator, pairSeparator string, trimSpaces bool) *KeyValueParser {
	return &KeyValueParser{
		BaseParser:        NewBaseParser(name),
		KeyValueSeparator: kvSeparator,
		PairSeparator:     pairSeparator,
		TrimSpaces:        trimSpaces,
	}
}

// Parse parses key-value output
func (p *KeyValueParser) Parse(ctx context.Context, output string) (interface{}, error) {
	ctx, span := p.tracer.Start(ctx, "key_value_parser.parse")
	defer span.End()

	span.SetAttributes(
		attribute.String("parser.name", p.Name),
		attribute.String("parser.kv_separator", p.KeyValueSeparator),
		attribute.String("parser.pair_separator", p.PairSeparator),
	)

	result := make(map[string]string)
	pairs := strings.Split(output, p.PairSeparator)

	for _, pair := range pairs {
		if p.TrimSpaces {
			pair = strings.TrimSpace(pair)
		}
		
		if pair == "" {
			continue
		}
		
		parts := strings.SplitN(pair, p.KeyValueSeparator, 2)
		if len(parts) == 2 {
			key := parts[0]
			value := parts[1]
			
			if p.TrimSpaces {
				key = strings.TrimSpace(key)
				value = strings.TrimSpace(value)
			}
			
			result[key] = value
		}
	}

	span.SetAttributes(attribute.Int("parsed.pairs", len(result)))
	return result, nil
}

// GetFormatInstructions returns key-value formatting instructions
func (p *KeyValueParser) GetFormatInstructions() string {
	return fmt.Sprintf("Please format your response as key-value pairs using '%s' to separate keys from values and '%s' to separate pairs.",
		p.KeyValueSeparator, p.PairSeparator)
}

// RegexParser parses output using regular expressions
type RegexParser struct {
	*BaseParser
	Pattern     string         `json:"pattern"`
	GroupNames  []string       `json:"group_names"`
	MultiMatch  bool           `json:"multi_match"`
	compiled    *regexp.Regexp `json:"-"`
}

// NewRegexParser creates a new regex parser
func NewRegexParser(name, pattern string, groupNames []string, multiMatch bool) *RegexParser {
	parser := &RegexParser{
		BaseParser: NewBaseParser(name),
		Pattern:    pattern,
		GroupNames: groupNames,
		MultiMatch: multiMatch,
	}
	
	// Compile regex
	if compiled, err := regexp.Compile(pattern); err == nil {
		parser.compiled = compiled
	}
	
	return parser
}

// Parse parses output using regex
func (p *RegexParser) Parse(ctx context.Context, output string) (interface{}, error) {
	ctx, span := p.tracer.Start(ctx, "regex_parser.parse")
	defer span.End()

	span.SetAttributes(
		attribute.String("parser.name", p.Name),
		attribute.String("parser.pattern", p.Pattern),
		attribute.Bool("parser.multi_match", p.MultiMatch),
	)

	if p.compiled == nil {
		err := fmt.Errorf("regex pattern not compiled")
		span.RecordError(err)
		return nil, err
	}

	if p.MultiMatch {
		matches := p.compiled.FindAllStringSubmatch(output, -1)
		result := make([]map[string]string, len(matches))
		
		for i, match := range matches {
			result[i] = p.createMatchMap(match)
		}
		
		span.SetAttributes(attribute.Int("parsed.matches", len(result)))
		return result, nil
	} else {
		match := p.compiled.FindStringSubmatch(output)
		if match == nil {
			return nil, fmt.Errorf("no match found for pattern")
		}
		
		result := p.createMatchMap(match)
		span.SetAttributes(attribute.Int("parsed.groups", len(result)))
		return result, nil
	}
}

// createMatchMap creates a map from regex match groups
func (p *RegexParser) createMatchMap(match []string) map[string]string {
	result := make(map[string]string)
	
	for i, value := range match {
		if i == 0 {
			result["full_match"] = value
		} else if i-1 < len(p.GroupNames) {
			result[p.GroupNames[i-1]] = value
		} else {
			result[fmt.Sprintf("group_%d", i-1)] = value
		}
	}
	
	return result
}

// GetFormatInstructions returns regex formatting instructions
func (p *RegexParser) GetFormatInstructions() string {
	instructions := "Please format your response to match the expected pattern."
	
	if len(p.GroupNames) > 0 {
		instructions += fmt.Sprintf(" Expected groups: %v", p.GroupNames)
	}
	
	return instructions
}

// Validate validates the regex parser
func (p *RegexParser) Validate() error {
	if err := p.BaseParser.Validate(); err != nil {
		return err
	}
	
	if p.Pattern == "" {
		return fmt.Errorf("regex pattern cannot be empty")
	}
	
	// Try to compile pattern
	_, err := regexp.Compile(p.Pattern)
	if err != nil {
		return fmt.Errorf("invalid regex pattern: %w", err)
	}
	
	return nil
}

// NumberParser parses numeric values from output
type NumberParser struct {
	*BaseParser
	Type        string  `json:"type"`         // "int", "float"
	DefaultValue interface{} `json:"default_value"`
}

// NewNumberParser creates a new number parser
func NewNumberParser(name, numberType string, defaultValue interface{}) *NumberParser {
	return &NumberParser{
		BaseParser:   NewBaseParser(name),
		Type:         numberType,
		DefaultValue: defaultValue,
	}
}

// Parse parses numeric output
func (p *NumberParser) Parse(ctx context.Context, output string) (interface{}, error) {
	ctx, span := p.tracer.Start(ctx, "number_parser.parse")
	defer span.End()

	span.SetAttributes(
		attribute.String("parser.name", p.Name),
		attribute.String("parser.type", p.Type),
	)

	// Extract first number from output
	re := regexp.MustCompile(`-?\d+(?:\.\d+)?`)
	match := re.FindString(strings.TrimSpace(output))
	
	if match == "" {
		if p.DefaultValue != nil {
			return p.DefaultValue, nil
		}
		return nil, fmt.Errorf("no number found in output")
	}

	switch p.Type {
	case "int":
		value, err := strconv.Atoi(match)
		if err != nil {
			span.RecordError(err)
			return p.DefaultValue, fmt.Errorf("failed to parse integer: %w", err)
		}
		return value, nil
		
	case "float":
		value, err := strconv.ParseFloat(match, 64)
		if err != nil {
			span.RecordError(err)
			return p.DefaultValue, fmt.Errorf("failed to parse float: %w", err)
		}
		return value, nil
		
	default:
		return match, nil
	}
}

// GetFormatInstructions returns number formatting instructions
func (p *NumberParser) GetFormatInstructions() string {
	switch p.Type {
	case "int":
		return "Please provide your answer as an integer number."
	case "float":
		return "Please provide your answer as a decimal number."
	default:
		return "Please provide your answer as a number."
	}
}

// ParserRegistry manages a collection of output parsers
type ParserRegistry struct {
	parsers map[string]OutputParser
	tracer  trace.Tracer
}

// NewParserRegistry creates a new parser registry
func NewParserRegistry() *ParserRegistry {
	return &ParserRegistry{
		parsers: make(map[string]OutputParser),
		tracer:  otel.Tracer("langchain.parser_registry"),
	}
}

// RegisterParser registers an output parser
func (pr *ParserRegistry) RegisterParser(parser OutputParser) error {
	if err := parser.Validate(); err != nil {
		return fmt.Errorf("parser validation failed: %w", err)
	}
	
	pr.parsers[parser.GetName()] = parser
	return nil
}

// GetParser retrieves a parser by name
func (pr *ParserRegistry) GetParser(name string) (OutputParser, error) {
	parser, exists := pr.parsers[name]
	if !exists {
		return nil, fmt.Errorf("parser '%s' not found", name)
	}
	return parser, nil
}

// ListParsers returns all registered parser names
func (pr *ParserRegistry) ListParsers() []string {
	names := make([]string, 0, len(pr.parsers))
	for name := range pr.parsers {
		names = append(names, name)
	}
	return names
}

// ParseWithParser parses output using a specific parser
func (pr *ParserRegistry) ParseWithParser(ctx context.Context, parserName, output string) (interface{}, error) {
	parser, err := pr.GetParser(parserName)
	if err != nil {
		return nil, err
	}
	
	return parser.Parse(ctx, output)
}
