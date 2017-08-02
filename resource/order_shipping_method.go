package resource

import (
	"context"
	"crypto/md5"
	"fmt"
	"github.com/segment-sources/stripe/api"
	"github.com/segment-sources/stripe/integration"
	"github.com/segment-sources/stripe/resource/dedupe"
	"github.com/segment-sources/stripe/resource/tr"
	"github.com/segmentio/go-source"
	"strings"
)

// OrderShippingMethod
// has no known discrepancies
type OrderShippingMethod struct {
	name      string
	apiClient api.Client
	objs      chan api.Object
	msgs      chan source.SetMessage
	errs      chan integration.CollectionError
	dedupe    dedupe.Interface
}

func (r *OrderShippingMethod) DesiredObjects() []string {
	return []string{"order"}
}

func (r *OrderShippingMethod) DesiredEvents() []string {
	return orderEvents
}

func (r *OrderShippingMethod) StartProducer(ctx context.Context, runContext integration.RunContext) error {
	// downloading events in incremental mode is handled by the bundle that this resource is a part of
	close(r.objs)
	close(r.errs)
	return nil
}

func (r *OrderShippingMethod) StartConsumer(ctx context.Context, ch <-chan api.Object) {
	defer close(r.msgs)
	for obj := range ch {
		switch tr.GetString(obj, "object") {
		case "event":
			if payload := tr.ExtractEventPayload(obj, "order"); payload != nil {
				r.consumeOrder(payload, true)
			}
		case "order":
			r.consumeOrder(obj, false)
		}
	}
}

func (r *OrderShippingMethod) consumeOrder(obj api.Object, fromEvent bool) {
	for _, method := range tr.GetMapList(obj, "shipping_methods") {
		if msg := r.transform(obj, method); msg != nil && !(fromEvent && r.dedupe.SeenBefore(msg.ID)) {
			r.msgs <- *msg
		}
	}
}

func (r *OrderShippingMethod) transform(order, method api.Object) *source.SetMessage {
	var orderId string
	if orderId = tr.GetString(order, "id"); orderId == "" {
		return nil
	}
	var methodId string
	if methodId = tr.GetString(method, "id"); methodId == "" {
		return nil
	}

	properties := map[string]interface{}{
		"order_id":    orderId,
		"shipping_id": methodId,
		"amount":      method["amount"],
		"currency":    method["currency"],
		"description": method["description"],
	}

	tr.Flatten(tr.GetMap(method, "delivery_estimate"), "delivery_estimate_", properties)

	internalIdSrc := strings.Join([]string{methodId, orderId}, ", ")
	hash := md5.New()
	fmt.Fprint(hash, internalIdSrc)

	return &source.SetMessage{
		ID:         fmt.Sprintf("%x", hash.Sum(nil)),
		Collection: r.name,
		Properties: properties,
	}
}

func (r *OrderShippingMethod) Collection() string {
	return r.name
}

func (r *OrderShippingMethod) Objects() <-chan api.Object {
	return r.objs
}

func (r *OrderShippingMethod) Messages() <-chan source.SetMessage {
	return r.msgs
}

func (r *OrderShippingMethod) CollectionErrors() <-chan integration.CollectionError {
	return r.errs
}

func (r *OrderShippingMethod) Consumers() []integration.Consumer {
	return []integration.Consumer{r}
}

func (r *OrderShippingMethod) Close() {
	r.dedupe.Close()
}

func NewOrderShippingMethod(apiClient api.Client) *OrderShippingMethod {
	return &OrderShippingMethod{
		name:      "order_shipping_methods",
		apiClient: apiClient,
		objs:      make(chan api.Object, 1000),
		msgs:      make(chan source.SetMessage),
		errs:      make(chan integration.CollectionError),
		dedupe:    dedupe.New(),
	}
}
