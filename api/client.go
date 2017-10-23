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
	"os"
	"time"
)

const helloWorldApiVersion = "2016-07-06"

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


func (c *clientImpl) get(ctx context.Context, req *Request, output interface{}) error {

	const resp_StatusCode  = 200
	uv4, err := uuid.NewV4()
	if err != nil {
		return urlog.WrapError(ctx, err, "failed to generate uuid")
	}

	ctx, logger := urlog.GetContextualLogger(ctx, nil, log.Fields{
		"request": log.Fields{
			"id":      uv4.String(),
			"url":     "helloworld:///",
			"headers": "",
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

	// Here's where we would make the HTTP request to the API
	// Instead, we emulate the response back from the API
	buffer := &bytes.Buffer{}
	buffer.WriteString("{ " +
		"\"object:\": \"helloworld\", " +
		"\"message\": \"Hello, World!\", " +
		"\"has_more\": \"false\" " +
		"}")

	// Fake the response metrics
	duration := time.Now().Sub(ts)
	metricTags = append(metricTags,
		fmt.Sprintf("status_code:%d", resp_StatusCode),
		fmt.Sprintf("status_code_bucket:%dxx", resp_StatusCode/100),
	)
	c.sourceClient.StatsIncrement("helloWorld.responses", 1, metricTags)
	c.sourceClient.StatsHistogram("helloWorld.response.payload_size", int64(buffer.Len()), metricTags)
	c.sourceClient.StatsHistogram("helloWorld.response.latency", duration.Nanoseconds()/1000000, metricTags)

	headersBuffer := &bytes.Buffer{}
	logMetadata := sourcelogger.Metadata{
		"uuid":    uv4.String(),
		"status":  resp_StatusCode,
		"headers": headersBuffer.String(),
	}
	c.sourceLogger.ResponseReceived(req.LogCollection, "helloworld:///", logMetadata, duration, buffer.String())

	ctx, logger = urlog.GetContextualLogger(ctx, logger, log.Fields{
		"response": log.Fields{
			"headers": "",
			"status":  resp_StatusCode,
			"body":    buffer.String(),
		},
	})
	logger.Debug("http response")

	if resp_StatusCode == 200 {
		decoder := json.NewDecoder(bytes.NewReader(buffer.Bytes()))
		decoder.UseNumber()
		if err := decoder.Decode(output); err != nil {
			return urlog.WrapError(ctx, err, "error decoding response")
		}
		return nil
	}

	err = urlog.WrapError(ctx, errors.New("unexpected response status code"), "")
	if resp_StatusCode >= 500 || resp_StatusCode == 429 {
		// transient errors
		return err
	}

	// non-transient errors
	isAuthRelated := resp_StatusCode == 401
	return &permanentError{
		wrappedError:  err.(wrappedError),
		isAuthRelated: isAuthRelated,
	}
}

func NewClient(opts *ClientOptions) Client {
	if opts.BaseUrl == "" {
		opts.BaseUrl = "helloworld:///" //we don't need an API URL for Hello, World
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
