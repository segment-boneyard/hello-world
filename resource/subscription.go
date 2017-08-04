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

var subscriptionEvents = []string{
	"customer.subscription.created",
	"customer.subscription.updated",
	"customer.subscription.trial_will_end",
	"customer.subscription.deleted",
}

// Subscription
// no known discrepancies
type Subscription struct {
	name      string
	apiClient api.Client
	objs      chan api.Object
	msgs      chan source.SetMessage
	errs      chan integration.CollectionError
	dedupe    dedupe.Interface
}

func (r *Subscription) DesiredObjects() []string {
	return []string{"subscription"}
}

func (r *Subscription) DesiredEvents() []string {
	return subscriptionEvents
}

func (r *Subscription) StartProducer(ctx context.Context, runContext integration.RunContext) error {
	defer close(r.objs)
	defer close(r.errs)
	if runContext.PreviousRunTimestamp.IsZero() {
		return downloader.New(r.apiClient).Do(ctx, &downloader.Task{
			Collection: r.name,
			Request: &api.Request{
				Url:           "/v1/subscriptions?status=all&limit=100",
				LogCollection: r.name,
			},
			Output: r.objs,
			Errors: r.errs,
			PostProcessors: []downloader.PostProcessor{
				processors.NewListExpander("items", r.apiClient),
			},
		})
	}

	// downloading events in incremental mode is handled by the bundle that this resource is a part of
	return nil
}

func (r *Subscription) GetEventProcessors() []downloader.PostProcessor {
	return []downloader.PostProcessor{
		processors.NewIsDeleted("customer.subscription.deleted"),
	}
}

func (r *Subscription) StartConsumer(ctx context.Context, ch <-chan api.Object) {
	defer close(r.msgs)
	for obj := range ch {
		switch tr.GetString(obj, "object") {
		case "event":
			if payload := tr.ExtractEventPayload(obj, "subscription"); payload != nil {
				r.consumeSubscription(payload, true)
			}
		case "subscription":
			r.consumeSubscription(obj, false)
		}
	}
}

func (r *Subscription) consumeSubscription(obj api.Object, fromEvent bool) {
	if msg := r.transform(obj); msg != nil && !(fromEvent && r.dedupe.SeenBefore(msg.ID)) {
		r.msgs <- *msg
	}
}

func (r *Subscription) transform(obj api.Object) *source.SetMessage {
	var id string
	if id = tr.GetString(obj, "id"); id == "" {
		return nil
	}

	properties := map[string]interface{}{
		"application_fee_percent": obj["application_fee_percent"],
		"cancel_at_period_end":    obj["cancel_at_period_end"],
		"customer_id":             obj["customer"],
		"quantity":                obj["quantity"],
		"status":                  obj["status"],
		"tax_percent":             obj["tax_percent"],
	}

	if v, ok := obj["is_deleted"].(bool); ok && v {
		properties["is_deleted"] = v
	}

	tr.Flatten(tr.GetMap(obj, "metadata"), "metadata_", properties)

	timestampFields := []string{
		"start",
		"created",
		"current_period_start",
		"current_period_end",
		"ended_at",
		"trial_start",
		"trial_end",
		"canceled_at",
	}
	for _, f := range timestampFields {
		if ts := tr.GetTimestamp(obj, f); ts != "" {
			properties[f] = ts
		}
	}

	if plan := tr.GetMap(obj, "plan"); plan != nil {
		if planId := tr.GetString(plan, "id"); planId != "" {
			properties["plan_id"] = planId
		}
	}

	if discount := tr.GetMap(obj, "discount"); discount != nil {
		if discountId := makeDiscountId(discount); discountId != "" {
			properties["discount_id"] = discountId
		}
	}

	return &source.SetMessage{
		ID:         id,
		Collection: r.name,
		Properties: properties,
	}
}

func (r *Subscription) Collection() string {
	return r.name
}

func (r *Subscription) Objects() <-chan api.Object {
	return r.objs
}

func (r *Subscription) Messages() <-chan source.SetMessage {
	return r.msgs
}

func (r *Subscription) CollectionErrors() <-chan integration.CollectionError {
	return r.errs
}

func (r *Subscription) Consumers() []integration.Consumer {
	return []integration.Consumer{r}
}

func (r *Subscription) Close() {
	r.dedupe.Close()
}

func NewSubscription(apiClient api.Client) *Subscription {
	return &Subscription{
		name:      "subscriptions",
		apiClient: apiClient,
		objs:      make(chan api.Object, 1000),
		msgs:      make(chan source.SetMessage),
		errs:      make(chan integration.CollectionError),
		dedupe:    dedupe.New(),
	}
}
