package core

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// BasePipelineError provides a default implementation of PipelineError
type BasePipelineError struct {
	message       string
	component     string
	errorType     ErrorType
	severity      Severity
	recoverable   bool
	context       map[string]interface{}
	originalError error
}

// NewPipelineError creates a new pipeline error
func NewPipelineError(message, component string, errorType ErrorType, severity Severity, recoverable bool) *BasePipelineError {
	return &BasePipelineError{
		message:     message,
		component:   component,
		errorType:   errorType,
		severity:    severity,
		recoverable: recoverable,
		context:     make(map[string]interface{}),
	}
}

// Error returns the error message
func (e *BasePipelineError) Error() string {
	return fmt.Sprintf("[%s] %s: %s", e.component, e.errorType.String(), e.message)
}

// Component returns the component where the error occurred
func (e *BasePipelineError) Component() string {
	return e.component
}

// ErrorType returns the type of error
func (e *BasePipelineError) ErrorType() ErrorType {
	return e.errorType
}

// Severity returns the severity level
func (e *BasePipelineError) Severity() Severity {
	return e.severity
}

// Recoverable indicates if the error can be recovered from
func (e *BasePipelineError) Recoverable() bool {
	return e.recoverable
}

// Context returns additional context information
func (e *BasePipelineError) Context() map[string]interface{} {
	return e.context
}

// WithContext adds context information to the error
func (e *BasePipelineError) WithContext(key string, value interface{}) *BasePipelineError {
	e.context[key] = value
	return e
}

// WithOriginalError sets the original error that caused this pipeline error
func (e *BasePipelineError) WithOriginalError(err error) *BasePipelineError {
	e.originalError = err
	e.context["original_error"] = err.Error()
	return e
}

// Unwrap returns the original error for error unwrapping
func (e *BasePipelineError) Unwrap() error {
	return e.originalError
}

// String methods for error types and severities
func (et ErrorType) String() string {
	switch et {
	case ValidationError:
		return "VALIDATION"
	case RuntimeError:
		return "RUNTIME"
	case ConfigurationError:
		return "CONFIGURATION"
	case ResourceError:
		return "RESOURCE"
	case NetworkError:
		return "NETWORK"
	default:
		return "UNKNOWN"
	}
}

func (s Severity) String() string {
	switch s {
	case Info:
		return "INFO"
	case Warning:
		return "WARNING"
	case Error:
		return "ERROR"
	case Critical:
		return "CRITICAL"
	default:
		return "UNKNOWN"
	}
}

func (ea ErrorAction) String() string {
	switch ea {
	case Continue:
		return "CONTINUE"
	case Retry:
		return "RETRY"
	case Skip:
		return "SKIP"
	case Abort:
		return "ABORT"
	default:
		return "UNKNOWN"
	}
}

// DefaultErrorHandler provides a default implementation of ErrorHandler
type DefaultErrorHandler struct {
	retryAttempts map[string]int
	maxRetries    int
	mutex         sync.RWMutex
}

// NewDefaultErrorHandler creates a new default error handler
func NewDefaultErrorHandler(maxRetries int) *DefaultErrorHandler {
	return &DefaultErrorHandler{
		retryAttempts: make(map[string]int),
		maxRetries:    maxRetries,
	}
}

// HandleError determines what action to take for a given error
func (h *DefaultErrorHandler) HandleError(ctx context.Context, err PipelineError) ErrorAction {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	key := fmt.Sprintf("%s:%s", err.Component(), err.ErrorType().String())

	switch err.Severity() {
	case Critical:
		return Abort
	case Error:
		if err.Recoverable() && h.retryAttempts[key] < h.maxRetries {
			h.retryAttempts[key]++
			return Retry
		}
		return Abort
	case Warning:
		if err.Recoverable() {
			return Continue
		}
		return Skip
	case Info:
		return Continue
	default:
		return Abort
	}
}

// CanRecover checks if an error can be recovered from
func (h *DefaultErrorHandler) CanRecover(err PipelineError) bool {
	return err.Recoverable() && err.Severity() != Critical
}

// ResetRetryCount resets the retry count for a specific component and error type
func (h *DefaultErrorHandler) ResetRetryCount(component string, errorType ErrorType) {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	
	key := fmt.Sprintf("%s:%s", component, errorType.String())
	delete(h.retryAttempts, key)
}

// BaseCircuitBreaker provides a default implementation of CircuitBreaker
type BaseCircuitBreaker struct {
	state           CircuitState
	failureCount    int
	successCount    int
	failureThreshold int
	successThreshold int
	timeout         time.Duration
	lastFailureTime time.Time
	mutex           sync.RWMutex
}

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(failureThreshold, successThreshold int, timeout time.Duration) *BaseCircuitBreaker {
	return &BaseCircuitBreaker{
		state:            Closed,
		failureThreshold: failureThreshold,
		successThreshold: successThreshold,
		timeout:          timeout,
	}
}

// Execute executes a function with circuit breaker protection
func (cb *BaseCircuitBreaker) Execute(ctx context.Context, fn func() (interface{}, error)) (interface{}, error) {
	cb.mutex.Lock()
	state := cb.state
	cb.mutex.Unlock()

	switch state {
	case Open:
		if time.Since(cb.lastFailureTime) > cb.timeout {
			cb.mutex.Lock()
			cb.state = HalfOpen
			cb.successCount = 0
			cb.mutex.Unlock()
		} else {
			return nil, fmt.Errorf("circuit breaker is open")
		}
	case HalfOpen:
		// Allow execution but monitor closely
	case Closed:
		// Normal execution
	}

	result, err := fn()

	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	if err != nil {
		cb.onFailure()
		return nil, err
	}

	cb.onSuccess()
	return result, nil
}

// State returns the current state of the circuit breaker
func (cb *BaseCircuitBreaker) State() CircuitState {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()
	return cb.state
}

// Reset resets the circuit breaker to closed state
func (cb *BaseCircuitBreaker) Reset() {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()
	
	cb.state = Closed
	cb.failureCount = 0
	cb.successCount = 0
}

// onFailure handles failure cases
func (cb *BaseCircuitBreaker) onFailure() {
	cb.failureCount++
	cb.lastFailureTime = time.Now()

	if cb.state == HalfOpen || cb.failureCount >= cb.failureThreshold {
		cb.state = Open
		cb.successCount = 0
	}
}

// onSuccess handles success cases
func (cb *BaseCircuitBreaker) onSuccess() {
	cb.failureCount = 0

	if cb.state == HalfOpen {
		cb.successCount++
		if cb.successCount >= cb.successThreshold {
			cb.state = Closed
		}
	}
}

func (cs CircuitState) String() string {
	switch cs {
	case Closed:
		return "CLOSED"
	case Open:
		return "OPEN"
	case HalfOpen:
		return "HALF_OPEN"
	default:
		return "UNKNOWN"
	}
}

// ErrorCollector collects and aggregates errors for analysis
type ErrorCollector struct {
	errors []PipelineError
	mutex  sync.RWMutex
}

// NewErrorCollector creates a new error collector
func NewErrorCollector() *ErrorCollector {
	return &ErrorCollector{
		errors: make([]PipelineError, 0),
	}
}

// Collect adds an error to the collection
func (ec *ErrorCollector) Collect(err PipelineError) {
	ec.mutex.Lock()
	defer ec.mutex.Unlock()
	ec.errors = append(ec.errors, err)
}

// GetErrors returns all collected errors
func (ec *ErrorCollector) GetErrors() []PipelineError {
	ec.mutex.RLock()
	defer ec.mutex.RUnlock()
	
	result := make([]PipelineError, len(ec.errors))
	copy(result, ec.errors)
	return result
}

// GetErrorsByComponent returns errors for a specific component
func (ec *ErrorCollector) GetErrorsByComponent(component string) []PipelineError {
	ec.mutex.RLock()
	defer ec.mutex.RUnlock()
	
	var result []PipelineError
	for _, err := range ec.errors {
		if err.Component() == component {
			result = append(result, err)
		}
	}
	return result
}

// GetErrorsBySeverity returns errors of a specific severity
func (ec *ErrorCollector) GetErrorsBySeverity(severity Severity) []PipelineError {
	ec.mutex.RLock()
	defer ec.mutex.RUnlock()
	
	var result []PipelineError
	for _, err := range ec.errors {
		if err.Severity() == severity {
			result = append(result, err)
		}
	}
	return result
}

// Clear removes all collected errors
func (ec *ErrorCollector) Clear() {
	ec.mutex.Lock()
	defer ec.mutex.Unlock()
	ec.errors = ec.errors[:0]
}

// Count returns the total number of collected errors
func (ec *ErrorCollector) Count() int {
	ec.mutex.RLock()
	defer ec.mutex.RUnlock()
	return len(ec.errors)
}