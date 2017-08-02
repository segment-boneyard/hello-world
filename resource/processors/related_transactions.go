package processors

import (
	"context"
	"github.com/segment-sources/stripe/api"
	"github.com/segment-sources/stripe/resource/dedupe"
	"github.com/segment-sources/stripe/resource/downloader"
	"github.com/segment-sources/stripe/resource/tr"
	"net/url"
	"sync"
)

func NewRelatedTransactions(apiClient api.Client) downloader.PostProcessor {
	d := downloader.New(apiClient)
	return func(ctx context.Context, obj api.Object) error {
		if tr.GetString(obj, "object") != "transfer" {
			return nil
		}
		return includeRelatedTransactions(ctx, d, obj)
	}
}

func NewRelatedTransactionsFromEvents(apiClient api.Client, dd dedupe.Interface) downloader.PostProcessor {
	dl := downloader.New(apiClient)
	return func(ctx context.Context, obj api.Object) error {
		transfer := tr.ExtractEventPayload(obj, "transfer")
		if transfer == nil {
			return nil
		}

		if id := tr.GetString(transfer, "id"); id == "" || dd.SeenBefore(id) {
			return nil
		}

		return includeRelatedTransactions(ctx, dl, transfer)
	}
}

// includeRelatedTransactions is a post-processor that sets "balance_transactions" property on each transfer
func includeRelatedTransactions(ctx context.Context, d *downloader.Client, obj api.Object) error {
	var transferId string
	if transferId = tr.GetString(obj, "id"); transferId == "" {
		return nil
	}

	transactions := []interface{}{}

	ch := make(chan api.Object)
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		for tx := range ch {
			transactions = append(transactions, map[string]interface{}(tx))
		}
	}()

	err := d.Do(ctx, &downloader.Task{
		Request: &api.Request{
			Url: "/v1/balance/history?limit=100",
			Qs: url.Values{
				"transfer": []string{transferId},
			},
		},
		Output: ch,
	})

	close(ch)
	wg.Wait()

	if err == nil {
		obj["balance_transactions"] = transactions
	}
	return err
}
