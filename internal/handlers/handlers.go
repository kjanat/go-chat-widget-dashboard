package handlers

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/gorilla/websocket"
	"golang.org/x/crypto/bcrypt"

	"github.com/kjanat/go-chat-widget-dashboard/internal/templates"

	"github.com/kjanat/go-chat-widget-dashboard/internal/models"
	"github.com/kjanat/go-chat-widget-dashboard/internal/services"
)

// Handler holds all the service dependencies
type Handler struct {
	customerService  *services.CustomerService
	chatService      *services.ChatService
	openaiService    *services.OpenAIService
	authService      *services.AuthService
	analyticsService *services.AnalyticsService
	store            *sessions.CookieStore
	upgrader         websocket.Upgrader
	templates        *template.Template
	startTime        time.Time
}

// New creates a new handler with all dependencies
func New(
	customerService *services.CustomerService,
	chatService *services.ChatService,
	openaiService *services.OpenAIService,
	authService *services.AuthService,
	analyticsService *services.AnalyticsService,
	store *sessions.CookieStore,
	templates *template.Template,
) *Handler {
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			origin := r.Header.Get("Origin")
			customerID := mux.Vars(r)["customerID"]
			return customerService.ValidateOrigin(origin, customerID)
		},
	}

	return &Handler{
		customerService:  customerService,
		chatService:      chatService,
		openaiService:    openaiService,
		authService:      authService,
		analyticsService: analyticsService,
		store:            store,
		upgrader:         upgrader,
		templates:        templates,
		startTime:        time.Now(),
	}
}

// WidgetJS serves the JavaScript widget
func (h *Handler) WidgetJS(w http.ResponseWriter, r *http.Request) {
	customerID := r.URL.Query().Get("customer")
	if customerID == "" {
		http.Error(w, "Customer ID required", http.StatusBadRequest)
		return
	}

	// Validate API key if provided
	apiKey := r.URL.Query().Get("key")

	customer, err := h.customerService.GetActiveByID(customerID)
	if err != nil {
		http.Error(w, "Invalid customer", http.StatusNotFound)
		return
	}

	if customer.APIKey != "" && customer.APIKey != apiKey {
		http.Error(w, "Invalid API key", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/javascript")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")

	config := models.WidgetConfig{
		CustomerID:  customer.ID,
		BrandColors: customer.BrandColorsToJSON(),
		LogoURL:     customer.LogoURL,
		ModelPath:   customer.ModelPath,
		Animations:  customer.Animations,
		WSEndpoint:  fmt.Sprintf("wss://%s/ws/%s", r.Host, customer.ID),
	}

	if err := h.templates.ExecuteTemplate(w, "widget.js", config); err != nil {
		log.Printf("error executing widget template: %v", err)
	}
}

// WebSocket handles chat WebSocket connections
func (h *Handler) WebSocket(w http.ResponseWriter, r *http.Request) {
	customerID := mux.Vars(r)["customerID"]

	customer, err := h.customerService.GetActiveByID(customerID)
	if err != nil {
		http.Error(w, "Invalid customer", http.StatusNotFound)
		return
	}

	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket upgrade error:", err)
		return
	}
	defer func() {
		if err := conn.Close(); err != nil {
			log.Printf("error closing websocket: %v", err)
		}
	}()

	// Create session
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		userID = fmt.Sprintf("user-%d", time.Now().UnixNano())
	}

	sessionID, err := h.chatService.CreateSession(customerID, userID)
	if err != nil {
		log.Println("Error creating session:", err)
		return
	}

	// Handle messages
	for {
		var msgReq models.ChatMessageRequest
		err := conn.ReadJSON(&msgReq)
		if err != nil {
			break
		}

		// Log user message
		_, err = h.chatService.CreateMessage(sessionID, "user", msgReq.Content, "")
		if err != nil {
			log.Println("Error logging message:", err)
			continue
		}

		// Generate AI response
		response, emotion := h.openaiService.GenerateResponse(customer.OpenAIPrompt, msgReq.Content)

		// Log assistant message
		messageID, err := h.chatService.CreateMessage(sessionID, "assistant", response, emotion)
		if err != nil {
			log.Println("Error logging assistant message:", err)
			continue
		}

		// Send response
		msgResp := models.ChatMessageResponse{
			Content:   response,
			Emotion:   emotion,
			MessageID: messageID,
		}
		if err := conn.WriteJSON(msgResp); err != nil {
			log.Println("error sending websocket message:", err)
			break
		}
	}

	// End session
	if err := h.chatService.EndSession(sessionID); err != nil {
		log.Println("error ending session:", err)
	}
}

// DashboardLogin handles admin login
func (h *Handler) DashboardLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		if err := templates.LoginPage(r).Render(r.Context(), w); err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")

	user, err := h.authService.GetAdminUser(username)
	if err != nil || bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)) != nil {
		http.Redirect(w, r, "/admin/login?error=1", http.StatusSeeOther)
		return
	}

	session, _ := h.store.Get(r, "admin-session")
	session.Values["user_id"] = user.ID
	session.Values["username"] = user.Username
	if err := session.Save(r, w); err != nil {
		log.Printf("error saving session: %v", err)
	}

	http.Redirect(w, r, "/admin/", http.StatusSeeOther)
}

// Dashboard handles the main admin dashboard
func (h *Handler) Dashboard(w http.ResponseWriter, r *http.Request) {
	session, _ := h.store.Get(r, "admin-session")
	if session.Values["user_id"] == nil {
		http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
		return
	}

	customers, err := h.customerService.GetAll()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"Username":  session.Values["username"],
		"Customers": customers,
	}

	if err := h.templates.ExecuteTemplate(w, "dashboard.html", data); err != nil {
		log.Printf("error executing dashboard template: %v", err)
	}
}

// DashboardLogout handles admin logout
func (h *Handler) DashboardLogout(w http.ResponseWriter, r *http.Request) {
	session, _ := h.store.Get(r, "admin-session")

	// Clear the session
	session.Values["user_id"] = nil
	session.Values["username"] = nil
	session.Options.MaxAge = -1 // This deletes the session
	if err := session.Save(r, w); err != nil {
		log.Printf("error saving session: %v", err)
	}

	http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
}

// CustomerEdit handles customer editing
func (h *Handler) CustomerEdit(w http.ResponseWriter, r *http.Request) {
	session, _ := h.store.Get(r, "admin-session")
	if session.Values["user_id"] == nil {
		http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
		return
	}

	customerID := mux.Vars(r)["id"]

	if r.Method == "GET" {
		customer, err := h.customerService.GetByID(customerID)
		if err != nil {
			http.Error(w, "Customer not found", http.StatusNotFound)
			return
		}

		if err := h.templates.ExecuteTemplate(w, "customer-edit.html", customer); err != nil {
			log.Printf("error executing customer edit template: %v", err)
		}
		return
	}

	// Handle POST - update customer
	customer := &models.Customer{
		Name:           r.FormValue("name"),
		Email:          r.FormValue("email"),
		BrandColors:    r.FormValue("brand_colors"),
		LogoURL:        r.FormValue("logo_url"),
		OpenAIPrompt:   r.FormValue("openai_prompt"),
		AllowedDomains: r.FormValue("allowed_domains"),
		Active:         r.FormValue("active") == "on",
	}

	err := h.customerService.Update(customerID, customer)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin/", http.StatusSeeOther)
}

// CustomerCreate handles customer creation
func (h *Handler) CustomerCreate(w http.ResponseWriter, r *http.Request) {
	session, _ := h.store.Get(r, "admin-session")
	if session.Values["user_id"] == nil {
		http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
		return
	}

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Create new customer
	customer := &models.Customer{
		Name:           r.FormValue("name"),
		Email:          r.FormValue("email"),
		BrandColors:    r.FormValue("brand_colors"),
		LogoURL:        r.FormValue("logo_url"),
		OpenAIPrompt:   r.FormValue("openai_prompt"),
		AllowedDomains: r.FormValue("allowed_domains"),
		Active:         true, // New customers are active by default
	}

	err := h.customerService.Create(customer)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin/", http.StatusSeeOther)
}

// ModelUpload handles 3D model uploads
func (h *Handler) ModelUpload(w http.ResponseWriter, r *http.Request) {
	session, _ := h.store.Get(r, "admin-session")
	if session.Values["user_id"] == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	customerID := mux.Vars(r)["id"]

	// Parse multipart form
	err := r.ParseMultipartForm(50 << 20) // 50MB max
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	file, handler, err := r.FormFile("model")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.Printf("error closing uploaded file: %v", err)
		}
	}()

	// Create models directory if it doesn't exist
	if err := os.MkdirAll("./uploads/models", 0755); err != nil {
		log.Printf("error creating models directory: %v", err)
	}

	// Save file
	filename := fmt.Sprintf("%s_%s", customerID, handler.Filename)
	filePath := filepath.Join("./uploads/models", filename)

	dst, err := os.Create(filePath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer func() {
		if err := dst.Close(); err != nil {
			log.Printf("error closing destination file: %v", err)
		}
	}()

	_, err = io.Copy(dst, file)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Update customer record
	modelPath := "/models/" + filename
	err = h.customerService.UpdateModelPath(customerID, modelPath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(map[string]string{
		"status": "success",
		"path":   modelPath,
	}); err != nil {
		log.Printf("error encoding model upload response: %v", err)
	}
}

// ModelDownload handles 3D model downloads
func (h *Handler) ModelDownload(w http.ResponseWriter, r *http.Request) {
	session, _ := h.store.Get(r, "admin-session")
	if session.Values["user_id"] == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	customerID := mux.Vars(r)["id"]

	customer, err := h.customerService.GetByID(customerID)
	if err != nil {
		http.Error(w, "Customer not found", http.StatusNotFound)
		return
	}

	if customer.ModelPath == "" {
		http.Error(w, "No model uploaded", http.StatusNotFound)
		return
	}

	// Get the actual file path
	filePath := filepath.Join("./uploads/models", filepath.Base(customer.ModelPath))

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		http.Error(w, "Model file not found", http.StatusNotFound)
		return
	}

	// Set headers for download
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filepath.Base(customer.ModelPath)))
	w.Header().Set("Content-Type", "application/octet-stream")

	// Serve the file
	http.ServeFile(w, r, filePath)
}

// ChatLogs handles chat log viewing
func (h *Handler) ChatLogs(w http.ResponseWriter, r *http.Request) {
	session, _ := h.store.Get(r, "admin-session")
	if session.Values["user_id"] == nil {
		http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
		return
	}

	customerID := r.URL.Query().Get("customer")

	sessions, err := h.chatService.GetSessions(customerID, 100)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"Sessions":   sessions,
		"CustomerID": customerID,
	}

	if err := h.templates.ExecuteTemplate(w, "chat-logs.html", data); err != nil {
		log.Printf("error executing chat logs template: %v", err)
	}
}

// Analytics handles the analytics dashboard
func (h *Handler) Analytics(w http.ResponseWriter, r *http.Request) {
	// Check authentication
	session, _ := h.store.Get(r, "admin-session")
	if session.Values["user_id"] == nil {
		http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
		return
	}

	// Get time range from query parameter
	timeRange := r.URL.Query().Get("range")
	if timeRange == "" {
		timeRange = "month" // default
	}

	// Get analytics data
	stats, err := h.analyticsService.GetDashboardStats(timeRange)
	if err != nil {
		log.Printf("Error getting analytics stats: %v", err)
		// Create empty stats to prevent template errors
		stats = &models.DashboardStats{
			TotalSessions:    0,
			ActiveSessions:   0,
			TotalMessages:    0,
			AvgResponseTime:  0,
			AvgSatisfaction:  0,
			ConversionRate:   0,
			TopCountries:     []models.CountryStats{},
			HourlyActivity:   []models.HourlyStats{},
			CustomerActivity: []models.CustomerStats{},
			DeviceBreakdown:  []models.DeviceStats{},
		}
	}

	data := map[string]interface{}{
		"Stats":     stats,
		"Username":  session.Values["username"],
		"TimeRange": timeRange,
	}

	// Add template functions for JSON serialization
	h.templates.Funcs(template.FuncMap{
		"json": func(v interface{}) string {
			b, _ := json.Marshal(v)
			return string(b)
		},
		"mul": func(a, b float64) float64 {
			return a * b
		},
	})

	err = h.templates.ExecuteTemplate(w, "analytics.html", data)
	if err != nil {
		log.Printf("Error executing analytics template: %v", err)
		http.Error(w, "Error rendering analytics page", http.StatusInternalServerError)
	}
}

// AnalyticsLiveStats provides real-time statistics for AJAX updates
func (h *Handler) AnalyticsLiveStats(w http.ResponseWriter, r *http.Request) {
	// Check authentication
	session, _ := h.store.Get(r, "admin-session")
	if session.Values["user_id"] == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get current active sessions
	activeSessions, err := h.analyticsService.GetActiveSessionsCount()
	if err != nil {
		log.Printf("Error getting active sessions: %v", err)
		activeSessions = 0
	}

	// Create mock recent activity for demonstration
	recentActivity := []map[string]interface{}{
		{
			"type":         "new_session",
			"description":  "New chat session started",
			"customerName": "Demo Customer",
			"timeAgo":      "Just now",
		},
	}

	response := map[string]interface{}{
		"activeSessions": activeSessions,
		"recentActivity": recentActivity,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("error encoding analytics live stats: %v", err)
	}
}

// GenerateSampleAnalytics creates sample data for demonstration
func (h *Handler) GenerateSampleAnalytics(w http.ResponseWriter, r *http.Request) {
	// Check authentication
	session, _ := h.store.Get(r, "admin-session")
	if session.Values["user_id"] == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	err := h.analyticsService.GenerateSampleData()
	if err != nil {
		http.Error(w, "Error generating sample data: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]string{"status": "success"}); err != nil {
		log.Printf("error encoding sample analytics response: %v", err)
	}
}

// Landing serves the marketing landing page
func (h *Handler) Landing(w http.ResponseWriter, r *http.Request) {
	// Get some basic stats for the landing page
	totalCustomers, _ := h.customerService.GetTotalCount()

	// Get analytics stats if available
	dashboardStats, _ := h.analyticsService.GetDashboardStats("all")

	data := map[string]interface{}{
		"TotalCustomers":  totalCustomers,
		"TotalSessions":   dashboardStats.TotalSessions,
		"ActiveSessions":  dashboardStats.ActiveSessions,
		"AvgResponseTime": dashboardStats.AvgResponseTime,
	}

	if err := h.templates.ExecuteTemplate(w, "landing.html", data); err != nil {
		log.Printf("error executing landing template: %v", err)
	}
}

// Health returns the application health status
func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	health := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().Unix(),
		"version":   "1.0.0",
		"uptime":    time.Since(h.startTime).String(),
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(health); err != nil {
		log.Printf("error encoding health response: %v", err)
	}
}

// APIStatus returns API statistics and system information
func (h *Handler) APIStatus(w http.ResponseWriter, r *http.Request) {
	// Get basic stats
	totalCustomers, _ := h.customerService.GetTotalCount()
	dashboardStats, _ := h.analyticsService.GetDashboardStats("all")

	status := map[string]interface{}{
		"api_version":       "v1.0.0",
		"status":            "operational",
		"timestamp":         time.Now().Unix(),
		"uptime":            time.Since(h.startTime).String(),
		"total_customers":   totalCustomers,
		"total_sessions":    dashboardStats.TotalSessions,
		"active_sessions":   dashboardStats.ActiveSessions,
		"avg_response_time": dashboardStats.AvgResponseTime,
		"database_status":   "connected",
		"memory_usage":      getMemoryUsage(),
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(status); err != nil {
		log.Printf("error encoding api status: %v", err)
	}
}

// APIDocs serves a simple API documentation page
func (h *Handler) APIDocs(w http.ResponseWriter, r *http.Request) {
	docs := map[string]interface{}{
		"title":       "Chat Widget API Documentation",
		"version":     "v1.0.0",
		"description": "REST API for the Go Chat Widget Dashboard",
		"endpoints": []map[string]interface{}{
			{
				"method":      "GET",
				"path":        "/health",
				"description": "Health check endpoint",
				"response":    "Returns application health status",
			},
			{
				"method":      "GET",
				"path":        "/api/status",
				"description": "System status and statistics",
				"response":    "Returns comprehensive system metrics",
			},
			{
				"method":      "GET",
				"path":        "/widget.js",
				"description": "Serve chat widget JavaScript",
				"parameters":  "customer (required), key (optional)",
				"response":    "JavaScript code for embedding chat widget",
			},
			{
				"method":      "WebSocket",
				"path":        "/ws/{customerID}",
				"description": "WebSocket connection for real-time chat",
				"response":    "Bidirectional chat messages",
			},
			{
				"method":      "GET",
				"path":        "/admin/analytics/live-stats",
				"description": "Live analytics data (authenticated)",
				"response":    "Real-time statistics for dashboard",
			},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(docs); err != nil {
		log.Printf("error encoding docs: %v", err)
	}
}

// WidgetUsage tracks widget usage statistics
func (h *Handler) WidgetUsage(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var usage struct {
		CustomerID string `json:"customer_id"`
		Event      string `json:"event"` // "load", "open", "message", "close"
		UserAgent  string `json:"user_agent"`
		PageURL    string `json:"page_url"`
		Timestamp  int64  `json:"timestamp"`
	}

	if err := json.NewDecoder(r.Body).Decode(&usage); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Record the usage event in analytics
	if h.analyticsService != nil {
		now := time.Now()
		metric := &models.ChatMetrics{
			ID:               fmt.Sprintf("usage-%d", now.UnixNano()),
			CustomerID:       usage.CustomerID,
			SessionID:        fmt.Sprintf("widget-%d", now.UnixNano()),
			MessageCount:     1,
			SessionStartTime: now,
			UserAgent:        usage.UserAgent,
			Country:          "Unknown", // Could be enhanced with IP geolocation
			DeviceType:       getDeviceFromUserAgent(usage.UserAgent),
			PageURL:          usage.PageURL,
			CreatedAt:        now,
		}

		// Don't block the response if analytics fails
		go func() {
			if err := h.analyticsService.RecordChatMetrics(metric); err != nil {
				log.Printf("error recording chat metrics: %v", err)
			}
		}()
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]string{"status": "recorded"}); err != nil {
		log.Printf("error encoding widget usage response: %v", err)
	}
}

func getDeviceFromUserAgent(userAgent string) string {
	userAgent = strings.ToLower(userAgent)
	if strings.Contains(userAgent, "mobile") || strings.Contains(userAgent, "android") || strings.Contains(userAgent, "iphone") {
		return "mobile"
	} else if strings.Contains(userAgent, "tablet") || strings.Contains(userAgent, "ipad") {
		return "tablet"
	}
	return "desktop"
}

func getMemoryUsage() map[string]interface{} {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return map[string]interface{}{
		"alloc_mb":       bToMb(m.Alloc),
		"total_alloc_mb": bToMb(m.TotalAlloc),
		"sys_mb":         bToMb(m.Sys),
		"num_gc":         m.NumGC,
	}
}

func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}
