package resource

import (
	"context"
	"github.com/segment-sources/stripe/api"
	"github.com/segment-sources/stripe/integration"
	"github.com/segment-sources/stripe/resource/dedupe"
	"github.com/segment-sources/stripe/resource/downloader"
	"github.com/segment-sources/stripe/resource/tr"
	"github.com/segmentio/go-source"
	"github.com/segmentio/ur-log"
)

type Account struct {
	name      string
	apiClient api.Client
	objs      chan api.Object
	msgs      chan source.SetMessage
	errs      chan integration.CollectionError
	dedupe    dedupe.Interface
}

func (r *Account) DesiredObjects() []string {
	return []string{"account"}
}

func (r *Account) DesiredEvents() []string {
	return nil
}

// StartProducer for Account always performs a full sync
func (r *Account) StartProducer(ctx context.Context, runContext integration.RunContext) error {
	defer close(r.objs)
	defer close(r.errs)

	// pull the primary account info
	primaryAcct, err := downloader.RetryGetObject(ctx, r.apiClient, &api.Request{
		Url:           "/v1/account",
		LogCollection: r.name,
	})
	if err != nil {
		r.errs <- integration.CollectionError{
			Collection: r.name,
			Message:    "HTTP request failed",
		}
		return urlog.WrapError(ctx, err, "failed to fetch primary account")
	}
	r.objs <- primaryAcct

	// pull all the Stripe Connect accounts
	d := downloader.New(r.apiClient)
	req := &api.Request{
		Url:           "/v1/accounts?limit=100",
		LogCollection: r.name,
	}
	task := &downloader.Task{
		Collection: r.name,
		Request:    req,
		Output:     r.objs,
		Errors:     r.errs,
	}
	if err := d.Do(ctx, task); err != nil {
		return err
	}

	return nil
}

func (r *Account) StartConsumer(ctx context.Context, ch <-chan api.Object) {
	defer close(r.msgs)
	for obj := range ch {
		if msg := r.transform(obj); msg != nil {
			r.msgs <- *msg
		}
	}
}

func (r *Account) transform(obj api.Object) *source.SetMessage {
	var id string
	if id = tr.GetString(obj, "id"); id == "" {
		return nil
	}

	properties := map[string]interface{}{
		"email":                   obj["email"],
		"statement_descriptor":    obj["statement_descriptor"],
		"display_name":            obj["display_name"],
		"timezone":                obj["timezone"],
		"details_submitted":       obj["details_submitted"],
		"charges_enabled":         obj["charges_enabled"],
		"transfers_enabled":       obj["transfers_enabled"],
		"default_currency":        obj["default_currency"],
		"country":                 obj["country"],
		"business_name":           obj["business_name"],
		"business_url":            obj["business_url"],
		"support_phone":           obj["support_phone"],
		"business_logo":           obj["business_logo"],
		"support_url":             obj["support_url"],
		"support_email":           obj["support_email"],
		"managed":                 obj["managed"],
		"product_description":     obj["product_description"],
		"debit_negative_balances": obj["debit_negative_balances"],
	}

	tr.Flatten(tr.GetMap(obj, "metadata"), "metadata_", properties)
	tr.Flatten(tr.GetMap(obj, "support_address"), "support_address_", properties)

	return &source.SetMessage{
		ID:         id,
		Collection: r.name,
		Properties: properties,
	}
}

func (r *Account) Collection() string {
	return r.name
}

func (r *Account) Objects() <-chan api.Object {
	return r.objs
}

func (r *Account) Messages() <-chan source.SetMessage {
	return r.msgs
}

func (r *Account) CollectionErrors() <-chan integration.CollectionError {
	return r.errs
}

func (r *Account) Consumers() []integration.Consumer {
	return []integration.Consumer{r}
}

func (r *Account) Close() {
	r.dedupe.Close()
}

func NewAccount(apiClient api.Client) *Account {
	return &Account{
		name:      "accounts",
		apiClient: apiClient,
		objs:      make(chan api.Object, 1000),
		msgs:      make(chan source.SetMessage),
		errs:      make(chan integration.CollectionError),
		dedupe:    dedupe.New(),
	}
}
