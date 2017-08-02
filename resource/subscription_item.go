package resource

import (
	"context"
	"github.com/segment-sources/stripe/api"
	"github.com/segment-sources/stripe/integration"
	"github.com/segment-sources/stripe/resource/dedupe"
	"github.com/segment-sources/stripe/resource/tr"
	"github.com/segmentio/go-source"
)

type SubscriptionItem struct {
	name      string
	apiClient api.Client
	objs      chan api.Object
	msgs      chan source.SetMessage
	errs      chan integration.CollectionError
	dedupe    dedupe.Interface
}

func (r *SubscriptionItem) DesiredObjects() []string {
	return []string{"subscription"}
}

func (r *SubscriptionItem) DesiredEvents() []string {
	return subscriptionEvents
}

func (r *SubscriptionItem) StartProducer(ctx context.Context, runContext integration.RunContext) error {
	// downloading events in incremental mode is handled by the bundle that this resource is a part of
	close(r.objs)
	close(r.errs)
	return nil
}

func (r *SubscriptionItem) StartConsumer(ctx context.Context, ch <-chan api.Object) {
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

func (r *SubscriptionItem) consumeSubscription(obj api.Object, fromEvent bool) {
	var subId string
	if subId = tr.GetString(obj, "id"); subId == "" || fromEvent && r.dedupe.SeenBefore(subId) {
		return
	}

	for _, line := range tr.GetMapList(tr.GetMap(obj, "items"), "data") {
		if msg := r.transform(obj, line); msg != nil {
			r.msgs <- *msg
		}
	}
}

func (r *SubscriptionItem) transform(subscription, item api.Object) *source.SetMessage {
	var itemId string
	if itemId = tr.GetString(item, "id"); itemId == "" {
		return nil
	}
	var subscriptionId string
	if subscriptionId = tr.GetString(subscription, "id"); itemId == "" {
		return nil
	}

	properties := map[string]interface{}{
		"subscription_id": subscriptionId,
		"quantity":        item["quantity"],
	}

	tr.Flatten(tr.GetMap(item, "metadata"), "metadata_", properties)
	if created := tr.GetTimestamp(item, "created"); created != "" {
		properties["created"] = created
	}

	if plan := tr.GetMap(item, "plan"); plan != nil {
		if planId := tr.GetString(plan, "id"); planId != "" {
			properties["plan_id"] = planId
		}
	}

	return &source.SetMessage{
		ID:         itemId,
		Collection: r.name,
		Properties: properties,
	}
}

func (r *SubscriptionItem) Collection() string {
	return r.name
}

func (r *SubscriptionItem) Objects() <-chan api.Object {
	return r.objs
}

func (r *SubscriptionItem) Messages() <-chan source.SetMessage {
	return r.msgs
}

func (r *SubscriptionItem) CollectionErrors() <-chan integration.CollectionError {
	return r.errs
}

func (r *SubscriptionItem) Consumers() []integration.Consumer {
	return []integration.Consumer{r}
}

func (r *SubscriptionItem) Close() {
	r.dedupe.Close()
}

func NewSubscriptionItem(apiClient api.Client) *SubscriptionItem {
	return &SubscriptionItem{
		name:      "subscription_items",
		apiClient: apiClient,
		objs:      make(chan api.Object, 1000),
		msgs:      make(chan source.SetMessage),
		errs:      make(chan integration.CollectionError),
		dedupe:    dedupe.New(),
	}
}
