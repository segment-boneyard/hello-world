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

var disputeEvents = []string{
	"charge.dispute.closed",
	"charge.dispute.created",
	"charge.dispute.funds_reinstated",
	"charge.dispute.funds_withdrawn",
	"charge.dispute.updated",
}

type Dispute struct {
	name      string
	apiClient api.Client
	objs      chan api.Object
	msgs      chan source.SetMessage
	errs      chan integration.CollectionError
	dedupe    dedupe.Interface
}

func (r *Dispute) DesiredObjects() []string {
	return []string{"dispute"}
}

func (r *Dispute) DesiredEvents() []string {
	return disputeEvents
}

func (r *Dispute) StartProducer(ctx context.Context, runContext integration.RunContext) error {
	defer close(r.objs)
	defer close(r.errs)
	var task *downloader.Task
	if runContext.PreviousRunTimestamp.IsZero() {
		task = &downloader.Task{
			Collection: r.name,
			Request: &api.Request{
				Url:           "/v1/disputes?limit=100",
				LogCollection: r.name,
			},
			Output: r.objs,
			Errors: r.errs,
		}
	} else {
		task = tasks.MakeIncremental(r, r.name, runContext.PreviousRunTimestamp, r.objs, r.errs)
	}

	return downloader.New(r.apiClient).Do(ctx, task)
}

func (r *Dispute) StartConsumer(ctx context.Context, ch <-chan api.Object) {
	defer close(r.msgs)
	defer r.dedupe.Close()
	for obj := range ch {
		switch tr.GetString(obj, "object") {
		case "event":
			if payload := tr.ExtractEventPayload(obj, "dispute"); payload != nil {
				r.consumeDispute(payload, true)
			}
		case "dispute":
			r.consumeDispute(obj, false)
		}
	}
}

func (r *Dispute) consumeDispute(obj api.Object, fromEvent bool) {
	if msg := r.transform(obj); msg != nil && !(fromEvent && r.dedupe.SeenBefore(msg.ID)) {
		r.msgs <- *msg
	}
}

func (r *Dispute) transform(obj api.Object) *source.SetMessage {
	var id string
	if id = tr.GetString(obj, "id"); id == "" {
		return nil
	}

	properties := map[string]interface{}{
		"charge_id":            obj["charge"],
		"amount":               obj["amount"],
		"status":               obj["status"],
		"currency":             obj["currency"],
		"reason":               obj["reason"],
		"is_charge_refundable": obj["is_charge_refundable"],
	}

	tr.Flatten(tr.GetMap(obj, "metadata"), "metadata_", properties)
	tr.Flatten(tr.GetMap(obj, "evidence_details"), "evidence_details_", properties)
	tr.Flatten(tr.GetMap(obj, "evidence"), "evidence_", properties)

	if ts := tr.GetTimestamp(obj, "created"); ts != "" {
		properties["created"] = ts
	}

	return &source.SetMessage{
		ID:         id,
		Collection: r.name,
		Properties: properties,
	}
}

func (r *Dispute) Collection() string {
	return r.name
}

func (r *Dispute) Objects() <-chan api.Object {
	return r.objs
}

func (r *Dispute) Messages() <-chan source.SetMessage {
	return r.msgs
}

func (r *Dispute) CollectionErrors() <-chan integration.CollectionError {
	return r.errs
}

func (r *Dispute) Consumers() []integration.Consumer {
	return []integration.Consumer{r}
}

func (r *Dispute) Close() {
	r.dedupe.Close()
}

func NewDispute(apiClient api.Client) *Dispute {
	return &Dispute{
		name:      "disputes",
		apiClient: apiClient,
		objs:      make(chan api.Object, 1000),
		msgs:      make(chan source.SetMessage),
		errs:      make(chan integration.CollectionError),
		dedupe:    dedupe.New(),
	}
}
