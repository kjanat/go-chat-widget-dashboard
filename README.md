# Go Chat Widget Dashboard

A complete, production-ready chat widget system built with Go, featuring a 3D avatar, admin dashboard, and real-time WebSocket communication. This system allows businesses to embed a customizable chat widget on their websites with OpenAI integration and 3D model support.

## 🚀 Features

### Core Features
- **🎯 Embeddable Chat Widget**: Simple 2-line embed code for any website
- **🤖 AI-Powered Responses**: OpenAI integration with customizable prompts per customer
- **👤 3D Avatar Support**: Upload and display custom 3D models (.glb/.gltf) with emotion-based animations
- **⚡ Real-time Communication**: WebSocket-based instant messaging
- **🎨 Custom Branding**: Per-customer theming (colors, logos, animations)
- **🔒 Secure Multi-tenancy**: Domain validation, API keys, and customer isolation

### Admin Dashboard
- **📊 Customer Management**: Create, edit, and manage multiple customers
- **📈 Chat Analytics**: View chat logs, session history, and usage statistics
- **🎬 3D Model Upload**: Easy upload and management of customer avatars
- **🌐 Domain Control**: Whitelist allowed domains for widget embedding
- **🔐 Secure Authentication**: Session-based admin authentication with bcrypt

### Technical Features
- **🏗️ Clean Architecture**: Well-structured codebase with separation of concerns
- **🧪 Test-Driven Development**: Comprehensive test coverage
- **🔄 Hot Reload**: Development mode with automatic reloading
- **🐳 Docker Ready**: Complete containerization with Docker Compose
- **📦 Zero Dependencies**: Self-contained with embedded static files

## 📁 Project Structure

```
go-chat-widget-dashboard/
├── cmd/server/                 # Main application entry point
│   └── main.go
├── internal/                   # Private application code
│   ├── database/              # Database layer and migrations
│   ├── handlers/              # HTTP handlers (controllers)
│   ├── models/                # Data models and DTOs
│   └── services/              # Business logic layer
├── web/                       # Web assets
│   ├── templates/             # HTML templates and widget JS
│   └── static/               # CSS, JS, and other static files
├── scripts/                   # Utility scripts
├── uploads/                   # File uploads (3D models, etc.)
├── docker-compose.yml         # Container orchestration
├── Dockerfile                 # Container definition
├── Makefile                   # Build and development commands
└── .air.toml                  # Hot reload configuration
```

## 🛠️ Quick Start

### Prerequisites
- Go 1.21+
- SQLite3
- Make (optional, for easier commands)

### Installation

1. **Clone the repository**
   ```bash
   git clone https://github.com/kjanat/go-chat-widget-dashboard.git
   cd go-chat-widget-dashboard
   ```

2. **Install dependencies**
   ```bash
   make install
   # or manually:
   go mod download
   ```

3. **Initialize database and create admin user**
   ```bash
   make init-db
   # or manually:
   go run scripts/init-db.go
   ```

4. **Run the server**
   ```bash
   make run
   # or manually:
   go build -o bin/chat-widget-server ./cmd/server && ./bin/chat-widget-server
   ```

5. **Access the admin dashboard**
   - URL: `http://localhost:8080/admin/login`
   - Username: `admin`
   - Password: `admin123`

## 🎮 Usage

### For Administrators

1. **Login to Admin Dashboard**
   - Navigate to `/admin/login`
   - Use default credentials (change them immediately!)

2. **Create a Customer**
   - Click "Add Customer" on the dashboard
   - Fill in customer details and allowed domains
   - Configure branding colors and OpenAI prompt

3. **Upload 3D Model (Optional)**
   - Edit the customer
   - Upload a .glb or .gltf file for the avatar
   - Configure animation mappings

4. **Get Embed Code**
   - Click "Embed" on any customer card
   - Copy the provided HTML code

### For Website Integration

Add these two lines to any webpage:

```html
<script src="https://yourserver.com/widget.js?customer=CUSTOMER_ID&key=API_KEY"></script>
<div id="chat-widget"></div>
```

## 🧪 Development

### Development Mode with Hot Reload
```bash
make dev
```

### Running Tests
```bash
make test
# With coverage:
make test-coverage
```

### Code Quality
```bash
make fmt     # Format code
make lint    # Lint code
make security # Security scan
```

### Building for Production
```bash
make build
```

## 🐳 Docker Deployment

### Simple Docker Run
```bash
make docker
```

### Docker Compose (Recommended)
```bash
docker-compose up -d
```

### With Nginx Proxy
```bash
docker-compose --profile nginx up -d
```

## ⚙️ Configuration

### Environment Variables
- `PORT`: Server port (default: 8080)
- `SESSION_SECRET`: Session encryption key
- `DATABASE_URL`: Database connection string (optional)

### Customer Configuration
Each customer can be configured with:
- **Brand Colors**: JSON object defining widget theme
- **Logo URL**: Customer logo displayed in widget header
- **OpenAI Prompt**: Custom system prompt for AI responses
- **Allowed Domains**: Comma-separated list of authorized domains
- **3D Model**: Custom avatar with emotion-based animations

### Example Brand Colors JSON
```json
{
  "primary": "#007bff",
  "background": "#ffffff",
  "primaryText": "#ffffff",
  "userMessage": "#e3f2fd",
  "assistantMessage": "#f5f5f5"
}
```

### Example Animation Mapping JSON
```json
{
  "happy": {"name": "smile"},
  "thinking": {"name": "think"},
  "neutral": {"name": "idle"},
  "excited": {"name": "wave"}
}
```

## 🔧 API Endpoints

### Widget Endpoints
- `GET /widget.js?customer=ID&key=KEY` - Serve widget JavaScript
- `WS /ws/{customerID}` - WebSocket chat connection

### Admin Endpoints
- `GET /admin/login` - Admin login page
- `GET /admin/` - Dashboard
- `GET /admin/customers/{id}` - Edit customer
- `POST /admin/customers/{id}/model` - Upload 3D model
- `GET /admin/chat-logs` - View chat logs

## 🏗️ Architecture

### Clean Architecture Principles
- **Models**: Define data structures and interfaces
- **Services**: Implement business logic and external integrations
- **Handlers**: Handle HTTP requests and WebSocket connections
- **Database**: Manage data persistence and migrations

### Key Design Patterns
- **Dependency Injection**: Services are injected into handlers
- **Repository Pattern**: Database operations abstracted through services
- **Template Rendering**: Server-side HTML generation with Go templates
- **WebSocket Management**: Real-time bidirectional communication

## 🔒 Security Features

- **Domain Validation**: Restrict widget usage to authorized domains
- **API Key Authentication**: Secure customer identification
- **Session Management**: Secure admin authentication with bcrypt
- **CORS Protection**: Configurable cross-origin request handling
- **Input Validation**: Sanitization of all user inputs

## 🚧 Roadmap

### Immediate Improvements
- [ ] **OpenAI Integration**: Replace mock with actual API calls
- [ ] **Rate Limiting**: Implement request rate limiting
- [ ] **Analytics Dashboard**: Add usage analytics and metrics
- [ ] **Customer Registration**: Self-service customer onboarding

### Advanced Features
- [ ] **Redis Integration**: Scale WebSocket connections across servers
- [ ] **S3/GCS Support**: Cloud storage for 3D models
- [ ] **Multi-language Support**: Internationalization
- [ ] **Advanced Analytics**: Chat sentiment analysis and insights
- [ ] **A/B Testing**: Test different prompts and avatars

### Enterprise Features
- [ ] **SSO Integration**: SAML/OIDC authentication
- [ ] **Multi-tenant Database**: PostgreSQL with tenant isolation
- [ ] **Kubernetes Deployment**: Helm charts and operators
- [ ] **Monitoring & Alerting**: Prometheus/Grafana integration

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Write tests for your changes
4. Commit your changes (`git commit -m 'Add amazing feature'`)
5. Push to the branch (`git push origin feature/amazing-feature`)
6. Open a Pull Request

### Development Guidelines
- Follow Go best practices and idioms
- Write tests for all new functionality
- Update documentation for API changes
- Use conventional commit messages
- Ensure all tests pass before submitting

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- [Gorilla WebSocket](https://github.com/gorilla/websocket) - WebSocket implementation
- [Gorilla Mux](https://github.com/gorilla/mux) - HTTP router
- [Three.js](https://threejs.org/) - 3D graphics library
- [Bootstrap](https://getbootstrap.com/) - UI framework

## 📞 Support

For support, email support@yourcompany.com or create an issue on GitHub.

---

**Built with ❤️ using Go and modern web technologies**
