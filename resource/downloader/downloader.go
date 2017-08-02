package downloader

import (
	"context"
	"github.com/apex/log"
	"github.com/segment-sources/stripe/api"
	"github.com/segment-sources/stripe/integration"
	"github.com/segmentio/ur-log"
	"net/url"
	"reflect"
)

// Client can download multiple subsequent pages of objects from Stripe API
// and send them to the output channel specified in a task
type Client struct {
	ApiClient api.Client
}

func (d *Client) Do(ctx context.Context, task *Task) error {
	first := task.Request

	if task.Collection != "" {
		ctx, _ = urlog.GetContextualLogger(ctx, nil, log.Fields{
			"collection": task.Collection,
		})
	}

	for next := first; next != nil; {
		req := next

		res, err := RetryGetList(ctx, d.ApiClient, req)
		if err != nil {
			if task.Collection != "" && task.Errors != nil {
				task.Errors <- integration.CollectionError{
					Collection: task.Collection,
					Message:    "HTTP request failed",
				}
			}
			return urlog.WrapError(ctx, err, "failed to fetch object list")
		}

		lastSeenId := ""
		for _, obj := range res.Objects {
			for _, p := range task.PostProcessors {
				procCtx, _ := urlog.GetContextualLogger(ctx, nil, log.Fields{
					"processor": reflect.TypeOf(p).String(),
				})
				if err := p(procCtx, obj); err != nil {
					if task.Collection != "" && task.Errors != nil {
						task.Errors <- integration.CollectionError{
							Collection: task.Collection,
							Message:    "HTTP request failed",
						}
					}
					return urlog.WrapError(ctx, err, "processor failed")
				}
			}

			task.Output <- obj
			if id, ok := obj["id"].(string); ok {
				lastSeenId = id
			}
		}

		if res.HasMore && lastSeenId != "" {
			next = &api.Request{
				Url:     req.Url,
				Qs:      url.Values{},
				Headers: req.Headers,
			}
			for key, value := range req.Qs {
				next.Qs[key] = value
			}
			next.Qs.Set("starting_after", lastSeenId)
		} else {
			next = nil
		}
	}

	return nil
}

func New(apiClient api.Client) *Client {
	return &Client{ApiClient: apiClient}
}
