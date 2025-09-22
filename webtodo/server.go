package webtodo

import (
	"context"
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"
	"time"

	"github.com/hdt3213/godis/config"
	"github.com/hdt3213/godis/lib/logger"
	"github.com/hdt3213/godis/redis/client"
)

// WebTodoServer represents the HTTP server for the web todo application
type WebTodoServer struct {
	httpServer   *http.Server
	redisClient  *client.Client
	staticDir    string
	port         int
	redisAddr    string
	ctx          context.Context
	cancel       context.CancelFunc
}

// Config holds configuration for the web todo server
type Config struct {
	Port      int    `json:"port"`
	RedisAddr string `json:"redis_addr"`
	StaticDir string `json:"static_dir"`
}

// NewWebTodoServer creates a new web todo server instance
func NewWebTodoServer(cfg *Config) *WebTodoServer {
	if cfg.Port == 0 {
		cfg.Port = 8080
	}
	if cfg.RedisAddr == "" {
		cfg.RedisAddr = fmt.Sprintf("%s:%d", config.Properties.Bind, config.Properties.Port)
	}
	if cfg.StaticDir == "" {
		cfg.StaticDir = "./webtodo/static"
	}

	ctx, cancel := context.WithCancel(context.Background())
	
	return &WebTodoServer{
		port:      cfg.Port,
		redisAddr: cfg.RedisAddr,
		staticDir: cfg.StaticDir,
		ctx:       ctx,
		cancel:    cancel,
	}
}

// Start initializes the Redis client connection and starts the HTTP server
func (s *WebTodoServer) Start() error {
	// Initialize Redis client
	var err error
	s.redisClient, err = client.MakeClient(s.redisAddr)
	if err != nil {
		return fmt.Errorf("failed to connect to Redis at %s: %v", s.redisAddr, err)
	}

	// Start Redis client
	err = s.redisClient.Start()
	if err != nil {
		return fmt.Errorf("failed to start Redis client: %v", err)
	}

	// Create HTTP server with router
	mux := s.setupRoutes()
	s.httpServer = &http.Server{
		Addr:         ":" + strconv.Itoa(s.port),
		Handler:      mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	logger.Info(fmt.Sprintf("Starting Web Todo Server on port %d", s.port))
	logger.Info(fmt.Sprintf("Connected to Redis at %s", s.redisAddr))
	logger.Info(fmt.Sprintf("Serving static files from %s", s.staticDir))

	// Start server in a goroutine
	go func() {
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Errorf("HTTP server error: %v", err)
		}
	}()

	return nil
}

// Stop gracefully shuts down the web server and Redis client
func (s *WebTodoServer) Stop() error {
	logger.Info("Shutting down Web Todo Server...")
	s.cancel()

	// Shutdown HTTP server
	if s.httpServer != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if err := s.httpServer.Shutdown(ctx); err != nil {
			logger.Errorf("Error shutting down HTTP server: %v", err)
		}
	}

	// Close Redis client
	if s.redisClient != nil {
		s.redisClient.Close()
	}

	logger.Info("Web Todo Server stopped")
	return nil
}

// setupRoutes configures HTTP routes and middleware
func (s *WebTodoServer) setupRoutes() *http.ServeMux {
	mux := http.NewServeMux()

	// API routes
	apiHandler := NewAPIHandler(s.redisClient)
	mux.Handle("/api/", http.StripPrefix("/api", s.withMiddleware(apiHandler)))

	// Static file serving
	staticPath, _ := filepath.Abs(s.staticDir)
	fileServer := http.FileServer(http.Dir(staticPath))
	mux.Handle("/static/", http.StripPrefix("/static/", fileServer))

	// Serve index.html for root and catch-all routes
	mux.HandleFunc("/", s.serveIndexHTML)

	return mux
}

// withMiddleware applies common middleware to handlers
func (s *WebTodoServer) withMiddleware(next http.Handler) http.Handler {
	return s.corsMiddleware(s.loggingMiddleware(s.errorHandlingMiddleware(next)))
}

// corsMiddleware handles CORS headers
func (s *WebTodoServer) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		
		next.ServeHTTP(w, r)
	})
}

// loggingMiddleware logs HTTP requests
func (s *WebTodoServer) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		logger.Info(fmt.Sprintf("%s %s %v", r.Method, r.URL.Path, time.Since(start)))
	})
}

// errorHandlingMiddleware handles panics and errors
func (s *WebTodoServer) errorHandlingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				logger.Errorf("Panic in HTTP handler: %v", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// serveIndexHTML serves the main HTML file
func (s *WebTodoServer) serveIndexHTML(w http.ResponseWriter, r *http.Request) {
	indexPath := filepath.Join(s.staticDir, "index.html")
	http.ServeFile(w, r, indexPath)
}