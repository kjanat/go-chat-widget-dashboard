package models

import (
	"time"
)

// ChatMetrics represents analytics data for chat interactions
type ChatMetrics struct {
	ID                string     `json:"id" db:"id"`
	CustomerID        string     `json:"customer_id" db:"customer_id"`
	SessionID         string     `json:"session_id" db:"session_id"`
	UserAgent         string     `json:"user_agent" db:"user_agent"`
	IPAddress         string     `json:"ip_address" db:"ip_address"`
	Country           string     `json:"country" db:"country"`
	City              string     `json:"city" db:"city"`
	SessionStartTime  time.Time  `json:"session_start_time" db:"session_start_time"`
	SessionEndTime    *time.Time `json:"session_end_time" db:"session_end_time"`
	MessageCount      int        `json:"message_count" db:"message_count"`
	ResponseTime      int        `json:"response_time" db:"response_time"`           // in milliseconds
	SatisfactionScore *int       `json:"satisfaction_score" db:"satisfaction_score"` // 1-5 rating
	ConversionEvent   string     `json:"conversion_event" db:"conversion_event"`     // e.g., "purchase", "signup"
	PageURL           string     `json:"page_url" db:"page_url"`
	ReferrerURL       string     `json:"referrer_url" db:"referrer_url"`
	DeviceType        string     `json:"device_type" db:"device_type"` // mobile, desktop, tablet
	BrowserName       string     `json:"browser_name" db:"browser_name"`
	CreatedAt         time.Time  `json:"created_at" db:"created_at"`
}

// DashboardStats represents aggregated statistics for the dashboard
type DashboardStats struct {
	TotalSessions    int64           `json:"total_sessions"`
	ActiveSessions   int64           `json:"active_sessions"`
	TotalMessages    int64           `json:"total_messages"`
	AvgResponseTime  float64         `json:"avg_response_time"`
	AvgSatisfaction  float64         `json:"avg_satisfaction"`
	ConversionRate   float64         `json:"conversion_rate"`
	TopCountries     []CountryStats  `json:"top_countries"`
	HourlyActivity   []HourlyStats   `json:"hourly_activity"`
	CustomerActivity []CustomerStats `json:"customer_activity"`
	DeviceBreakdown  []DeviceStats   `json:"device_breakdown"`
}

// CountryStats represents chat statistics by country
type CountryStats struct {
	Country      string `json:"country"`
	SessionCount int64  `json:"session_count"`
	MessageCount int64  `json:"message_count"`
}

// HourlyStats represents chat activity by hour
type HourlyStats struct {
	Hour         int   `json:"hour"`
	SessionCount int64 `json:"session_count"`
	MessageCount int64 `json:"message_count"`
}

// CustomerStats represents activity statistics per customer
type CustomerStats struct {
	CustomerID      string  `json:"customer_id"`
	CustomerName    string  `json:"customer_name"`
	SessionCount    int64   `json:"session_count"`
	MessageCount    int64   `json:"message_count"`
	AvgSatisfaction float64 `json:"avg_satisfaction"`
	ConversionRate  float64 `json:"conversion_rate"`
}

// DeviceStats represents usage statistics by device type
type DeviceStats struct {
	DeviceType   string  `json:"device_type"`
	SessionCount int64   `json:"session_count"`
	Percentage   float64 `json:"percentage"`
}

// RealtimeEvent represents a real-time event for the dashboard
type RealtimeEvent struct {
	Type       string      `json:"type"` // "new_session", "new_message", "session_end", "conversion"
	CustomerID string      `json:"customer_id"`
	SessionID  string      `json:"session_id"`
	Data       interface{} `json:"data"`
	Timestamp  time.Time   `json:"timestamp"`
}
