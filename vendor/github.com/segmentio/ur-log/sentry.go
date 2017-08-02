package urlog

import (
	"fmt"
	"github.com/apex/log"
	"github.com/getsentry/raven-go"
	"github.com/pkg/errors"
	"reflect"
	"runtime"
)

// BuildErrorPacket builds a sentry packet based on err.
// Works best if err implements log.Fielder and stackTracer interfaces
func BuildErrorPacket(err error) *raven.Packet {
	// unwrap the cause error
	cause := errors.Cause(err)

	exception := &raven.Exception{
		Type:  cause.Error(),
		Value: reflect.TypeOf(cause).String(),
	}

	// add reversed stacktrace to the packet
	if stackTracerValue, ok := err.(stackTracer); ok {
		stackTrace := stackTracerValue.StackTrace()
		exception.Stacktrace = &raven.Stacktrace{}

		for i := len(stackTrace) - 1; i >= 0; i-- {
			pc := uintptr(stackTrace[i])
			filename, line := runtime.FuncForPC(pc - 1).FileLine(pc - 1)
			sentryFrame := raven.NewStacktraceFrame(pc, filename, line, 3, raven.IncludePaths())
			if sentryFrame == nil {
				continue
			}
			exception.Stacktrace.Frames = append(exception.Stacktrace.Frames, sentryFrame)
		}
	}

	packet := raven.NewPacket(cause.Error(), exception)

	// add tags from the fields to the packet
	if fielder, ok := err.(log.Fielder); ok {
		for key, value := range fielder.Fields() {
			packet.Extra[key] = fmt.Sprintf("%v", value)
		}
	}

	return packet
}
