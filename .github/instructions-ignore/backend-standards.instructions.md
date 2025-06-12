---
applyTo: '**/*.go'
---

# Backend Development Standards

This document covers Go backend architecture, database patterns, and API development standards for the Go Chat Widget Dashboard.

## Architecture & Structure

### Clean Architecture Principles
- Follow clean architecture with clear separation of concerns
- Use dependency injection for services and handlers
- Structure code in layers: handlers → services → database
- Keep business logic in services, not handlers
- Use interfaces for testability and loose coupling

### Project Structure
```
internal/
├── handlers/     # HTTP handlers (controllers)
├── services/     # Business logic
├── models/       # Data models and entities
└── database/     # Data access layer
```

### Dependency Injection Pattern
```go
type Server struct {
    userService    services.UserService
    chatService    services.ChatService
    db            *gorm.DB
    logger        *logrus.Logger
}

func NewServer(db *gorm.DB, logger *logrus.Logger) *Server {
    return &Server{
        userService: services.NewUserService(db, logger),
        chatService: services.NewChatService(db, logger),
        db:         db,
        logger:     logger,
    }
}
```

## Code Standards

### Formatting and Naming
- Use `gofmt` and `goimports` for consistent formatting
- Follow effective Go naming conventions:
  - CamelCase for exported functions/types
  - camelCase for private functions/variables
  - Use descriptive names over comments where possible

### Context Usage
- Use `context.Context` for request scoping and cancellation
- Pass context as the first parameter in service methods
- Use context for database operations and HTTP requests

```go
func (s *UserService) CreateUser(ctx context.Context, user *models.User) error {
    // Use context for database operations
    return s.db.WithContext(ctx).Create(user).Error
}
```

### Error Handling
- Implement custom error types for better error handling
- Wrap errors with context using `fmt.Errorf`
- Use sentinel errors for common error conditions

```go
// Custom error types
type AppError struct {
    Code    int    `json:"code"`
    Message string `json:"message"`
    Err     error  `json:"-"`
}

func (e *AppError) Error() string {
    return e.Message
}

// Wrap errors with context
func (s *UserService) GetUser(ctx context.Context, id uint) (*models.User, error) {
    var user models.User
    if err := s.db.WithContext(ctx).First(&user, id).Error; err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, &AppError{Code: 404, Message: "User not found", Err: err}
        }
        return nil, fmt.Errorf("failed to get user %d: %w", id, err)
    }
    return &user, nil
}
```

### Logging
- Use structured logging with logrus or zap
- Include request IDs for tracing
- Log at appropriate levels (Debug, Info, Warn, Error)

```go
logger.WithFields(logrus.Fields{
    "user_id":    userID,
    "request_id": requestID,
    "action":     "create_user",
}).Info("Creating new user")
```

## Database Patterns

### GORM Best Practices
- Use GORM for ORM operations with proper migrations
- Implement database connection pooling
- Use transactions for multi-table operations
- Follow repository pattern for data access

### Model Definitions
```go
type User struct {
    ID        uint      `gorm:"primarykey" json:"id"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
    DeletedAt *time.Time `gorm:"index" json:"-"`
    
    Name     string `gorm:"not null" json:"name"`
    Email    string `gorm:"uniqueIndex;not null" json:"email"`
    Password string `gorm:"not null" json:"-"`
    
    // Relationships
    ChatSessions []ChatSession `json:"chat_sessions,omitempty"`
}
```

### Repository Pattern
```go
type UserRepository interface {
    Create(ctx context.Context, user *models.User) error
    GetByID(ctx context.Context, id uint) (*models.User, error)
    GetByEmail(ctx context.Context, email string) (*models.User, error)
    Update(ctx context.Context, user *models.User) error
    Delete(ctx context.Context, id uint) error
}

type userRepository struct {
    db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
    return &userRepository{db: db}
}
```

### Transactions
```go
func (s *ChatService) CreateChatSession(ctx context.Context, userID uint, message string) error {
    return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
        // Create chat session
        session := &models.ChatSession{UserID: userID}
        if err := tx.Create(session).Error; err != nil {
            return err
        }
        
        // Create initial message
        msg := &models.Message{
            ChatSessionID: session.ID,
            Content:       message,
            IsFromUser:    true,
        }
        return tx.Create(msg).Error
    })
}
```

## API Development

### HTTP Handlers
- Keep handlers thin - delegate to services
- Use proper HTTP status codes
- Implement consistent error responses
- Validate input data

```go
func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
    var user models.User
    if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
        http.Error(w, "Invalid JSON", http.StatusBadRequest)
        return
    }
    
    if err := h.userService.CreateUser(r.Context(), &user); err != nil {
        var appErr *AppError
        if errors.As(err, &appErr) {
            http.Error(w, appErr.Message, appErr.Code)
            return
        }
        http.Error(w, "Internal server error", http.StatusInternalServerError)
        return
    }
    
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(user)
}
```

### HTMX Integration
- Return partial HTML for HTMX requests
- Use appropriate HTMX headers
- Handle both JSON API and HTMX responses

```go
func (h *ChatHandler) GetChatLogs(w http.ResponseWriter, r *http.Request) {
    logs, err := h.chatService.GetRecentLogs(r.Context(), 50)
    if err != nil {
        if isHTMXRequest(r) {
            w.WriteHeader(http.StatusInternalServerError)
            fmt.Fprint(w, "<div class='error'>Failed to load chat logs</div>")
            return
        }
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
    if isHTMXRequest(r) {
        tmpl := template.Must(template.ParseFiles("web/templates/chat-logs-partial.html"))
        tmpl.Execute(w, logs)
        return
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(logs)
}

func isHTMXRequest(r *http.Request) bool {
    return r.Header.Get("HX-Request") == "true"
}
```

### Middleware
- Implement logging middleware
- Add authentication/authorization middleware
- Use recovery middleware for panic handling

```go
func LoggingMiddleware(logger *logrus.Logger) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            start := time.Now()
            requestID := uuid.New().String()
            
            ctx := context.WithValue(r.Context(), "request_id", requestID)
            r = r.WithContext(ctx)
            
            logger.WithFields(logrus.Fields{
                "method":     r.Method,
                "path":       r.URL.Path,
                "request_id": requestID,
            }).Info("Request started")
            
            next.ServeHTTP(w, r)
            
            logger.WithFields(logrus.Fields{
                "method":     r.Method,
                "path":       r.URL.Path,
                "request_id": requestID,
                "duration":   time.Since(start),
            }).Info("Request completed")
        })
    }
}
```

## WebSocket Implementation

### Real-time Chat
- Use gorilla/websocket for WebSocket connections
- Implement connection pooling for multiple clients
- Handle connection cleanup properly

```go
type Hub struct {
    clients    map[*Client]bool
    broadcast  chan []byte
    register   chan *Client
    unregister chan *Client
}

type Client struct {
    hub  *Hub
    conn *websocket.Conn
    send chan []byte
}

func (h *Hub) Run() {
    for {
        select {
        case client := <-h.register:
            h.clients[client] = true
            
        case client := <-h.unregister:
            if _, ok := h.clients[client]; ok {
                delete(h.clients, client)
                close(client.send)
            }
            
        case message := <-h.broadcast:
            for client := range h.clients {
                select {
                case client.send <- message:
                default:
                    close(client.send)
                    delete(h.clients, client)
                }
            }
        }
    }
}
```

## Configuration Management

### Environment Variables
- Use environment variables for configuration
- Implement configuration validation
- Use default values where appropriate

```go
type Config struct {
    Port        string `env:"PORT" envDefault:"8080"`
    DatabaseURL string `env:"DATABASE_URL" envDefault:"postgres://localhost/chatwidget"`
    JWTSecret   string `env:"JWT_SECRET,required"`
    LogLevel    string `env:"LOG_LEVEL" envDefault:"info"`
}

func LoadConfig() (*Config, error) {
    cfg := &Config{}
    if err := env.Parse(cfg); err != nil {
        return nil, fmt.Errorf("failed to parse config: %w", err)
    }
    return cfg, nil
}
```

### Graceful Shutdown
- Implement graceful shutdown for the server
- Close database connections properly
- Handle ongoing requests gracefully

```go
func (s *Server) Run() error {
    srv := &http.Server{
        Addr:    ":" + s.config.Port,
        Handler: s.router,
    }
    
    go func() {
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            s.logger.Fatalf("Server failed to start: %v", err)
        }
    }()
    
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit
    
    s.logger.Info("Shutting down server...")
    
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    return srv.Shutdown(ctx)
}
```
