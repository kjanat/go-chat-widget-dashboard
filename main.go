package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"

	"github.com/kjanat/go-chat-widget-dashboard/internal/database"
	"github.com/kjanat/go-chat-widget-dashboard/internal/handlers"
	"github.com/kjanat/go-chat-widget-dashboard/internal/services"
)

//go:embed web
var staticFiles embed.FS

//go:embed web
var templateFiles embed.FS

func main() {
	log.Println("Starting Go Chat Widget Dashboard...")
	
	// Initialize database
	log.Println("Initializing database...")
	db, err := database.New("./db/chat_widget.db")
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()
	log.Println("Database initialized successfully")

	// Initialize services
	log.Println("Initializing services...")
	customerService := services.NewCustomerService(db.DB)
	chatService := services.NewChatService(db.DB)
	openaiService := services.NewOpenAIService()
	authService := services.NewAuthService(db.DB)
	analyticsService := services.NewAnalyticsService(db.DB)
	log.Println("Services initialized successfully")

	// Initialize session store
	log.Println("Initializing session store...")
	store := sessions.NewCookieStore([]byte(getEnvOrDefault("SESSION_SECRET", "your-secret-key")))
	log.Println("Session store initialized")

	// Load templates
	log.Println("Loading templates...")
	templates, err := loadTemplates()
	if err != nil {
		log.Fatal("Failed to load templates:", err)
	}
	log.Println("Templates loaded successfully")

	// Initialize handlers
	log.Println("Initializing handlers...")
	handler := handlers.New(
		customerService,
		chatService,
		openaiService,
		authService,
		analyticsService,
		store,
		templates,
	)
	log.Println("Handlers initialized successfully")

	// Setup routes
	log.Println("Setting up routes...")
	router := setupRoutes(handler)
	log.Println("Routes setup complete")

	// Start server
	port := getEnvOrDefault("PORT", "3000")
	log.Printf("Server starting on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}

func setupRoutes(h *handlers.Handler) *mux.Router {
	router := mux.NewRouter()

	// Widget routes
	router.HandleFunc("/widget.js", h.WidgetJS).Methods("GET")
	router.HandleFunc("/ws/{customerID}", h.WebSocket)
	router.HandleFunc("/api/widget/usage", h.WidgetUsage).Methods("POST")

	// Health and status endpoints
	router.HandleFunc("/health", h.Health).Methods("GET")
	router.HandleFunc("/api/status", h.APIStatus).Methods("GET")
	router.HandleFunc("/api/docs", h.APIDocs).Methods("GET")

	// Static files
	staticFS, _ := fs.Sub(staticFiles, "web/static")
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.FS(staticFS))))
	router.PathPrefix("/models/").Handler(http.StripPrefix("/models/", http.FileServer(http.Dir("./uploads/models/"))))

	// Admin routes
	admin := router.PathPrefix("/admin").Subrouter()
	admin.HandleFunc("/login", h.DashboardLogin).Methods("GET", "POST")
	admin.HandleFunc("/logout", h.DashboardLogout).Methods("GET")
	admin.HandleFunc("/", h.Dashboard).Methods("GET")
	admin.HandleFunc("/customers", h.CustomerCreate).Methods("POST")
	admin.HandleFunc("/customers/{id}", h.CustomerEdit).Methods("GET", "POST")
	admin.HandleFunc("/customers/{id}/model", h.ModelUpload).Methods("POST")
	admin.HandleFunc("/customers/{id}/model/download", h.ModelDownload).Methods("GET")
	admin.HandleFunc("/chat-logs", h.ChatLogs).Methods("GET")
	admin.HandleFunc("/analytics", h.Analytics).Methods("GET")
	admin.HandleFunc("/analytics/live-stats", h.AnalyticsLiveStats).Methods("GET")
	admin.HandleFunc("/analytics/generate-sample", h.GenerateSampleAnalytics).Methods("POST")

	// Landing page
	router.HandleFunc("/", h.Landing).Methods("GET")
	
	// Direct admin access
	router.HandleFunc("/dashboard", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/admin/", http.StatusSeeOther)
	})

	return router
}

func loadTemplates() (*template.Template, error) {
	// Try to use ParseFS approach
	templateFS, err := fs.Sub(templateFiles, "web/templates")
	if err != nil {
		return nil, fmt.Errorf("failed to create template filesystem: %v", err)
	}

	// Create template with custom functions
	tmpl := template.New("").Funcs(template.FuncMap{
		"toNewlines": func(s string) string {
			// Convert comma-separated values to newlines, or just return as-is if already has newlines
			if len(s) == 0 {
				return s
			}
			// If it already has newlines, return as-is
			if strings.Contains(s, "\n") {
				return s
			}
			// If it has commas but no newlines, convert commas to newlines
			if strings.Contains(s, ",") {
				return strings.ReplaceAll(s, ",", "\n")
			}
			return s
		},
		"json": func(v interface{}) string {
			// Simple JSON marshal for template use
			if v == nil {
				return "null"
			}
			b, err := json.Marshal(v)
			if err != nil {
				return "null"
			}
			return string(b)
		},
		"mul": func(a, b float64) float64 {
			return a * b
		},
	})

	// Parse all template files
	tmpl, err = tmpl.ParseFS(templateFS, "*.html", "*.js")
	if err != nil {
		return nil, fmt.Errorf("failed to parse templates: %v", err)
	}

	return tmpl, nil
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
