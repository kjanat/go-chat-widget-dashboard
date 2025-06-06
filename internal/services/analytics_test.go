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
		if _, err := svc.getTopCountries(clause); err != nil {
			t.Errorf("getTopCountries failed with clause %q: %v", clause, err)
		}
		if _, err := svc.getDeviceBreakdown(clause); err != nil {
			t.Errorf("getDeviceBreakdown failed with clause %q: %v", clause, err)
		}
	}
}
