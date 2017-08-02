package tasks

import (
	"fmt"
	"github.com/segment-sources/stripe/api"
	"github.com/segment-sources/stripe/integration"
	"github.com/segment-sources/stripe/resource/downloader"
	"net/url"
	"sort"
	"time"
)

type HasEventProcessors interface {
	GetEventProcessors() []downloader.PostProcessor
}

// MakeIncremental is a shortcut for creating a downloader.Task
func MakeIncremental(res integration.Resource, collection string, previousRunTimestamp time.Time, ch chan api.Object, errs chan integration.CollectionError) *downloader.Task {
	allEventsSet := map[string]bool{}
	for _, con := range res.Consumers() {
		for _, eventType := range con.DesiredEvents() {
			allEventsSet[eventType] = true
		}
	}
	if len(allEventsSet) < 1 {
		return nil
	}
	allDesiredEvents := []string{}
	for eventType := range allEventsSet {
		allDesiredEvents = append(allDesiredEvents, eventType)
	}
	sort.Strings(allDesiredEvents)

	var postProcessors []downloader.PostProcessor
	if i, ok := res.(HasEventProcessors); ok {
		postProcessors = i.GetEventProcessors()
	}

	return &downloader.Task{
		Collection: collection,
		Request: &api.Request{
			Url: "/v1/events",
			Qs: url.Values{
				"limit":       []string{"100"},
				"types[]":     allDesiredEvents,
				"created[gt]": []string{fmt.Sprintf("%d", previousRunTimestamp.Add(-time.Hour).Unix())},
			},
		},
		PostProcessors: postProcessors,
		Output:         ch,
		Errors:         errs,
	}
}
