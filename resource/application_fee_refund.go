package resource

import (
	"context"
	"github.com/segment-sources/stripe/api"
	"github.com/segment-sources/stripe/integration"
	"github.com/segment-sources/stripe/resource/dedupe"
	"github.com/segment-sources/stripe/resource/tr"
	"github.com/segmentio/go-source"
)

var feeRefundEvents = []string{
	"application_fee.refund.updated",
}

type ApplicationFeeRefund struct {
	name      string
	apiClient api.Client
	objs      chan api.Object
	msgs      chan source.SetMessage
	errs      chan integration.CollectionError
	dedupe    dedupe.Interface
}

func (r *ApplicationFeeRefund) DesiredObjects() []string {
	return []string{"application_fee"}
}

func (r *ApplicationFeeRefund) DesiredEvents() []string {
	return append(feeRefundEvents, applicationFeeEvents...)
}

func (r *ApplicationFeeRefund) StartProducer(ctx context.Context, runContext integration.RunContext) error {
	// downloading events in incremental mode is handled by the bundle that this resource is a part of
	close(r.objs)
	close(r.errs)
	return nil
}

func (r *ApplicationFeeRefund) StartConsumer(ctx context.Context, ch <-chan api.Object) {
	defer close(r.msgs)
	for obj := range ch {
		switch tr.GetString(obj, "object") {
		case "event":
			if payload := tr.ExtractEventPayload(obj, "fee_refund", "application_fee"); payload != nil {
				switch tr.GetString(payload, "object") {
				case "fee_refund":
					r.consumeFeeRefund(payload, true)
				case "application_fee":
					r.consumeFee(payload, true)
				}
			}
		case "application_fee":
			r.consumeFee(obj, false)
		}
	}
}

func (r *ApplicationFeeRefund) consumeFeeRefund(obj api.Object, fromEvent bool) {
	if msg := r.transform(obj); msg != nil && !(fromEvent && r.dedupe.SeenBefore(msg.ID)) {
		r.msgs <- *msg
	}
}

func (r *ApplicationFeeRefund) consumeFee(obj api.Object, fromEvent bool) {
	for _, refund := range tr.GetMapList(tr.GetMap(obj, "refunds"), "data") {
		r.consumeFeeRefund(refund, fromEvent)
	}
}

func (r *ApplicationFeeRefund) transform(obj api.Object) *source.SetMessage {
	var id string
	if id = tr.GetString(obj, "id"); id == "" {
		return nil
	}

	properties := map[string]interface{}{
		"amount":                 obj["amount"],
		"currency":               obj["currency"],
		"fee_id":                 obj["fee"],
	}

	if metadata := tr.GetMap(obj, "metadata"); metadata != nil {
		tr.Flatten(metadata, "metadata_", properties)
	}

	if created := tr.GetTimestamp(obj, "created"); created != "" {
		properties["created"] = created
	}

	return &source.SetMessage{
		ID:         id,
		Collection: r.name,
		Properties: properties,
	}
}

func (r *ApplicationFeeRefund) Collection() string {
	return r.name
}

func (r *ApplicationFeeRefund) Objects() <-chan api.Object {
	return r.objs
}

func (r *ApplicationFeeRefund) Messages() <-chan source.SetMessage {
	return r.msgs
}

func (r *ApplicationFeeRefund) CollectionErrors() <-chan integration.CollectionError {
	return r.errs
}

func (r *ApplicationFeeRefund) Consumers() []integration.Consumer {
	return []integration.Consumer{r}
}

func (r *ApplicationFeeRefund) Close() {
	r.dedupe.Close()
}

func NewApplicationFeeRefund(apiClient api.Client) *ApplicationFeeRefund {
	return &ApplicationFeeRefund{
		name:      "application_fee_refunds",
		apiClient: apiClient,
		objs:      make(chan api.Object, 1000),
		msgs:      make(chan source.SetMessage),
		errs:      make(chan integration.CollectionError),
		dedupe:    dedupe.New(),
	}
}
