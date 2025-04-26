package component

import (
	"fmt"
	"sync"
	"time"

	"github.com/fatih/color"
)

var (
	ErrorCache = make(map[string]*CustomErrorImpl)
	errLock    sync.Mutex
	errKey     = "recent_errors" // Define errKey
)

type CustomError interface {
	Error() string
	GetCaller() string
	GetTimestamp() time.Time
	Unwrap() error // Added Unwrap for Go 1.13+ error chaining
}

type CustomErrorImpl struct {
	Message   string
	Caller    string
	Timestamp time.Time
	Err       error
	lock      sync.Mutex // Use 'lock' as defined in the second NewCustomError
}

func (c *CustomErrorImpl) Error() string {
	if c.Err != nil {
		return fmt.Sprintf("%s: %v", c.Message, c.Err)
	}
	return c.Message
}

func (c *CustomErrorImpl) GetCaller() string {
	return c.Caller
}

func (c *CustomErrorImpl) GetTimestamp() time.Time {
	return c.Timestamp
}

func (c *CustomErrorImpl) Unwrap() error {
	return c.Err
}

func NewCustomError(msg, caller string, err error) *CustomErrorImpl {
	return &CustomErrorImpl{
		Message:   msg,
		Caller:    caller,
		Timestamp: time.Now(),
		Err:       err,
		lock:      sync.Mutex{},
	}
}

type ErrorHandler interface {
	Capture(err error)
	Report()
}

type LogFormatter interface {
	Colorize(string) string
}

type ErrorHandlerImpl struct {
	LogFormatter
	cache map[string]interface{}
	mu    sync.Mutex
}

func NewErrorHandler() *ErrorHandlerImpl {
	return &ErrorHandlerImpl{
		cache: make(map[string]interface{}),
		mu:    sync.Mutex{},
	}
}

func (h *ErrorHandlerImpl) Capture(err error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if err != nil {
		// Simplified logging in Capture as caller info is part of CustomError
		color.Red("\n❌ ERROR OCCURRED\n")
		color.Yellow("Timestamp: %v\nMessage: %v", time.Now(), err)
		// You might want to store the error in the cache here if needed for Report
		// h.cache[errKey] = append(h.cache[errKey].([]error), err) // Example caching
	}
}

func (h *ErrorHandlerImpl) Report() {
	h.mu.Lock()
	defer h.mu.Unlock()

	// This Report function seems intended to report cached errors.
	// The current implementation is basic. You might want to iterate
	// through cached errors and format them nicely.
	if h.cache[errKey] != nil {
		color.Yellow("\nRECENT ERRORS:\n%v", h.cache[errKey]) // Use %v for potentially complex types
	}
	// Clear the cache after reporting
	h.cache[errKey] = nil
}

func LogAndCache(err *CustomErrorImpl) {
	err.lock.Lock() // Use 'lock' field
	defer err.lock.Unlock()

	ErrorCache[err.Message] = err

	color.Red("\n❌ ERROR OCCURRED\n")
	color.Yellow("Timestamp: %v\nCaller: %s\nMessage: %v", err.Timestamp, err.Caller, err.Message)
	if err.Err != nil {
		color.Yellow("\nORIGINAL ERROR:\n%s", err.Err)
	}
}

func HandleError(f func() error) error {
	handler := NewErrorHandler()

	defer func() {
		r := recover()
		if r != nil {
			// Capture panic as an error
			handler.Capture(fmt.Errorf("panic: %v", r))
			// Re-panic after capturing
			panic(r)
		}
	}()

	resErr := f()

	// This logging seems out of place in a generic error handler
	// color.Yellow("\n⏱️ %s completed in %.2fs\n", "ERROR HANDLING", elapse.Seconds())

	return resErr
}

func RecoverError() {
	handler := NewErrorHandler()
	recovered := recover()
	if recovered != nil {
		// Capture recovered panic as an error
		handler.Capture(fmt.Errorf("panic: %v", recovered))
		// Report captured errors (if Capture was modified to cache)
		handler.Report()
		// Optionally re-panic if the panic should not be fully recovered
		// panic(recovered)
	} else {
		// If no panic occurred, still report any cached errors
		handler.Report()
	}
}
