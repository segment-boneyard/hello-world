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

var applicationFeeEvents = []string{
	"application_fee.created",
	"application_fee.refunded",
}

type ApplicationFee struct {
	name      string
	apiClient api.Client
	objs      chan api.Object
	msgs      chan source.SetMessage
	errs      chan integration.CollectionError
	dedupe    dedupe.Interface
}

func (r *ApplicationFee) DesiredObjects() []string {
	return []string{"application_fee"}
}

func (r *ApplicationFee) DesiredEvents() []string {
	return applicationFeeEvents
}

func (r *ApplicationFee) StartProducer(ctx context.Context, runContext integration.RunContext) error {
	defer close(r.objs)
	defer close(r.errs)
	if runContext.PreviousRunTimestamp.IsZero() {
		return downloader.New(r.apiClient).Do(ctx, &downloader.Task{
			Collection: r.name,
			Request: &api.Request{
				Url:           "/v1/application_fees?limit=100",
				LogCollection: r.name,
			},
			Output: r.objs,
			Errors: r.errs,
			PostProcessors: []downloader.PostProcessor{
				processors.NewListExpander("refunds", r.apiClient),
			},
		})
	}

	// downloading events in incremental mode is handled by the bundle that this resource is a part of
	return nil
}

func (r *ApplicationFee) StartConsumer(ctx context.Context, ch <-chan api.Object) {
	defer close(r.msgs)
	for obj := range ch {
		switch tr.GetString(obj, "object") {
		case "event":
			if payload := tr.ExtractEventPayload(obj, "application_fee"); payload != nil {
				r.consumeFee(payload, true)
			}
		case "application_fee":
			r.consumeFee(obj, false)
		}
	}
}

func (r *ApplicationFee) consumeFee(obj api.Object, fromEvent bool) {
	if msg := r.transform(obj); msg != nil && !(fromEvent && r.dedupe.SeenBefore(msg.ID)) {
		r.msgs <- *msg
	}
}

func (r *ApplicationFee) transform(obj api.Object) *source.SetMessage {
	var id string
	if id = tr.GetString(obj, "id"); id == "" {
		return nil
	}

	properties := map[string]interface{}{
		"account_id":              obj["account"],
		"amount":                  obj["amount"],
		"amount_refunded":         obj["amount_refunded"],
		"application_id":          obj["application"],
		"balance_transaction_id":  obj["balance_transaction"],
		"charge_id":               obj["charge"],
		"currency":                obj["currency"],
		"originating_transaction": obj["originating_transaction"],
		"refunded":                obj["refunded"],
	}

	tr.Flatten(tr.GetMap(obj, "metadata"), "metadata_", properties)

	if created := tr.GetTimestamp(obj, "created"); created != "" {
		properties["created"] = created
	}

	return &source.SetMessage{
		ID:         id,
		Collection: r.name,
		Properties: properties,
	}
}

func (r *ApplicationFee) Collection() string {
	return r.name
}

func (r *ApplicationFee) Objects() <-chan api.Object {
	return r.objs
}

func (r *ApplicationFee) Messages() <-chan source.SetMessage {
	return r.msgs
}

func (r *ApplicationFee) CollectionErrors() <-chan integration.CollectionError {
	return r.errs
}

func (r *ApplicationFee) Consumers() []integration.Consumer {
	return []integration.Consumer{r}
}

func (r *ApplicationFee) Close() {
	r.dedupe.Close()
}

func NewApplicationFee(apiClient api.Client) *ApplicationFee {
	return &ApplicationFee{
		name:      "application_fees",
		apiClient: apiClient,
		objs:      make(chan api.Object, 1000),
		msgs:      make(chan source.SetMessage),
		errs:      make(chan integration.CollectionError),
		dedupe:    dedupe.New(),
	}
}
