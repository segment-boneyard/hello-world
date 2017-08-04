package processors

import (
	"context"
	"github.com/segment-sources/stripe/api"
	"github.com/segment-sources/stripe/resource/downloader"
	"github.com/segment-sources/stripe/resource/tr"
)

func NewIsDeleted(deletedEvents ...string) downloader.PostProcessor {
	return func(ctx context.Context, obj api.Object, task *downloader.Task) error {
		return isDeleted(ctx, deletedEvents, obj)
	}
}

// isDeleted is a post-processor that sets is_deleted=true on an event's payload object
// if event type matches on of the types specified in the parameters
func isDeleted(_ context.Context, deletedEvents []string, obj api.Object) error {
	eventType := tr.GetString(obj, "type")
	for _, t := range deletedEvents {
		if eventType == t {
			if payload := tr.ExtractEventPayload(obj); payload != nil {
				payload["is_deleted"] = true
			}
			return nil
		}
	}
	return nil
}
