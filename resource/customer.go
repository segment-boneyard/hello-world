package resource

import (
	"context"
	"github.com/segment-sources/stripe/api"
	"github.com/segment-sources/stripe/integration"
	"github.com/segment-sources/stripe/resource/dedupe"
	"github.com/segment-sources/stripe/resource/downloader"
	"github.com/segment-sources/stripe/resource/processors"
	"github.com/segment-sources/stripe/resource/tasks"
	"github.com/segment-sources/stripe/resource/tr"
	"github.com/segmentio/go-source"
)

var customerEvents = []string{
	"customer.created",
	"customer.updated",
	"customer.deleted",
}

type Customer struct {
	name      string
	apiClient api.Client
	objs      chan api.Object
	msgs      chan source.SetMessage
	errs      chan integration.CollectionError
	dedupe    dedupe.Interface
}

func (r *Customer) DesiredObjects() []string {
	return []string{"customer"}
}

func (r *Customer) DesiredEvents() []string {
	return customerEvents
}

func (r *Customer) StartProducer(ctx context.Context, runContext integration.RunContext) error {
	defer close(r.objs)
	defer close(r.errs)
	var task *downloader.Task
	if runContext.PreviousRunTimestamp.IsZero() {
		task = &downloader.Task{
			Collection: r.name,
			Request: &api.Request{
				Url:           "/v1/customers?limit=100",
				LogCollection: r.name,
			},
			Output: r.objs,
			Errors: r.errs,
			PostProcessors: []downloader.PostProcessor{
				processors.NewListExpander("sources", r.apiClient),
			},
		}
	} else {
		task = tasks.MakeIncremental(r, r.name, runContext.PreviousRunTimestamp, r.objs, r.errs)
	}

	return downloader.New(r.apiClient).Do(ctx, task)
}

func (r *Customer) GetEventProcessors() []downloader.PostProcessor {
	return []downloader.PostProcessor{
		processors.NewIsDeleted("customer.deleted"),
	}
}

func (r *Customer) StartConsumer(ctx context.Context, ch <-chan api.Object) {
	defer close(r.msgs)
	for obj := range ch {
		switch tr.GetString(obj, "object") {
		case "event":
			if payload := tr.ExtractEventPayload(obj, "customer"); payload != nil {
				r.consumeCustomer(payload, true)
			}
		case "customer":
			r.consumeCustomer(obj, false)
		}
	}
}

func (r *Customer) consumeCustomer(obj api.Object, fromEvent bool) {
	if msg := r.transform(obj); msg != nil && !(fromEvent && r.dedupe.SeenBefore(msg.ID)) {
		r.msgs <- *msg
	}
}

func (r *Customer) transform(obj api.Object) *source.SetMessage {
	var id string
	if id = tr.GetString(obj, "id"); id == "" {
		return nil
	}

	properties := map[string]interface{}{
		"account_balance": obj["account_balance"],
		"currency":        obj["currency"],
		"delinquent":      obj["delinquent"],
		"description":     obj["description"],
		"email":           obj["email"],
	}

	if v, ok := obj["is_deleted"].(bool); ok && v {
		properties["is_deleted"] = v
	}

	tr.Flatten(tr.GetMap(obj, "metadata"), "metadata_", properties)

	if created := tr.GetTimestamp(obj, "created"); created != "" {
		properties["created"] = created
	}

	return &source.SetMessage{
		ID:         id,
		Collection: r.name,
		Properties: properties,
	}
}

func (r *Customer) Collection() string {
	return r.name
}

func (r *Customer) Objects() <-chan api.Object {
	return r.objs
}

func (r *Customer) Messages() <-chan source.SetMessage {
	return r.msgs
}

func (r *Customer) CollectionErrors() <-chan integration.CollectionError {
	return r.errs
}

func (r *Customer) Consumers() []integration.Consumer {
	return []integration.Consumer{r}
}

func (r *Customer) Close() {
	r.dedupe.Close()
}

func NewCustomer(apiClient api.Client) *Customer {
	return &Customer{
		name:      "customers",
		apiClient: apiClient,
		objs:      make(chan api.Object, 1000),
		msgs:      make(chan source.SetMessage),
		errs:      make(chan integration.CollectionError),
		dedupe:    dedupe.New(),
	}
}
