package downloader

import (
	"context"
	"github.com/segment-sources/stripe/api"
	"github.com/segment-sources/stripe/integration"
)

// PostProcessor is a function called by the downloader on each object retrieved from the API
type PostProcessor func(ctx context.Context, obj api.Object, task *Task) error

// Task describes which objects
type Task struct {
	// Request is an API request that should be performed to retrieve the first page of objects
	Request *api.Request
	// Output will receive the downloaded objects
	Output chan api.Object
	// Errors will receive an error if an error happens
	Errors chan integration.CollectionError
	// PostProcessors is a list of functions that should be called on each retrieved object
	PostProcessors []PostProcessor
	// Collection is a name that will be used when reporting collection errors
	Collection string
}
