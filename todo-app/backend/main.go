package main

import (
	"log"
	"net/http"
	"time"

	"todo-backend/handlers"
	"todo-backend/repository"
	"todo-backend/service"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/hdt3213/godis/database"
)

func main() {
	// Initialize Godis database server
	server := database.NewStandaloneServer()
	
	// Create Godis adapter
	godisClient := repository.NewGodisClientAdapter(server)
	
	// Initialize repository and service
	todoRepo := repository.NewTodoRepository(godisClient)
	todoService := service.NewTodoService(todoRepo)
	
	// Initialize handlers
	todoHandler := handlers.NewTodoHandler(todoService)
	
	// Create Gin router
	router := gin.Default()
	
	// Configure CORS
	config := cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "http://localhost:5173"}, // React dev servers
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}
	router.Use(cors.New(config))
	
	// Add logging middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	
	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"timestamp": time.Now(),
			"service":   "todo-backend",
		})
	})
	
	// API routes group
	api := router.Group("/api")
	{
		// Todo routes
		todos := api.Group("/todos")
		{
			todos.GET("", todoHandler.GetAllTodos)         // GET /api/todos
			todos.POST("", todoHandler.CreateTodo)         // POST /api/todos
			todos.GET("/:id", todoHandler.GetTodoByID)     // GET /api/todos/:id
			todos.PUT("/:id", todoHandler.UpdateTodo)      // PUT /api/todos/:id
			todos.DELETE("/:id", todoHandler.DeleteTodo)   // DELETE /api/todos/:id
			todos.PATCH("/:id/toggle", todoHandler.ToggleTodoComplete) // PATCH /api/todos/:id/toggle
		}
		
		// Stats route
		api.GET("/stats", todoHandler.GetTodoStats) // GET /api/stats
	}
	
	// Start server
	port := ":8081"
	log.Printf("Starting Todo Backend Server on port %s", port)
	log.Printf("Health check: http://localhost%s/health", port)
	log.Printf("API endpoint: http://localhost%s/api", port)
	
	if err := router.Run(port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}