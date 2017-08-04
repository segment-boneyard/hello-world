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

var refundEvents = []string{
	"charge.refund.updated",
}

type Refund struct {
	name      string
	apiClient api.Client
	objs      chan api.Object
	msgs      chan source.SetMessage
	errs      chan integration.CollectionError
	dedupe    dedupe.Interface
}

func (r *Refund) DesiredObjects() []string {
	return []string{"refund"}
}

func (r *Refund) DesiredEvents() []string {
	return append(refundEvents, chargeEvents...)
}

func (r *Refund) StartProducer(ctx context.Context, runContext integration.RunContext) error {
	defer close(r.objs)
	defer close(r.errs)
	if runContext.PreviousRunTimestamp.IsZero() {
		return downloader.New(r.apiClient).Do(ctx, &downloader.Task{
			Collection: r.name,
			Request: &api.Request{
				Url:           "/v1/refunds?limit=100",
				LogCollection: r.name,
			},
			Output: r.objs,
			Errors: r.errs,
		})
	}

	// downloading events in incremental mode is handled by the bundle that this resource is a part of
	return nil
}

func (r *Refund) StartConsumer(ctx context.Context, ch <-chan api.Object) {
	defer close(r.msgs)
	for obj := range ch {
		switch tr.GetString(obj, "object") {
		case "event":
			if payload := tr.ExtractEventPayload(obj, "refund", "charge"); payload != nil {
				switch tr.GetString(payload, "object") {
				case "refund":
					r.consumeRefund(payload, true)
				case "charge":
					r.consumeCharge(payload, true)
				}
			}
		case "refund":
			r.consumeRefund(obj, false)
		}
	}
}

func (r *Refund) consumeRefund(obj api.Object, fromEvent bool) {
	if msg := r.transform(obj); msg != nil && !(fromEvent && r.dedupe.SeenBefore(msg.ID)) {
		r.msgs <- *msg
	}
}

func (r *Refund) consumeCharge(obj api.Object, fromEvent bool) {
	var refundsMap map[string]interface{}
	if refundsMap = tr.GetMap(obj, "refunds"); refundsMap == nil {
		return
	}

	var refundsList []map[string]interface{}
	if refundsList = tr.GetMapList(refundsMap, "data"); refundsList == nil {
		return
	}

	for _, refund := range refundsList {
		r.consumeRefund(refund, fromEvent)
	}
}

func (r *Refund) transform(obj api.Object) *source.SetMessage {
	var id string
	if id = tr.GetString(obj, "id"); id == "" {
		return nil
	}

	properties := map[string]interface{}{
		"amount":                 obj["amount"],
		"currency":               obj["currency"],
		"balance_transaction_id": obj["balance_transaction"],
		"charge_id":              obj["charge"],
		"receipt_number":         obj["receipt_number"],
		"reason":                 obj["reason"],
	}

	if created := tr.GetTimestamp(obj, "created"); created != "" {
		properties["created"] = created
	}

	tr.Flatten(tr.GetMap(obj, "metadata"), "metadata_", properties)

	return &source.SetMessage{
		ID:         id,
		Collection: r.name,
		Properties: properties,
	}
}

func (r *Refund) Collection() string {
	return r.name
}

func (r *Refund) Objects() <-chan api.Object {
	return r.objs
}

func (r *Refund) Messages() <-chan source.SetMessage {
	return r.msgs
}

func (r *Refund) CollectionErrors() <-chan integration.CollectionError {
	return r.errs
}

func (r *Refund) Consumers() []integration.Consumer {
	return []integration.Consumer{r}
}

func (r *Refund) Close() {
	r.dedupe.Close()
}

func NewRefund(apiClient api.Client) *Refund {
	return &Refund{
		name:      "refunds",
		apiClient: apiClient,
		objs:      make(chan api.Object, 1000),
		msgs:      make(chan source.SetMessage),
		errs:      make(chan integration.CollectionError),
		dedupe:    dedupe.New(),
	}
}
