package store

import (
	"errors"
	"sort"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"gofr.dev/pkg/gofr"

	"github.com/sksmagr23/url-shortener-gofr/model"
)

type AnalyticsStore struct{}

func NewAnalyticsStore() *AnalyticsStore {
	return &AnalyticsStore{}
}

func (s *AnalyticsStore) InsertClick(ctx *gofr.Context, click *model.ClickEvent) error {
	click.ID = primitive.NewObjectID().Hex()
	if click.Timestamp.IsZero() {
		click.Timestamp = time.Now().UTC()
	}
	_, err := ctx.Mongo.InsertOne(ctx, "click_events", click)
	return err
}

func (s *AnalyticsStore) HasIPClicked(ctx *gofr.Context, shortCode, ip string) (bool, error) {
	var result model.ClickEvent
	err := ctx.Mongo.FindOne(ctx, "click_events", bson.M{"short_code": shortCode, "ip_address": ip}, &result)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func (s *AnalyticsStore) GetFieldBreakdown(ctx *gofr.Context, shortCode, fieldName string, limit int) ([]model.BreakdownItem, error) {
	if limit <= 0 {
		limit = 10
	}

	var events []model.ClickEvent
	err := ctx.Mongo.Find(ctx, "click_events", bson.M{"short_code": shortCode}, &events)
	if err != nil {
		return nil, err
	}

	counts := make(map[string]int64)
	for _, event := range events {
		var val string
		switch fieldName {
		case "browser":
			val = event.Browser
		case "os":
			val = event.OS
		case "device_type":
			val = event.DeviceType
		case "country":
			val = event.Country
		case "referrer":
			val = event.Referrer
		}
		if val == "" {
			val = "Unknown"
		}
		counts[val]++
	}

	results := make([]model.BreakdownItem, 0, len(counts))
	for name, count := range counts {
		results = append(results, model.BreakdownItem{Name: name, Count: count})
	}

	sort.SliceStable(results, func(i, j int) bool {
		if results[i].Count == results[j].Count {
			return results[i].Name < results[j].Name
		}
		return results[i].Count > results[j].Count
	})

	if len(results) > limit {
		results = results[:limit]
	}

	return results, nil
}

func (s *AnalyticsStore) GetTimeseries(ctx *gofr.Context, shortCode, unit string, limit int) ([]model.TimeseriesPoint, error) {
	if limit <= 0 {
		limit = 30
	}

	var events []model.ClickEvent
	err := ctx.Mongo.Find(ctx, "click_events", bson.M{"short_code": shortCode}, &events)
	if err != nil {
		return nil, err
	}

	counts := make(map[string]int64)
	for _, event := range events {
		var key string
		if unit == "hour" {
			key = event.Timestamp.Format("2006-01-02T15:00:00Z")
		} else {
			key = event.Timestamp.Format("2006-01-02")
		}
		counts[key]++
	}

	results := make([]model.TimeseriesPoint, 0, len(counts))
	for ts, count := range counts {
		results = append(results, model.TimeseriesPoint{Timestamp: ts, Clicks: count})
	}

	sort.SliceStable(results, func(i, j int) bool {
		return results[i].Timestamp < results[j].Timestamp
	})

	if len(results) > limit {
		results = results[:limit]
	}

	return results, nil
}
