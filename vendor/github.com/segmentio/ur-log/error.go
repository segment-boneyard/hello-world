package urlog

import (
	"github.com/apex/log"
	"github.com/pkg/errors"
)

type stackTracer interface {
	StackTrace() errors.StackTrace
}

type errorWrapper struct {
	cause         error
	prefixMessage string
	stacktrace    errors.StackTrace
	fields        log.Fields
}

func (e *errorWrapper) Fields() log.Fields {
	return e.fields
}

func (e *errorWrapper) StackTrace() errors.StackTrace {
	return e.stacktrace
}

func (e *errorWrapper) Cause() error {
	return errors.Cause(e.cause)
}

func (e *errorWrapper) Error() string {
	if e.prefixMessage == "" {
		return e.cause.Error()
	}

	return e.prefixMessage + ": " + e.cause.Error()
}
