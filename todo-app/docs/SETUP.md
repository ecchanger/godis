# Todo List Application Configuration

## Development Environment Setup

### Backend Configuration
- **Port**: 8080
- **Database**: Godis (in-memory Redis-compatible storage)
- **CORS**: Enabled for frontend development
- **Logging**: Console logging enabled

### Frontend Configuration
- **Port**: 3000
- **Proxy**: API requests proxied to backend (localhost:8080)
- **Hot Reload**: Enabled for development

### Environment Variables

#### Backend (.env file in backend directory - optional)
```bash
# Server Configuration
PORT=8080
GIN_MODE=debug

# Database Configuration
DB_TYPE=godis

# CORS Configuration
CORS_ORIGINS=http://localhost:3000,http://localhost:5173
```

#### Frontend (.env file in frontend directory - optional)
```bash
# API Configuration
VITE_API_BASE_URL=http://localhost:8080/api

# Development Configuration
VITE_APP_NAME=Todo List App
VITE_APP_VERSION=1.0.0
```

## Production Deployment

### Backend Deployment
1. Build the binary:
   ```bash
   cd backend
   CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o todo-backend .
   ```

2. Set production environment variables:
   ```bash
   export GIN_MODE=release
   export PORT=8080
   ```

3. Run the binary:
   ```bash
   ./todo-backend
   ```

### Frontend Deployment
1. Build for production:
   ```bash
   cd frontend
   npm run build
   ```

2. Serve static files using nginx, Apache, or any static file server.

### Docker Deployment (Coming Soon)

```yaml
# docker-compose.yml
version: '3.8'
services:
  backend:
    build: ./backend
    ports:
      - "8080:8080"
    environment:
      - GIN_MODE=release
  
  frontend:
    build: ./frontend
    ports:
      - "80:80"
    depends_on:
      - backend
```

## Troubleshooting

### Common Issues

1. **Port already in use**:
   ```bash
   # Find process using port 8080
   lsof -i :8080
   # Kill the process
   kill -9 <PID>
   ```

2. **Go module issues**:
   ```bash
   cd backend
   go mod tidy
   go mod download
   ```

3. **Node.js dependency issues**:
   ```bash
   cd frontend
   rm -rf node_modules package-lock.json
   npm install
   ```

4. **CORS issues**:
   - Ensure backend CORS is configured correctly
   - Check frontend proxy configuration in vite.config.js

### Performance Optimization

#### Backend
- Use `gin.SetMode(gin.ReleaseMode)` in production
- Implement proper logging levels
- Add rate limiting for production use

#### Frontend
- Implement code splitting for large applications
- Use React.memo for expensive components
- Optimize bundle size with tree shaking

### Security Considerations

1. **Input Validation**: All user inputs are validated on both client and server
2. **CORS**: Configure appropriate origins for production
3. **Rate Limiting**: Implement rate limiting in production
4. **HTTPS**: Use HTTPS in production environments

### Monitoring and Logging

#### Backend Logging
```go
// Use structured logging in production
log.Printf("[%s] %s %s - %d", method, path, ip, statusCode)
```

#### Frontend Error Tracking
```javascript
// Integrate with error tracking services
window.addEventListener('error', (event) => {
  console.error('Unhandled error:', event.error);
});
```