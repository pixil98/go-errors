package errors

import (
	"fmt"
	"sync"
)

type errorList struct {
	errors []error
	mu     sync.RWMutex
}

func NewErrorList() *errorList {
	return &errorList{
		errors: []error{},
	}
}

func (el *errorList) Error() string {
	el.mu.RLock()
	defer el.mu.RUnlock()

	if len(el.errors) == 0 {
		return ""
	}

	if len(el.errors) == 1 {
		return el.errors[0].Error()
	}

	str := fmt.Sprintf("%d errors:\n", len(el.errors))
	for _, err := range el.errors {
		str += fmt.Sprintf("\t%s\n", err.Error())
	}

	return string(str)
}

func (el *errorList) Err() error {
	el.mu.RLock()
	defer el.mu.RUnlock()

	switch len(el.errors) {
	case 0:
		return nil
	case 1:
		return el.errors[0]
	default:
		return el
	}
}

func (el *errorList) Add(err error) {
	if err == nil {
		return
	}

	el.mu.Lock()
	defer el.mu.Unlock()

	el.errors = append(el.errors, err)
}
