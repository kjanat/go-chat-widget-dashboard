package services

import (
	"testing"
	"time"

	"github.com/kjanat/go-chat-widget-dashboard/internal/database"
	"github.com/kjanat/go-chat-widget-dashboard/internal/models"
)

// helper to setup DB with one customer and analytics service
func setupAnalyticsTest(t *testing.T) (*AnalyticsService, *database.DB) {
	db, err := database.New(":memory:")
	if err != nil {
		t.Fatalf("failed to create database: %v", err)
	}

	_, err = db.Exec(`INSERT INTO customers (id, name, email, created_at, brand_colors, logo_url, model_path, animations, openai_prompt, allowed_domains, api_key, active) VALUES ('c1','Test','test@example.com', datetime('now'),'','','','','','','key',1)`)
	if err != nil {
		t.Fatalf("failed to insert customer: %v", err)
	}

	return NewAnalyticsService(db.DB), db
}

func insertMetric(t *testing.T, svc *AnalyticsService, m *models.ChatMetrics) {
	if err := svc.RecordChatMetrics(m); err != nil {
		t.Fatalf("failed to insert metric: %v", err)
	}
}

func TestGetTopCountriesAndDeviceBreakdown(t *testing.T) {
	svc, db := setupAnalyticsTest(t)
	defer db.Close()

	now := time.Now()
	insertMetric(t, svc, &models.ChatMetrics{
		ID:               "m1",
		CustomerID:       "c1",
		SessionID:        "s1",
		Country:          "USA",
		DeviceType:       "desktop",
		SessionStartTime: now,
		CreatedAt:        now,
	})
	insertMetric(t, svc, &models.ChatMetrics{
		ID:               "m2",
		CustomerID:       "c1",
		SessionID:        "s2",
		Country:          "Canada",
		DeviceType:       "mobile",
		SessionStartTime: now,
		CreatedAt:        now,
	})

	cases := []string{"", "WHERE created_at >= date('now', '-1 day')"}
	for _, clause := range cases {
		countries, err := svc.getTopCountries(clause)
		if err != nil {
			t.Errorf("getTopCountries failed with clause %q: %v", clause, err)
			continue
		}
		if len(countries) != 2 {
			t.Errorf("expected 2 countries for clause %q, got %d", clause, len(countries))
		}

		counts := map[string]int64{}
		for _, c := range countries {
			counts[c.Country] = c.SessionCount
		}
		if counts["USA"] != 1 || counts["Canada"] != 1 {
			t.Errorf("unexpected country counts for clause %q: %+v", clause, counts)
		}

		devices, err := svc.getDeviceBreakdown(clause)
		if err != nil {
			t.Errorf("getDeviceBreakdown failed with clause %q: %v", clause, err)
			continue
		}
		if len(devices) != 2 {
			t.Errorf("expected 2 device types for clause %q, got %d", clause, len(devices))
		}

		devCounts := map[string]int64{}
		for _, d := range devices {
			devCounts[d.DeviceType] = d.SessionCount
		}
		if devCounts["desktop"] != 1 || devCounts["mobile"] != 1 {
			t.Errorf("unexpected device counts for clause %q: %+v", clause, devCounts)
		}
	}
}
