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

var cardEvents = []string{
	"customer.source.created",
	"customer.source.updated",
	"customer.source.deleted",
}

type Card struct {
	name      string
	apiClient api.Client
	objs      chan api.Object
	msgs      chan source.SetMessage
	errs      chan integration.CollectionError
	dedupe    dedupe.Interface
}

func (r *Card) DesiredObjects() []string {
	return []string{"customer", "charge"}
}

func (r *Card) DesiredEvents() []string {
	return append(cardEvents, chargeEvents...)
}

func (r *Card) StartProducer(ctx context.Context, runContext integration.RunContext) error {
	// downloading events in incremental mode is handled by the bundle that this resource is a part of
	close(r.objs)
	close(r.errs)
	return nil
}

func (r *Card) GetEventProcessors() []downloader.PostProcessor {
	return []downloader.PostProcessor{
		processors.NewIsDeleted("customer.source.deleted"),
	}
}

func (r *Card) StartConsumer(ctx context.Context, ch <-chan api.Object) {
	defer close(r.msgs)
	for obj := range ch {
		switch tr.GetString(obj, "object") {
		case "event":
			if payload := tr.ExtractEventPayload(obj, "card", "charge"); payload != nil {
				switch tr.GetString(payload, "object") {
				case "card":
					r.consumeCard(payload, true)
				case "charge":
					r.consumeCharge(payload, true)
				}
			}
		case "charge":
			r.consumeCharge(obj, false)
		case "customer":
			r.consumeCustomer(obj, false)
		}
	}
}

func (r *Card) consumeCard(obj api.Object, fromEvent bool) {
	if msg := r.transform(obj); msg != nil && !(fromEvent && r.dedupe.SeenBefore(msg.ID)) {
		r.msgs <- *msg
	}
}

func (r *Card) consumeCharge(obj api.Object, fromEvent bool) {
	src := tr.GetMap(obj, "source")
	if src != nil && tr.GetString(src, "object") == "card" {
		r.consumeCard(src, fromEvent)
	}
}

func (r *Card) consumeCustomer(obj api.Object, fromEvent bool) {
	for _, src := range tr.GetMapList(tr.GetMap(obj, "sources"), "data") {
		if tr.GetString(src, "object") == "card" {
			r.consumeCard(src, fromEvent)
		}
	}
}

func (r *Card) transform(obj api.Object) *source.SetMessage {
	var id string
	if id = tr.GetString(obj, "id"); id == "" {
		return nil
	}

	properties := map[string]interface{}{
		"address_city":        obj["address_city"],
		"address_country":     obj["address_country"],
		"address_line1":       obj["address_line1"],
		"address_line1_check": obj["address_line1_check"],
		"address_line2":       obj["address_line2"],
		"address_state":       obj["address_state"],
		"address_zip":         obj["address_zip"],
		"address_zip_check":   obj["address_zip_check"],
		"brand":               obj["brand"],
		"country":             obj["country"],
		"cvc_check":           obj["cvc_check"],
		"exp_month":           obj["exp_month"],
		"exp_year":            obj["exp_year"],
		"funding":             obj["funding"],
		"name":                obj["name"],
		"last4":               obj["last4"],
		"dynamic_last4":       obj["dynamic_last4"],
		"fingerprint":         obj["fingerprint"],
		"tokenization_method": obj["tokenization_method"],
	}

	if obj["customer"] != nil {
		properties["customer_id"] = obj["customer"]
	}

	if v, ok := obj["is_deleted"].(bool); ok && v {
		properties["is_deleted"] = obj["is_deleted"]
	}

	tr.Flatten(tr.GetMap(obj, "metadata"), "metadata_", properties)

	return &source.SetMessage{
		ID:         id,
		Collection: r.name,
		Properties: properties,
	}
}

func (r *Card) Collection() string {
	return r.name
}

func (r *Card) Objects() <-chan api.Object {
	return r.objs
}

func (r *Card) Messages() <-chan source.SetMessage {
	return r.msgs
}

func (r *Card) CollectionErrors() <-chan integration.CollectionError {
	return r.errs
}

func (r *Card) Consumers() []integration.Consumer {
	return []integration.Consumer{r}
}

func (r *Card) Close() {
	r.dedupe.Close()
}

func NewCard(apiClient api.Client) *Card {
	return &Card{
		name:      "cards",
		apiClient: apiClient,
		objs:      make(chan api.Object, 1000),
		msgs:      make(chan source.SetMessage),
		errs:      make(chan integration.CollectionError),
		dedupe:    dedupe.New(),
	}
}
