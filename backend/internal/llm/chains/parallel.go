package chains

import (
	"context"
	"fmt"
	"sync"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// ParallelChain executes steps in parallel
type ParallelChain struct {
	*BaseChain
	maxConcurrency int
}

// NewParallelChain creates a new parallel chain
func NewParallelChain(name, description string, maxConcurrency int) *ParallelChain {
	if maxConcurrency <= 0 {
		maxConcurrency = 10 // Default max concurrency
	}
	
	return &ParallelChain{
		BaseChain:      NewBaseChain(name, description),
		maxConcurrency: maxConcurrency,
	}
}

// ParallelResult represents the result of a parallel step execution
type ParallelResult struct {
	StepIndex int
	StepName  string
	Output    *ChainOutput
	Error     error
}

// Execute executes all steps in parallel
func (c *ParallelChain) Execute(ctx context.Context, input *ChainInput) (*ChainOutput, error) {
	ctx, span := c.tracer.Start(ctx, "parallel_chain.execute")
	defer span.End()
	
	span.SetAttributes(
		attribute.String("chain.name", c.name),
		attribute.String("chain.type", "parallel"),
		attribute.Int("chain.steps", len(c.steps)),
		attribute.Int("chain.max_concurrency", c.maxConcurrency),
	)
	
	if len(c.steps) == 0 {
		return &ChainOutput{
			Result:  "No steps to execute",
			Context: input.Context,
		}, nil
	}
	
	// Create a semaphore to limit concurrency
	semaphore := make(chan struct{}, c.maxConcurrency)
	results := make(chan ParallelResult, len(c.steps))
	var wg sync.WaitGroup
	
	// Execute all steps in parallel
	for i, step := range c.steps {
		wg.Add(1)
		go func(index int, s ChainStep) {
			defer wg.Done()
			
			// Acquire semaphore
			semaphore <- struct{}{}
			defer func() { <-semaphore }()
			
			stepSpan := fmt.Sprintf("parallel_step_%d_%s", index, s.GetName())
			stepCtx, stepSpanTracer := c.tracer.Start(ctx, stepSpan)
			defer stepSpanTracer.End()
			
			stepSpanTracer.SetAttributes(
				attribute.String("step.name", s.GetName()),
				attribute.String("step.description", s.GetDescription()),
				attribute.Int("step.index", index),
			)
			
			// Execute the step
			output, err := s.Execute(stepCtx, input)
			if err != nil {
				stepSpanTracer.RecordError(err)
			}
			
			// Send result
			results <- ParallelResult{
				StepIndex: index,
				StepName:  s.GetName(),
				Output:    output,
				Error:     err,
			}
		}(i, step)
	}
	
	// Wait for all goroutines to complete
	go func() {
		wg.Wait()
		close(results)
	}()
	
	// Collect results
	stepResults := make(map[int]ParallelResult)
	var errors []error
	
	for result := range results {
		stepResults[result.StepIndex] = result
		if result.Error != nil {
			errors = append(errors, fmt.Errorf("step %d (%s) failed: %w", result.StepIndex, result.StepName, result.Error))
		}
		
		// Store successful results in memory if available
		if c.memory != nil && result.Error == nil {
			memoryKey := fmt.Sprintf("%s_parallel_step_%d", c.name, result.StepIndex)
			if err := c.memory.Store(ctx, memoryKey, result.Output); err != nil {
				// Log error but don't fail the chain
				span.AddEvent("memory_store_failed", trace.WithAttributes(attribute.String("error", err.Error())))
			}
		}
	}
	
	// Check if any steps failed
	if len(errors) > 0 {
		// Return the first error (could be enhanced to return all errors)
		span.RecordError(errors[0])
		return nil, errors[0]
	}
	
	// Combine results
	combinedOutput := c.combineResults(stepResults, input.Context)
	
	// Enhance output with chain metadata
	if combinedOutput.Metadata == nil {
		combinedOutput.Metadata = make(map[string]interface{})
	}
	combinedOutput.Metadata["chain_name"] = c.name
	combinedOutput.Metadata["chain_type"] = "parallel"
	combinedOutput.Metadata["steps_executed"] = len(c.steps)
	combinedOutput.Metadata["max_concurrency"] = c.maxConcurrency
	
	return combinedOutput, nil
}

// combineResults combines the results from parallel step execution
func (c *ParallelChain) combineResults(stepResults map[int]ParallelResult, originalContext map[string]interface{}) *ChainOutput {
	// Collect all results in order
	results := make([]interface{}, len(stepResults))
	var allMessages []Message
	combinedContext := make(map[string]interface{})
	combinedMetadata := make(map[string]interface{})
	
	// Copy original context
	for k, v := range originalContext {
		combinedContext[k] = v
	}
	
	// Process results in step order
	for i := 0; i < len(stepResults); i++ {
		if result, exists := stepResults[i]; exists && result.Output != nil {
			results[i] = result.Output.Result
			
			// Combine messages
			if result.Output.Messages != nil {
				allMessages = append(allMessages, result.Output.Messages...)
			}
			
			// Combine context
			if result.Output.Context != nil {
				for k, v := range result.Output.Context {
					combinedContext[k] = v
				}
			}
			
			// Combine metadata with step prefix
			if result.Output.Metadata != nil {
				stepKey := fmt.Sprintf("step_%d_%s", i, result.StepName)
				combinedMetadata[stepKey] = result.Output.Metadata
			}
		}
	}
	
	return &ChainOutput{
		Result:   results,
		Messages: allMessages,
		Context:  combinedContext,
		Metadata: combinedMetadata,
	}
}

// MapReduceChain executes a map phase in parallel followed by a reduce phase
type MapReduceChain struct {
	*BaseChain
	mapStep        ChainStep
	reduceStep     ChainStep
	maxConcurrency int
	inputSplitter  func(*ChainInput) ([]*ChainInput, error)
}

// NewMapReduceChain creates a new map-reduce chain
func NewMapReduceChain(name, description string, maxConcurrency int) *MapReduceChain {
	if maxConcurrency <= 0 {
		maxConcurrency = 10
	}
	
	return &MapReduceChain{
		BaseChain:      NewBaseChain(name, description),
		maxConcurrency: maxConcurrency,
	}
}

// SetMapStep sets the map step
func (c *MapReduceChain) SetMapStep(step ChainStep) *MapReduceChain {
	c.mapStep = step
	return c
}

// SetReduceStep sets the reduce step
func (c *MapReduceChain) SetReduceStep(step ChainStep) *MapReduceChain {
	c.reduceStep = step
	return c
}

// SetInputSplitter sets the function to split input for map phase
func (c *MapReduceChain) SetInputSplitter(splitter func(*ChainInput) ([]*ChainInput, error)) *MapReduceChain {
	c.inputSplitter = splitter
	return c
}

// Execute executes the map-reduce chain
func (c *MapReduceChain) Execute(ctx context.Context, input *ChainInput) (*ChainOutput, error) {
	ctx, span := c.tracer.Start(ctx, "mapreduce_chain.execute")
	defer span.End()
	
	span.SetAttributes(
		attribute.String("chain.name", c.name),
		attribute.String("chain.type", "mapreduce"),
		attribute.Int("chain.max_concurrency", c.maxConcurrency),
	)
	
	if c.mapStep == nil {
		return nil, fmt.Errorf("map step is required for map-reduce chain")
	}
	
	if c.reduceStep == nil {
		return nil, fmt.Errorf("reduce step is required for map-reduce chain")
	}
	
	if c.inputSplitter == nil {
		return nil, fmt.Errorf("input splitter is required for map-reduce chain")
	}
	
	// Split input for map phase
	mapInputs, err := c.inputSplitter(input)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to split input for map phase: %w", err)
	}
	
	span.SetAttributes(attribute.Int("map.inputs", len(mapInputs)))
	
	// Execute map phase in parallel
	mapResults, err := c.executeMapPhase(ctx, mapInputs)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("map phase failed: %w", err)
	}
	
	// Prepare input for reduce phase
	reduceInput := &ChainInput{
		Variables: map[string]interface{}{
			"map_results": mapResults,
		},
		Context: input.Context,
	}
	
	// Execute reduce phase
	reduceOutput, err := c.reduceStep.Execute(ctx, reduceInput)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("reduce phase failed: %w", err)
	}
	
	// Enhance output with chain metadata
	if reduceOutput.Metadata == nil {
		reduceOutput.Metadata = make(map[string]interface{})
	}
	reduceOutput.Metadata["chain_name"] = c.name
	reduceOutput.Metadata["chain_type"] = "mapreduce"
	reduceOutput.Metadata["map_inputs"] = len(mapInputs)
	reduceOutput.Metadata["map_results"] = len(mapResults)
	
	return reduceOutput, nil
}

// executeMapPhase executes the map phase in parallel
func (c *MapReduceChain) executeMapPhase(ctx context.Context, inputs []*ChainInput) ([]*ChainOutput, error) {
	ctx, span := c.tracer.Start(ctx, "mapreduce_chain.map_phase")
	defer span.End()
	
	if len(inputs) == 0 {
		return []*ChainOutput{}, nil
	}
	
	// Create a semaphore to limit concurrency
	semaphore := make(chan struct{}, c.maxConcurrency)
	results := make(chan ParallelResult, len(inputs))
	var wg sync.WaitGroup
	
	// Execute map step for each input in parallel
	for i, mapInput := range inputs {
		wg.Add(1)
		go func(index int, input *ChainInput) {
			defer wg.Done()
			
			// Acquire semaphore
			semaphore <- struct{}{}
			defer func() { <-semaphore }()
			
			stepSpan := fmt.Sprintf("map_step_%d", index)
			stepCtx, stepSpanTracer := c.tracer.Start(ctx, stepSpan)
			defer stepSpanTracer.End()
			
			stepSpanTracer.SetAttributes(
				attribute.String("step.name", c.mapStep.GetName()),
				attribute.Int("map.index", index),
			)
			
			// Execute the map step
			output, err := c.mapStep.Execute(stepCtx, input)
			if err != nil {
				stepSpanTracer.RecordError(err)
			}
			
			// Send result
			results <- ParallelResult{
				StepIndex: index,
				StepName:  c.mapStep.GetName(),
				Output:    output,
				Error:     err,
			}
		}(i, mapInput)
	}
	
	// Wait for all goroutines to complete
	go func() {
		wg.Wait()
		close(results)
	}()
	
	// Collect results
	mapResults := make([]*ChainOutput, len(inputs))
	var errors []error
	
	for result := range results {
		if result.Error != nil {
			errors = append(errors, fmt.Errorf("map step %d failed: %w", result.StepIndex, result.Error))
		} else {
			mapResults[result.StepIndex] = result.Output
		}
	}
	
	// Check if any map steps failed
	if len(errors) > 0 {
		return nil, errors[0]
	}
	
	return mapResults, nil
}
