package resource

import (
	"context"
	"github.com/segment-sources/stripe/api"
	"github.com/segment-sources/stripe/integration"
	"github.com/segment-sources/stripe/resource/dedupe"
	"github.com/segment-sources/stripe/resource/downloader"
	"github.com/segment-sources/stripe/resource/processors"
	"github.com/segment-sources/stripe/resource/tr"
	"github.com/segmentio/go-source"
)

var planEvents = []string{
	"plan.created",
	"plan.updated",
	"plan.deleted",
}

type Plan struct {
	name      string
	apiClient api.Client
	objs      chan api.Object
	msgs      chan source.SetMessage
	errs      chan integration.CollectionError
	dedupe    dedupe.Interface
}

func (r *Plan) DesiredObjects() []string {
	return []string{"plan", "subscription", "invoice"}
}

func (r *Plan) DesiredEvents() []string {
	allEventTypes := planEvents
	allEventTypes = append(allEventTypes, subscriptionEvents...)
	allEventTypes = append(allEventTypes, invoiceEvents...)
	return allEventTypes
}

func (r *Plan) StartProducer(ctx context.Context, runContext integration.RunContext) error {
	defer close(r.objs)
	defer close(r.errs)
	if runContext.PreviousRunTimestamp.IsZero() {
		return downloader.New(r.apiClient).Do(ctx, &downloader.Task{
			Collection: r.name,
			Request: &api.Request{
				Url: "/v1/plans?limit=100",
			},
			Output: r.objs,
			Errors: r.errs,
		})
	}

	// downloading events in incremental mode is handled by the bundle that this resource is a part of
	return nil
}

func (r *Plan) GetEventProcessors() []downloader.PostProcessor {
	return []downloader.PostProcessor{
		processors.NewIsDeleted("plan.deleted"),
	}
}

func (r *Plan) StartConsumer(ctx context.Context, ch <-chan api.Object) {
	defer close(r.msgs)
	for obj := range ch {
		switch tr.GetString(obj, "object") {
		case "event":
			if payload := tr.ExtractEventPayload(obj, "plan", "subscription"); payload != nil {
				switch tr.GetString(payload, "object") {
				case "plan":
					r.consumePlan(payload, true)
				case "subscription":
					r.consumeSubscription(obj, true)
				case "invoice":
					r.consumeInvoice(obj, true)
				}

			}
		case "plan":
			r.consumePlan(obj, false)
		case "subscription":
			r.consumeSubscription(obj, false)
		case "invoice":
			r.consumeInvoice(obj, false)
		}
	}
}

func (r *Plan) consumePlan(obj api.Object, fromEvent bool) {
	if msg := r.transform(obj); msg != nil && !(fromEvent && r.dedupe.SeenBefore(msg.ID)) {
		r.msgs <- *msg
	}
}

func (r *Plan) consumeSubscription(obj api.Object, fromEvent bool) {
	if plan := tr.GetMap(obj, "plan"); plan != nil {
		r.consumePlan(plan, fromEvent)
	}
}

func (r *Plan) consumeInvoice(obj api.Object, fromEvent bool) {
	for _, line := range tr.GetMapList(tr.GetMap(obj, "lines"), "data") {
		r.consumeInvoiceLine(line, fromEvent)
	}
}

func (r *Plan) consumeInvoiceLine(obj api.Object, fromEvent bool) {
	if plan := tr.GetMap(obj, "plan"); plan != nil {
		r.consumePlan(plan, fromEvent)
	}
}

func (r *Plan) transform(obj api.Object) *source.SetMessage {
	var id string
	if id = tr.GetString(obj, "id"); id == "" {
		return nil
	}

	properties := map[string]interface{}{
		"interval":             obj["interval"],
		"name":                 obj["name"],
		"amount":               obj["amount"],
		"currency":             obj["currency"],
		"interval_count":       obj["interval_count"],
		"trial_period_days":    obj["trial_period_days"],
		"statement_descriptor": obj["statement_descriptor"],
	}

	if v, ok := obj["is_deleted"].(bool); ok && v {
		properties["is_deleted"] = v
	}

	if metadata := tr.GetMap(obj, "metadata"); metadata != nil {
		tr.Flatten(metadata, "metadata_", properties)
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

func (r *Plan) Collection() string {
	return r.name
}

func (r *Plan) Objects() <-chan api.Object {
	return r.objs
}

func (r *Plan) Messages() <-chan source.SetMessage {
	return r.msgs
}

func (r *Plan) CollectionErrors() <-chan integration.CollectionError {
	return r.errs
}

func (r *Plan) Consumers() []integration.Consumer {
	return []integration.Consumer{r}
}

func (r *Plan) Close() {
	r.dedupe.Close()
}

func NewPlan(apiClient api.Client) *Plan {
	return &Plan{
		name:      "plans",
		apiClient: apiClient,
		objs:      make(chan api.Object, 1000),
		msgs:      make(chan source.SetMessage),
		errs:      make(chan integration.CollectionError),
		dedupe:    dedupe.New(),
	}
}
