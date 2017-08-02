package processors

import (
	"context"
	"encoding/json"
	"github.com/segment-sources/stripe/api"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestListExpander(t *testing.T) {
	client := &MockClient{GetListPayloads: map[string]*api.ObjectList{
		"/v1/customers/cus_1/sources?limit=100&starting_after=card_2": {
			Objects: []api.Object{
				{
					"id":     "card_3",
					"object": "card",
				},
				{
					"id":     "card_4",
					"object": "card",
				},
			},
			HasMore: true,
		},
		"/v1/customers/cus_1/sources?limit=100&starting_after=card_4": {
			Objects: []api.Object{
				{
					"id":     "card_5",
					"object": "card",
				},
			},
			HasMore: false,
		},
	}}

	proc := NewListExpander("sources", client)
	obj := api.Object{
		"id":      "cus_1",
		"object":  "customer",
		"created": json.Number("1501174932"),
		"sources": map[string]interface{}{
			"object": "list",
			"data": []interface{}{
				map[string]interface{}{
					"id":     "card_1",
					"object": "card",
				},
				map[string]interface{}{
					"id":     "card_2",
					"object": "card",
				},
			},
			"has_more":    true,
			"total_count": 5,
			"url":         "/v1/customers/cus_1/sources",
		},
	}

	a := assert.New(t)

	err := proc(context.Background(), obj)
	if !a.NoError(err) {
		return
	}

	expected := api.Object{
		"id":      "cus_1",
		"object":  "customer",
		"created": json.Number("1501174932"),
		"sources": map[string]interface{}{
			"object": "list",
			"data": []interface{}{
				map[string]interface{}{
					"id":     "card_1",
					"object": "card",
				},
				map[string]interface{}{
					"id":     "card_2",
					"object": "card",
				},
				map[string]interface{}{
					"id":     "card_3",
					"object": "card",
				},
				map[string]interface{}{
					"id":     "card_4",
					"object": "card",
				},
				map[string]interface{}{
					"id":     "card_5",
					"object": "card",
				},
			},
			"has_more":    true,
			"total_count": 5,
			"url":         "/v1/customers/cus_1/sources",
		},
	}

	a.Equal(obj, expected)
}
