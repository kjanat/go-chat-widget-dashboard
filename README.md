# Go-based chat widget system

Here's what's included:

## Key Features

### 1. **Chat Widget Management**
- Dynamic widget generation based on customer configuration
- Custom branding (colors, logos) per customer
- Three.js model integration with emotion-based animations
- WebSocket-based real-time chat
- Domain validation for security

### 2. **Data Storage & Logging**
- SQLite database for all data (easily switchable to PostgreSQL/MySQL)
- Complete chat session tracking
- Message logging with timestamps and emotions
- User tracking across sessions

### 3. **Admin Dashboard**
- Secure login system with bcrypt password hashing
- Customer management (create, edit, activate/deactivate)
- Upload custom 3D models per customer
- View chat logs and session history
- Generate embed codes with API keys
- Domain whitelist configuration

### 4. **Security Features**
- Origin validation against allowed domains
- API key authentication per customer
- Session-based admin authentication
- CORS protection for WebSocket connections

## How It Works

### For Customers
Customers embed the widget with just two lines:
```html
<script src="https://yourserver.com/widget.js?customer=CUSTOMER_ID&key=API_KEY"></script>
<div id="chat-widget"></div>
```

### For Admins
1. Login at `/admin/` (default: admin/admin123)
2. Create/manage customers
3. Upload 3D models
4. Set allowed domains
5. Configure OpenAI prompts
6. View chat logs

## Setup Instructions

1. **Install dependencies:**
```bash
go get github.com/gorilla/mux
go get github.com/gorilla/websocket
go get github.com/gorilla/sessions
go get github.com/mattn/go-sqlite3
go get golang.org/x/crypto/bcrypt
```

2. **Create directories:**
```bash
mkdir -p uploads/models
mkdir -p static
mkdir -p templates
```

3. **Add Three.js files to static directory:**
- Download three.min.js
- Download GLTFLoader.js

4. **Run the server:**
```bash
go run main.go
```

## Next Steps

To make this production-ready:

1. **Replace mock OpenAI calls** with actual API integration
2. **Add Redis** for WebSocket scaling across multiple servers
3. **Implement proper file storage** (S3/GCS) for models
4. **Add monitoring and analytics**
5. **Set up proper TLS/SSL**
6. **Add rate limiting**
7. **Implement backup strategies**

The entire system is self-contained in Go, serving all HTML/CSS/JS dynamically through templates. The admin dashboard provides complete control over customer configurations, and the widget automatically adapts to each customer's branding and 3D model settings.
