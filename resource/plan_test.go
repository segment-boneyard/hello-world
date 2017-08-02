package resource

import (
	"context"
	"encoding/json"
	"github.com/segment-sources/stripe/api"
	"github.com/segmentio/go-source"
	"github.com/stretchr/testify/suite"
	"sync"
	"testing"
)

type PlanConsumerSuite struct {
	suite.Suite

	consumer        *Plan
	inputPlan       map[string]interface{}
	expectedMessage source.SetMessage
}

func TestPlanConsumerSuite(t *testing.T) {
	suite.Run(t, new(PlanConsumerSuite))
}

func (s *PlanConsumerSuite) SetupTest() {
	s.inputPlan = map[string]interface{}{
		"id":             "warehouses-standard-$199-1-month",
		"object":         "plan",
		"amount":         json.Number("19900"),
		"created":        json.Number("1459956536"),
		"currency":       "usd",
		"interval":       "month",
		"interval_count": json.Number("1"),
		"livemode":       true,
		"metadata": map[string]interface{}{
			"service": "warehouses",
			"tier":    "standard",
		},
		"name":                 "Warehouses Standard",
		"statement_descriptor": "Segment Warehouses",
		"trial_period_days":    nil,
	}
	s.expectedMessage = source.SetMessage{
		ID:         "warehouses-standard-$199-1-month",
		Collection: "plans",
		Properties: map[string]interface{}{
			"amount":               json.Number("19900"),
			"created":              "2016-04-06T15:28:56.000Z",
			"currency":             "usd",
			"interval":             "month",
			"interval_count":       json.Number("1"),
			"metadata_service":     "warehouses",
			"metadata_tier":        "standard",
			"name":                 "Warehouses Standard",
			"statement_descriptor": "Segment Warehouses",
			"trial_period_days":    nil,
		},
	}
	s.consumer = NewPlan(nil)
}

func (s *PlanConsumerSuite) testConsumer(input api.Object, output source.SetMessage) {
	objs := make(chan api.Object, 1)
	objs <- input
	close(objs)

	wg := sync.WaitGroup{}

	msgs := []source.SetMessage{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		for msg := range s.consumer.Messages() {
			msgs = append(msgs, msg)
		}
	}()

	s.consumer.StartConsumer(context.Background(), objs)
	wg.Wait()

	if !s.Len(msgs, 1) {
		return
	}

	if !s.Equal(output, msgs[0]) {
		return
	}
}

func (s *PlanConsumerSuite) TestInputPlan() {
	input := api.Object(s.inputPlan)
	s.testConsumer(input, s.expectedMessage)
}

func (s *PlanConsumerSuite) TestInputSubscription() {
	input := api.Object{
		"id":      "sub_B6WtmJlmJwmzUo",
		"object":  "subscription",
		"created": json.Number("1501192079"),
		"plan":    s.inputPlan,
	}
	s.testConsumer(input, s.expectedMessage)
}

func (s *PlanConsumerSuite) TestInputInvoice() {
	input := api.Object{
		"id":     "in_veTwovH2fLLnPo",
		"object": "invoice",
		"date":   1501192880,
		"lines": map[string]interface{}{
			"object": "list",
			"data": []interface{}{
				map[string]interface{}{
					"id":     "sub_7tzPoid1f8Qd7d",
					"object": "line_item",
					"plan":   s.inputPlan,
				},
			},
			"has_more":    false,
			"total_count": 1,
			"url":         "/v1/invoices/in_veTwovH2fLLnPo/lines",
		},
	}

	s.testConsumer(input, s.expectedMessage)
}
