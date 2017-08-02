package processors

import (
	"context"
	"fmt"
	"github.com/segment-sources/stripe/api"
	"net/url"
)

type MockClient struct {
	GetListPayloads map[string]*api.ObjectList
}

func (c *MockClient) GetList(ctx context.Context, req *api.Request) (*api.ObjectList, error) {
	u, err := url.Parse(req.Url)
	if err != nil {
		return nil, err
	}
	qs := u.Query()
	for k, v := range req.Qs {
		qs[k] = v
	}
	u.RawQuery = qs.Encode()

	payload, ok := c.GetListPayloads[u.String()]
	if !ok {
		return nil, fmt.Errorf("unknown url requested: %s", u.String())
	}

	return payload, nil
}

func (c *MockClient) GetObject(context.Context, *api.Request) (api.Object, error) {
	panic("implement me")
}
