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

var bankAccountEvents = []string{
	"customer.source.created",
	"customer.source.updated",
	"customer.source.deleted",
	"account.external_account.created",
	"account.external_account.updated",
	"account.external_account.deleted",
}

type BankAccount struct {
	name      string
	apiClient api.Client
	objs      chan api.Object
	msgs      chan source.SetMessage
	errs      chan integration.CollectionError
	dedupe    dedupe.Interface
}

func (r *BankAccount) DesiredObjects() []string {
	return []string{"customer", "charge"}
}

func (r *BankAccount) DesiredEvents() []string {
	return append(bankAccountEvents, chargeEvents...)
}

func (r *BankAccount) StartProducer(ctx context.Context, runContext integration.RunContext) error {
	// downloading events in incremental mode is handled by the bundle that this resource is a part of
	close(r.objs)
	close(r.errs)
	return nil
}

func (r *BankAccount) GetEventProcessors() []downloader.PostProcessor {
	return []downloader.PostProcessor{
		processors.NewIsDeleted("customer.source.deleted", "account.external_account.deleted"),
	}
}

func (r *BankAccount) StartConsumer(ctx context.Context, ch <-chan api.Object) {
	defer close(r.msgs)
	for obj := range ch {
		switch tr.GetString(obj, "object") {
		case "event":
			if payload := tr.ExtractEventPayload(obj, "bank_account", "charge"); payload != nil {
				switch tr.GetString(payload, "object") {
				case "bank_account":
					r.consumeBankAccount(payload, true)
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

func (r *BankAccount) consumeBankAccount(obj api.Object, fromEvent bool) {
	if msg := r.transform(obj); msg != nil && !(fromEvent && r.dedupe.SeenBefore(msg.ID)) {
		r.msgs <- *msg
	}
}

func (r *BankAccount) consumeCharge(obj api.Object, fromEvent bool) {
	src := tr.GetMap(obj, "source")
	if src != nil && tr.GetString(src, "object") == "bank_account" {
		r.consumeBankAccount(src, fromEvent)
	}
}

func (r *BankAccount) consumeCustomer(obj api.Object, fromEvent bool) {
	for _, src := range tr.GetMapList(tr.GetMap(obj, "sources"), "data") {
		if tr.GetString(src, "object") == "bank_account" {
			r.consumeBankAccount(src, fromEvent)
		}
	}
}

func (r *BankAccount) transform(obj api.Object) *source.SetMessage {
	var id string
	if id = tr.GetString(obj, "id"); id == "" {
		return nil
	}

	properties := map[string]interface{}{
		"bank_name":            obj["bank_name"],
		"country":              obj["country"],
		"currency":             obj["currency"],
		"default_for_currency": obj["default_for_currency"],
		"status":               obj["status"],
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

func (r *BankAccount) Collection() string {
	return r.name
}

func (r *BankAccount) Objects() <-chan api.Object {
	return r.objs
}

func (r *BankAccount) Messages() <-chan source.SetMessage {
	return r.msgs
}

func (r *BankAccount) CollectionErrors() <-chan integration.CollectionError {
	return r.errs
}

func (r *BankAccount) Consumers() []integration.Consumer {
	return []integration.Consumer{r}
}

func (r *BankAccount) Close() {
	r.dedupe.Close()
}

func NewBankAccount(apiClient api.Client) *BankAccount {
	return &BankAccount{
		name:      "bank_accounts",
		apiClient: apiClient,
		objs:      make(chan api.Object, 1000),
		msgs:      make(chan source.SetMessage),
		errs:      make(chan integration.CollectionError),
		dedupe:    dedupe.New(),
	}
}
