package urlog

import (
	"context"
	"github.com/apex/log"
	"github.com/pkg/errors"
)

// WrapError returns a wrapped copy of err if it doesn't have a stacktrace, lacks some fields available in ctx
// or needs to be prefixed with prefixMessage. If neither of those is true, original err is returned.
func WrapError(ctx context.Context, err error, prefixMessage string) error {
	mergedFields, newFields := mergeContextFields(ctx, err)
	stacktracer, isStacktracer := err.(stackTracer)

	if isStacktracer && !newFields && prefixMessage == "" {
		return err
	}

	wrapper := &errorWrapper{
		cause:         err,
		prefixMessage: prefixMessage,
		fields:        mergedFields,
	}
	// save err's stacktrace or create a new one omitting the current frame
	if isStacktracer {
		wrapper.stacktrace = stacktracer.StackTrace()
	} else {
		wrapper.stacktrace = errors.WithStack(err).(stackTracer).StackTrace()[1:]
	}

	return wrapper
}

func mergeContextFields(ctx context.Context, err error) (result log.Fields, newFields bool) {
	var currentFields log.Fields
	if fielder, isFielder := err.(log.Fielder); isFielder {
		currentFields = fielder.Fields()
	} else {
		newFields = true
	}

	if ctx == nil {
		return currentFields, newFields
	}

	ctxFields := loadFields(ctx)
	for key, _ := range ctxFields {
		if _, ok := currentFields[key]; !ok {
			newFields = true
			break
		}
	}

	if !newFields {
		return currentFields, newFields
	}

	result = make(log.Fields)
	for key, value := range ctxFields {
		result[key] = value
	}
	for key, value := range currentFields {
		result[key] = value
	}

	return
}
