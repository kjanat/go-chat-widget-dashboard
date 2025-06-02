package main

import (
	"embed"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"os"

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
	// Initialize database
	db, err := database.New("./db/chat_widget.db")
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	defer db.Close()

	// Initialize services
	customerService := services.NewCustomerService(db.DB)
	chatService := services.NewChatService(db.DB)
	openaiService := services.NewOpenAIService()
	authService := services.NewAuthService(db.DB)

	// Initialize session store
	store := sessions.NewCookieStore([]byte(getEnvOrDefault("SESSION_SECRET", "your-secret-key")))

	// Load templates
	templates, err := loadTemplates()
	if err != nil {
		log.Fatal("Failed to load templates:", err)
	}

	// Initialize handlers
	handler := handlers.New(
		customerService,
		chatService,
		openaiService,
		authService,
		store,
		templates,
	)

	// Setup routes
	router := setupRoutes(handler)

	// Start server
	port := getEnvOrDefault("PORT", "8080")
	log.Printf("Server starting on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}

func setupRoutes(h *handlers.Handler) *mux.Router {
	router := mux.NewRouter()

	// Widget routes
	router.HandleFunc("/widget.js", h.WidgetJS).Methods("GET")
	router.HandleFunc("/ws/{customerID}", h.WebSocket)

	// Static files
	staticFS, _ := fs.Sub(staticFiles, "web/static")
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.FS(staticFS))))
	router.PathPrefix("/models/").Handler(http.StripPrefix("/models/", http.FileServer(http.Dir("./uploads/models/"))))

	// Admin routes
	admin := router.PathPrefix("/admin").Subrouter()
	admin.HandleFunc("/login", h.DashboardLogin).Methods("GET", "POST")
	admin.HandleFunc("/", h.Dashboard).Methods("GET")
	admin.HandleFunc("/customers", h.CustomerCreate).Methods("POST")
	admin.HandleFunc("/customers/{id}", h.CustomerEdit).Methods("GET", "POST")
	admin.HandleFunc("/customers/{id}/model", h.ModelUpload).Methods("POST")
	admin.HandleFunc("/chat-logs", h.ChatLogs).Methods("GET")

	// Root redirect
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
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

	// Parse all template files
	tmpl, err := template.ParseFS(templateFS, "*.html", "*.js")
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
