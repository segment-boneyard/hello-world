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

var invoiceEvents = []string{
	"invoice.created",
	"invoice.payment_failed",
	"invoice.payment_succeeded",
	"invoice.updated",
}

type Invoice struct {
	name      string
	apiClient api.Client
	objs      chan api.Object
	msgs      chan source.SetMessage
	errs      chan integration.CollectionError
	dedupe    dedupe.Interface
}

func (r *Invoice) DesiredObjects() []string {
	return []string{"invoice"}
}

func (r *Invoice) DesiredEvents() []string {
	return invoiceEvents
}

func (r *Invoice) StartProducer(ctx context.Context, runContext integration.RunContext) error {
	defer close(r.objs)
	defer close(r.errs)
	if runContext.PreviousRunTimestamp.IsZero() {
		return downloader.New(r.apiClient).Do(ctx, &downloader.Task{
			Collection: r.name,
			Request: &api.Request{
				Url: "/v1/invoices?limit=100",
			},
			PostProcessors: []downloader.PostProcessor{
				processors.NewListExpander("lines", r.apiClient),
			},
			Output: r.objs,
			Errors: r.errs,
		})
	}

	// downloading events in incremental mode is handled by the bundle that this resource is a part of
	return nil
}

func (r *Invoice) StartConsumer(ctx context.Context, ch <-chan api.Object) {
	defer close(r.msgs)
	for obj := range ch {
		switch tr.GetString(obj, "object") {
		case "event":
			if payload := tr.ExtractEventPayload(obj, "invoice"); payload != nil {
				r.consumeInvoice(payload, true)
			}
		case "invoice":
			r.consumeInvoice(obj, false)
		}
	}
}

func (r *Invoice) consumeInvoice(obj api.Object, fromEvent bool) {
	if msg := r.transform(obj); msg != nil && !(fromEvent && r.dedupe.SeenBefore(msg.ID)) {
		r.msgs <- *msg
	}
}

func (r *Invoice) transform(obj api.Object) *source.SetMessage {
	var id string
	if id = tr.GetString(obj, "id"); id == "" {
		return nil
	}

	properties := map[string]interface{}{
		"amount_due":           obj["amount_due"],
		"application_fee":      obj["application_fee"],
		"attempt_count":        obj["attempt_count"],
		"attempted":            obj["attempted"],
		"charge_id":            obj["charge"],
		"closed":               obj["closed"],
		"currency":             obj["currency"],
		"customer_id":          obj["customer"],
		"description":          obj["description"],
		"ending_balance":       obj["ending_balance"],
		"forgiven":             obj["forgiven"],
		"paid":                 obj["paid"],
		"receipt_number":       obj["receipt_number"],
		"starting_balance":     obj["starting_balance"],
		"statement_descriptor": obj["statement_descriptor"],
		"subscription_id":      obj["subscription"],
		"subtotal":             obj["subtotal"],
		"tax":                  obj["tax"],
		"tax_percent":          obj["tax_percent"],
		"total":                obj["total"],
	}

	tr.Flatten(tr.GetMap(obj, "metadata"), "metadata_", properties)

	timestampFields := []string{
		"date",
		"period_start",
		"period_end",
		"next_payment_attempt",
	}
	for _, f := range timestampFields {
		if ts := tr.GetTimestamp(obj, f); ts != "" {
			properties[f] = ts
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

func (r *Invoice) Collection() string {
	return r.name
}

func (r *Invoice) Messages() <-chan source.SetMessage {
	return r.msgs
}

func (r *Invoice) CollectionErrors() <-chan integration.CollectionError {
	return r.errs
}

func (r *Invoice) Objects() <-chan api.Object {
	return r.objs
}

func (r *Invoice) Consumers() []integration.Consumer {
	return []integration.Consumer{r}
}

func (r *Invoice) Close() {
	r.dedupe.Close()
}

func NewInvoice(apiClient api.Client) *Invoice {
	return &Invoice{
		name:      "invoices",
		apiClient: apiClient,
		objs:      make(chan api.Object, 1000),
		msgs:      make(chan source.SetMessage),
		errs:      make(chan integration.CollectionError),
		dedupe:    dedupe.New(),
	}
}
