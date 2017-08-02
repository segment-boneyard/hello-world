package resource

import (
	"context"
	"github.com/segment-sources/stripe/api"
	"github.com/segment-sources/stripe/integration"
	"github.com/segment-sources/stripe/resource/dedupe"
	"github.com/segment-sources/stripe/resource/tr"
	"github.com/segmentio/go-source"
)

type TransferReversal struct {
	name      string
	apiClient api.Client
	objs      chan api.Object
	msgs      chan source.SetMessage
	errs      chan integration.CollectionError
	dedupe    dedupe.Interface
}

func (r *TransferReversal) DesiredObjects() []string {
	return []string{"transfer"}
}

func (r *TransferReversal) DesiredEvents() []string {
	return transferEvents
}

func (r *TransferReversal) StartProducer(ctx context.Context, runContext integration.RunContext) error {
	// downloading events in incremental mode is handled by the bundle that this resource is a part of
	close(r.objs)
	close(r.errs)
	return nil
}

func (r *TransferReversal) StartConsumer(ctx context.Context, ch <-chan api.Object) {
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

func (r *TransferReversal) consumeTransfer(obj api.Object, fromEvent bool) {
	var transferId string
	if transferId = tr.GetString(obj, "id"); transferId == "" || fromEvent && r.dedupe.SeenBefore(transferId) {
		return
	}

	for _, reversal := range tr.GetMapList(tr.GetMap(obj, "reversals"), "data") {
		if msg := r.transform(reversal); msg != nil {
			r.msgs <- *msg
		}
	}
}

func (r *TransferReversal) transform(obj api.Object) *source.SetMessage {
	var id string
	if id = tr.GetString(obj, "id"); id == "" {
		return nil
	}

	properties := map[string]interface{}{
		"amount":                 obj["amount"],
		"currency":               obj["currency"],
		"balance_transaction_id": obj["balance_transaction"],
		"transfer_id":            obj["transfer"],
	}

	tr.Flatten(tr.GetMap(obj, "metadata"), "metadata_", properties)
	if ts := tr.GetTimestamp(obj, "created"); ts != "" {
		properties["created"] = ts
	}

	return &source.SetMessage{
		ID:         id,
		Collection: r.name,
		Properties: properties,
	}
}

func (r *TransferReversal) Collection() string {
	return r.name
}

func (r *TransferReversal) Objects() <-chan api.Object {
	return r.objs
}

func (r *TransferReversal) Messages() <-chan source.SetMessage {
	return r.msgs
}

func (r *TransferReversal) CollectionErrors() <-chan integration.CollectionError {
	return r.errs
}

func (r *TransferReversal) Consumers() []integration.Consumer {
	return []integration.Consumer{r}
}

func (r *TransferReversal) Close() {
	r.dedupe.Close()
}

func NewTransferReversal(apiClient api.Client) *TransferReversal {
	return &TransferReversal{
		name:      "transfer_reversals",
		apiClient: apiClient,
		objs:      make(chan api.Object, 1000),
		msgs:      make(chan source.SetMessage),
		errs:      make(chan integration.CollectionError),
		dedupe:    dedupe.New(),
	}
}
