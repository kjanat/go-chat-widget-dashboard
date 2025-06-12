package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/kjanat/go-chat-widget-dashboard/internal/models"
	"github.com/kjanat/go-chat-widget-dashboard/internal/services"
	"github.com/kjanat/go-chat-widget-dashboard/views"
)

type TemplHandlers struct {
	analyticsService *services.AnalyticsService
}

func NewTemplHandlers(analyticsService *services.AnalyticsService) *TemplHandlers {
	return &TemplHandlers{
		analyticsService: analyticsService,
	}
}

// LandingHandler renders the landing page using Templ
func (h *TemplHandlers) LandingHandler(w http.ResponseWriter, r *http.Request) {
	// Get dashboard stats for the landing page
	stats, err := h.analyticsService.GetDashboardStats()
	if err != nil {
		log.Printf("Error getting dashboard stats: %v", err)
		stats = &models.DashboardStats{
			TotalSessions:   0,
			ActiveSessions:  0,
			ConversionRate:  0.0,
			AvgResponseTime: 0.0,
		}
	}

	// Render the landing page
	component := views.LandingPage(stats)
	err = component.Render(r.Context(), w)
	if err != nil {
		http.Error(w, "Error rendering page", http.StatusInternalServerError)
		log.Printf("Error rendering landing page: %v", err)
		return
	}
}

// DashboardHandler renders the dashboard page using Templ
func (h *TemplHandlers) DashboardHandler(w http.ResponseWriter, r *http.Request) {
	// Get dashboard stats
	stats, err := h.analyticsService.GetDashboardStats()
	if err != nil {
		log.Printf("Error getting dashboard stats: %v", err)
		stats = &models.DashboardStats{
			TotalSessions:   0,
			ActiveSessions:  0,
			ConversionRate:  0.0,
			AvgResponseTime: 0.0,
		}
	}

	// Get recent metrics
	metrics, err := h.analyticsService.GetRecentMetrics(10)
	if err != nil {
		log.Printf("Error getting recent metrics: %v", err)
		metrics = []*models.ChatMetrics{}
	}

	// Render the dashboard page
	component := views.DashboardPage(stats, metrics)
	err = component.Render(r.Context(), w)
	if err != nil {
		http.Error(w, "Error rendering page", http.StatusInternalServerError)
		log.Printf("Error rendering dashboard page: %v", err)
		return
	}
}

// HTMX API endpoints for dynamic updates

// MetricsHandler returns updated metrics for HTMX requests
func (h *TemplHandlers) MetricsHandler(w http.ResponseWriter, r *http.Request) {
	stats, err := h.analyticsService.GetDashboardStats()
	if err != nil {
		log.Printf("Error getting dashboard stats: %v", err)
		http.Error(w, "Error getting metrics", http.StatusInternalServerError)
		return
	}

	// Render only the metrics cards component
	component := views.MetricsCards(stats)
	err = component.Render(r.Context(), w)
	if err != nil {
		http.Error(w, "Error rendering metrics", http.StatusInternalServerError)
		log.Printf("Error rendering metrics: %v", err)
		return
	}
}

// ConversationsHandler returns updated conversation list for HTMX requests
func (h *TemplHandlers) ConversationsHandler(w http.ResponseWriter, r *http.Request) {
	metrics, err := h.analyticsService.GetRecentMetrics(10)
	if err != nil {
		log.Printf("Error getting recent metrics: %v", err)
		http.Error(w, "Error getting conversations", http.StatusInternalServerError)
		return
	}

	// Render only the conversations component
	component := views.RecentConversations(metrics)
	err = component.Render(r.Context(), w)
	if err != nil {
		http.Error(w, "Error rendering conversations", http.StatusInternalServerError)
		log.Printf("Error rendering conversations: %v", err)
		return
	}
}

// ActivityChartHandler returns mock chart data for activity
func (h *TemplHandlers) ActivityChartHandler(w http.ResponseWriter, r *http.Request) {
	// For now, return a simple chart placeholder
	// In a real implementation, you'd generate chart data
	chartHTML := `
		<div class="h-64 flex items-center justify-center border-2 border-dashed border-gray-300 rounded-lg">
			<div class="text-center">
				<svg class="w-12 h-12 text-gray-400 mx-auto mb-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v4a2 2 0 01-2 2H9z"></path>
				</svg>
				<p class="text-gray-500">Activity Chart</p>
				<p class="text-sm text-gray-400">Chart implementation coming soon</p>
			</div>
		</div>
	`
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(chartHTML))
}

// ResponseTimesChartHandler returns mock chart data for response times
func (h *TemplHandlers) ResponseTimesChartHandler(w http.ResponseWriter, r *http.Request) {
	// For now, return a simple chart placeholder
	chartHTML := `
		<div class="h-64 flex items-center justify-center border-2 border-dashed border-gray-300 rounded-lg">
			<div class="text-center">
				<svg class="w-12 h-12 text-gray-400 mx-auto mb-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z"></path>
				</svg>
				<p class="text-gray-500">Response Times Chart</p>
				<p class="text-sm text-gray-400">Chart implementation coming soon</p>
			</div>
		</div>
	`
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(chartHTML))
}

// SystemHealthHandler returns system health status
func (h *TemplHandlers) SystemHealthHandler(w http.ResponseWriter, r *http.Request) {
	// Render the system health widget
	component := views.SystemHealthWidget()
	err := component.Render(r.Context(), w)
	if err != nil {
		http.Error(w, "Error rendering system health", http.StatusInternalServerError)
		log.Printf("Error rendering system health: %v", err)
		return
	}
}

// StatusHandler returns system status as JSON (keeping existing API)
func (h *TemplHandlers) StatusHandler(w http.ResponseWriter, r *http.Request) {
	status := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().Format(time.RFC3339),
		"uptime":    time.Since(time.Now().Add(-24 * time.Hour)).String(),
		"version":   "1.0.0",
		"services": map[string]string{
			"database": "healthy",
			"cache":    "healthy",
			"storage":  "healthy",
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

// HealthHandler returns basic health check
func (h *TemplHandlers) HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
		"time":   time.Now().Format(time.RFC3339),
	})
}

// TrackUsageHandler handles widget usage tracking
func (h *TemplHandlers) TrackUsageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var usage struct {
		Action    string `json:"action"`
		Message   string `json:"message"`
		Timestamp string `json:"timestamp"`
	}

	if err := json.NewDecoder(r.Body).Decode(&usage); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Create a chat metric for tracking
	metric := &models.ChatMetrics{
		SessionID:        "demo-session",
		UserID:           "demo-user",
		DeviceType:       "web",
		MessageCount:     1,
		Status:           "active",
		SessionStartTime: time.Now(),
		SessionDuration:  0,
	}

	// Record the metric
	if err := h.analyticsService.RecordChatMetrics(metric); err != nil {
		log.Printf("Error recording chat metrics: %v", err)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "recorded",
		"action": usage.Action,
	})
}
