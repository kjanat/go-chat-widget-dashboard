package services

import (
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/kjanat/go-chat-widget-dashboard/internal/models"
)

// AnalyticsService handles analytics and metrics operations
type AnalyticsService struct {
	db *sql.DB
}

// NewAnalyticsService creates a new analytics service
func NewAnalyticsService(db *sql.DB) *AnalyticsService {
	return &AnalyticsService{db: db}
}

// RecordChatMetrics records analytics data for a chat session
func (s *AnalyticsService) RecordChatMetrics(metrics *models.ChatMetrics) error {
	query := `
		INSERT INTO chat_metrics (
			id, customer_id, session_id, user_agent, ip_address, country, city,
			session_start_time, session_end_time, message_count, response_time,
			satisfaction_score, conversion_event, page_url, referrer_url,
			device_type, browser_name, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := s.db.Exec(query,
		metrics.ID, metrics.CustomerID, metrics.SessionID, metrics.UserAgent,
		metrics.IPAddress, metrics.Country, metrics.City, metrics.SessionStartTime,
		metrics.SessionEndTime, metrics.MessageCount, metrics.ResponseTime,
		metrics.SatisfactionScore, metrics.ConversionEvent, metrics.PageURL,
		metrics.ReferrerURL, metrics.DeviceType, metrics.BrowserName, metrics.CreatedAt,
	)

	return err
}

// GetDashboardStats returns comprehensive dashboard statistics
func (s *AnalyticsService) GetDashboardStats(timeRange string) (*models.DashboardStats, error) {
	var whereClause string
	switch timeRange {
	case "today":
		whereClause = "WHERE DATE(created_at) = DATE('now')"
	case "week":
		whereClause = "WHERE created_at >= date('now', '-7 days')"
	case "month":
		whereClause = "WHERE created_at >= date('now', '-30 days')"
	default:
		whereClause = "" // All time
	}

	stats := &models.DashboardStats{}

	// Get basic statistics
	err := s.getBasicStats(stats, whereClause)
	if err != nil {
		return nil, err
	}

	// Get top countries
	stats.TopCountries, err = s.getTopCountries(whereClause)
	if err != nil {
		return nil, err
	}

	// Get hourly activity
	stats.HourlyActivity, err = s.getHourlyActivity(whereClause)
	if err != nil {
		return nil, err
	}

	// Get customer activity
	stats.CustomerActivity, err = s.getCustomerActivity(whereClause)
	if err != nil {
		return nil, err
	}

	// Get device breakdown
	stats.DeviceBreakdown, err = s.getDeviceBreakdown(whereClause)
	if err != nil {
		return nil, err
	}

	return stats, nil
}

// prependWhereOrAnd returns "WHERE" if the provided clause is empty,
// otherwise it appends " AND" to the existing clause. This helps build
// conditional SQL statements without duplicating logic across queries.
func (s *AnalyticsService) prependWhereOrAnd(whereClause string) string {
	if whereClause == "" {
		return "WHERE"
	}
	return whereClause + " AND"
}

func (s *AnalyticsService) getBasicStats(stats *models.DashboardStats, whereClause string) error {
	query := fmt.Sprintf(`
		SELECT 
			COUNT(DISTINCT session_id) as total_sessions,
			COUNT(DISTINCT CASE WHEN session_end_time IS NULL THEN session_id END) as active_sessions,
			COALESCE(SUM(message_count), 0) as total_messages,
			COALESCE(AVG(response_time), 0) as avg_response_time,
			COALESCE(AVG(satisfaction_score), 0) as avg_satisfaction,
			COALESCE(
				COUNT(CASE WHEN conversion_event != '' THEN 1 END) * 100.0 / 
				NULLIF(COUNT(DISTINCT session_id), 0), 0
			) as conversion_rate
		FROM chat_metrics %s
	`, whereClause)

	row := s.db.QueryRow(query)
	return row.Scan(
		&stats.TotalSessions,
		&stats.ActiveSessions,
		&stats.TotalMessages,
		&stats.AvgResponseTime,
		&stats.AvgSatisfaction,
		&stats.ConversionRate,
	)
}

func (s *AnalyticsService) getTopCountries(whereClause string) ([]models.CountryStats, error) {
	fullClause := s.prependWhereOrAnd(whereClause)

	query := fmt.Sprintf(`
               SELECT
                       country,
                       COUNT(DISTINCT session_id) as session_count,
                       COALESCE(SUM(message_count), 0) as message_count
               FROM chat_metrics
               %s country != ''
               GROUP BY country
               ORDER BY session_count DESC
               LIMIT 10
       `, fullClause)

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("error closing rows: %v", err)
		}
	}()

	var countries []models.CountryStats
	for rows.Next() {
		var country models.CountryStats
		err := rows.Scan(&country.Country, &country.SessionCount, &country.MessageCount)
		if err != nil {
			return nil, err
		}
		countries = append(countries, country)
	}

	return countries, nil
}

func (s *AnalyticsService) getHourlyActivity(whereClause string) ([]models.HourlyStats, error) {
	query := fmt.Sprintf(`
		SELECT 
			CAST(strftime('%%H', created_at) AS INTEGER) as hour,
			COUNT(DISTINCT session_id) as session_count,
			COALESCE(SUM(message_count), 0) as message_count
		FROM chat_metrics 
		%s
		GROUP BY hour
		ORDER BY hour
	`, whereClause)

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("error closing rows: %v", err)
		}
	}()

	hourlyMap := make(map[int]models.HourlyStats)
	for rows.Next() {
		var hour int
		var sessionCount, messageCount int64
		err := rows.Scan(&hour, &sessionCount, &messageCount)
		if err != nil {
			return nil, err
		}
		hourlyMap[hour] = models.HourlyStats{
			Hour:         hour,
			SessionCount: sessionCount,
			MessageCount: messageCount,
		}
	}

	// Fill in missing hours with zero values
	var hourlyActivity []models.HourlyStats
	for i := 0; i < 24; i++ {
		if stats, exists := hourlyMap[i]; exists {
			hourlyActivity = append(hourlyActivity, stats)
		} else {
			hourlyActivity = append(hourlyActivity, models.HourlyStats{
				Hour:         i,
				SessionCount: 0,
				MessageCount: 0,
			})
		}
	}

	return hourlyActivity, nil
}

func (s *AnalyticsService) getCustomerActivity(whereClause string) ([]models.CustomerStats, error) {
	query := fmt.Sprintf(`
		SELECT 
			cm.customer_id,
			c.name as customer_name,
			COUNT(DISTINCT cm.session_id) as session_count,
			COALESCE(SUM(cm.message_count), 0) as message_count,
			COALESCE(AVG(cm.satisfaction_score), 0) as avg_satisfaction,
			COALESCE(
				COUNT(CASE WHEN cm.conversion_event != '' THEN 1 END) * 100.0 / 
				NULLIF(COUNT(DISTINCT cm.session_id), 0), 0
			) as conversion_rate
		FROM chat_metrics cm
		JOIN customers c ON cm.customer_id = c.id
		%s
		GROUP BY cm.customer_id, c.name
		ORDER BY session_count DESC
	`, whereClause)

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("error closing rows: %v", err)
		}
	}()

	var customers []models.CustomerStats
	for rows.Next() {
		var customer models.CustomerStats
		err := rows.Scan(
			&customer.CustomerID,
			&customer.CustomerName,
			&customer.SessionCount,
			&customer.MessageCount,
			&customer.AvgSatisfaction,
			&customer.ConversionRate,
		)
		if err != nil {
			return nil, err
		}
		customers = append(customers, customer)
	}

	return customers, nil
}

func (s *AnalyticsService) getDeviceBreakdown(whereClause string) ([]models.DeviceStats, error) {
	fullClause := s.prependWhereOrAnd(whereClause)

	query := fmt.Sprintf(`
                SELECT
                        device_type,
                        COUNT(DISTINCT session_id) as session_count
               FROM chat_metrics
               %s device_type != ''
               GROUP BY device_type
               ORDER BY session_count DESC
       `, fullClause)

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("error closing rows: %v", err)
		}
	}()

	var devices []models.DeviceStats
	var totalSessions int64

	for rows.Next() {
		var device models.DeviceStats
		err := rows.Scan(&device.DeviceType, &device.SessionCount)
		if err != nil {
			return nil, err
		}
		devices = append(devices, device)
		totalSessions += device.SessionCount
	}

	// Calculate percentages
	for i := range devices {
		if totalSessions > 0 {
			devices[i].Percentage = float64(devices[i].SessionCount) * 100 / float64(totalSessions)
		}
	}

	return devices, nil
}

// GetActiveSessionsCount returns the current number of active sessions
func (s *AnalyticsService) GetActiveSessionsCount() (int64, error) {
	query := `
		SELECT COUNT(DISTINCT session_id)
		FROM chat_metrics 
		WHERE session_end_time IS NULL
	`

	var count int64
	err := s.db.QueryRow(query).Scan(&count)
	return count, err
}

// GenerateSampleData creates sample analytics data for demonstration
func (s *AnalyticsService) GenerateSampleData() error {
	log.Println("Generating sample analytics data...")

	// Get all customers
	customers, err := s.getCustomers()
	if err != nil {
		return err
	}

	if len(customers) == 0 {
		log.Println("No customers found, skipping sample data generation")
		return nil
	}

	countries := []string{"United States", "Canada", "United Kingdom", "Germany", "France", "Japan", "Australia", "Brazil", "India", "China"}
	cities := []string{"New York", "London", "Toronto", "Berlin", "Paris", "Tokyo", "Sydney", "São Paulo", "Mumbai", "Beijing"}
	devices := []string{"desktop", "mobile", "tablet"}
	browsers := []string{"Chrome", "Firefox", "Safari", "Edge", "Opera"}
	events := []string{"", "", "", "purchase", "signup", "download"} // More empty for realistic conversion rates

	// Generate data for the last 30 days
	now := time.Now()
	for i := 0; i < 500; i++ {
		// Random time in the last 30 days
		randomHours := rand.Intn(30 * 24)
		sessionStart := now.Add(-time.Duration(randomHours) * time.Hour)

		// Random session duration (some ongoing)
		var sessionEnd *time.Time
		if rand.Float32() < 0.8 { // 80% of sessions are completed
			duration := time.Duration(rand.Intn(3600)) * time.Second
			endTime := sessionStart.Add(duration)
			sessionEnd = &endTime
		}

		customer := customers[rand.Intn(len(customers))]

		metrics := &models.ChatMetrics{
			ID:               fmt.Sprintf("metric_%d_%d", i, time.Now().UnixNano()),
			CustomerID:       customer.ID,
			SessionID:        fmt.Sprintf("session_%d_%d", i, time.Now().UnixNano()),
			UserAgent:        fmt.Sprintf("Mozilla/5.0 (%s)", browsers[rand.Intn(len(browsers))]),
			IPAddress:        fmt.Sprintf("192.168.%d.%d", rand.Intn(255), rand.Intn(255)),
			Country:          countries[rand.Intn(len(countries))],
			City:             cities[rand.Intn(len(cities))],
			SessionStartTime: sessionStart,
			SessionEndTime:   sessionEnd,
			MessageCount:     rand.Intn(20) + 1,
			ResponseTime:     rand.Intn(5000) + 500, // 500-5500ms
			PageURL:          fmt.Sprintf("https://example.com/page%d", rand.Intn(10)),
			ReferrerURL:      fmt.Sprintf("https://google.com/search?q=query%d", rand.Intn(100)),
			DeviceType:       devices[rand.Intn(len(devices))],
			BrowserName:      browsers[rand.Intn(len(browsers))],
			CreatedAt:        sessionStart,
		}

		// Add satisfaction score for completed sessions
		if sessionEnd != nil && rand.Float32() < 0.6 { // 60% provide feedback
			score := rand.Intn(5) + 1
			metrics.SatisfactionScore = &score
		}

		// Add conversion events occasionally
		if rand.Float32() < 0.1 { // 10% conversion rate
			metrics.ConversionEvent = events[rand.Intn(len(events))]
		}

		err := s.RecordChatMetrics(metrics)
		if err != nil {
			log.Printf("Error inserting sample metric: %v", err)
		}
	}

	log.Println("Sample analytics data generated successfully!")
	return nil
}

func (s *AnalyticsService) getCustomers() ([]models.Customer, error) {
	query := `SELECT id, name FROM customers WHERE active = 1`
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("error closing rows: %v", err)
		}
	}()

	var customers []models.Customer
	for rows.Next() {
		var customer models.Customer
		err := rows.Scan(&customer.ID, &customer.Name)
		if err != nil {
			return nil, err
		}
		customers = append(customers, customer)
	}

	return customers, nil
}
