package resource

import (
	"context"
	"github.com/segment-sources/stripe/api"
	"github.com/segment-sources/stripe/integration"
	"github.com/segment-sources/stripe/resource/dedupe"
	"github.com/segment-sources/stripe/resource/downloader"
	"github.com/segment-sources/stripe/resource/tasks"
	"github.com/segment-sources/stripe/resource/tr"
	"github.com/segmentio/go-source"
)

var orderReturnEvents = []string{
	"order_return.created",
}

type OrderReturn struct {
	name      string
	apiClient api.Client
	objs      chan api.Object
	msgs      chan source.SetMessage
	errs      chan integration.CollectionError
	dedupe    dedupe.Interface
}

func (r *OrderReturn) DesiredObjects() []string {
	return []string{"order_return"}
}

func (r *OrderReturn) DesiredEvents() []string {
	return orderReturnEvents
}

func (r *OrderReturn) StartProducer(ctx context.Context, runContext integration.RunContext) error {
	defer close(r.objs)
	defer close(r.errs)
	var task *downloader.Task
	if runContext.PreviousRunTimestamp.IsZero() {
		task = &downloader.Task{
			Collection: r.name,
			Request: &api.Request{
				Url: "/v1/order_returns?limit=100",
			},
			Output: r.objs,
			Errors: r.errs,
		}
	} else {
		task = tasks.MakeIncremental(r, r.name, runContext.PreviousRunTimestamp, r.objs, r.errs)
	}

	return downloader.New(r.apiClient).Do(ctx, task)
}

func (r *OrderReturn) StartConsumer(ctx context.Context, ch <-chan api.Object) {
	defer close(r.msgs)
	for obj := range ch {
		switch tr.GetString(obj, "object") {
		case "event":
			if payload := tr.ExtractEventPayload(obj, "order_return"); payload != nil {
				r.consumeReturn(payload, true)
			}
		case "order_return":
			r.consumeReturn(obj, false)
		}
	}
}

func (r *OrderReturn) consumeReturn(obj api.Object, fromEvent bool) {
	if msg := r.transform(obj); msg != nil && !(fromEvent && r.dedupe.SeenBefore(msg.ID)) {
		r.msgs <- *msg
	}
}

func (r *OrderReturn) Collection() string {
	return r.name
}

func (r *OrderReturn) Objects() <-chan api.Object {
	return r.objs
}

func (r *OrderReturn) Messages() <-chan source.SetMessage {
	return r.msgs
}

func (r *OrderReturn) CollectionErrors() <-chan integration.CollectionError {
	return r.errs
}

func (r *OrderReturn) Consumers() []integration.Consumer {
	return []integration.Consumer{r}
}

func (r *OrderReturn) transform(obj api.Object) *source.SetMessage {
	var id string
	if id = tr.GetString(obj, "id"); id == "" {
		return nil
	}

	properties := map[string]interface{}{
		"amount":      obj["amount"],
		"currency":    obj["currency"],
		"description": obj["description"],
		"parent_id":   obj["parent"],
		"quantity":    obj["quantity"],
		"type":        obj["type"],
		"livemode":    obj["livemode"],
		"order_id":    obj["order"],
	}
	if ts := tr.GetTimestamp(obj, "created"); ts != "" {
		properties["created"] = ts
	}
	return &source.SetMessage{
		ID:         id,
		Collection: r.name,
		Properties: properties,
	}
}

func (r *OrderReturn) Close() {
	r.dedupe.Close()
}

func NewOrderReturn(apiClient api.Client) *OrderReturn {
	return &OrderReturn{
		name:      "order_returns",
		apiClient: apiClient,
		objs:      make(chan api.Object, 1000),
		msgs:      make(chan source.SetMessage),
		errs:      make(chan integration.CollectionError),
		dedupe:    dedupe.New(),
	}
}
