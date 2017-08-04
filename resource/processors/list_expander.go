package processors

import (
	"context"
	"github.com/segment-sources/stripe/api"
	"github.com/segment-sources/stripe/resource/downloader"
	"github.com/segment-sources/stripe/resource/tr"
	"net/url"
	"strings"
	"sync"
)

func NewListExpander(path string, apiClient api.Client) downloader.PostProcessor {
	d := downloader.New(apiClient)

	return func(ctx context.Context, obj api.Object, task *downloader.Task) error {
		return expandList(ctx, d, obj, task, path)
	}
}

// expandList is a post-processor that expands item lists inside objects
// by downloading and appending extra items that didn't fit in the initial response
func expandList(ctx context.Context, d *downloader.Client, obj api.Object, task *downloader.Task, path string) error {
	target := obj
	for _, component := range strings.Split(path, ".") {
		if v, ok := target[component].(map[string]interface{}); ok {
			target = v
		} else {
			return nil
		}
	}
	if tr.GetString(target, "object") != "list" {
		return nil
	}
	if !tr.GetBool(target, "has_more") {
		return nil
	}

	lastSeenId := ""
	for _, o := range tr.GetMapList(target, "data") {
		if id := tr.GetString(o, "id"); id != "" {
			lastSeenId = id
		}
	}

	data, dataIsList := target["data"].([]interface{})
	endpoint := tr.GetString(target, "url")
	if !dataIsList || endpoint == "" || lastSeenId == "" {
		return nil
	}

	ch := make(chan api.Object)
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		for extraObj := range ch {
			data = append(data, map[string]interface{}(extraObj))
		}
	}()

	err := d.Do(ctx, &downloader.Task{
		Request: &api.Request{
			Url: endpoint,
			Qs: url.Values{
				"limit":          []string{"100"},
				"starting_after": []string{lastSeenId},
			},
			LogCollection: task.Collection,
		},
		Output: ch,
	})

	close(ch)
	wg.Wait()

	if err == nil {
		target["data"] = data
	}

	return err
}
