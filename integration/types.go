package integration

import (
	"context"
	"github.com/segment-sources/stripe/api"
	"github.com/segmentio/go-source"
	"time"
)

type Producer interface {
	StartProducer(ctx context.Context, runContext RunContext) error
	CollectionErrors() <-chan CollectionError
	Objects() <-chan api.Object
}

type Consumer interface {
	Collection() string
	StartConsumer(ctx context.Context, ch <-chan api.Object)
	DesiredEvents() []string
	DesiredObjects() []string

	Messages() <-chan source.SetMessage
}

type Resource interface {
	Producer
	Close()
	Consumers() []Consumer
}

type RunContext struct {
	PreviousRunTimestamp time.Time `json:"previous_run_timestamp"`
	Version              int       `json:"version"`
}

type subscription struct {
	ch       chan api.Object
	consumer Consumer
}

type CollectionError struct {
	Collection string
	Message    string
}
