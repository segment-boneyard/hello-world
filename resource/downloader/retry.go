package downloader

import (
	"context"
	"github.com/apex/log"
	"github.com/segment-sources/stripe/api"
	"github.com/segmentio/backo-go"
	"time"
)

var retryBackoff = backo.NewBacko(time.Second, 2, 0, time.Second*30)

const retryMaxAttempts = 5

func RetryGetList(ctx context.Context, client api.Client, req *api.Request) (*api.ObjectList, error) {
	res, err := RetryApiCall(func() (interface{}, error) {
		return client.GetList(ctx, req)
	})

	if err != nil {
		return nil, err
	}

	return res.(*api.ObjectList), nil
}

func RetryGetObject(ctx context.Context, client api.Client, req *api.Request) (api.Object, error) {
	res, err := RetryApiCall(func() (interface{}, error) {
		return client.GetObject(ctx, req)
	})

	if err != nil {
		return nil, err
	}

	return res.(api.Object), nil
}

func RetryApiCall(f func() (interface{}, error)) (resp interface{}, err error) {
	attemptsLeft := retryMaxAttempts
	for attemptsLeft > 0 {
		resp, err = f()
		if err == nil {
			return
		}

		delay := retryBackoff.Duration(retryMaxAttempts - attemptsLeft)
		attemptsLeft--

		if api.IsErrorPermanent(err) {
			return nil, err
		}

		logger := log.WithError(err).WithFields(log.Fields{
			"attempts_left": attemptsLeft,
		})

		if attemptsLeft > 0 {
			logger = logger.WithField("delay", delay.String())
			logger.Warn("api call failed, will retry after delay")
			time.Sleep(delay)
		}
	}

	return
}
