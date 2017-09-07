package processors

import (
	"context"
	"github.com/apex/log"
	"github.com/segment-sources/stripe/api"
	"github.com/segment-sources/stripe/resource/dedupe"
	"github.com/segment-sources/stripe/resource/downloader"
	"github.com/segment-sources/stripe/resource/tr"
	"net/url"
	"sync"
)

func NewRelatedTransactions(apiClient api.Client) downloader.PostProcessor {
	d := downloader.New(apiClient)
	return func(ctx context.Context, obj api.Object, task *downloader.Task) error {
		if tr.GetString(obj, "object") != "transfer" {
			return nil
		}
		return fetchRelatedTransactions(ctx, d, obj, task)
	}
}

func NewRelatedTransactionsFromEvents(apiClient api.Client, dd dedupe.Interface) downloader.PostProcessor {
	dl := downloader.New(apiClient)
	return func(ctx context.Context, obj api.Object, task *downloader.Task) error {
		transfer := tr.ExtractEventPayload(obj, "transfer")
		if transfer == nil {
			return nil
		}

		if id := tr.GetString(transfer, "id"); id == "" || dd.SeenBefore(id) {
			return nil
		}

		return fetchRelatedTransactions(ctx, dl, transfer, task)
	}
}

// fetchRelatedTransactions is a post-processor that downloads balance transactions related to a transfer object,
// sets transfer_id property on the transactions and sends them to the task's output channel
func fetchRelatedTransactions(ctx context.Context, d *downloader.Client, obj api.Object, task *downloader.Task) error {
	var transferId string
	if transferId = tr.GetString(obj, "id"); transferId == "" {
		return nil
	}

	transactionCount := 0

	ch := make(chan api.Object)
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		for tx := range ch {
			tx["transfer_id"] = transferId
			task.Output <- tx
			transactionCount++
		}
	}()

	err := d.Do(ctx, &downloader.Task{
		Request: &api.Request{
			Url: "/v1/balance/history?limit=100",
			Qs: url.Values{
				"transfer": []string{transferId},
			},
			LogCollection: task.Collection,
		},
		Output: ch,
	})

	close(ch)
	wg.Wait()

	if transactionCount > 0 {
		logger := log.WithFields(log.Fields{"transaction_count": transactionCount, "transfer_id": transferId})
		logger.Info("Published additional transactions for a transfer")
	}

	return err
}
