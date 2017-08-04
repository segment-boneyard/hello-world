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

var transferEvents = []string{
	"transfer.created",
	"transfer.updated",
	"transfer.reversed",
}

type Transfer struct {
	name              string
	apiClient         api.Client
	objs              chan api.Object
	msgs              chan source.SetMessage
	errs              chan integration.CollectionError
	dedupe            dedupe.Interface
	processorDedupe   dedupe.Interface
	enableTransferIds bool
}

func (r *Transfer) DesiredObjects() []string {
	return []string{"transfer"}
}

func (r *Transfer) DesiredEvents() []string {
	return transferEvents
}

func (r *Transfer) StartProducer(ctx context.Context, runContext integration.RunContext) error {
	defer close(r.objs)
	defer close(r.errs)
	if runContext.PreviousRunTimestamp.IsZero() {
		postProcessors := []downloader.PostProcessor{
			processors.NewListExpander("reversals", r.apiClient),
		}

		if r.enableTransferIds {
			postProcessors = append(postProcessors, processors.NewRelatedTransactions(r.apiClient))
		}

		return downloader.New(r.apiClient).Do(ctx, &downloader.Task{
			Collection: r.name,
			Request: &api.Request{
				Url:           "/v1/transfers?limit=100",
				LogCollection: r.name,
			},
			PostProcessors: postProcessors,
			Output:         r.objs,
			Errors:         r.errs,
		})
	}

	// downloading events in incremental mode is handled by the bundle that this resource is a part of
	return nil
}

func (r *Transfer) StartConsumer(ctx context.Context, ch <-chan api.Object) {
	defer close(r.msgs)
	for obj := range ch {
		switch tr.GetString(obj, "object") {
		case "event":
			if payload := tr.ExtractEventPayload(obj, "transfer"); payload != nil {
				r.consumeTransfer(payload, true)
			}
		case "transfer":
			r.consumeTransfer(obj, false)
		}
	}
}

func (r *Transfer) consumeTransfer(obj api.Object, fromEvent bool) {
	if msg := r.transform(obj); msg != nil && !(fromEvent && r.dedupe.SeenBefore(msg.ID)) {
		r.msgs <- *msg
	}
}

func (r *Transfer) transform(obj api.Object) *source.SetMessage {
	var id string
	if id = tr.GetString(obj, "id"); id == "" {
		return nil
	}

	properties := map[string]interface{}{
		"amount":                 obj["amount"],
		"amount_reversed":        obj["amount_reversed"],
		"application_fee":        obj["application_fee"],
		"balance_transaction_id": obj["balance_transaction"],
		"currency":               obj["currency"],
		"description":            obj["description"],
		"destination_id":         obj["destination"],
		"destination_payment":    obj["destination_payment"],
		"failure_code":           obj["failure_code"],
		"failure_message":        obj["failure_message"],
		"reversed":               obj["reversed"],
		"source_transaction":     obj["source_transaction"],
		"statement_descriptor":   obj["statement_descriptor"],
		"status":                 obj["status"],
		"type":                   obj["type"],
	}

	if bankAccount := tr.GetMap(obj, "bank_account"); bankAccount != nil {
		if bankAccountId := tr.GetString(bankAccount, "id"); bankAccountId != "" {
			properties["bank_account_id"] = bankAccountId
		}
	}

	tr.Flatten(tr.GetMap(obj, "metadata"), "metadata_", properties)
	if ts := tr.GetTimestamp(obj, "created"); ts != "" {
		properties["created"] = ts
	}
	if ts := tr.GetTimestamp(obj, "date"); ts != "" {
		properties["date"] = ts
	}

	return &source.SetMessage{
		ID:         id,
		Collection: r.name,
		Properties: properties,
	}
}

func (r *Transfer) Collection() string {
	return r.name
}

func (r *Transfer) Objects() <-chan api.Object {
	return r.objs
}

func (r *Transfer) Messages() <-chan source.SetMessage {
	return r.msgs
}

func (r *Transfer) CollectionErrors() <-chan integration.CollectionError {
	return r.errs
}

func (r *Transfer) Consumers() []integration.Consumer {
	return []integration.Consumer{r}
}

func (r *Transfer) GetEventProcessors() []downloader.PostProcessor {
	if r.enableTransferIds {
		return []downloader.PostProcessor{processors.NewRelatedTransactionsFromEvents(r.apiClient, r.processorDedupe)}
	}

	return nil
}

func (r *Transfer) Close() {
	r.dedupe.Close()
	r.processorDedupe.Close()
}

func NewTransfer(apiClient api.Client, enableTransferIds bool) *Transfer {
	return &Transfer{
		name:              "transfers",
		apiClient:         apiClient,
		objs:              make(chan api.Object, 1000),
		msgs:              make(chan source.SetMessage),
		errs:              make(chan integration.CollectionError),
		dedupe:            dedupe.New(),
		processorDedupe:   dedupe.New(),
		enableTransferIds: enableTransferIds,
	}
}
