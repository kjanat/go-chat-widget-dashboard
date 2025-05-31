package main

import (
    "database/sql"
    "embed"
    "encoding/json"
    "fmt"
    "html/template"
    "io"
    "log"
    "net/http"
    "os"
    "path/filepath"
    "strings"
    "time"

    "github.com/gorilla/mux"
    "github.com/gorilla/sessions"
    "github.com/gorilla/websocket"
    _ "github.com/mattn/go-sqlite3"
    "golang.org/x/crypto/bcrypt"
)

//go:embed static/*
var staticFiles embed.FS

//go:embed templates/*
var templateFiles embed.FS

var (
    upgrader = websocket.Upgrader{
        CheckOrigin: func(r *http.Request) bool {
            // We'll validate origin against allowed domains
            origin := r.Header.Get("Origin")
            customerID := mux.Vars(r)["customerID"]
            return validateOrigin(origin, customerID)
        },
    }
    store = sessions.NewCookieStore([]byte("your-secret-key"))
)

// Database Models
type Customer struct {
    ID            string    `json:"id"`
    Name          string    `json:"name"`
    Email         string    `json:"email"`
    CreatedAt     time.Time `json:"created_at"`
    BrandColors   string    `json:"brand_colors"` // JSON string
    LogoURL       string    `json:"logo_url"`
    ModelPath     string    `json:"model_path"`
    Animations    string    `json:"animations"` // JSON string
    OpenAIPrompt  string    `json:"openai_prompt"`
    AllowedDomains string   `json:"allowed_domains"` // Comma-separated
    APIKey        string    `json:"api_key"`
    Active        bool      `json:"active"`
}

type ChatSession struct {
    ID         string    `json:"id"`
    CustomerID string    `json:"customer_id"`
    UserID     string    `json:"user_id"`
    StartedAt  time.Time `json:"started_at"`
    EndedAt    *time.Time `json:"ended_at"`
    Metadata   string    `json:"metadata"` // JSON string
}

type ChatMessage struct {
    ID         string    `json:"id"`
    SessionID  string    `json:"session_id"`
    Role       string    `json:"role"` // "user" or "assistant"
    Content    string    `json:"content"`
    Emotion    string    `json:"emotion"`
    CreatedAt  time.Time `json:"created_at"`
}

type AdminUser struct {
    ID           string
    Username     string
    PasswordHash string
    CreatedAt    time.Time
}

type Server struct {
    db     *sql.DB
    router *mux.Router
}

// Database initialization
func initDB() (*sql.DB, error) {
    db, err := sql.Open("sqlite3", "./chat_widget.db")
    if err != nil {
        return nil, err
    }

    // Create tables
    queries := []string{
        `CREATE TABLE IF NOT EXISTS customers (
            id TEXT PRIMARY KEY,
            name TEXT NOT NULL,
            email TEXT NOT NULL,
            created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
            brand_colors TEXT,
            logo_url TEXT,
            model_path TEXT,
            animations TEXT,
            openai_prompt TEXT,
            allowed_domains TEXT,
            api_key TEXT UNIQUE,
            active BOOLEAN DEFAULT 1
        )`,
        `CREATE TABLE IF NOT EXISTS chat_sessions (
            id TEXT PRIMARY KEY,
            customer_id TEXT NOT NULL,
            user_id TEXT NOT NULL,
            started_at DATETIME DEFAULT CURRENT_TIMESTAMP,
            ended_at DATETIME,
            metadata TEXT,
            FOREIGN KEY (customer_id) REFERENCES customers(id)
        )`,
        `CREATE TABLE IF NOT EXISTS chat_messages (
            id TEXT PRIMARY KEY,
            session_id TEXT NOT NULL,
            role TEXT NOT NULL,
            content TEXT NOT NULL,
            emotion TEXT,
            created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
            FOREIGN KEY (session_id) REFERENCES chat_sessions(id)
        )`,
        `CREATE TABLE IF NOT EXISTS admin_users (
            id TEXT PRIMARY KEY,
            username TEXT UNIQUE NOT NULL,
            password_hash TEXT NOT NULL,
            created_at DATETIME DEFAULT CURRENT_TIMESTAMP
        )`,
        `CREATE INDEX IF NOT EXISTS idx_sessions_customer ON chat_sessions(customer_id)`,
        `CREATE INDEX IF NOT EXISTS idx_messages_session ON chat_messages(session_id)`,
    }

    for _, query := range queries {
        if _, err := db.Exec(query); err != nil {
            return nil, err
        }
    }

    return db, nil
}

// Widget Handler - serves the JavaScript widget
func (s *Server) HandleWidgetJS(w http.ResponseWriter, r *http.Request) {
    customerID := r.URL.Query().Get("customer")
    if customerID == "" {
        http.Error(w, "Customer ID required", http.StatusBadRequest)
        return
    }

    // Validate API key if provided
    apiKey := r.URL.Query().Get("key")
    
    var customer Customer
    err := s.db.QueryRow(`
        SELECT id, brand_colors, logo_url, model_path, animations, active, api_key
        FROM customers WHERE id = ? AND active = 1
    `, customerID).Scan(
        &customer.ID, &customer.BrandColors, &customer.LogoURL,
        &customer.ModelPath, &customer.Animations, &customer.Active, &customer.APIKey,
    )
    
    if err != nil || !customer.Active {
        http.Error(w, "Invalid customer", http.StatusNotFound)
        return
    }

    if customer.APIKey != "" && customer.APIKey != apiKey {
        http.Error(w, "Invalid API key", http.StatusUnauthorized)
        return
    }

    // Parse template and inject customer config
    tmpl := template.Must(template.New("widget.js").Parse(widgetJSTemplate))
    
    w.Header().Set("Content-Type", "application/javascript")
    w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
    
    data := map[string]interface{}{
        "CustomerID":   customer.ID,
        "BrandColors":  customer.BrandColors,
        "LogoURL":      customer.LogoURL,
        "ModelPath":    customer.ModelPath,
        "Animations":   customer.Animations,
        "WSEndpoint":   fmt.Sprintf("wss://%s/ws/%s", r.Host, customer.ID),
    }
    
    tmpl.Execute(w, data)
}

// WebSocket handler for chat
func (s *Server) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
    customerID := mux.Vars(r)["customerID"]
    
    // Verify customer exists and is active
    var customer Customer
    err := s.db.QueryRow(`
        SELECT id, openai_prompt, active FROM customers WHERE id = ? AND active = 1
    `, customerID).Scan(&customer.ID, &customer.OpenAIPrompt, &customer.Active)
    
    if err != nil || !customer.Active {
        http.Error(w, "Invalid customer", http.StatusNotFound)
        return
    }

    conn, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        log.Println("WebSocket upgrade error:", err)
        return
    }
    defer conn.Close()

    // Create session
    sessionID := generateID()
    userID := r.Header.Get("X-User-ID")
    if userID == "" {
        userID = generateID()
    }

    _, err = s.db.Exec(`
        INSERT INTO chat_sessions (id, customer_id, user_id, metadata)
        VALUES (?, ?, ?, ?)
    `, sessionID, customerID, userID, "{}")
    
    if err != nil {
        log.Println("Error creating session:", err)
        return
    }

    // Handle messages
    for {
        var msg map[string]string
        err := conn.ReadJSON(&msg)
        if err != nil {
            break
        }

        // Log user message
        messageID := generateID()
        _, err = s.db.Exec(`
            INSERT INTO chat_messages (id, session_id, role, content)
            VALUES (?, ?, ?, ?)
        `, messageID, sessionID, "user", msg["content"])

        if err != nil {
            log.Println("Error logging message:", err)
            continue
        }

        // Call OpenAI (mock implementation)
        response, emotion := s.callOpenAI(customer.OpenAIPrompt, msg["content"])

        // Log assistant message
        assistantMessageID := generateID()
        _, err = s.db.Exec(`
            INSERT INTO chat_messages (id, session_id, role, content, emotion)
            VALUES (?, ?, ?, ?, ?)
        `, assistantMessageID, sessionID, "assistant", response, emotion)

        // Send response
        conn.WriteJSON(map[string]interface{}{
            "content": response,
            "emotion": emotion,
            "messageId": assistantMessageID,
        })
    }

    // End session
    s.db.Exec("UPDATE chat_sessions SET ended_at = CURRENT_TIMESTAMP WHERE id = ?", sessionID)
}

// Dashboard handlers
func (s *Server) HandleDashboardLogin(w http.ResponseWriter, r *http.Request) {
    if r.Method == "GET" {
        tmpl := template.Must(template.New("login").Parse(loginTemplate))
        tmpl.Execute(w, nil)
        return
    }

    username := r.FormValue("username")
    password := r.FormValue("password")

    var user AdminUser
    err := s.db.QueryRow("SELECT id, username, password_hash FROM admin_users WHERE username = ?", username).
        Scan(&user.ID, &user.Username, &user.PasswordHash)

    if err != nil || bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)) != nil {
        http.Redirect(w, r, "/admin/login?error=1", http.StatusSeeOther)
        return
    }

    session, _ := store.Get(r, "admin-session")
    session.Values["user_id"] = user.ID
    session.Values["username"] = user.Username
    session.Save(r, w)

    http.Redirect(w, r, "/admin/", http.StatusSeeOther)
}

func (s *Server) HandleDashboard(w http.ResponseWriter, r *http.Request) {
    session, _ := store.Get(r, "admin-session")
    if session.Values["user_id"] == nil {
        http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
        return
    }

    // Get all customers
    rows, err := s.db.Query("SELECT id, name, email, created_at, active FROM customers ORDER BY created_at DESC")
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    defer rows.Close()

    var customers []Customer
    for rows.Next() {
        var c Customer
        rows.Scan(&c.ID, &c.Name, &c.Email, &c.CreatedAt, &c.Active)
        customers = append(customers, c)
    }

    tmpl := template.Must(template.New("dashboard").Parse(dashboardTemplate))
    tmpl.Execute(w, map[string]interface{}{
        "Username":  session.Values["username"],
        "Customers": customers,
    })
}

func (s *Server) HandleCustomerEdit(w http.ResponseWriter, r *http.Request) {
    session, _ := store.Get(r, "admin-session")
    if session.Values["user_id"] == nil {
        http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
        return
    }

    customerID := mux.Vars(r)["id"]
    
    if r.Method == "GET" {
        var customer Customer
        err := s.db.QueryRow(`
            SELECT * FROM customers WHERE id = ?
        `, customerID).Scan(
            &customer.ID, &customer.Name, &customer.Email, &customer.CreatedAt,
            &customer.BrandColors, &customer.LogoURL, &customer.ModelPath,
            &customer.Animations, &customer.OpenAIPrompt, &customer.AllowedDomains,
            &customer.APIKey, &customer.Active,
        )

        if err != nil {
            http.Error(w, "Customer not found", http.StatusNotFound)
            return
        }

        tmpl := template.Must(template.New("edit").Parse(customerEditTemplate))
        tmpl.Execute(w, customer)
        return
    }

    // Handle POST - update customer
    _, err := s.db.Exec(`
        UPDATE customers SET
            name = ?, email = ?, brand_colors = ?, logo_url = ?,
            openai_prompt = ?, allowed_domains = ?, active = ?
        WHERE id = ?
    `,
        r.FormValue("name"), r.FormValue("email"), r.FormValue("brand_colors"),
        r.FormValue("logo_url"), r.FormValue("openai_prompt"),
        r.FormValue("allowed_domains"), r.FormValue("active") == "on",
        customerID,
    )

    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    http.Redirect(w, r, "/admin/", http.StatusSeeOther)
}

func (s *Server) HandleModelUpload(w http.ResponseWriter, r *http.Request) {
    session, _ := store.Get(r, "admin-session")
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
    defer file.Close()

    // Create models directory if it doesn't exist
    os.MkdirAll("./uploads/models", 0755)

    // Save file
    filename := fmt.Sprintf("%s_%s", customerID, handler.Filename)
    filepath := filepath.Join("./uploads/models", filename)
    
    dst, err := os.Create(filepath)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    defer dst.Close()

    _, err = io.Copy(dst, file)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    // Update customer record
    _, err = s.db.Exec("UPDATE customers SET model_path = ? WHERE id = ?", 
        "/models/"+filename, customerID)
    
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]string{"status": "success", "path": "/models/" + filename})
}

func (s *Server) HandleChatLogs(w http.ResponseWriter, r *http.Request) {
    session, _ := store.Get(r, "admin-session")
    if session.Values["user_id"] == nil {
        http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
        return
    }

    customerID := r.URL.Query().Get("customer")
    
    query := `
        SELECT s.id, s.user_id, s.started_at, s.ended_at, 
               COUNT(m.id) as message_count
        FROM chat_sessions s
        LEFT JOIN chat_messages m ON s.id = m.session_id
        WHERE 1=1
    `
    args := []interface{}{}
    
    if customerID != "" {
        query += " AND s.customer_id = ?"
        args = append(args, customerID)
    }
    
    query += " GROUP BY s.id ORDER BY s.started_at DESC LIMIT 100"
    
    rows, err := s.db.Query(query, args...)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    defer rows.Close()

    var sessions []map[string]interface{}
    for rows.Next() {
        var s ChatSession
        var messageCount int
        rows.Scan(&s.ID, &s.UserID, &s.StartedAt, &s.EndedAt, &messageCount)
        
        sessions = append(sessions, map[string]interface{}{
            "ID":           s.ID,
            "UserID":       s.UserID,
            "StartedAt":    s.StartedAt,
            "EndedAt":      s.EndedAt,
            "MessageCount": messageCount,
        })
    }

    tmpl := template.Must(template.New("logs").Parse(chatLogsTemplate))
    tmpl.Execute(w, map[string]interface{}{
        "Sessions":   sessions,
        "CustomerID": customerID,
    })
}

// Utility functions
func validateOrigin(origin, customerID string) bool {
    var allowedDomains string
    db, _ := sql.Open("sqlite3", "./chat_widget.db")
    defer db.Close()
    
    err := db.QueryRow("SELECT allowed_domains FROM customers WHERE id = ?", customerID).
        Scan(&allowedDomains)
    
    if err != nil {
        return false
    }

    // Check if origin matches any allowed domain
    domains := strings.Split(allowedDomains, ",")
    for _, domain := range domains {
        domain = strings.TrimSpace(domain)
        if domain == "*" || strings.Contains(origin, domain) {
            return true
        }
    }
    
    return false
}

func (s *Server) callOpenAI(prompt, message string) (string, string) {
    // Mock implementation - replace with actual OpenAI API call
    emotions := []string{"happy", "thinking", "neutral", "excited"}
    emotion := emotions[time.Now().Unix()%int64(len(emotions))]
    
    response := fmt.Sprintf("Response to: %s (with prompt: %s)", message, prompt[:20])
    return response, emotion
}

func generateID() string {
    return fmt.Sprintf("%d", time.Now().UnixNano())
}

// Templates
const widgetJSTemplate = `
(function() {
    // Configuration
    const config = {
        customerID: "{{.CustomerID}}",
        wsEndpoint: "{{.WSEndpoint}}",
        brandColors: {{.BrandColors}},
        logoURL: "{{.LogoURL}}",
        modelPath: "{{.ModelPath}}",
        animations: {{.Animations}}
    };

    // Create chat widget container
    const widgetContainer = document.getElementById('chat-widget');
    if (!widgetContainer) {
        console.error('Chat widget container not found');
        return;
    }

    // Inject styles
    const style = document.createElement('style');
    const colors = JSON.parse(config.brandColors || '{}');
    style.textContent = ` + "`" + `
        .chat-widget {
            position: fixed;
            bottom: 20px;
            right: 20px;
            width: 350px;
            height: 500px;
            background: ${colors.background || '#ffffff'};
            border-radius: 10px;
            box-shadow: 0 0 20px rgba(0,0,0,0.1);
            display: flex;
            flex-direction: column;
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
        }
        .chat-header {
            background: ${colors.primary || '#007bff'};
            color: ${colors.primaryText || '#ffffff'};
            padding: 15px;
            border-radius: 10px 10px 0 0;
            display: flex;
            align-items: center;
            gap: 10px;
        }
        .chat-logo {
            width: 30px;
            height: 30px;
            border-radius: 50%;
        }
        .chat-messages {
            flex: 1;
            overflow-y: auto;
            padding: 15px;
        }
        .chat-message {
            margin-bottom: 10px;
            padding: 10px;
            border-radius: 8px;
            max-width: 80%;
        }
        .chat-message.user {
            background: ${colors.userMessage || '#e3f2fd'};
            margin-left: auto;
        }
        .chat-message.assistant {
            background: ${colors.assistantMessage || '#f5f5f5'};
        }
        .chat-input-container {
            padding: 15px;
            border-top: 1px solid #eee;
        }
        .chat-input {
            width: 100%;
            padding: 10px;
            border: 1px solid #ddd;
            border-radius: 5px;
            font-size: 14px;
        }
        .chat-3d-container {
            height: 150px;
            background: #f0f0f0;
            position: relative;
        }
    ` + "`" + `;
    document.head.appendChild(style);

    // Create chat HTML
    widgetContainer.innerHTML = ` + "`" + `
        <div class="chat-widget">
            <div class="chat-header">
                ${config.logoURL ? '<img src="' + config.logoURL + '" class="chat-logo">' : ''}
                <h3>Chat Assistant</h3>
            </div>
            <div class="chat-3d-container" id="three-container"></div>
            <div class="chat-messages" id="chat-messages"></div>
            <div class="chat-input-container">
                <input type="text" class="chat-input" id="chat-input" placeholder="Type your message...">
            </div>
        </div>
    ` + "`" + `;

    // Initialize Three.js
    const threeContainer = document.getElementById('three-container');
    const scene = new THREE.Scene();
    const camera = new THREE.PerspectiveCamera(75, threeContainer.clientWidth / threeContainer.clientHeight, 0.1, 1000);
    const renderer = new THREE.WebGLRenderer({ alpha: true });
    renderer.setSize(threeContainer.clientWidth, threeContainer.clientHeight);
    threeContainer.appendChild(renderer.domElement);

    // Load 3D model
    let model, mixer;
    const loader = new THREE.GLTFLoader();
    if (config.modelPath) {
        loader.load(config.modelPath, (gltf) => {
            model = gltf.scene;
            scene.add(model);
            
            // Setup animations
            if (gltf.animations && gltf.animations.length) {
                mixer = new THREE.AnimationMixer(model);
                const animations = JSON.parse(config.animations || '{}');
                
                // Store animation clips
                window.chatAnimations = {};
                gltf.animations.forEach(clip => {
                    window.chatAnimations[clip.name] = mixer.clipAction(clip);
                });
            }
        });
    }

    // Setup camera and lights
    camera.position.z = 5;
    const light = new THREE.DirectionalLight(0xffffff, 1);
    light.position.set(0, 1, 1);
    scene.add(light);
    scene.add(new THREE.AmbientLight(0x404040));

    // Animation loop
    const clock = new THREE.Clock();
    function animate() {
        requestAnimationFrame(animate);
        if (mixer) mixer.update(clock.getDelta());
        renderer.render(scene, camera);
    }
    animate();

    // WebSocket connection
    const ws = new WebSocket(config.wsEndpoint);
    const messagesContainer = document.getElementById('chat-messages');
    const chatInput = document.getElementById('chat-input');

    // Generate or retrieve user ID
    let userID = localStorage.getItem('chat-user-id');
    if (!userID) {
        userID = 'user-' + Date.now() + '-' + Math.random().toString(36).substr(2, 9);
        localStorage.setItem('chat-user-id', userID);
    }

    ws.onopen = () => {
        console.log('Connected to chat server');
    };

    ws.onmessage = (event) => {
        const data = JSON.parse(event.data);
        
        // Add message to chat
        const messageEl = document.createElement('div');
        messageEl.className = 'chat-message assistant';
        messageEl.textContent = data.content;
        messagesContainer.appendChild(messageEl);
        messagesContainer.scrollTop = messagesContainer.scrollHeight;

        // Play animation based on emotion
        if (data.emotion && window.chatAnimations) {
            const animations = JSON.parse(config.animations || '{}');
            const animConfig = animations[data.emotion];
            if (animConfig && window.chatAnimations[animConfig.name]) {
                const action = window.chatAnimations[animConfig.name];
                action.reset();
                action.play();
            }
        }
    };

    // Handle input
    chatInput.addEventListener('keypress', (e) => {
        if (e.key === 'Enter' && chatInput.value.trim()) {
            // Add user message to chat
            const messageEl = document.createElement('div');
            messageEl.className = 'chat-message user';
            messageEl.textContent = chatInput.value;
            messagesContainer.appendChild(messageEl);
            messagesContainer.scrollTop = messagesContainer.scrollHeight;

            // Send message
            ws.send(JSON.stringify({
                content: chatInput.value,
                userID: userID
            }));

            chatInput.value = '';
        }
    });
})();
`

const loginTemplate = `
<!DOCTYPE html>
<html>
<head>
    <title>Admin Login</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            background: #f5f5f5;
            display: flex;
            justify-content: center;
            align-items: center;
            height: 100vh;
            margin: 0;
        }
        .login-form {
            background: white;
            padding: 40px;
            border-radius: 10px;
            box-shadow: 0 0 20px rgba(0,0,0,0.1);
            width: 300px;
        }
        h2 { margin-top: 0; }
        input {
            width: 100%;
            padding: 10px;
            margin-bottom: 15px;
            border: 1px solid #ddd;
            border-radius: 5px;
            box-sizing: border-box;
        }
        button {
            width: 100%;
            padding: 10px;
            background: #007bff;
            color: white;
            border: none;
            border-radius: 5px;
            cursor: pointer;
        }
        button:hover { background: #0056b3; }
        .error { color: red; margin-bottom: 15px; }
    </style>
</head>
<body>
    <div class="login-form">
        <h2>Admin Login</h2>
        {{if .Error}}<div class="error">Invalid credentials</div>{{end}}
        <form method="POST">
            <input type="text" name="username" placeholder="Username" required>
            <input type="password" name="password" placeholder="Password" required>
            <button type="submit">Login</button>
        </form>
    </div>
</body>
</html>
`

const dashboardTemplate = `
<!DOCTYPE html>
<html>
<head>
    <title>Chat Widget Dashboard</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            margin: 0;
            background: #f5f5f5;
        }
        .header {
            background: #007bff;
            color: white;
            padding: 20px;
            display: flex;
            justify-content: space-between;
            align-items: center;
        }
        .container {
            max-width: 1200px;
            margin: 20px auto;
            padding: 0 20px;
        }
        .card {
            background: white;
            border-radius: 10px;
            box-shadow: 0 0 10px rgba(0,0,0,0.1);
            padding: 20px;
            margin-bottom: 20px;
        }
        table {
            width: 100%;
            border-collapse: collapse;
        }
        th, td {
            padding: 12px;
            text-align: left;
            border-bottom: 1px solid #ddd;
        }
        th { background: #f8f9fa; }
        .btn {
            padding: 8px 16px;
            border-radius: 5px;
            text-decoration: none;
            display: inline-block;
            margin-right: 5px;
        }
        .btn-primary {
            background: #007bff;
            color: white;
        }
        .btn-secondary {
            background: #6c757d;
            color: white;
        }
        .btn-danger {
            background: #dc3545;
            color: white;
        }
        .status-active { color: #28a745; }
        .status-inactive { color: #dc3545; }
    </style>
</head>
<body>
    <div class="header">
        <h1>Chat Widget Dashboard</h1>
        <div>
            Welcome, {{.Username}} |
            <a href="/admin/logout" style="color: white;">Logout</a>
        </div>
    </div>
    
    <div class="container">
        <div class="card">
            <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 20px;">
                <h2>Customers</h2>
                <a href="/admin/customers/new" class="btn btn-primary">Add New Customer</a>
            </div>
            
            <table>
                <thead>
                    <tr>
                        <th>ID</th>
                        <th>Name</th>
                        <th>Email</th>
                        <th>Created</th>
                        <th>Status</th>
                        <th>Actions</th>
                    </tr>
                </thead>
                <tbody>
                    {{range .Customers}}
                    <tr>
                        <td>{{.ID}}</td>
                        <td>{{.Name}}</td>
                        <td>{{.Email}}</td>
                        <td>{{.CreatedAt.Format "2006-01-02"}}</td>
                        <td>
                            {{if .Active}}
                                <span class="status-active">Active</span>
                            {{else}}
                                <span class="status-inactive">Inactive</span>
                            {{end}}
                        </td>
                        <td>
                            <a href="/admin/customers/{{.ID}}/edit" class="btn btn-secondary">Edit</a>
                            <a href="/admin/customers/{{.ID}}/logs" class="btn btn-secondary">Logs</a>
                            <a href="#" onclick="copyWidget('{{.ID}}')" class="btn btn-primary">Get Code</a>
                        </td>
                    </tr>
                    {{end}}
                </tbody>
            </table>
        </div>
        
        <div class="card">
            <h2>Quick Stats</h2>
            <div style="display: grid; grid-template-columns: repeat(3, 1fr); gap: 20px;">
                <div>
                    <h3>Total Customers</h3>
                    <p style="font-size: 2em; margin: 0;">{{len .Customers}}</p>
                </div>
                <div>
                    <h3>Active Chats Today</h3>
                    <p style="font-size: 2em; margin: 0;">-</p>
                </div>
                <div>
                    <h3>Total Messages</h3>
                    <p style="font-size: 2em; margin: 0;">-</p>
                </div>
            </div>
        </div>
    </div>
    
    <script>
    function copyWidget(customerID) {
        const code = `<script src="${window.location.origin}/widget.js?customer=${customerID}"></script>\n<div id="chat-widget"></div>`;
        navigator.clipboard.writeText(code).then(() => {
            alert('Widget code copied to clipboard!');
        });
    }
    </script>
</body>
</html>
`

const customerEditTemplate = `
<!DOCTYPE html>
<html>
<head>
    <title>Edit Customer</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            margin: 0;
            background: #f5f5f5;
        }
        .header {
            background: #007bff;
            color: white;
            padding: 20px;
        }
        .container {
            max-width: 800px;
            margin: 20px auto;
            padding: 0 20px;
        }
        .card {
            background: white;
            border-radius: 10px;
            box-shadow: 0 0 10px rgba(0,0,0,0.1);
            padding: 30px;
        }
        .form-group {
            margin-bottom: 20px;
        }
        label {
            display: block;
            margin-bottom: 5px;
            font-weight: 500;
        }
        input[type="text"], input[type="email"], textarea {
            width: 100%;
            padding: 10px;
            border: 1px solid #ddd;
            border-radius: 5px;
            box-sizing: border-box;
        }
        textarea { min-height: 100px; resize: vertical; }
        .checkbox-group {
            display: flex;
            align-items: center;
            gap: 10px;
        }
        .btn {
            padding: 10px 20px;
            border-radius: 5px;
            border: none;
            cursor: pointer;
            text-decoration: none;
            display: inline-block;
            margin-right: 10px;
        }
        .btn-primary {
            background: #007bff;
            color: white;
        }
        .btn-secondary {
            background: #6c757d;
            color: white;
        }
        .code-block {
            background: #f8f9fa;
            padding: 15px;
            border-radius: 5px;
            font-family: monospace;
            margin-top: 10px;
        }
        .upload-section {
            border: 2px dashed #ddd;
            padding: 20px;
            text-align: center;
            margin-top: 10px;
            border-radius: 5px;
        }
    </style>
</head>
<body>
    <div class="header">
        <h1>Edit Customer: {{.Name}}</h1>
    </div>
    
    <div class="container">
        <div class="card">
            <form method="POST">
                <div class="form-group">
                    <label>Customer ID</label>
                    <input type="text" value="{{.ID}}" disabled>
                </div>
                
                <div class="form-group">
                    <label>Name</label>
                    <input type="text" name="name" value="{{.Name}}" required>
                </div>
                
                <div class="form-group">
                    <label>Email</label>
                    <input type="email" name="email" value="{{.Email}}" required>
                </div>
                
                <div class="form-group">
                    <label>Brand Colors (JSON)</label>
                    <textarea name="brand_colors">{{.BrandColors}}</textarea>
                    <small>Example: {"primary": "#007bff", "background": "#ffffff"}</small>
                </div>
                
                <div class="form-group">
                    <label>Logo URL</label>
                    <input type="text" name="logo_url" value="{{.LogoURL}}">
                </div>
                
                <div class="form-group">
                    <label>OpenAI Prompt</label>
                    <textarea name="openai_prompt">{{.OpenAIPrompt}}</textarea>
                </div>
                
                <div class="form-group">
                    <label>Allowed Domains (comma-separated)</label>
                    <input type="text" name="allowed_domains" value="{{.AllowedDomains}}">
                    <small>Use * to allow all domains</small>
                </div>
                
                <div class="form-group">
                    <label>API Key</label>
                    <input type="text" value="{{.APIKey}}" disabled>
                    <button type="button" onclick="regenerateKey()" class="btn btn-secondary">Regenerate</button>
                </div>
                
                <div class="form-group checkbox-group">
                    <input type="checkbox" name="active" id="active" {{if .Active}}checked{{end}}>
                    <label for="active">Active</label>
                </div>
                
                <button type="submit" class="btn btn-primary">Save Changes</button>
                <a href="/admin/" class="btn btn-secondary">Cancel</a>
            </form>
            
            <hr style="margin: 30px 0;">
            
            <h3>3D Model</h3>
            <div class="upload-section">
                <p>Current model: {{if .ModelPath}}{{.ModelPath}}{{else}}None{{end}}</p>
                <input type="file" id="modelFile" accept=".glb,.gltf">
                <button onclick="uploadModel()" class="btn btn-primary">Upload Model</button>
            </div>
            
            <h3>Widget Code</h3>
            <div class="code-block">
                &lt;script src="{{.Host}}/widget.js?customer={{.ID}}&key={{.APIKey}}"&gt;&lt;/script&gt;<br>
                &lt;div id="chat-widget"&gt;&lt;/div&gt;
            </div>
        </div>
    </div>
    
    <script>
    function uploadModel() {
        const fileInput = document.getElementById('modelFile');
        const file = fileInput.files[0];
        if (!file) return;
        
        const formData = new FormData();
        formData.append('model', file);
        
        fetch('/admin/customers/{{.ID}}/upload-model', {
            method: 'POST',
            body: formData
        })
        .then(res => res.json())
        .then(data => {
            alert('Model uploaded successfully!');
            location.reload();
        })
        .catch(err => alert('Upload failed: ' + err));
    }
    
    function regenerateKey() {
        if (confirm('This will invalidate the current API key. Continue?')) {
            fetch('/admin/customers/{{.ID}}/regenerate-key', {method: 'POST'})
                .then(() => location.reload());
        }
    }
    </script>
</body>
</html>
`

const chatLogsTemplate = `
<!DOCTYPE html>
<html>
<head>
    <title>Chat Logs</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            margin: 0;
            background: #f5f5f5;
        }
        .header {
            background: #007bff;
            color: white;
            padding: 20px;
            display: flex;
            justify-content: space-between;
            align-items: center;
        }
        .container {
            max-width: 1200px;
            margin: 20px auto;
            padding: 0 20px;
        }
        .card {
            background: white;
            border-radius: 10px;
            box-shadow: 0 0 10px rgba(0,0,0,0.1);
            padding: 20px;
            margin-bottom: 20px;
        }
        table {
            width: 100%;
            border-collapse: collapse;
        }
        th, td {
            padding: 12px;
            text-align: left;
            border-bottom: 1px solid #ddd;
        }
        th { background: #f8f9fa; }
        .btn {
            padding: 8px 16px;
            border-radius: 5px;
            text-decoration: none;
            display: inline-block;
        }
        .btn-secondary {
            background: #6c757d;
            color: white;
        }
        .session-row:hover {
            background: #f8f9fa;
            cursor: pointer;
        }
        .messages-panel {
            display: none;
            margin-top: 20px;
            padding: 20px;
            background: #f8f9fa;
            border-radius: 5px;
        }
        .message {
            margin-bottom: 15px;
            padding: 10px;
            border-radius: 5px;
        }
        .message.user {
            background: #e3f2fd;
            margin-left: 20%;
        }
        .message.assistant {
            background: #f5f5f5;
            margin-right: 20%;
        }
    </style>
</head>
<body>
    <div class="header">
        <h1>Chat Logs</h1>
        <a href="/admin/" style="color: white;">Back to Dashboard</a>
    </div>
    
    <div class="container">
        <div class="card">
            <h2>Recent Sessions</h2>
            <table>
                <thead>
                    <tr>
                        <th>Session ID</th>
                        <th>User ID</th>
                        <th>Started</th>
                        <th>Duration</th>
                        <th>Messages</th>
                        <th>Actions</th>
                    </tr>
                </thead>
                <tbody>
                    {{range .Sessions}}
                    <tr class="session-row" onclick="toggleMessages('{{.ID}}')">
                        <td>{{.ID}}</td>
                        <td>{{.UserID}}</td>
                        <td>{{.StartedAt.Format "2006-01-02 15:04:05"}}</td>
                        <td>
                            {{if .EndedAt}}
                                {{.Duration}} min
                            {{else}}
                                Active
                            {{end}}
                        </td>
                        <td>{{.MessageCount}}</td>
                        <td>
                            <a href="#" onclick="loadMessages('{{.ID}}')" class="btn btn-secondary">View</a>
                        </td>
                    </tr>
                    <tr>
                        <td colspan="6">
                            <div id="messages-{{.ID}}" class="messages-panel"></div>
                        </td>
                    </tr>
                    {{end}}
                </tbody>
            </table>
        </div>
    </div>
    
    <script>
    function loadMessages(sessionId) {
        fetch(`/admin/sessions/${sessionId}/messages`)
            .then(res => res.json())
            .then(messages => {
                const panel = document.getElementById(`messages-${sessionId}`);
                panel.innerHTML = messages.map(m => `
                    <div class="message ${m.role}">
                        <strong>${m.role}:</strong> ${m.content}
                        <br><small>${new Date(m.created_at).toLocaleString()}</small>
                        ${m.emotion ? `<br><small>Emotion: ${m.emotion}</small>` : ''}
                    </div>
                `).join('');
                panel.style.display = 'block';
            });
    }
    
    function toggleMessages(sessionId) {
        const panel = document.getElementById(`messages-${sessionId}`);
        if (panel.style.display === 'block') {
            panel.style.display = 'none';
        } else {
            loadMessages(sessionId);
        }
    }
    </script>
</body>
</html>
`

// Main function
func main() {
    // Initialize database
    db, err := initDB()
    if err != nil {
        log.Fatal("Failed to initialize database:", err)
    }
    defer db.Close()

    // Create admin user if none exists
    var adminCount int
    db.QueryRow("SELECT COUNT(*) FROM admin_users").Scan(&adminCount)
    if adminCount == 0 {
        hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
        db.Exec(`
            INSERT INTO admin_users (id, username, password_hash)
            VALUES (?, ?, ?)
        `, generateID(), "admin", string(hashedPassword))
        log.Println("Created default admin user - username: admin, password: admin123")
    }

    // Create server
    server := &Server{
        db:     db,
        router: mux.NewRouter(),
    }

    // Public routes
    server.router.HandleFunc("/widget.js", server.HandleWidgetJS).Methods("GET")
    server.router.HandleFunc("/ws/{customerID}", server.HandleWebSocket)
    server.router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.FS(staticFiles))))
    server.router.PathPrefix("/models/").Handler(http.StripPrefix("/models/", http.FileServer(http.Dir("./uploads/models"))))

    // Admin routes
    admin := server.router.PathPrefix("/admin").Subrouter()
    admin.HandleFunc("/login", server.HandleDashboardLogin).Methods("GET", "POST")
    admin.HandleFunc("/logout", func(w http.ResponseWriter, r *http.Request) {
        session, _ := store.Get(r, "admin-session")
        session.Values = make(map[interface{}]interface{})
        session.Save(r, w)
        http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
    })
    admin.HandleFunc("/", server.HandleDashboard).Methods("GET")
    admin.HandleFunc("/customers/{id}/edit", server.HandleCustomerEdit).Methods("GET", "POST")
    admin.HandleFunc("/customers/{id}/upload-model", server.HandleModelUpload).Methods("POST")
    admin.HandleFunc("/customers/{id}/logs", server.HandleChatLogs).Methods("GET")
    admin.HandleFunc("/sessions/{id}/messages", func(w http.ResponseWriter, r *http.Request) {
        sessionID := mux.Vars(r)["id"]
        
        rows, err := server.db.Query(`
            SELECT id, role, content, emotion, created_at
            FROM chat_messages
            WHERE session_id = ?
            ORDER BY created_at ASC
        `, sessionID)
        
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        defer rows.Close()

        var messages []ChatMessage
        for rows.Next() {
            var m ChatMessage
            rows.Scan(&m.ID, &m.Role, &m.Content, &m.Emotion, &m.CreatedAt)
            messages = append(messages, m)
        }

        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(messages)
    })

    // Start server
    log.Println("Server starting on :8080")
    log.Fatal(http.ListenAndServe(":8080", server.router))
}
