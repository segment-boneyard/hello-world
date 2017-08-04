package bundle

import (
	"context"
	"github.com/apex/log"
	"github.com/pkg/errors"
	"github.com/segment-sources/stripe/api"
	"github.com/segment-sources/stripe/integration"
	"github.com/segment-sources/stripe/resource/downloader"
	"github.com/segment-sources/stripe/resource/tasks"
	"github.com/segmentio/ur-log"
	"sync"
	"sync/atomic"
)

// ResourceBundle is a group of resources that behaves like resource itself.
// It's used to request all member resources' events in a single ordered stream avoiding data races.
type ResourceBundle struct {
	apiClient      api.Client
	objs           chan api.Object
	errs           chan integration.CollectionError
	resources      []integration.Resource
	producerErrors int32
}

func (b *ResourceBundle) forwardProducer(ctx context.Context, res integration.Resource, runContext integration.RunContext) error {
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		for obj := range res.Objects() {
			b.objs <- obj
		}
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		for e := range res.CollectionErrors() {
			b.errs <- e
		}
	}()

	err := res.StartProducer(ctx, runContext)
	wg.Wait()
	return err
}

// eventsProducer downloads a joined stream of events desired by all member's consumers
func (b *ResourceBundle) eventsProducer(ctx context.Context, runContext integration.RunContext) error {
	consumers := b.Consumers()

	collectionNames := []string{}
	for _, con := range consumers {
		collectionNames = append(collectionNames, con.Collection())
	}
	ctx, _ = urlog.GetContextualLogger(ctx, nil, log.Fields{
		"bundle_collections": collectionNames,
	})

	colErrors := make(chan integration.CollectionError)
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		for e := range colErrors {
			// if downloading of the joined stream fails, a separate error will be created
			// for every collection in the bundle
			for _, name := range collectionNames {
				e.Collection = name
				b.errs <- e
			}
		}
	}()

	err := downloader.New(b.apiClient).Do(ctx, tasks.MakeIncremental(
		b,
		"events",
		runContext.PreviousRunTimestamp,
		b.objs,
		colErrors,
	))

	close(colErrors)
	wg.Wait()

	return err
}

// StartProducer starts every member resources' producer and a special event producer
// that downloads all member resources' events in a single stream
func (b *ResourceBundle) StartProducer(ctx context.Context, runContext integration.RunContext) error {
	defer close(b.objs)
	defer close(b.errs)
	wg := sync.WaitGroup{}

	// start forwarding producers
	for _, res := range b.resources {
		wg.Add(1)
		go func(res integration.Resource) {
			defer wg.Done()
			if err := b.forwardProducer(ctx, res, runContext); err != nil {
				log.WithError(err).Error("bundle forwarded producer failed")
				atomic.AddInt32(&b.producerErrors, 1)
			}
		}(res)
	}

	// if incremental mode enabled, start event producer
	if !runContext.PreviousRunTimestamp.IsZero() {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := b.eventsProducer(ctx, runContext); err != nil {
				log.WithError(err).Error("bundle event producer failed")
				atomic.AddInt32(&b.producerErrors, 1)
			}
		}()
	}

	wg.Wait()

	if b.producerErrors > 0 {
		return errors.New("One or more bundle producers failed")
	}
	return nil
}

func (b *ResourceBundle) Objects() <-chan api.Object {
	return b.objs
}

func (b *ResourceBundle) CollectionErrors() <-chan integration.CollectionError {
	return b.errs
}

// Consumers returns a joint list of all member resource's event processors
func (b *ResourceBundle) GetEventProcessors() []downloader.PostProcessor {
	postProcessors := []downloader.PostProcessor{}
	for _, res := range b.resources {
		if i, ok := res.(tasks.HasEventProcessors); ok {
			postProcessors = append(postProcessors, i.GetEventProcessors()...)
		}
	}
	return postProcessors
}

// Consumers returns a joint list of all member resource's consumers
func (b *ResourceBundle) Consumers() []integration.Consumer {
	result := []integration.Consumer{}
	for _, res := range b.resources {
		for _, con := range res.Consumers() {
			result = append(result, con)
		}
	}
	return result
}

func (b *ResourceBundle) Close() {
	for _, res := range b.resources {
		res.Close()
	}
}

func New(apiClient api.Client, resources ...integration.Resource) *ResourceBundle {
	return &ResourceBundle{
		apiClient: apiClient,
		objs:      make(chan api.Object, 1000),
		errs:      make(chan integration.CollectionError),
		resources: resources,
	}
}
