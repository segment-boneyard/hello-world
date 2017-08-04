package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/apex/log"
	"github.com/pkg/errors"
	"github.com/segment-sources/stripe/api"
	"github.com/segmentio/go-source"
	"github.com/segmentio/ur-log"
	"reflect"
	"sync"
	"sync/atomic"
	"time"
)

type Dispatcher struct {
	sourceClient        source.Client
	resources           []Resource
	subscriptions       []subscription
	eventSubscriptions  map[string][]subscription
	objectSubscriptions map[string][]subscription
	startedAt           time.Time
	producerFailures    int32
	collectionErrors    int32
	runContext          RunContext
}

const contextVersion = 1
const staleContextThreshold = time.Hour * 24 * 30 // 30 days

func (d *Dispatcher) Register(res Resource) {
	d.resources = append(d.resources, res)
	for _, con := range res.Consumers() {
		sub := subscription{
			ch:       make(chan api.Object),
			consumer: con,
		}
		d.subscriptions = append(d.subscriptions, sub)
		for _, eventType := range con.DesiredEvents() {
			d.eventSubscriptions[eventType] = append(d.eventSubscriptions[eventType], sub)
		}
		for _, objectType := range con.DesiredObjects() {
			d.objectSubscriptions[objectType] = append(d.objectSubscriptions[objectType], sub)
		}
	}
}

func (d *Dispatcher) routerWorker(res Resource) {
	for obj := range res.Objects() {
		if objectType, ok := (obj["object"]).(string); ok && objectType == "event" {
			if eventType, ok := (obj["type"]).(string); ok {
				for _, con := range d.eventSubscriptions[eventType] {
					con.ch <- obj
				}
			}
		} else if objectType != "" {
			for _, con := range d.objectSubscriptions[objectType] {
				con.ch <- obj
			}
		}
	}
}

func (d *Dispatcher) setWorker(sub subscription) {
	for msg := range sub.consumer.Messages() {
		if err := d.sourceClient.Set(msg.Collection, msg.ID, msg.Properties); err != nil {
			log.WithError(err).Fatal("Set call failed, aborting the sync")
		}
	}
}

func (d *Dispatcher) runProducers(ctx context.Context) *sync.WaitGroup {
	producerWg := sync.WaitGroup{}
	for _, res := range d.resources {
		producerWg.Add(1)
		go func(res Resource) {
			defer producerWg.Done()
			d.routerWorker(res)
		}(res)
		producerWg.Add(1)
		go func(res Resource) {
			defer producerWg.Done()
			if err := res.StartProducer(ctx, d.runContext); err != nil {
				atomic.AddInt32(&d.producerFailures, 1)
				operation := fmt.Sprintf("running producer %s", reflect.TypeOf(res).String())
				d.sourceClient.Log().Error("", operation, err)
				log.WithError(err).Error("producer failed")
			}
		}(res)
		producerWg.Add(1)
		go func(res Resource) {
			defer producerWg.Done()
			for err := range res.CollectionErrors() {
				atomic.AddInt32(&d.collectionErrors, 1)
				d.sourceClient.ReportError(err.Message, err.Collection)
				d.sourceClient.Log().Error(err.Collection, "syncing colleciton", errors.New(err.Message))
			}
		}(res)
	}

	return &producerWg
}

func (d *Dispatcher) runConsumers(ctx context.Context) *sync.WaitGroup {
	consumerWg := sync.WaitGroup{}
	for _, sub := range d.subscriptions {
		consumerWg.Add(1)
		go func(sub subscription) {
			defer consumerWg.Done()
			d.setWorker(sub)
		}(sub)
		consumerWg.Add(1)
		go func(sub subscription) {
			defer consumerWg.Done()
			sub.consumer.StartConsumer(ctx, sub.ch)
		}(sub)
	}
	return &consumerWg
}

func (d *Dispatcher) initContext(ctx context.Context) error {
	d.startedAt = time.Now().UTC()

	doc, err := d.sourceClient.GetContext(source.GetContextOptions{AllowFailed: false})
	if err != nil {
		d.sourceClient.Log().Error("", "loading context", err)
		return urlog.WrapError(ctx, err, "GetContext call failed")
	}

	if len(doc) < 1 {
		log.Info("no run context available")
		return nil
	}

	value := RunContext{}
	if err := json.Unmarshal(doc, &value); err != nil {
		d.sourceClient.Log().Error("", "decoding context", err)
		log.WithError(err).WithField("context", string(doc)).Error("failed to unmarshal context")
		return urlog.WrapError(ctx, err, "unmarshalling context failed")
	}

	if value.Version != contextVersion {
		log.WithFields(log.Fields{
			"required_version": contextVersion,
			"loaded_version":   value.Version,
		}).Infof("context version mismatch")
		return nil
	}

	if value.PreviousRunTimestamp.IsZero() {
		log.Info("run context doesn't contain a timestamp")
		return nil
	}

	if time.Now().UTC().Sub(value.PreviousRunTimestamp) > staleContextThreshold {
		log.Infof("discarding context as it is older than %s", staleContextThreshold.String())
		return nil
	}

	d.runContext = value
	return nil
}

func (d *Dispatcher) saveContext(ctx context.Context) error {
	value := RunContext{
		Version:              contextVersion,
		PreviousRunTimestamp: d.startedAt,
	}

	doc, _ := json.Marshal(value)
	if err := d.sourceClient.SetContext(doc); err != nil {
		d.sourceClient.Log().Error("", "saving context", err)
		return urlog.WrapError(ctx, err, "SetContext call failed")
	}

	return nil
}

func (d *Dispatcher) Run() error {
	ctx := context.Background()

	if err := d.initContext(ctx); err != nil {
		return err
	}

	consumerWg := d.runConsumers(ctx)
	producerWg := d.runProducers(ctx)

	producerWg.Wait()

	for _, sub := range d.subscriptions {
		close(sub.ch)
	}

	consumerWg.Wait()

	if err := d.saveContext(ctx); err != nil {
		return err
	}

	if d.producerFailures > 0 {
		return errors.New("One or more producers failed")
	}

	if d.collectionErrors > 0 {
		return errors.New("One or more collection errors happened")
	}

	return nil
}

func (d *Dispatcher) Close() {
	for _, res := range d.resources {
		res.Close()
	}
}

func NewDispatcher(sourceClient source.Client) *Dispatcher {
	return &Dispatcher{
		sourceClient:        sourceClient,
		eventSubscriptions:  make(map[string][]subscription),
		objectSubscriptions: make(map[string][]subscription),
	}
}
