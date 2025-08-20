package langgraph

import (
	"context"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// Edge represents a connection between two nodes
type Edge struct {
	From        string     `json:"from"`
	To          string     `json:"to"`
	Description string     `json:"description"`
	Condition   Condition  `json:"condition,omitempty"`
	Weight      float64    `json:"weight,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// NewEdge creates a new edge
func NewEdge(from, to, description string) Edge {
	return Edge{
		From:        from,
		To:          to,
		Description: description,
		Weight:      1.0,
		Metadata:    make(map[string]interface{}),
	}
}

// NewConditionalEdge creates a new edge with a condition
func NewConditionalEdge(from, to, description string, condition Condition) Edge {
	return Edge{
		From:        from,
		To:          to,
		Description: description,
		Condition:   condition,
		Weight:      1.0,
		Metadata:    make(map[string]interface{}),
	}
}

// Condition represents a condition that can be evaluated
type Condition interface {
	// Evaluate evaluates the condition against the given state
	Evaluate(ctx context.Context, state *State) (bool, error)
	
	// GetDescription returns a human-readable description of the condition
	GetDescription() string
}

// BaseCondition provides common functionality for conditions
type BaseCondition struct {
	Description string `json:"description"`
	tracer      trace.Tracer `json:"-"`
}

// NewBaseCondition creates a new base condition
func NewBaseCondition(description string) *BaseCondition {
	return &BaseCondition{
		Description: description,
		tracer:      otel.Tracer("langgraph.condition"),
	}
}

// GetDescription returns the condition description
func (c *BaseCondition) GetDescription() string {
	return c.Description
}

// StateKeyCondition checks if a state key exists and optionally matches a value
type StateKeyCondition struct {
	*BaseCondition
	Key           string      `json:"key"`
	ExpectedValue interface{} `json:"expected_value,omitempty"`
	Operator      string      `json:"operator"` // "exists", "equals", "not_equals", "greater", "less", "contains"
}

// NewStateKeyCondition creates a condition that checks if a state key exists
func NewStateKeyCondition(key string) *StateKeyCondition {
	return &StateKeyCondition{
		BaseCondition: NewBaseCondition(fmt.Sprintf("key '%s' exists", key)),
		Key:           key,
		Operator:      "exists",
	}
}

// NewStateValueCondition creates a condition that checks if a state key equals a value
func NewStateValueCondition(key string, expectedValue interface{}, operator string) *StateKeyCondition {
	return &StateKeyCondition{
		BaseCondition: NewBaseCondition(fmt.Sprintf("key '%s' %s %v", key, operator, expectedValue)),
		Key:           key,
		ExpectedValue: expectedValue,
		Operator:      operator,
	}
}

// Evaluate evaluates the state key condition
func (c *StateKeyCondition) Evaluate(ctx context.Context, state *State) (bool, error) {
	ctx, span := c.tracer.Start(ctx, "state_key_condition.evaluate")
	defer span.End()

	span.SetAttributes(
		attribute.String("condition.key", c.Key),
		attribute.String("condition.operator", c.Operator),
	)

	value, exists := state.Get(c.Key)
	
	switch c.Operator {
	case "exists":
		result := exists
		span.SetAttributes(attribute.Bool("condition.result", result))
		return result, nil
		
	case "not_exists":
		result := !exists
		span.SetAttributes(attribute.Bool("condition.result", result))
		return result, nil
		
	case "equals":
		if !exists {
			span.SetAttributes(attribute.Bool("condition.result", false))
			return false, nil
		}
		result := reflect.DeepEqual(value, c.ExpectedValue)
		span.SetAttributes(attribute.Bool("condition.result", result))
		return result, nil
		
	case "not_equals":
		if !exists {
			span.SetAttributes(attribute.Bool("condition.result", true))
			return true, nil
		}
		result := !reflect.DeepEqual(value, c.ExpectedValue)
		span.SetAttributes(attribute.Bool("condition.result", result))
		return result, nil
		
	case "greater":
		if !exists {
			span.SetAttributes(attribute.Bool("condition.result", false))
			return false, nil
		}
		result, err := c.compareNumbers(value, c.ExpectedValue, ">")
		if err != nil {
			span.RecordError(err)
			return false, err
		}
		span.SetAttributes(attribute.Bool("condition.result", result))
		return result, nil
		
	case "less":
		if !exists {
			span.SetAttributes(attribute.Bool("condition.result", false))
			return false, nil
		}
		result, err := c.compareNumbers(value, c.ExpectedValue, "<")
		if err != nil {
			span.RecordError(err)
			return false, err
		}
		span.SetAttributes(attribute.Bool("condition.result", result))
		return result, nil
		
	case "contains":
		if !exists {
			span.SetAttributes(attribute.Bool("condition.result", false))
			return false, nil
		}
		result, err := c.checkContains(value, c.ExpectedValue)
		if err != nil {
			span.RecordError(err)
			return false, err
		}
		span.SetAttributes(attribute.Bool("condition.result", result))
		return result, nil
		
	default:
		err := fmt.Errorf("unknown operator: %s", c.Operator)
		span.RecordError(err)
		return false, err
	}
}

// compareNumbers compares two numeric values
func (c *StateKeyCondition) compareNumbers(a, b interface{}, operator string) (bool, error) {
	aFloat, err := c.toFloat64(a)
	if err != nil {
		return false, fmt.Errorf("cannot convert %v to number", a)
	}
	
	bFloat, err := c.toFloat64(b)
	if err != nil {
		return false, fmt.Errorf("cannot convert %v to number", b)
	}
	
	switch operator {
	case ">":
		return aFloat > bFloat, nil
	case "<":
		return aFloat < bFloat, nil
	default:
		return false, fmt.Errorf("unsupported numeric operator: %s", operator)
	}
}

// toFloat64 converts a value to float64
func (c *StateKeyCondition) toFloat64(value interface{}) (float64, error) {
	switch v := value.(type) {
	case float64:
		return v, nil
	case float32:
		return float64(v), nil
	case int:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case int32:
		return float64(v), nil
	case string:
		return strconv.ParseFloat(v, 64)
	default:
		return 0, fmt.Errorf("cannot convert %T to float64", value)
	}
}

// checkContains checks if a value contains another value
func (c *StateKeyCondition) checkContains(container, item interface{}) (bool, error) {
	switch cont := container.(type) {
	case string:
		itemStr, ok := item.(string)
		if !ok {
			return false, fmt.Errorf("cannot check if string contains %T", item)
		}
		return strings.Contains(cont, itemStr), nil
		
	case []interface{}:
		for _, element := range cont {
			if reflect.DeepEqual(element, item) {
				return true, nil
			}
		}
		return false, nil
		
	case map[string]interface{}:
		itemStr, ok := item.(string)
		if !ok {
			return false, fmt.Errorf("cannot check if map contains %T key", item)
		}
		_, exists := cont[itemStr]
		return exists, nil
		
	default:
		return false, fmt.Errorf("cannot check contains for type %T", container)
	}
}

// FunctionCondition evaluates a custom function
type FunctionCondition struct {
	*BaseCondition
	Function func(ctx context.Context, state *State) (bool, error) `json:"-"`
}

// NewFunctionCondition creates a condition based on a custom function
func NewFunctionCondition(description string, fn func(ctx context.Context, state *State) (bool, error)) *FunctionCondition {
	return &FunctionCondition{
		BaseCondition: NewBaseCondition(description),
		Function:      fn,
	}
}

// Evaluate evaluates the function condition
func (c *FunctionCondition) Evaluate(ctx context.Context, state *State) (bool, error) {
	ctx, span := c.tracer.Start(ctx, "function_condition.evaluate")
	defer span.End()

	if c.Function == nil {
		err := fmt.Errorf("condition function is nil")
		span.RecordError(err)
		return false, err
	}

	result, err := c.Function(ctx, state)
	if err != nil {
		span.RecordError(err)
		return false, fmt.Errorf("condition function failed: %w", err)
	}

	span.SetAttributes(attribute.Bool("condition.result", result))
	return result, nil
}

// AndCondition combines multiple conditions with AND logic
type AndCondition struct {
	*BaseCondition
	Conditions []Condition `json:"conditions"`
}

// NewAndCondition creates a condition that requires all sub-conditions to be true
func NewAndCondition(conditions ...Condition) *AndCondition {
	descriptions := make([]string, len(conditions))
	for i, cond := range conditions {
		descriptions[i] = cond.GetDescription()
	}
	
	return &AndCondition{
		BaseCondition: NewBaseCondition(fmt.Sprintf("(%s)", strings.Join(descriptions, " AND "))),
		Conditions:    conditions,
	}
}

// Evaluate evaluates the AND condition
func (c *AndCondition) Evaluate(ctx context.Context, state *State) (bool, error) {
	ctx, span := c.tracer.Start(ctx, "and_condition.evaluate")
	defer span.End()

	span.SetAttributes(attribute.Int("condition.sub_conditions", len(c.Conditions)))

	for i, condition := range c.Conditions {
		result, err := condition.Evaluate(ctx, state)
		if err != nil {
			span.RecordError(err)
			return false, fmt.Errorf("sub-condition %d failed: %w", i, err)
		}
		
		if !result {
			span.SetAttributes(attribute.Bool("condition.result", false))
			return false, nil
		}
	}

	span.SetAttributes(attribute.Bool("condition.result", true))
	return true, nil
}

// OrCondition combines multiple conditions with OR logic
type OrCondition struct {
	*BaseCondition
	Conditions []Condition `json:"conditions"`
}

// NewOrCondition creates a condition that requires at least one sub-condition to be true
func NewOrCondition(conditions ...Condition) *OrCondition {
	descriptions := make([]string, len(conditions))
	for i, cond := range conditions {
		descriptions[i] = cond.GetDescription()
	}
	
	return &OrCondition{
		BaseCondition: NewBaseCondition(fmt.Sprintf("(%s)", strings.Join(descriptions, " OR "))),
		Conditions:    conditions,
	}
}

// Evaluate evaluates the OR condition
func (c *OrCondition) Evaluate(ctx context.Context, state *State) (bool, error) {
	ctx, span := c.tracer.Start(ctx, "or_condition.evaluate")
	defer span.End()

	span.SetAttributes(attribute.Int("condition.sub_conditions", len(c.Conditions)))

	for i, condition := range c.Conditions {
		result, err := condition.Evaluate(ctx, state)
		if err != nil {
			span.RecordError(err)
			return false, fmt.Errorf("sub-condition %d failed: %w", i, err)
		}
		
		if result {
			span.SetAttributes(attribute.Bool("condition.result", true))
			return true, nil
		}
	}

	span.SetAttributes(attribute.Bool("condition.result", false))
	return false, nil
}

// NotCondition negates a condition
type NotCondition struct {
	*BaseCondition
	Condition Condition `json:"condition"`
}

// NewNotCondition creates a condition that negates another condition
func NewNotCondition(condition Condition) *NotCondition {
	return &NotCondition{
		BaseCondition: NewBaseCondition(fmt.Sprintf("NOT (%s)", condition.GetDescription())),
		Condition:     condition,
	}
}

// Evaluate evaluates the NOT condition
func (c *NotCondition) Evaluate(ctx context.Context, state *State) (bool, error) {
	ctx, span := c.tracer.Start(ctx, "not_condition.evaluate")
	defer span.End()

	result, err := c.Condition.Evaluate(ctx, state)
	if err != nil {
		span.RecordError(err)
		return false, fmt.Errorf("negated condition failed: %w", err)
	}

	negatedResult := !result
	span.SetAttributes(attribute.Bool("condition.result", negatedResult))
	return negatedResult, nil
}

// AlwaysTrueCondition always returns true
type AlwaysTrueCondition struct {
	*BaseCondition
}

// NewAlwaysTrueCondition creates a condition that always returns true
func NewAlwaysTrueCondition() *AlwaysTrueCondition {
	return &AlwaysTrueCondition{
		BaseCondition: NewBaseCondition("always true"),
	}
}

// Evaluate always returns true
func (c *AlwaysTrueCondition) Evaluate(ctx context.Context, state *State) (bool, error) {
	return true, nil
}

// AlwaysFalseCondition always returns false
type AlwaysFalseCondition struct {
	*BaseCondition
}

// NewAlwaysFalseCondition creates a condition that always returns false
func NewAlwaysFalseCondition() *AlwaysFalseCondition {
	return &AlwaysFalseCondition{
		BaseCondition: NewBaseCondition("always false"),
	}
}

// Evaluate always returns false
func (c *AlwaysFalseCondition) Evaluate(ctx context.Context, state *State) (bool, error) {
	return false, nil
}
