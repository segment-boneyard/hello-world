package resource

import (
	"context"
	"github.com/segment-sources/stripe/api"
	"github.com/segment-sources/stripe/integration"
	"github.com/segment-sources/stripe/resource/dedupe"
	"github.com/segment-sources/stripe/resource/downloader"
	"github.com/segment-sources/stripe/resource/processors"
	"github.com/segment-sources/stripe/resource/tasks"
	"github.com/segment-sources/stripe/resource/tr"
	"github.com/segmentio/go-source"
	"strings"
)

var productEvents = []string{
	"product.created",
	"product.updated",
	"product.deleted",
}

type Product struct {
	name      string
	apiClient api.Client
	objs      chan api.Object
	msgs      chan source.SetMessage
	errs      chan integration.CollectionError
	dedupe    dedupe.Interface
}

func (r *Product) DesiredObjects() []string {
	return []string{"product"}
}

func (r *Product) DesiredEvents() []string {
	return productEvents
}

func (r *Product) StartProducer(ctx context.Context, runContext integration.RunContext) error {
	defer close(r.objs)
	defer close(r.errs)
	var task *downloader.Task
	if runContext.PreviousRunTimestamp.IsZero() {
		task = &downloader.Task{
			Collection: r.name,
			Request: &api.Request{
				Url: "/v1/products?limit=100",
			},
			Output: r.objs,
			Errors: r.errs,
		}
	} else {
		task = tasks.MakeIncremental(r, r.name, runContext.PreviousRunTimestamp, r.objs, r.errs)
	}

	return downloader.New(r.apiClient).Do(ctx, task)
}

func (r *Product) GetEventProcessors() []downloader.PostProcessor {
	return []downloader.PostProcessor{
		processors.NewIsDeleted("product.deleted"),
	}
}

func (r *Product) StartConsumer(ctx context.Context, ch <-chan api.Object) {
	defer close(r.msgs)
	for obj := range ch {
		switch tr.GetString(obj, "object") {
		case "event":
			if payload := tr.ExtractEventPayload(obj, "product"); payload != nil {
				r.consumeProduct(payload, true)
			}
		case "product":
			r.consumeProduct(obj, false)
		}
	}
}

func (r *Product) consumeProduct(obj api.Object, fromEvent bool) {
	if msg := r.transform(obj); msg != nil && !(fromEvent && r.dedupe.SeenBefore(msg.ID)) {
		r.msgs <- *msg
	}
}

func (r *Product) transform(obj api.Object) *source.SetMessage {
	var id string
	if id = tr.GetString(obj, "id"); id == "" {
		return nil
	}

	properties := map[string]interface{}{
		"active":        obj["active"],
		"attributes":    strings.Join(tr.GetStringList(obj, "attributes"), ","),
		"caption":       obj["caption"],
		"deactivate_on": strings.Join(tr.GetStringList(obj, "deactivate_on"), ","),
		"description":   obj["description"],
		"images":        strings.Join(tr.GetStringList(obj, "images"), ","),
		"livemode":      obj["livemode"],
		"name":          obj["name"],
		"shippable":     obj["shippable"],
		"url":           obj["url"],
	}

	if v, ok := obj["is_deleted"].(bool); ok && v {
		properties["is_deleted"] = v
	}

	tr.Flatten(tr.GetMap(obj, "metadata"), "metadata_", properties)
	tr.Flatten(tr.GetMap(obj, "package_dimensions"), "package_dimensions_", properties)

	if ts := tr.GetTimestamp(obj, "created"); ts != "" {
		properties["created"] = ts
	}
	if ts := tr.GetTimestamp(obj, "updated"); ts != "" {
		properties["updated"] = ts
	}

	return &source.SetMessage{
		ID:         id,
		Collection: r.name,
		Properties: properties,
	}
}

func (r *Product) Collection() string {
	return r.name
}

func (r *Product) Objects() <-chan api.Object {
	return r.objs
}

func (r *Product) Messages() <-chan source.SetMessage {
	return r.msgs
}

func (r *Product) CollectionErrors() <-chan integration.CollectionError {
	return r.errs
}

func (r *Product) Consumers() []integration.Consumer {
	return []integration.Consumer{r}
}

func (r *Product) Close() {
	r.dedupe.Close()
}

func NewProduct(apiClient api.Client) *Product {
	return &Product{
		name:      "products",
		apiClient: apiClient,
		objs:      make(chan api.Object, 1000),
		msgs:      make(chan source.SetMessage),
		errs:      make(chan integration.CollectionError),
		dedupe:    dedupe.New(),
	}
}
