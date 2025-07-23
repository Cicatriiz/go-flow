package execution

import (
	"context"
	"fmt"
	"sync"

	"github.com/forrest/go-flow/core"
	"github.com/prometheus/client_golang/prometheus"
)

func init() {
	core.SetDefaultEngineCreator(func() core.ExecutionEngine {
		return NewConcurrentEngine()
	})
}

// DefaultEngine is the default implementation of the ExecutionEngine.
// It executes components sequentially based on their dependencies.
type DefaultEngine struct{}

// NewDefaultEngine creates a new DefaultEngine.
func NewDefaultEngine() *DefaultEngine {
	return &DefaultEngine{}
}

// Run executes the pipeline sequentially.
func (e *DefaultEngine) Run(ctx context.Context, p *core.Pipeline, inputs, outputs map[string]chan interface{}) error {
	graph := NewGraph(p)
	sorted, err := graph.TopologicalSort()
	if err != nil {
		return fmt.Errorf("error sorting pipeline graph: %w", err)
	}

	fmt.Println("Running pipeline sequentially:")
	components := p.GetComponents()
	connections := p.GetConnections()
	data := make(map[string]interface{})

	for _, name := range sorted {
		component := components[name]
		compInputs := make(map[string]interface{})

		for _, port := range component.InputPorts() {
			// Check for external inputs
			if ch, ok := inputs[port.Name()]; ok {
				compInputs[port.Name()] = <-ch
				continue
			}
			// Check for internal connections
			for _, conn := range connections {
				if conn.ToComponent == name && conn.ToPort == port.Name() {
					dataKey := fmt.Sprintf("%s.%s", conn.FromComponent, conn.FromPort)
					compInputs[port.Name()] = data[dataKey]
				}
			}
		}

		fmt.Printf("Executing component: %s\n", component.Name())
		compOutputs, err := component.Process(ctx, compInputs)
		if err != nil {
			return fmt.Errorf("error executing component %s: %w", component.Name(), err)
		}

		for portName, outData := range compOutputs {
			dataKey := fmt.Sprintf("%s.%s", name, portName)
			data[dataKey] = outData

			// Check for external outputs
			if ch, ok := outputs[portName]; ok {
				ch <- outData
			}
		}
	}

	fmt.Println("Pipeline execution complete.")
	return nil
}

// Close gracefully shuts down the engine.
func (e *DefaultEngine) Close() error {
	return nil
}

// ConcurrentEngine executes the pipeline with concurrency.
type ConcurrentEngine struct{}

// NewConcurrentEngine creates a new ConcurrentEngine.
func NewConcurrentEngine() *ConcurrentEngine {
	return &ConcurrentEngine{}
}

// Run executes the pipeline with concurrency.
func (e *ConcurrentEngine) Run(ctx context.Context, p *core.Pipeline, inputs, outputs map[string]chan interface{}) error {
	fmt.Println("Running pipeline concurrently:")

	var wg sync.WaitGroup
	components := p.GetComponents()
	connections := p.GetConnections()
	channels := make(map[string]chan interface{})
	errCh := make(chan error, len(components))

	// Create channels for all internal connections
	for _, conn := range connections {
		channels[fmt.Sprintf("%s.%s", conn.FromComponent, conn.FromPort)] = make(chan interface{})
	}

	// Start each component in a goroutine
	for name, component := range components {
		wg.Add(1)
		go func(name string, component core.Component) {
			defer wg.Done()

			compInputs := make(map[string]interface{})
			for _, port := range component.InputPorts() {
				var data interface{}
				// Check if this is an external input
				if ch, ok := inputs[port.Name()]; ok {
					data = <-ch
				} else {
					// Check if this is an internal connection
					for _, conn := range connections {
						if conn.ToComponent == name && conn.ToPort == port.Name() {
							chName := fmt.Sprintf("%s.%s", conn.FromComponent, conn.FromPort)
							var ok bool
							data, ok = <-channels[chName]
							if !ok {
								errCh <- fmt.Errorf("channel closed for %s.%s", name, port.Name())
								return
							}
						}
					}
				}
				compInputs[port.Name()] = data
			}

			timer := prometheus.NewTimer(core.ComponentLatency.WithLabelValues(name))
			compOutputs, err := component.Process(ctx, compInputs)
			timer.ObserveDuration()
			if err != nil {
				core.ComponentErrors.WithLabelValues(name).Inc()
				errCh <- fmt.Errorf("error executing component %s: %w", name, err)
				return
			}

			for portName, data := range compOutputs {
				// Check if this is an external output
				if ch, ok := outputs[portName]; ok {
					ch <- data
					continue
				}
				// Check if this is an internal connection
				chName := fmt.Sprintf("%s.%s", name, portName)
				if ch, ok := channels[chName]; ok {
					ch <- data
				}
			}

		}(name, component)
	}

	go func() {
		wg.Wait()
		close(errCh)
		for _, ch := range channels {
			close(ch)
		}
	}()

	// Collect errors
	for err := range errCh {
		if err != nil {
			// For now, we just return the first error.
			// A more advanced implementation could handle multiple errors.
			return err
		}
	}

	fmt.Println("Pipeline execution complete.")
	return nil
}

// Close gracefully shuts down the engine.
func (e *ConcurrentEngine) Close() error {
	return nil
}
