package resource

import (
	"context"
	"github.com/segment-sources/stripe/api"
	"github.com/segment-sources/stripe/integration"
	"github.com/segment-sources/stripe/resource/dedupe"
	"github.com/segment-sources/stripe/resource/tr"
	"github.com/segmentio/go-source"
	"strings"
)

var discountEvents = []string{
	"customer.discount.created",
	"customer.discount.updated",
	"customer.discount.deleted",
}

type Discount struct {
	name      string
	apiClient api.Client
	objs      chan api.Object
	msgs      chan source.SetMessage
	errs      chan integration.CollectionError
	dedupe    dedupe.Interface
}

func (r *Discount) DesiredObjects() []string {
	return []string{"customer", "invoice", "subscription"}
}

func (r *Discount) DesiredEvents() []string {
	allEventTypes := discountEvents
	allEventTypes = append(allEventTypes, invoiceEvents...)
	allEventTypes = append(allEventTypes, subscriptionEvents...)
	return allEventTypes
}

func (r *Discount) StartProducer(ctx context.Context, runContext integration.RunContext) error {
	// downloading events in incremental mode is handled by the bundle that this resource is a part of
	close(r.objs)
	close(r.errs)
	return nil
}

func (r *Discount) StartConsumer(ctx context.Context, ch <-chan api.Object) {
	defer close(r.msgs)
	for obj := range ch {
		switch tr.GetString(obj, "object") {
		case "event":
			if payload := tr.ExtractEventPayload(obj, "discount", "invoice", "subscription"); payload != nil {
				switch tr.GetString(payload, "object") {
				case "discount":
					r.consumeDiscount(payload, true)
				case "invoice":
					r.consumeParent(payload, true)
				case "subscription":
					r.consumeParent(payload, true)
				}
			}
		case "customer":
			r.consumeParent(obj, false)
		case "invoice":
			r.consumeParent(obj, false)
		case "subscription":
			r.consumeParent(obj, false)
		}
	}
}

func (r *Discount) consumeDiscount(obj api.Object, fromEvent bool) {
	if msg := r.transform(obj); msg != nil && !(fromEvent && r.dedupe.SeenBefore(msg.ID)) {
		r.msgs <- *msg
	}
}

func (r *Discount) consumeParent(obj api.Object, fromEvent bool) {
	if discount := tr.GetMap(obj, "discount"); discount != nil {
		r.consumeDiscount(discount, fromEvent)
	}
}

func (r *Discount) transform(obj api.Object) *source.SetMessage {
	var id string
	if id = makeDiscountId(obj); id == "" {
		return nil
	}

	properties := map[string]interface{}{
		"customer_id":  obj["customer"],
		"subscription": obj["subscription"],
	}

	if v := tr.GetTimestamp(obj, "start"); v != "" {
		properties["start"] = v
	}

	if v := tr.GetTimestamp(obj, "end"); v != "" {
		properties["end"] = v
	}

	if m := tr.GetMap(obj, "coupon"); m != nil {
		properties["coupon_id"] = m["id"]
	}

	return &source.SetMessage{
		ID:         id,
		Collection: r.name,
		Properties: properties,
	}
}

func (r *Discount) Collection() string {
	return r.name
}

func (r *Discount) Objects() <-chan api.Object {
	return r.objs
}

func (r *Discount) Messages() <-chan source.SetMessage {
	return r.msgs
}

func (r *Discount) CollectionErrors() <-chan integration.CollectionError {
	return r.errs
}

func (r *Discount) Consumers() []integration.Consumer {
	return []integration.Consumer{r}
}

func (r *Discount) Close() {
	r.dedupe.Close()
}

func NewDiscount(apiClient api.Client) *Discount {
	return &Discount{
		name:      "discounts",
		apiClient: apiClient,
		objs:      make(chan api.Object, 1000),
		msgs:      make(chan source.SetMessage),
		errs:      make(chan integration.CollectionError),
		dedupe:    dedupe.New(),
	}
}

func makeDiscountId(discountObj map[string]interface{}) string {
	var customerId string
	if customerId = tr.GetString(discountObj, "customer"); customerId == "" {
		return ""
	}

	var coupon map[string]interface{}
	if coupon = tr.GetMap(discountObj, "coupon"); coupon == nil {
		return ""
	}

	var couponId string
	if couponId = tr.GetString(coupon, "id"); couponId == "" {
		return ""
	}

	return strings.Join([]string{customerId, couponId}, "_")
}
