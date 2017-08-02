package resource

import (
	"context"
	"crypto/md5"
	"fmt"
	"github.com/segment-sources/stripe/api"
	"github.com/segment-sources/stripe/integration"
	"github.com/segment-sources/stripe/resource/tr"
	"github.com/segmentio/go-source"
	"strings"
)

type BalanceTransactionFeeDetail struct {
	name      string
	apiClient api.Client
	objs      chan api.Object
	msgs      chan source.SetMessage
	errs      chan integration.CollectionError
}

func (r *BalanceTransactionFeeDetail) DesiredObjects() []string {
	return []string{"balance_transaction"}
}

func (r *BalanceTransactionFeeDetail) DesiredEvents() []string {
	return []string{}
}

func (r *BalanceTransactionFeeDetail) StartProducer(ctx context.Context, runContext integration.RunContext) error {
	// downloading events in incremental mode is handled by the bundle that this resource is a part of
	close(r.objs)
	close(r.errs)
	return nil
}

func (r *BalanceTransactionFeeDetail) StartConsumer(ctx context.Context, ch <-chan api.Object) {
	defer close(r.msgs)
	for obj := range ch {
		for _, feeDetail := range tr.GetMapList(obj, "fee_details") {
			if msg := r.transform(obj, feeDetail); msg != nil {
				r.msgs <- *msg
			}
		}
	}
}

func (r *BalanceTransactionFeeDetail) transform(tx, feeDetail api.Object) *source.SetMessage {
	var txId string
	if txId = tr.GetString(tx, "id"); txId == "" {
		return nil
	}

	properties := map[string]interface{}{
		"balance_transaction_id": txId,
		"amount":                 feeDetail["amount"],
		"application":            feeDetail["application"],
		"currency":               feeDetail["currency"],
		"description":            feeDetail["description"],
		"type":                   feeDetail["type"],
	}

	internalIdComponents := []string{}
	if v := tr.GetNumber(feeDetail, "amount"); v != 0 {
		internalIdComponents = append(internalIdComponents, fmt.Sprintf("%d", v))
	}
	for _, prop := range []string{"currency", "application", "type"} {
		if v := tr.GetString(feeDetail, prop); v != "" {
			internalIdComponents = append(internalIdComponents, v)
		}
	}
	internalIdComponents = append(internalIdComponents, txId)

	hash := md5.New()
	fmt.Fprint(hash, strings.Join(internalIdComponents, ", "))

	return &source.SetMessage{
		ID:         fmt.Sprintf("%x", hash.Sum(nil)),
		Collection: r.name,
		Properties: properties,
	}
}

func (r *BalanceTransactionFeeDetail) Collection() string {
	return r.name
}

func (r *BalanceTransactionFeeDetail) Objects() <-chan api.Object {
	return r.objs
}

func (r *BalanceTransactionFeeDetail) Messages() <-chan source.SetMessage {
	return r.msgs
}

func (r *BalanceTransactionFeeDetail) CollectionErrors() <-chan integration.CollectionError {
	return r.errs
}

func (r *BalanceTransactionFeeDetail) Consumers() []integration.Consumer {
	return []integration.Consumer{r}
}

func (r *BalanceTransactionFeeDetail) Close() {
}

func NewBalanceTransactionFeeDetail(apiClient api.Client) *BalanceTransactionFeeDetail {
	return &BalanceTransactionFeeDetail{
		name:      "balance_transaction_fee_details",
		apiClient: apiClient,
		objs:      make(chan api.Object, 1000),
		msgs:      make(chan source.SetMessage),
		errs:      make(chan integration.CollectionError),
	}
}
