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
)

var skuEvents = []string{
	"sku.created",
	"sku.updated",
	"sku.deleted",
}

type Sku struct {
	name      string
	apiClient api.Client
	objs      chan api.Object
	msgs      chan source.SetMessage
	errs      chan integration.CollectionError
	dedupe    dedupe.Interface
}

func (r *Sku) DesiredObjects() []string {
	return []string{"sku"}
}

func (r *Sku) DesiredEvents() []string {
	return skuEvents
}

func (r *Sku) StartProducer(ctx context.Context, runContext integration.RunContext) error {
	defer close(r.objs)
	defer close(r.errs)
	var task *downloader.Task
	if runContext.PreviousRunTimestamp.IsZero() {
		task = &downloader.Task{
			Collection: r.name,
			Request: &api.Request{
				Url:           "/v1/skus?limit=100",
				LogCollection: r.name,
			},
			Output: r.objs,
			Errors: r.errs,
		}
	} else {
		task = tasks.MakeIncremental(r, r.name, runContext.PreviousRunTimestamp, r.objs, r.errs)
	}

	return downloader.New(r.apiClient).Do(ctx, task)
}

func (r *Sku) GetEventProcessors() []downloader.PostProcessor {
	return []downloader.PostProcessor{
		processors.NewIsDeleted("sku.deleted"),
	}
}

func (r *Sku) StartConsumer(ctx context.Context, ch <-chan api.Object) {
	defer close(r.msgs)
	for obj := range ch {
		switch tr.GetString(obj, "object") {
		case "event":
			if payload := tr.ExtractEventPayload(obj, "sku"); payload != nil {
				r.consumeSku(payload, true)
			}
		case "sku":
			r.consumeSku(obj, false)
		}
	}
}

func (r *Sku) consumeSku(obj api.Object, fromEvent bool) {
	if msg := r.transform(obj); msg != nil && !(fromEvent && r.dedupe.SeenBefore(msg.ID)) {
		r.msgs <- *msg
	}
}

func (r *Sku) transform(obj api.Object) *source.SetMessage {
	var id string
	if id = tr.GetString(obj, "id"); id == "" {
		return nil
	}

	properties := map[string]interface{}{
		"product_id": obj["product"],
		"active":     obj["active"],
		"currency":   obj["currency"],
		"image":      obj["image"],
		"livemode":   obj["livemode"],
		"price":      obj["price"],
	}

	if v, ok := obj["is_deleted"].(bool); ok && v {
		properties["is_deleted"] = v
	}

	tr.Flatten(tr.GetMap(obj, "metadata"), "metadata_", properties)
	tr.Flatten(tr.GetMap(obj, "inventory"), "inventory_", properties)
	tr.Flatten(tr.GetMap(obj, "package_dimensions"), "package_dimensions_", properties)
	tr.Flatten(tr.GetMap(obj, "attributes"), "attributes_", properties)

	if created := tr.GetTimestamp(obj, "created"); created != "" {
		properties["created"] = created
	}
	if updated := tr.GetTimestamp(obj, "updated"); updated != "" {
		properties["updated"] = updated
	}

	return &source.SetMessage{
		ID:         id,
		Collection: r.name,
		Properties: properties,
	}
}

func (r *Sku) Collection() string {
	return r.name
}

func (r *Sku) Objects() <-chan api.Object {
	return r.objs
}

func (r *Sku) Messages() <-chan source.SetMessage {
	return r.msgs
}

func (r *Sku) CollectionErrors() <-chan integration.CollectionError {
	return r.errs
}

func (r *Sku) Consumers() []integration.Consumer {
	return []integration.Consumer{r}
}

func (r *Sku) Close() {
	r.dedupe.Close()
}

func NewSku(apiClient api.Client) *Sku {
	return &Sku{
		name:      "skus",
		apiClient: apiClient,
		objs:      make(chan api.Object, 1000),
		msgs:      make(chan source.SetMessage),
		errs:      make(chan integration.CollectionError),
		dedupe:    dedupe.New(),
	}
}
