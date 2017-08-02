package api

import (
	"github.com/apex/log"
	"github.com/pkg/errors"
)

type wrappedError interface {
	error
	log.Fielder
	StackTrace() errors.StackTrace
	Cause() error
}

type causer interface {
	Cause() error
}

type errorInterface interface {
	IsPermanent() bool
	IsAuthRelated() bool
}

// permanentError wraps a urlog-compatible error so that it would still implement urlog interface
// and could be used with IsErrorPermanent and IsErrorAuthRelated methods
type permanentError struct {
	wrappedError
	isAuthRelated bool
}

func (e *permanentError) IsPermanent() bool {
	return true
}

func (e *permanentError) IsAuthRelated() bool {
	return e.isAuthRelated
}

func getErrorInterface(err error) errorInterface {
	for {
		if i, ok := err.(errorInterface); ok {
			return i
		}

		if causer, ok := err.(causer); ok {
			err = causer.Cause()
		} else {
			return nil
		}
	}
}

// IsErrorPermanent returns true if any error in a wrapper chain is a permanent error
func IsErrorPermanent(err error) bool {
	if i := getErrorInterface(err); i != nil {
		return i.IsPermanent()
	}

	return false
}

// IsErrorAuthRelated returns true if any error in a wrapper chain is a permanent error
// and has isAuthRelated flag set
func IsErrorAuthRelated(err error) bool {
	if i := getErrorInterface(err); i != nil {
		return i.IsAuthRelated()
	}

	return false
}
