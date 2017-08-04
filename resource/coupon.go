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

var couponEvents = []string{
	"coupon.created",
	"coupon.updated",
	"coupon.deleted",
}

type Coupon struct {
	name      string
	apiClient api.Client
	objs      chan api.Object
	msgs      chan source.SetMessage
	errs      chan integration.CollectionError
	dedupe    dedupe.Interface
}

func (r *Coupon) DesiredObjects() []string {
	return []string{"coupon"}
}

func (r *Coupon) DesiredEvents() []string {
	allEventTypes := couponEvents
	allEventTypes = append(allEventTypes, invoiceEvents...)
	allEventTypes = append(allEventTypes, subscriptionEvents...)
	allEventTypes = append(allEventTypes, discountEvents...)
	return allEventTypes
}

func (r *Coupon) StartProducer(ctx context.Context, runContext integration.RunContext) error {
	defer close(r.objs)
	defer close(r.errs)
	if runContext.PreviousRunTimestamp.IsZero() {
		return downloader.New(r.apiClient).Do(ctx, &downloader.Task{
			Collection: r.name,
			Request: &api.Request{
				Url:           "/v1/coupons?limit=100",
				LogCollection: r.name,
			},
			Output: r.objs,
			Errors: r.errs,
		})
	}

	// downloading events in incremental mode is handled by the bundle that this resource is a part of
	return nil
}

func (r *Coupon) GetEventProcessors() []downloader.PostProcessor {
	return []downloader.PostProcessor{
		processors.NewIsDeleted("coupon.deleted"),
	}
}

func (r *Coupon) StartConsumer(ctx context.Context, ch <-chan api.Object) {
	defer close(r.msgs)
	for obj := range ch {
		switch tr.GetString(obj, "object") {
		case "event":
			if payload := tr.ExtractEventPayload(obj, "coupon", "discount", "invoice", "subscription"); payload != nil {
				switch tr.GetString(payload, "object") {
				case "coupon":
					r.consumeCoupon(payload, true)
				case "discount":
					r.consumeDiscount(payload, true)
				case "invoice":
					r.consumeDiscountParent(payload, true)
				case "subscription":
					r.consumeDiscountParent(payload, true)
				}
			}
		case "coupon":
			r.consumeCoupon(obj, false)
		}
	}
}

func (r *Coupon) consumeCoupon(obj api.Object, fromEvent bool) {
	if msg := r.transform(obj); msg != nil && !(fromEvent && r.dedupe.SeenBefore(msg.ID)) {
		r.msgs <- *msg
	}
}

func (r *Coupon) consumeDiscount(obj api.Object, fromEvent bool) {
	if coupon := tr.GetMap(obj, "coupon"); coupon != nil {
		r.consumeCoupon(coupon, fromEvent)
	}
}

func (r *Coupon) consumeDiscountParent(obj api.Object, fromEvent bool) {
	if discount := tr.GetMap(obj, "discount"); discount != nil {
		r.consumeDiscount(discount, fromEvent)
	}
}

func (r *Coupon) transform(obj api.Object) *source.SetMessage {
	var id string
	if id = tr.GetString(obj, "id"); id == "" {
		return nil
	}

	properties := map[string]interface{}{
		"percent_off":        obj["percent_off"],
		"amount_off":         obj["amount_off"],
		"currency":           obj["currency"],
		"duration":           obj["duration"],
		"max_redemptions":    obj["max_redemptions"],
		"times_redeemed":     obj["times_redeemed"],
		"valid":              obj["valid"],
		"duration_in_months": obj["duration_in_months"],
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

	if ts := tr.GetTimestamp(obj, "redeem_by"); ts != "" {
		properties["redeem_by"] = ts
	}

	return &source.SetMessage{
		ID:         id,
		Collection: r.name,
		Properties: properties,
	}
}

func (r *Coupon) Collection() string {
	return r.name
}

func (r *Coupon) Objects() <-chan api.Object {
	return r.objs
}

func (r *Coupon) Messages() <-chan source.SetMessage {
	return r.msgs
}

func (r *Coupon) CollectionErrors() <-chan integration.CollectionError {
	return r.errs
}

func (r *Coupon) Consumers() []integration.Consumer {
	return []integration.Consumer{r}
}

func (r *Coupon) Close() {
	r.dedupe.Close()
}

func NewCoupon(apiClient api.Client) *Coupon {
	return &Coupon{
		name:      "coupons",
		apiClient: apiClient,
		objs:      make(chan api.Object, 1000),
		msgs:      make(chan source.SetMessage),
		errs:      make(chan integration.CollectionError),
		dedupe:    dedupe.New(),
	}
}
