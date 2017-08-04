package resource

import (
	"context"
	"fmt"
	"github.com/segment-sources/stripe/api"
	"github.com/segment-sources/stripe/integration"
	"github.com/segment-sources/stripe/resource/downloader"
	"github.com/segment-sources/stripe/resource/tr"
	"github.com/segmentio/go-source"
	"net/url"
	"time"
)

type BalanceTransaction struct {
	name              string
	apiClient         api.Client
	objs              chan api.Object
	msgs              chan source.SetMessage
	errs              chan integration.CollectionError
	enableTransferIds bool
}

func (r *BalanceTransaction) DesiredObjects() []string {
	return []string{"balance_transaction", "transfer"}
}

func (r *BalanceTransaction) DesiredEvents() []string {
	return transferEvents
}

func (r *BalanceTransaction) StartProducer(ctx context.Context, runContext integration.RunContext) error {
	defer close(r.objs)
	defer close(r.errs)
	req := &api.Request{
		Url: "/v1/balance/history",
		Qs: url.Values{
			"limit": []string{"100"},
		},
		LogCollection: r.name,
	}

	if !runContext.PreviousRunTimestamp.IsZero() {
		// incremental sync mode - download balance transactions created since the previous sync
		timestampLimit := runContext.PreviousRunTimestamp.Add(-time.Hour)
		req.Qs.Set("created[gt]", fmt.Sprintf("%d", timestampLimit.Unix()))
	} else if r.enableTransferIds {
		// in enableTransferIds full sync mode we only pull balance transactions created in the last 10 days
		// the rest of them will be received via transfers
		timestampLimit := time.Now().UTC().Add(-time.Hour * 24 * 10)
		req.Qs.Set("created[gt]", fmt.Sprintf("%d", timestampLimit.Unix()))
	}

	return downloader.New(r.apiClient).Do(ctx, &downloader.Task{
		Collection: r.name,
		Request:    req,
		Output:     r.objs,
		Errors:     r.errs,
	})
}

func (r *BalanceTransaction) StartConsumer(ctx context.Context, ch <-chan api.Object) {
	defer close(r.msgs)
	for obj := range ch {
		switch tr.GetString(obj, "object") {
		case "event":
			if payload := tr.ExtractEventPayload(obj, "transfer"); payload != nil {
				r.consumeTransfer(payload, true)
			}
		case "balance_transaction":
			r.consumeTransaction(obj, false)
		case "transfer":
			r.consumeTransfer(obj, false)
		}
	}
}

func (r *BalanceTransaction) consumeTransaction(obj api.Object, fromEvent bool) {
	if msg := r.transform(obj); msg != nil {
		r.msgs <- *msg
	}
}

func (r *BalanceTransaction) consumeTransfer(obj api.Object, fromEvent bool) {
	var transferId string
	if transferId = tr.GetString(obj, "id"); transferId == "" {
		return
	}

	for _, obj := range tr.GetMapList(obj, "balance_transactions") {
		newobj := api.Object{}
		for k, v := range obj {
			newobj[k] = v
		}
		newobj["transfer_id"] = transferId
		r.consumeTransaction(newobj, fromEvent)
	}
}

func (r *BalanceTransaction) transform(obj api.Object) *source.SetMessage {
	var id string
	if id = tr.GetString(obj, "id"); id == "" {
		return nil
	}

	properties := map[string]interface{}{
		"amount":      obj["amount"],
		"currency":    obj["currency"],
		"description": obj["description"],
		"fee":         obj["fee"],
		"net":         obj["net"],
		"status":      obj["status"],
		"type":        obj["type"],
		"source":      obj["source"],
	}

	if transferId := tr.GetString(obj, "transfer_id"); transferId != "" {
		properties["transfer_id"] = transferId
	}

	tr.Flatten(tr.GetMap(obj, "metadata"), "metadata_", properties)

	if ts := tr.GetTimestamp(obj, "created"); ts != "" {
		properties["created"] = ts
	}
	if ts := tr.GetTimestamp(obj, "available_on"); ts != "" {
		properties["available"] = ts
	}

	return &source.SetMessage{
		ID:         id,
		Collection: r.name,
		Properties: properties,
	}
}

func (r *BalanceTransaction) Collection() string {
	return r.name
}

func (r *BalanceTransaction) Objects() <-chan api.Object {
	return r.objs
}

func (r *BalanceTransaction) Messages() <-chan source.SetMessage {
	return r.msgs
}

func (r *BalanceTransaction) CollectionErrors() <-chan integration.CollectionError {
	return r.errs
}

func (r *BalanceTransaction) Consumers() []integration.Consumer {
	return []integration.Consumer{r}
}

func (r *BalanceTransaction) Close() {
}

func NewBalanceTransaction(apiClient api.Client, enableTransferIds bool) *BalanceTransaction {
	return &BalanceTransaction{
		name:              "balance_transactions",
		apiClient:         apiClient,
		enableTransferIds: enableTransferIds,
		objs:              make(chan api.Object, 1000),
		msgs:              make(chan source.SetMessage),
		errs:              make(chan integration.CollectionError),
	}
}
