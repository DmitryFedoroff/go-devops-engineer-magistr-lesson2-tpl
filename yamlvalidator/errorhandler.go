package yamlvalidator

import "fmt"

type ValidationError struct {
	FilePath string
	Line     int
	Message  string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s:%d %s", e.FilePath, e.Line, e.Message)
}

type ErrorCollector struct {
	Errors []error
}

func (ec *ErrorCollector) Add(err error) {
	ec.Errors = append(ec.Errors, err)
}

func (ec *ErrorCollector) HasErrors() bool {
	return len(ec.Errors) > 0
}
