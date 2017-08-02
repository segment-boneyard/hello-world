package source

import (
	"encoding/json"
	"sync"

	"golang.org/x/net/context"

	"google.golang.org/grpc"

	"github.com/segmentio/analytics-go"
	"github.com/segmentio/source-runner/domain"
)

var (
	c    *client
	once sync.Once
)

type client struct {
	client domain.SourceClient
}

func newClient(config *Config) (*client, error) {
	var url string
	if config.URL == "" {
		url = "localhost:4000"
	} else {
		url = config.URL
	}

	var err error
	once.Do(func() {
		var cc *grpc.ClientConn
		cc, err = grpc.Dial(url, grpc.WithInsecure())
		c = &client{client: domain.NewSourceClient(cc)}
	})

	return c, err
}

func (c *client) Set(collection string, id string, properties map[string]interface{}) error {
	var err error
	setRequest := &domain.SetRequest{
		Collection: collection,
		Id:         id,
	}
	setRequest.Properties, err = json.Marshal(properties)
	if err != nil {
		return err
	}

	_, err = c.client.Set(context.Background(), setRequest)
	return err
}

func (c *client) SetBatch(reqs []*SetMessage) error {
	for _, req := range reqs {
		if err := c.Set(req.Collection, req.ID, req.Properties); err != nil {
			return err
		}
	}
	return nil
}

func (c *client) Track(track *analytics.Track) error {
	var err error

	trackRequest := &domain.TrackRequest{
		AnonymousId: track.AnonymousId,
		UserId:      track.UserId,
		Event:       track.Event,
		MessageId:   track.MessageId,
		Timestamp:   track.Timestamp,
	}

	trackRequest.Context, err = json.Marshal(track.Context)
	if err != nil {
		return err
	}

	trackRequest.Integrations, err = json.Marshal(track.Integrations)
	if err != nil {
		return err
	}

	trackRequest.Properties, err = json.Marshal(track.Properties)
	if err != nil {
		return err
	}

	_, err = c.client.Track(context.Background(), trackRequest)
	return err
}

func (c *client) Identify(identify *analytics.Identify) error {
	var err error

	identifyRequest := &domain.IdentifyRequest{
		AnonymousId: identify.AnonymousId,
		UserId:      identify.UserId,
		MessageId:   identify.MessageId,
		Timestamp:   identify.Timestamp,
	}

	identifyRequest.Context, err = json.Marshal(identify.Context)
	if err != nil {
		return err
	}

	identifyRequest.Integrations, err = json.Marshal(identify.Integrations)
	if err != nil {
		return err
	}

	identifyRequest.Traits, err = json.Marshal(identify.Traits)
	if err != nil {
		return err
	}

	_, err = c.client.Identify(context.Background(), identifyRequest)
	return err
}

func (c *client) Group(group *analytics.Group) error {
	var err error

	groupRequest := &domain.GroupRequest{
		AnonymousId: group.AnonymousId,
		UserId:      group.UserId,
		GroupId:     group.GroupId,
		MessageId:   group.MessageId,
		Timestamp:   group.Timestamp,
	}

	groupRequest.Context, err = json.Marshal(group.Context)
	if err != nil {
		return err
	}

	groupRequest.Integrations, err = json.Marshal(group.Integrations)
	if err != nil {
		return err
	}

	groupRequest.Traits, err = json.Marshal(group.Traits)
	if err != nil {
		return err
	}

	_, err = c.client.Group(context.Background(), groupRequest)
	return err
}

func (c *client) GetContext(options GetContextOptions) ([]byte, error) {
	res, err := c.client.GetContext(context.Background(), &domain.GetContextRequest{AllowFailed: options.AllowFailed})
	if err != nil {
		return nil, err
	}

	return res.Data, nil
}

func (c *client) GetContextIntoFile(options GetContextOptions) (string, error) {
	res, err := c.client.GetContextIntoFile(context.Background(), &domain.GetContextIntoFileRequest{AllowFailed: options.AllowFailed})
	if err != nil {
		return "", err
	}

	return res.Filename, nil
}

func (c *client) SetContext(data []byte) error {
	req := &domain.StoreContextRequest{Payload: data}
	_, err := c.client.StoreContext(context.Background(), req)
	return err
}

func (c *client) SetContextFromFile(filename string) error {
	req := &domain.StoreContextFromFileRequest{Filename: filename}
	_, err := c.client.StoreContextFromFile(context.Background(), req)
	return err
}

func (c *client) ReportError(message string, collection string) error {
	_, err := c.client.ReportError(context.Background(), &domain.ReportErrorRequest{Collection: collection, Message: message})
	return err
}

func (c *client) ReportWarning(message string, collection string) error {
	_, err := c.client.ReportWarning(context.Background(), &domain.ReportWarningRequest{Collection: collection, Message: message})
	return err
}

func (c *client) StatsIncrement(name string, value int64, tags []string) error {
	_, err := c.client.StatsIncrement(context.Background(), &domain.StatsRequest{Name: name, Value: value, Tags: tags})
	return err
}

func (c *client) StatsHistogram(name string, value int64, tags []string) error {
	_, err := c.client.StatsHistogram(context.Background(), &domain.StatsRequest{Name: name, Value: value, Tags: tags})
	return err
}

func (c *client) StatsGauge(name string, value int64, tags []string) error {
	_, err := c.client.StatsGauge(context.Background(), &domain.StatsRequest{Name: name, Value: value, Tags: tags})
	return err
}

func (c *client) KeepAlive() error {
	_, err := c.client.KeepAlive(context.Background(), &domain.Empty{})
	return err
}
