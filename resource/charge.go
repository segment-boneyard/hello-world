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

var chargeEvents = []string{
	"charge.captured",
	"charge.failed",
	"charge.pending",
	"charge.refunded",
	"charge.succeeded",
	"charge.updated",
}

type Charge struct {
	name      string
	apiClient api.Client
	objs      chan api.Object
	msgs      chan source.SetMessage
	errs      chan integration.CollectionError
	dedupe    dedupe.Interface
}

func (r *Charge) DesiredObjects() []string {
	return []string{"charge"}
}

func (r *Charge) DesiredEvents() []string {
	return chargeEvents
}

func (r *Charge) StartProducer(ctx context.Context, runContext integration.RunContext) error {
	defer close(r.objs)
	defer close(r.errs)
	if runContext.PreviousRunTimestamp.IsZero() {
		return downloader.New(r.apiClient).Do(ctx, &downloader.Task{
			Collection: r.name,
			Request: &api.Request{
				Url:           "/v1/charges?limit=100",
				LogCollection: r.name,
			},
			Output: r.objs,
			Errors: r.errs,
		})
	}

	// downloading events in incremental mode is handled by the bundle that this resource is a part of
	return nil
}

func (r *Charge) StartConsumer(ctx context.Context, ch <-chan api.Object) {
	defer close(r.msgs)
	for obj := range ch {
		switch tr.GetString(obj, "object") {
		case "event":
			if payload := tr.ExtractEventPayload(obj, "charge"); payload != nil {
				r.consumeCharge(payload, true)
			}
		case "charge":
			r.consumeCharge(obj, false)
		}
	}
}

func (r *Charge) consumeCharge(obj api.Object, fromEvent bool) {
	if msg := r.transform(obj); msg != nil && !(fromEvent && r.dedupe.SeenBefore(msg.ID)) {
		r.msgs <- *msg
	}
}

func (r *Charge) transform(obj api.Object) *source.SetMessage {
	var id string
	if id = tr.GetString(obj, "id"); id == "" {
		return nil
	}

	properties := map[string]interface{}{
		"amount":                 obj["amount"],
		"amount_refunded":        obj["amount_refunded"],
		"application_fee":        obj["application_fee"],
		"balance_transaction_id": obj["balance_transaction"],
		"captured":               obj["captured"],
		"currency":               obj["currency"],
		"customer_id":            obj["customer"],
		"description":            obj["description"],
		"destination":            obj["destination"],
		"failure_code":           obj["failure_code"],
		"failure_message":        obj["failure_message"],
		"invoice_id":             obj["invoice"],
		"paid":                   obj["paid"],
		"receipt_email":          obj["receipt_email"],
		"receipt_number":         obj["receipt_number"],
		"refunded":               obj["refunded"],
		"statement_descriptor":   obj["statement_descriptor"],
		"status":                 obj["status"],
	}

	tr.Flatten(tr.GetMap(obj, "metadata"), "metadata_", properties)
	tr.Flatten(tr.GetMap(obj, "fraud_details"), "fraud_details_", properties)
	tr.Flatten(tr.GetMap(obj, "shipping"), "shipping_", properties)

	if created := tr.GetTimestamp(obj, "created"); created != "" {
		properties["created"] = created
	}

	if src := tr.GetMap(obj, "source"); src != nil {
		srcId := tr.GetString(src, "id")
		if srcType := tr.GetString(src, "object"); srcType == "card" {
			properties["card_id"] = srcId
		} else if srcType == "bank_account" {
			properties["bank_account_id"] = srcId
		}
	}

	return &source.SetMessage{
		ID:         id,
		Collection: r.name,
		Properties: properties,
	}
}

func (r *Charge) Collection() string {
	return r.name
}

func (r *Charge) Objects() <-chan api.Object {
	return r.objs
}

func (r *Charge) Messages() <-chan source.SetMessage {
	return r.msgs
}

func (r *Charge) CollectionErrors() <-chan integration.CollectionError {
	return r.errs
}

func (r *Charge) Consumers() []integration.Consumer {
	return []integration.Consumer{r}
}

func (r *Charge) Close() {
	r.dedupe.Close()
}

func NewCharge(apiClient api.Client) *Charge {
	return &Charge{
		name:      "charges",
		apiClient: apiClient,
		objs:      make(chan api.Object, 1000),
		msgs:      make(chan source.SetMessage),
		errs:      make(chan integration.CollectionError),
		dedupe:    dedupe.New(),
	}
}
