package errors

import (
	"errors"
	"strings"
	"sync"
	"testing"
)

func TestNewErrorList(t *testing.T) {
	el := NewErrorList()
	if el == nil {
		t.Fatal("NewErrorList returned nil")
	}
	if len(el.errors) != 0 {
		t.Errorf("expected empty error list, got %d errors", len(el.errors))
	}
}

func TestErrorList_Error(t *testing.T) {
	tests := map[string]struct {
		errors   []error
		expected string
	}{
		"no errors": {
			errors:   []error{},
			expected: "",
		},
		"single error": {
			errors:   []error{errors.New("first error")},
			expected: "first error",
		},
		"multiple errors": {
			errors:   []error{errors.New("error one"), errors.New("error two")},
			expected: "2 errors:\n\terror one\n\terror two\n",
		},
		"three errors": {
			errors:   []error{errors.New("a"), errors.New("b"), errors.New("c")},
			expected: "3 errors:\n\ta\n\tb\n\tc\n",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			el := NewErrorList()
			for _, err := range tc.errors {
				el.Add(err)
			}

			got := el.Error()
			if got != tc.expected {
				t.Errorf("expected %q, got %q", tc.expected, got)
			}
		})
	}
}

func TestErrorList_Err(t *testing.T) {
	err1 := errors.New("first error")
	err2 := errors.New("second error")

	tests := map[string]struct {
		errors      []error
		expectNil   bool
		expectSame  error
		expectMulti bool
	}{
		"no errors returns nil": {
			errors:    []error{},
			expectNil: true,
		},
		"single error returns that error": {
			errors:     []error{err1},
			expectSame: err1,
		},
		"multiple errors returns errorList": {
			errors:      []error{err1, err2},
			expectMulti: true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			el := NewErrorList()
			for _, err := range tc.errors {
				el.Add(err)
			}

			got := el.Err()

			if tc.expectNil {
				if got != nil {
					t.Errorf("expected nil, got %v", got)
				}
				return
			}

			if tc.expectSame != nil {
				if got != tc.expectSame {
					t.Errorf("expected same error instance, got different")
				}
				return
			}

			if tc.expectMulti {
				if got != el {
					t.Errorf("expected errorList itself, got %T", got)
				}
			}
		})
	}
}

func TestErrorList_Add(t *testing.T) {
	tests := map[string]struct {
		toAdd         []error
		expectedCount int
	}{
		"add nil does nothing": {
			toAdd:         []error{nil},
			expectedCount: 0,
		},
		"add single error": {
			toAdd:         []error{errors.New("error")},
			expectedCount: 1,
		},
		"add multiple errors": {
			toAdd:         []error{errors.New("one"), errors.New("two"), errors.New("three")},
			expectedCount: 3,
		},
		"add mix of nil and errors": {
			toAdd:         []error{errors.New("one"), nil, errors.New("two"), nil},
			expectedCount: 2,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			el := NewErrorList()
			for _, err := range tc.toAdd {
				el.Add(err)
			}

			if len(el.errors) != tc.expectedCount {
				t.Errorf("expected %d errors, got %d", tc.expectedCount, len(el.errors))
			}
		})
	}
}

func TestErrorList_Concurrency(t *testing.T) {
	el := NewErrorList()
	var wg sync.WaitGroup
	numGoroutines := 100

	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func(n int) {
			defer wg.Done()
			el.Add(errors.New("concurrent error"))
		}(i)
	}
	wg.Wait()

	if len(el.errors) != numGoroutines {
		t.Errorf("expected %d errors, got %d", numGoroutines, len(el.errors))
	}

	// Test concurrent reads
	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			_ = el.Error()
			_ = el.Err()
		}()
	}
	wg.Wait()

	// Verify the error string contains expected content
	errStr := el.Error()
	if !strings.Contains(errStr, "100 errors:") {
		t.Errorf("expected error string to mention 100 errors, got %q", errStr)
	}
}
