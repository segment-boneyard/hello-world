package resource

import (
	"context"
	"github.com/segment-sources/stripe/api"
	"github.com/segment-sources/stripe/integration"
	"github.com/segment-sources/stripe/resource/dedupe"
	"github.com/segment-sources/stripe/resource/downloader"
	"github.com/segment-sources/stripe/resource/processors"
	"github.com/segment-sources/stripe/resource/tasks"
	"github.com/segment-sources/stripe/resource/tr"
	"github.com/segmentio/go-source"
)

var invoiceitemEvents = []string{
	"invoiceitem.created",
	"invoiceitem.updated",
	"invoiceitem.deleted",
}

type InvoiceItem struct {
	name      string
	apiClient api.Client
	objs      chan api.Object
	msgs      chan source.SetMessage
	errs      chan integration.CollectionError
	dedupe    dedupe.Interface
}

func (r *InvoiceItem) DesiredObjects() []string {
	return []string{"invoiceitem"}
}

func (r *InvoiceItem) DesiredEvents() []string {
	return invoiceitemEvents
}

func (r *InvoiceItem) StartProducer(ctx context.Context, runContext integration.RunContext) error {
	defer close(r.objs)
	defer close(r.errs)
	var task *downloader.Task
	if runContext.PreviousRunTimestamp.IsZero() {
		task = &downloader.Task{
			Collection: r.name,
			Request: &api.Request{
				Url: "/v1/invoiceitems?limit=100",
			},
			Output: r.objs,
			Errors: r.errs,
		}
	} else {
		task = tasks.MakeIncremental(r, r.name, runContext.PreviousRunTimestamp, r.objs, r.errs)
	}

	return downloader.New(r.apiClient).Do(ctx, task)
}

func (r *InvoiceItem) GetEventProcessors() []downloader.PostProcessor {
	return []downloader.PostProcessor{
		processors.NewIsDeleted("invoiceitem.deleted"),
	}
}

func (r *InvoiceItem) StartConsumer(ctx context.Context, ch <-chan api.Object) {
	defer close(r.msgs)
	for obj := range ch {
		switch tr.GetString(obj, "object") {
		case "event":
			if payload := tr.ExtractEventPayload(obj, "invoiceitem"); payload != nil {
				r.consumeObject(payload, true)
			}
		case "invoiceitem":
			r.consumeObject(obj, false)
		}
	}
}

func (r *InvoiceItem) consumeObject(obj api.Object, fromEvent bool) {
	if msg := r.transform(obj); msg != nil && !(fromEvent && r.dedupe.SeenBefore(msg.ID)) {
		r.msgs <- *msg
	}
}

func (r *InvoiceItem) transform(obj api.Object) *source.SetMessage {
	var id string
	if id = tr.GetString(obj, "id"); id == "" {
		return nil
	}

	properties := map[string]interface{}{
		"amount":          obj["amount"],
		"currency":        obj["currency"],
		"customer_id":     obj["customer"],
		"description":     obj["description"],
		"discountable":    obj["discountable"],
		"invoice_id":      obj["invoice"],
		"proration":       obj["proration"],
		"quantity":        obj["quantity"],
		"subscription_id": obj["subscription"],
	}

	if v, ok := obj["is_deleted"].(bool); ok && v {
		properties["is_deleted"] = v
	}

	tr.Flatten(tr.GetMap(obj, "metadata"), "metadata_", properties)

	if date := tr.GetTimestamp(obj, "date"); date != "" {
		properties["date"] = date
	}
	period := tr.GetMap(obj, "period")
	if periodStart := tr.GetTimestamp(period, "start"); periodStart != "" {
		properties["period_start"] = periodStart
	}
	if periodEnd := tr.GetTimestamp(period, "end"); periodEnd != "" {
		properties["period_end"] = periodEnd
	}

	if plan := tr.GetMap(obj, "plan"); plan != nil {
		if planId := tr.GetString(plan, "id"); planId != "" {
			properties["plan_id"] = planId
		}
	}

	return &source.SetMessage{
		ID:         id,
		Collection: r.name,
		Properties: properties,
	}
}

func (r *InvoiceItem) Collection() string {
	return r.name
}

func (r *InvoiceItem) Objects() <-chan api.Object {
	return r.objs
}

func (r *InvoiceItem) Messages() <-chan source.SetMessage {
	return r.msgs
}

func (r *InvoiceItem) CollectionErrors() <-chan integration.CollectionError {
	return r.errs
}

func (r *InvoiceItem) Consumers() []integration.Consumer {
	return []integration.Consumer{r}
}

func (r *InvoiceItem) Close() {
	r.dedupe.Close()
}

func NewInvoiceItem(apiClient api.Client) *InvoiceItem {
	return &InvoiceItem{
		name:      "invoice_items",
		apiClient: apiClient,
		objs:      make(chan api.Object, 1000),
		msgs:      make(chan source.SetMessage),
		errs:      make(chan integration.CollectionError),
		dedupe:    dedupe.New(),
	}
}
