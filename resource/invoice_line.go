package resource

import (
	"context"
	"crypto/md5"
	"fmt"
	"github.com/segment-sources/stripe/api"
	"github.com/segment-sources/stripe/integration"
	"github.com/segment-sources/stripe/resource/dedupe"
	"github.com/segment-sources/stripe/resource/tr"
	"github.com/segmentio/go-source"
	"strings"
	"time"
)

const jsTimeFormat = "Mon Jan 02 2006 15:04:05 GMT-0700 (MST)"

type InvoiceLine struct {
	name      string
	apiClient api.Client
	objs      chan api.Object
	msgs      chan source.SetMessage
	errs      chan integration.CollectionError
	dedupe    dedupe.Interface
}

func (r *InvoiceLine) DesiredObjects() []string {
	return []string{"invoice"}
}

func (r *InvoiceLine) DesiredEvents() []string {
	return invoiceEvents
}

func (r *InvoiceLine) StartProducer(ctx context.Context, runContext integration.RunContext) error {
	// downloading events in incremental mode is handled by the bundle that this resource is a part of
	close(r.objs)
	close(r.errs)
	return nil
}

func (r *InvoiceLine) StartConsumer(ctx context.Context, ch <-chan api.Object) {
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

func (r *InvoiceLine) consumeInvoice(obj api.Object, fromEvent bool) {
	var invoiceId string
	if invoiceId = tr.GetString(obj, "id"); invoiceId == "" || fromEvent && r.dedupe.SeenBefore(invoiceId) {
		return
	}

	for _, line := range tr.GetMapList(tr.GetMap(obj, "lines"), "data") {
		if msg := r.transform(obj, line); msg != nil {
			r.msgs <- *msg
		}
	}
}

func (r *InvoiceLine) transform(invoice, line api.Object) *source.SetMessage {
	var invoiceId string
	if invoiceId = tr.GetString(invoice, "id"); invoiceId == "" {
		return nil
	}

	var id string
	if id = tr.GetString(line, "id"); id == "" {
		return nil
	}

	internalIdComponents := []string{id}

	if lineType := tr.GetString(line, "type"); lineType != "" {
		internalIdComponents = append(internalIdComponents, lineType)
	}

	properties := map[string]interface{}{
		"amount":          line["amount"],
		"currency":        line["currency"],
		"description":     line["description"],
		"discountable":    line["discountable"],
		"proration":       line["proration"],
		"quantity":        line["quantity"],
		"subscription_id": line["subscription"],
		"type":            line["type"],
		"invoice_id":      invoiceId,
	}

	period := tr.GetMap(line, "period")
	if periodStart := tr.GetTimestamp(period, "start"); periodStart != "" {
		properties["period_start"] = periodStart
		ts := time.Unix(tr.GetNumber(period, "start"), 0)
		internalIdComponents = append(internalIdComponents, ts.Format(jsTimeFormat))
	}
	if periodEnd := tr.GetTimestamp(period, "end"); periodEnd != "" {
		properties["period_end"] = periodEnd
		ts := time.Unix(tr.GetNumber(period, "end"), 0)
		internalIdComponents = append(internalIdComponents, ts.Format(jsTimeFormat))
	}

	if plan := tr.GetMap(line, "plan"); plan != nil {
		if planId := tr.GetString(plan, "id"); planId != "" {
			properties["plan_id"] = planId
		}
	}

	if strings.HasPrefix(id, "sub_") {
		properties["subscription_id"] = id
	} else if strings.HasPrefix(id, "ii_") {
		properties["item_id"] = id
	}

	hash := md5.New()
	fmt.Fprint(hash, strings.Join(internalIdComponents, ", "))

	return &source.SetMessage{
		ID:         fmt.Sprintf("%x", hash.Sum(nil)),
		Collection: r.name,
		Properties: properties,
	}
}

func (r *InvoiceLine) Collection() string {
	return r.name
}

func (r *InvoiceLine) Messages() <-chan source.SetMessage {
	return r.msgs
}

func (r *InvoiceLine) CollectionErrors() <-chan integration.CollectionError {
	return r.errs
}

func (r *InvoiceLine) Objects() <-chan api.Object {
	return r.objs
}

func (r *InvoiceLine) Consumers() []integration.Consumer {
	return []integration.Consumer{r}
}

func (r *InvoiceLine) Close() {
	r.dedupe.Close()
}

func NewInvoiceLine(apiClient api.Client) *InvoiceLine {
	return &InvoiceLine{
		name:      "invoice_lines",
		apiClient: apiClient,
		objs:      make(chan api.Object, 1000),
		msgs:      make(chan source.SetMessage),
		errs:      make(chan integration.CollectionError),
		dedupe:    dedupe.New(),
	}
}
