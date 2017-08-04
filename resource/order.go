package resource

import (
	"context"
	"github.com/segment-sources/stripe/api"
	"github.com/segment-sources/stripe/integration"
	"github.com/segment-sources/stripe/resource/dedupe"
	"github.com/segment-sources/stripe/resource/downloader"
	"github.com/segment-sources/stripe/resource/tr"
	"github.com/segmentio/go-source"
)

var orderEvents = []string{
	"order.created",
	"order.payment_failed",
	"order.payment_succeeded",
	"order.updated",
}

type Order struct {
	name      string
	apiClient api.Client
	objs      chan api.Object
	msgs      chan source.SetMessage
	errs      chan integration.CollectionError
	dedupe    dedupe.Interface
}

func (r *Order) DesiredObjects() []string {
	return []string{"order"}
}

func (r *Order) DesiredEvents() []string {
	return orderEvents
}

func (r *Order) StartProducer(ctx context.Context, runContext integration.RunContext) error {
	defer close(r.objs)
	defer close(r.errs)
	if runContext.PreviousRunTimestamp.IsZero() {
		return downloader.New(r.apiClient).Do(ctx, &downloader.Task{
			Collection: r.name,
			Request: &api.Request{
				Url:           "/v1/orders?limit=100",
				LogCollection: r.name,
			},
			Output: r.objs,
			Errors: r.errs,
		})
	}

	// downloading events in incremental mode is handled by the bundle that this resource is a part of
	return nil
}

func (r *Order) StartConsumer(ctx context.Context, ch <-chan api.Object) {
	defer close(r.msgs)
	for obj := range ch {
		switch tr.GetString(obj, "object") {
		case "event":
			if payload := tr.ExtractEventPayload(obj, "order"); payload != nil {
				r.consumeOrder(payload, true)
			}
		case "order":
			r.consumeOrder(obj, false)
		}
	}
}

func (r *Order) consumeOrder(obj api.Object, fromEvent bool) {
	if msg := r.transform(obj); msg != nil && !(fromEvent && r.dedupe.SeenBefore(msg.ID)) {
		r.msgs <- *msg
	}
}

func (r *Order) transform(obj api.Object) *source.SetMessage {
	var id string
	if id = tr.GetString(obj, "id"); id == "" {
		return nil
	}

	properties := map[string]interface{}{
		"amount":                   obj["amount"],
		"amount_returned":          obj["amount_returned"],
		"application":              obj["application"],
		"application_fee":          obj["application_fee"],
		"charge_id":                obj["charge"],
		"currency":                 obj["currency"],
		"customer_id":              obj["customer"],
		"email":                    obj["email"],
		"livemode":                 obj["livemode"],
		"selected_shipping_method": obj["selected_shipping_method"],
		"status":                   obj["status"],
	}

	tr.Flatten(tr.GetMap(obj, "metadata"), "metadata_", properties)
	tr.Flatten(tr.GetMap(obj, "shipping"), "shipping_", properties)

	if ts := tr.GetTimestamp(obj, "created"); ts != "" {
		properties["created"] = ts
	}
	if ts := tr.GetTimestamp(obj, "updated"); ts != "" {
		properties["updated"] = ts
	}

	return &source.SetMessage{
		ID:         id,
		Collection: r.name,
		Properties: properties,
	}
}

func (r *Order) Collection() string {
	return r.name
}

func (r *Order) Objects() <-chan api.Object {
	return r.objs
}

func (r *Order) Messages() <-chan source.SetMessage {
	return r.msgs
}

func (r *Order) CollectionErrors() <-chan integration.CollectionError {
	return r.errs
}

func (r *Order) Consumers() []integration.Consumer {
	return []integration.Consumer{r}
}

func (r *Order) Close() {
	r.dedupe.Close()
}

func NewOrder(apiClient api.Client) *Order {
	return &Order{
		name:      "orders",
		apiClient: apiClient,
		objs:      make(chan api.Object, 1000),
		msgs:      make(chan source.SetMessage),
		errs:      make(chan integration.CollectionError),
		dedupe:    dedupe.New(),
	}
}
