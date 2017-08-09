package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/apex/log"
	"github.com/nu7hatch/gouuid"
	"github.com/pkg/errors"
	"github.com/segmentio/go-source"
	"github.com/segmentio/go-source/source-logger"
	"github.com/segmentio/ur-log"
	"net/http"
	"os"
	"time"
)

const stripeApiVersion = "2016-07-06"

type clientImpl struct {
	httpClient   HttpClient
	baseUrl      string
	secret       string
	throttler    *Throttler
	sourceClient source.Client
	sourceLogger SourceLogger
}

func (c *clientImpl) GetList(ctx context.Context, req *Request) (*ObjectList, error) {
	output := &listResponse{}
	if err := c.get(ctx, req, output); err != nil {
		return nil, err
	}

	result := &ObjectList{
		HasMore: output.HasMore,
	}
	for _, obj := range output.Data {
		result.Objects = append(result.Objects, obj)
	}

	return result, nil
}

func (c *clientImpl) GetObject(ctx context.Context, req *Request) (Object, error) {
	output := Object{}
	if err := c.get(ctx, req, &output); err != nil {
		return nil, err
	}

	return output, nil
}

func (c *clientImpl) prepareRequest(req *Request) (*http.Request, error) {
	url := req.Url
	if url[0] == '/' {
		url = c.baseUrl + url
	}

	httpReq, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	newQs := httpReq.URL.Query()
	for key, value := range req.Qs {
		newQs[key] = value
	}
	httpReq.URL.RawQuery = newQs.Encode()

	httpReq.Header.Set("Stripe-Version", stripeApiVersion)
	for key, value := range req.Headers {
		httpReq.Header[key] = value
	}
	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.secret))

	return httpReq, nil
}

func (c *clientImpl) get(ctx context.Context, req *Request, output interface{}) error {
	httpReq, err := c.prepareRequest(req)
	if err != nil {
		ctx, _ := urlog.GetContextualLogger(ctx, nil, log.Fields{"request": req})
		return urlog.WrapError(ctx, err, "failed to prepare request")
	}

	uv4, err := uuid.NewV4()
	if err != nil {
		return urlog.WrapError(ctx, err, "failed to generate uuid")
	}

	ctx, logger := urlog.GetContextualLogger(ctx, nil, log.Fields{
		"request": log.Fields{
			"id":      uv4.String(),
			"url":     httpReq.URL.String(),
			"headers": httpReq.Header,
		},
	})

	c.throttler.Use()

	ts := time.Now()
	logger.Info("http request")
	metricTags := []string{}
	workspaceSlug := os.Getenv("SEGMENT_WORKSPACE_SLUG")
	projectSlug := os.Getenv("SEGMENT_SOURCE_SLUG")
	if projectSlug != "" && workspaceSlug != "" {
		metricTags = append(metricTags, fmt.Sprintf("project:%s/%s", workspaceSlug, projectSlug))
	}
	c.sourceClient.StatsIncrement("stripe.requests", 1, metricTags)
	resp, err := c.httpClient.Do(httpReq)
	c.sourceLogger.RequestSent(req.LogCollection, httpReq.URL.String(), sourcelogger.Metadata{"uuid": uv4.String()})
	if err != nil {
		return urlog.WrapError(ctx, err, "error performing request")
	}
	defer resp.Body.Close()

	buffer := &bytes.Buffer{}
	if _, err := buffer.ReadFrom(resp.Body); err != nil {
		return urlog.WrapError(ctx, err, "error reading response")
	}

	duration := time.Now().Sub(ts)
	metricTags = append(metricTags,
		fmt.Sprintf("status_code:%d", resp.StatusCode),
		fmt.Sprintf("status_code_bucket:%dxx", resp.StatusCode/100),
	)
	c.sourceClient.StatsIncrement("stripe.responses", 1, metricTags)
	c.sourceClient.StatsHistogram("stripe.response.payload_size", int64(buffer.Len()), metricTags)
	c.sourceClient.StatsHistogram("stripe.response.latency", duration.Nanoseconds()/1000000, metricTags)

	headersBuffer := &bytes.Buffer{}
	resp.Header.Write(headersBuffer)
	logMetadata := sourcelogger.Metadata{
		"uuid":    uv4.String(),
		"status":  resp.Status,
		"headers": headersBuffer.String(),
	}
	c.sourceLogger.ResponseReceived(req.LogCollection, httpReq.URL.String(), logMetadata, duration, buffer.String())

	ctx, logger = urlog.GetContextualLogger(ctx, logger, log.Fields{
		"response": log.Fields{
			"headers": resp.Header,
			"status":  resp.Status,
			"body":    buffer.String(),
		},
	})
	logger.Debug("http response")

	if resp.StatusCode == 200 {
		decoder := json.NewDecoder(bytes.NewReader(buffer.Bytes()))
		decoder.UseNumber()
		if err := decoder.Decode(output); err != nil {
			return urlog.WrapError(ctx, err, "error decoding response")
		}
		return nil
	}

	err = urlog.WrapError(ctx, errors.New("unexpected response status code"), "")
	if resp.StatusCode >= 500 || resp.StatusCode == 429 {
		// transient errors
		return err
	}

	// non-transient errors
	isAuthRelated := resp.StatusCode == 401
	return &permanentError{
		wrappedError:  err.(wrappedError),
		isAuthRelated: isAuthRelated,
	}
}

func NewClient(opts *ClientOptions) Client {
	if opts.BaseUrl == "" {
		opts.BaseUrl = "https://api.stripe.com"
	}

	return &clientImpl{
		httpClient:   opts.HttpClient,
		baseUrl:      opts.BaseUrl,
		secret:       opts.Secret,
		throttler:    NewThrottler(opts.MaxRps, time.Second),
		sourceClient: opts.SourceClient,
		sourceLogger: opts.SourceClient.Log(),
	}
}
