# Godis Web Todo Feature

A modern, responsive web-based todo list application built on top of the Godis Redis server. This feature demonstrates practical usage of Redis data structures while providing a user-friendly interface for task management.

## Features

### Core Functionality
- ✅ **CRUD Operations**: Create, read, update, and delete todos
- ✅ **Real-time Updates**: Instant UI updates with Redis persistence
- ✅ **Priority Management**: High, medium, and low priority levels
- ✅ **Due Date Support**: Set and track due dates with visual indicators
- ✅ **Search & Filter**: Full-text search and filtering by status/priority
- ✅ **Batch Operations**: Clear completed todos, bulk actions

### User Interface
- 🎨 **Modern Design**: Clean, responsive interface with dark/light themes
- 📱 **Mobile-First**: Optimized for all screen sizes
- ♿ **Accessibility**: WCAG compliant with keyboard navigation
- 🔔 **Toast Notifications**: User-friendly feedback system
- ⚡ **Progressive Enhancement**: Works without JavaScript (basic functionality)

### Technical Features
- 🚀 **High Performance**: Efficient Redis operations with caching
- 🔄 **Offline Support**: Queue operations when offline
- 🔐 **Security**: XSS protection, input validation
- 📊 **Analytics**: Real-time statistics and metrics
- 🐳 **Docker Ready**: Container support for easy deployment

## Architecture

### System Overview

```
┌─────────────────┐    HTTP    ┌──────────────────┐    Redis    ┌─────────────────┐
│   Web Browser   │ ◄────────► │   Web Server     │ ◄─────────► │   Godis Redis   │
│                 │            │   (Port 8080)    │             │   (Port 6399)   │
└─────────────────┘            └──────────────────┘             └─────────────────┘
│                              │                               │
├─ HTML/CSS/JS                 ├─ REST API                    ├─ Hash Storage
├─ Responsive UI               ├─ Static Files                ├─ List Indexing  
├─ PWA Features                ├─ CORS Handling               ├─ Priority Sets
└─ Offline Support             └─ Error Handling              └─ Expiration TTL
```

### Data Storage Strategy

The application uses Redis data structures optimally:

- **Todos**: Stored as Redis hashes (`todo:{user_id}:{todo_id}`)
- **User Lists**: Maintained as Redis lists (`todos:user:{user_id}`)
- **Counters**: Auto-increment IDs (`todo:counter:{user_id}`)
- **Indexes**: Priority and due date sorted sets for efficient querying

## Quick Start

### Option 1: Docker Compose (Recommended)

1. **Clone and start services**:
   ```bash
   git clone https://github.com/HDT3213/godis.git
   cd godis
   ./scripts/deploy.sh start
   ```

2. **Access the application**:
   - Web Todo: http://localhost:8080
   - Redis Server: localhost:6399

### Option 2: Manual Build

1. **Build the server**:
   ```bash
   ./scripts/build.sh --single
   ```

2. **Start the server**:
   ```bash
   ./bin/godis-webtodo-linux-amd64 --enable-web-todo
   ```

### Option 3: Development Mode

1. **Start Redis server** (existing Godis instance):
   ```bash
   go run main.go
   ```

2. **Start Web Todo server**:
   ```bash
   go run cmd/webtodo-server/main.go --enable-web-todo
   ```

## Configuration

### Command Line Options

```bash
godis-webtodo [options]

Options:
  --enable-web-todo          Enable Web Todo feature
  --web-todo-port PORT       Web server port (default: 8080)
  --config FILE              Redis configuration file
  --help                     Show help message
```

### Environment Variables

```bash
# Web Todo Configuration
ENABLE_WEB_TODO=true
WEB_TODO_PORT=8080
WEB_TODO_REDIS_ADDR=127.0.0.1:6399
WEB_TODO_STATIC_DIR=./webtodo/static

# Redis Configuration
REDIS_HOST=127.0.0.1
REDIS_PORT=6399
REDIS_PASSWORD=
```

### Configuration File

Create `webtodo.conf`:

```ini
# Redis Settings
bind 0.0.0.0
port 6399
databases 16

# Persistence
appendonly yes
appendfilename "appendonly.aof"

# Performance
maxclients 1000
```

## API Reference

### Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/todos` | List todos with filtering |
| POST | `/api/todos` | Create new todo |
| GET | `/api/todos/{id}` | Get specific todo |
| PUT | `/api/todos/{id}` | Update todo |
| DELETE | `/api/todos/{id}` | Delete todo |

### Request Examples

**Create Todo**:
```bash
curl -X POST http://localhost:8080/api/todos \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Complete project",
    "description": "Finish the Godis web todo feature",
    "priority": "high",
    "due_date": "2024-12-31T23:59:59Z"
  }'
```

**List Todos**:
```bash
curl "http://localhost:8080/api/todos?completed=false&priority=high&limit=10"
```

**Update Todo**:
```bash
curl -X PUT http://localhost:8080/api/todos/1 \
  -H "Content-Type: application/json" \
  -d '{"completed": true}'
```

## Development

### Project Structure

```
webtodo/
├── server.go              # HTTP server and routing
├── api_handler.go         # REST API handlers
├── todo_service.go        # Business logic and Redis operations
├── static/               # Frontend assets
│   ├── index.html        # Main HTML page
│   ├── css/              # Stylesheets
│   │   ├── main.css      # Base styles and variables
│   │   ├── components.css # UI components
│   │   └── responsive.css # Mobile responsiveness
│   └── js/               # JavaScript modules
│       ├── utils.js      # Utility functions
│       ├── api.js        # API client
│       ├── components.js # UI components
│       └── app.js        # Main application
└── cmd/
    └── webtodo-server/   # Main server entry point
```

### Building from Source

**Prerequisites**:
- Go 1.21+
- Node.js (for development tools)
- Docker (optional)

**Build Commands**:
```bash
# Build for current platform
./scripts/build.sh --single

# Build for all platforms
./scripts/build.sh

# Build with Docker
./scripts/build.sh --docker

# Development build with tests
./scripts/build.sh --mode debug --test --clean
```

### Development Workflow

1. **Start Redis server**:
   ```bash
   go run main.go
   ```

2. **Start web server with hot reload**:
   ```bash
   go run cmd/webtodo-server/main.go --enable-web-todo
   ```

3. **Make changes** to Go code or static files

4. **Test changes**:
   ```bash
   go test ./webtodo/...
   ```

## Deployment

### Production Deployment

1. **Using Docker Compose**:
   ```bash
   ./scripts/deploy.sh start --profile production
   ```

2. **Manual deployment**:
   ```bash
   # Build production binary
   ./scripts/build.sh --mode release
   
   # Deploy to server
   scp bin/godis-webtodo-linux-amd64 user@server:/opt/godis/
   scp -r webtodo/static user@server:/opt/godis/webtodo/
   
   # Start service
   ssh user@server 'cd /opt/godis && ./godis-webtodo-linux-amd64 --enable-web-todo'
   ```

### Nginx Configuration

For production deployment with Nginx:

```nginx
server {
    listen 80;
    server_name your-domain.com;
    
    location / {
        proxy_pass http://localhost:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_cache_bypass $http_upgrade;
    }
    
    location /static/ {
        alias /opt/godis/webtodo/static/;
        expires 1y;
        add_header Cache-Control "public, immutable";
    }
}
```

### Monitoring and Logging

Monitor the application:

```bash
# View logs
./scripts/deploy.sh logs --follow

# Check service status
./scripts/deploy.sh status

# Monitor Redis with RedisInsight
./scripts/deploy.sh start --profile monitoring
```

## Security Considerations

### Input Validation
- All user inputs are validated and sanitized
- XSS protection through HTML escaping
- SQL injection prevention (N/A for Redis)

### Network Security
- CORS configured for web requests
- Optional HTTPS support
- Redis AUTH support

### Data Protection
- No sensitive data stored in localStorage
- Session-based user identification
- Configurable data retention policies

## Performance Optimization

### Frontend Optimizations
- CSS and JS minification
- Image optimization
- Lazy loading for large lists
- Client-side caching with TTL

### Backend Optimizations
- Redis connection pooling
- Request/response caching
- Batch operations for bulk updates
- Efficient pagination

### Monitoring Metrics
- Response time tracking
- Error rate monitoring
- Redis operation metrics
- User engagement analytics

## Troubleshooting

### Common Issues

**Connection refused**:
```bash
# Check if Redis is running
redis-cli -p 6399 ping

# Check if web server is running
curl http://localhost:8080/api/todos
```

**Static files not loading**:
```bash
# Verify static directory exists
ls -la webtodo/static/

# Check file permissions
chmod -R 755 webtodo/static/
```

**Build failures**:
```bash
# Update dependencies
go mod tidy
go mod download

# Clean and rebuild
./scripts/build.sh --clean --build
```

### Debug Mode

Enable debug logging:
```bash
export DEBUG=true
go run cmd/webtodo-server/main.go --enable-web-todo
```

## Contributing

We welcome contributions! Please see our [Contributing Guide](CONTRIBUTING.md) for details.

### Development Setup

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

### Code Style

- Follow Go conventions
- Use `gofmt` for formatting
- Add comments for public functions
- Write tests for new features

## License

This project is licensed under the MIT License. See [LICENSE](LICENSE) for details.

## Acknowledgments

- Built on top of [Godis](https://github.com/HDT3213/godis) Redis server
- UI inspired by modern todo applications
- Icons from [Font Awesome](https://fontawesome.com/)
- Fonts from [Google Fonts](https://fonts.google.com/)

## Changelog

### v1.0.0 (2024-01-15)
- Initial release
- Core CRUD functionality
- Responsive web interface
- Docker support
- API documentation

---

For more information, visit the [Godis repository](https://github.com/HDT3213/godis) or check our [documentation](https://godis.io).