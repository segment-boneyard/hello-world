package urlog

import (
	"context"
	"github.com/apex/log"
)

// unique key holding logger fields.
type key struct{ string }

// holdes log.Fields in context.
var logFieldsKey = &key{"log fields"}

// GetContextualLogger returns updated versions of ctx and logger. If fielder is not nil, fields from it are merged
// into fields already stored in ctx. If logger is nil, a default value of log.Log is assumed. Updated logger contains
// its old fields merged with fields from fielder.
func GetContextualLogger(ctx context.Context, logger log.Interface, fielder log.Fielder) (context.Context, log.Interface) {
	ctxFields := loadFields(ctx)

	if fielder != nil && len(fielder.Fields()) > 0 {
		mergedFields := make(log.Fields)
		for key, value := range ctxFields {
			mergedFields[key] = value
		}
		for key, value := range fielder.Fields() {
			mergedFields[key] = value
		}

		ctxFields = mergedFields
		ctx = context.WithValue(ctx, logFieldsKey, ctxFields)
	}

	if logger == nil {
		logger = log.Log
	}

	if len(ctxFields) > 0 {
		logger = logger.WithFields(ctxFields)
	}

	return ctx, logger
}

func loadFields(ctx context.Context) log.Fields {
	interfaceValue := ctx.Value(logFieldsKey)
	if interfaceValue == nil {
		return nil
	}

	if fields, ok := interfaceValue.(log.Fields); ok {
		return fields
	}

	return nil
}
