package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"todolist/internal/handler"
	"todolist/internal/service"
)

func main() {
	// 创建服务
	todoService := service.NewTodoService()
	todoHandler := handler.NewTodoHandler(todoService)

	// 创建路由器
	r := mux.NewRouter()

	// 静态文件服务
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("web/static/"))))

	// 首页路由
	r.HandleFunc("/", todoHandler.ServeIndex).Methods("GET")

	// API路由
	api := r.PathPrefix("/api").Subrouter()
	api.HandleFunc("/todos", todoHandler.GetAllTodos).Methods("GET")
	api.HandleFunc("/todos", todoHandler.CreateTodo).Methods("POST")
	api.HandleFunc("/todos/{id:[0-9]+}", todoHandler.GetTodo).Methods("GET")
	api.HandleFunc("/todos/{id:[0-9]+}", todoHandler.UpdateTodo).Methods("PUT")
	api.HandleFunc("/todos/{id:[0-9]+}", todoHandler.DeleteTodo).Methods("DELETE")
	api.HandleFunc("/todos/{id:[0-9]+}/toggle", todoHandler.ToggleComplete).Methods("PATCH")

	// 配置CORS
	c := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
		AllowedHeaders: []string{"*"},
	})

	handler := c.Handler(r)

	// 获取端口
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	log.Printf("Open http://localhost:%s in your browser", port)
	log.Fatal(http.ListenAndServe(":"+port, handler))
}