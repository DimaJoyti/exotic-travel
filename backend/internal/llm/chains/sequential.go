package chains

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// SequentialChain executes steps one after another
type SequentialChain struct {
	*BaseChain
}

// NewSequentialChain creates a new sequential chain
func NewSequentialChain(name, description string) *SequentialChain {
	return &SequentialChain{
		BaseChain: NewBaseChain(name, description),
	}
}

// Execute executes all steps in sequence
func (c *SequentialChain) Execute(ctx context.Context, input *ChainInput) (*ChainOutput, error) {
	ctx, span := c.tracer.Start(ctx, "sequential_chain.execute")
	defer span.End()

	span.SetAttributes(
		attribute.String("chain.name", c.name),
		attribute.String("chain.type", "sequential"),
		attribute.Int("chain.steps", len(c.steps)),
	)

	if len(c.steps) == 0 {
		return &ChainOutput{
			Result:  "No steps to execute",
			Context: input.Context,
		}, nil
	}

	// Initialize current input with the provided input
	currentInput := input
	var finalOutput *ChainOutput

	// Execute each step in sequence
	for i, step := range c.steps {
		stepSpan := fmt.Sprintf("step_%d_%s", i, step.GetName())
		stepCtx, stepSpanTracer := c.tracer.Start(ctx, stepSpan)

		stepSpanTracer.SetAttributes(
			attribute.String("step.name", step.GetName()),
			attribute.String("step.description", step.GetDescription()),
			attribute.Int("step.index", i),
		)

		// Execute the step
		output, err := step.Execute(stepCtx, currentInput)
		if err != nil {
			stepSpanTracer.RecordError(err)
			stepSpanTracer.End()
			span.RecordError(err)
			return nil, fmt.Errorf("step %d (%s) failed: %w", i, step.GetName(), err)
		}

		stepSpanTracer.End()

		// Store in memory if available
		if c.memory != nil {
			memoryKey := fmt.Sprintf("%s_step_%d", c.name, i)
			if err := c.memory.Store(ctx, memoryKey, output); err != nil {
				// Log error but don't fail the chain
				span.AddEvent("memory_store_failed", trace.WithAttributes(attribute.String("error", err.Error())))
			}
		}

		// Prepare input for next step
		// The output of current step becomes part of the input for the next step
		currentInput = &ChainInput{
			Variables: mergeVariables(currentInput.Variables, map[string]interface{}{
				"previous_result": output.Result,
				"step_output":     output,
			}),
			Messages: output.Messages,
			Context:  mergeContext(currentInput.Context, output.Context),
		}

		finalOutput = output
	}

	// Enhance final output with chain metadata
	if finalOutput != nil {
		if finalOutput.Metadata == nil {
			finalOutput.Metadata = make(map[string]interface{})
		}
		finalOutput.Metadata["chain_name"] = c.name
		finalOutput.Metadata["chain_type"] = "sequential"
		finalOutput.Metadata["steps_executed"] = len(c.steps)
	}

	return finalOutput, nil
}

// mergeVariables merges two variable maps, with the second map taking precedence
func mergeVariables(vars1, vars2 map[string]interface{}) map[string]interface{} {
	if vars1 == nil {
		vars1 = make(map[string]interface{})
	}
	if vars2 == nil {
		return vars1
	}

	result := make(map[string]interface{})

	// Copy from first map
	for k, v := range vars1 {
		result[k] = v
	}

	// Override with second map
	for k, v := range vars2 {
		result[k] = v
	}

	return result
}

// mergeContext merges two context maps, with the second map taking precedence
func mergeContext(ctx1, ctx2 map[string]interface{}) map[string]interface{} {
	if ctx1 == nil {
		ctx1 = make(map[string]interface{})
	}
	if ctx2 == nil {
		return ctx1
	}

	result := make(map[string]interface{})

	// Copy from first map
	for k, v := range ctx1 {
		result[k] = v
	}

	// Override with second map
	for k, v := range ctx2 {
		result[k] = v
	}

	return result
}

// ConditionalChain executes steps based on conditions
type ConditionalChain struct {
	*BaseChain
	conditions []ChainCondition
}

// ChainCondition represents a condition for executing a step
type ChainCondition struct {
	Step      ChainStep
	Condition func(ctx context.Context, input *ChainInput) (bool, error)
}

// NewConditionalChain creates a new conditional chain
func NewConditionalChain(name, description string) *ConditionalChain {
	return &ConditionalChain{
		BaseChain:  NewBaseChain(name, description),
		conditions: make([]ChainCondition, 0),
	}
}

// AddConditionalStep adds a step with a condition
func (c *ConditionalChain) AddConditionalStep(step ChainStep, condition func(ctx context.Context, input *ChainInput) (bool, error)) *ConditionalChain {
	c.conditions = append(c.conditions, ChainCondition{
		Step:      step,
		Condition: condition,
	})
	c.steps = append(c.steps, step)
	return c
}

// Execute executes steps based on their conditions
func (c *ConditionalChain) Execute(ctx context.Context, input *ChainInput) (*ChainOutput, error) {
	ctx, span := c.tracer.Start(ctx, "conditional_chain.execute")
	defer span.End()

	span.SetAttributes(
		attribute.String("chain.name", c.name),
		attribute.String("chain.type", "conditional"),
		attribute.Int("chain.conditions", len(c.conditions)),
	)

	if len(c.conditions) == 0 {
		return &ChainOutput{
			Result:  "No conditional steps to execute",
			Context: input.Context,
		}, nil
	}

	currentInput := input
	var finalOutput *ChainOutput
	executedSteps := 0

	// Check each condition and execute matching steps
	for i, condition := range c.conditions {
		conditionSpan := fmt.Sprintf("condition_%d_%s", i, condition.Step.GetName())
		conditionCtx, conditionSpanTracer := c.tracer.Start(ctx, conditionSpan)

		conditionSpanTracer.SetAttributes(
			attribute.String("step.name", condition.Step.GetName()),
			attribute.Int("condition.index", i),
		)

		// Evaluate condition
		shouldExecute, err := condition.Condition(conditionCtx, currentInput)
		if err != nil {
			conditionSpanTracer.RecordError(err)
			conditionSpanTracer.End()
			span.RecordError(err)
			return nil, fmt.Errorf("condition %d evaluation failed: %w", i, err)
		}

		conditionSpanTracer.SetAttributes(attribute.Bool("condition.result", shouldExecute))

		if shouldExecute {
			// Execute the step
			output, err := condition.Step.Execute(conditionCtx, currentInput)
			if err != nil {
				conditionSpanTracer.RecordError(err)
				conditionSpanTracer.End()
				span.RecordError(err)
				return nil, fmt.Errorf("conditional step %d (%s) failed: %w", i, condition.Step.GetName(), err)
			}

			executedSteps++

			// Store in memory if available
			if c.memory != nil {
				memoryKey := fmt.Sprintf("%s_conditional_step_%d", c.name, i)
				if err := c.memory.Store(ctx, memoryKey, output); err != nil {
					// Log error but don't fail the chain
					span.AddEvent("memory_store_failed", trace.WithAttributes(attribute.String("error", err.Error())))
				}
			}

			// Update input for next potential step
			currentInput = &ChainInput{
				Variables: mergeVariables(currentInput.Variables, map[string]interface{}{
					"previous_result": output.Result,
					"step_output":     output,
				}),
				Messages: output.Messages,
				Context:  mergeContext(currentInput.Context, output.Context),
			}

			finalOutput = output
		}

		conditionSpanTracer.End()
	}

	// If no steps were executed, return a default output
	if finalOutput == nil {
		finalOutput = &ChainOutput{
			Result:  "No conditions were met, no steps executed",
			Context: input.Context,
			Metadata: map[string]interface{}{
				"chain_name":         c.name,
				"chain_type":         "conditional",
				"steps_executed":     0,
				"conditions_checked": len(c.conditions),
			},
		}
	} else {
		// Enhance final output with chain metadata
		if finalOutput.Metadata == nil {
			finalOutput.Metadata = make(map[string]interface{})
		}
		finalOutput.Metadata["chain_name"] = c.name
		finalOutput.Metadata["chain_type"] = "conditional"
		finalOutput.Metadata["steps_executed"] = executedSteps
		finalOutput.Metadata["conditions_checked"] = len(c.conditions)
	}

	span.SetAttributes(attribute.Int("chain.steps_executed", executedSteps))

	return finalOutput, nil
}
