package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

// HelloResponse 定义Hello API的响应结构
type HelloResponse struct {
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
	Name      string    `json:"name,omitempty"`
}

// HelloRequest 定义POST请求的结构
type HelloRequest struct {
	Name string `json:"name"`
}

// ErrorResponse 定义错误响应结构
type ErrorResponse struct {
	Error   string `json:"error"`
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// sendJSONResponse 发送JSON响应的辅助函数
func sendJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("JSON编码错误: %v", err)
	}
}

// sendErrorResponse 发送错误响应的辅助函数
func sendErrorResponse(w http.ResponseWriter, statusCode int, message string) {
	response := ErrorResponse{
		Error:   http.StatusText(statusCode),
		Code:    statusCode,
		Message: message,
	}
	sendJSONResponse(w, statusCode, response)
}

// loggingMiddleware 日志中间件
func loggingMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		log.Printf("请求开始: %s %s", r.Method, r.URL.Path)
		
		next.ServeHTTP(w, r)
		
		duration := time.Since(start)
		log.Printf("请求完成: %s %s - 耗时: %v", r.Method, r.URL.Path, duration)
	}
}

// mainHandler 主路由处理器
func mainHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("收到请求: %s %s", r.Method, r.URL.Path)
	
	switch {
	case r.Method == "GET" && r.URL.Path == "/":
		// 根路径
		response := HelloResponse{
			Message:   "欢迎使用Hello World API服务器!",
			Timestamp: time.Now(),
		}
		sendJSONResponse(w, http.StatusOK, response)
		
	case r.Method == "GET" && r.URL.Path == "/hello":
		// 基本hello
		response := HelloResponse{
			Message:   "Hello, World!",
			Timestamp: time.Now(),
		}
		sendJSONResponse(w, http.StatusOK, response)
		
	case r.Method == "GET" && strings.HasPrefix(r.URL.Path, "/hello/"):
		// 带参数的hello
		name := strings.TrimPrefix(r.URL.Path, "/hello/")
		name = strings.TrimSpace(name)
		
		if name == "" {
			sendErrorResponse(w, http.StatusBadRequest, "名字参数不能为空")
			return
		}
		
		if len(name) > 50 {
			sendErrorResponse(w, http.StatusBadRequest, "名字长度不能超过50个字符")
			return
		}
		
		response := HelloResponse{
			Message:   fmt.Sprintf("你好, %s!", name),
			Timestamp: time.Now(),
			Name:      name,
		}
		sendJSONResponse(w, http.StatusOK, response)
		
	case r.Method == "POST" && r.URL.Path == "/hello":
		// POST hello
		var req HelloRequest
		
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			sendErrorResponse(w, http.StatusBadRequest, "无效的JSON格式")
			return
		}
		
		if req.Name == "" {
			sendErrorResponse(w, http.StatusBadRequest, "名字字段是必需的")
			return
		}
		
		req.Name = strings.TrimSpace(req.Name)
		if len(req.Name) > 50 {
			sendErrorResponse(w, http.StatusBadRequest, "名字长度不能超过50个字符")
			return
		}
		
		response := HelloResponse{
			Message:   fmt.Sprintf("你好, %s! 感谢您使用POST请求!", req.Name),
			Timestamp: time.Now(),
			Name:      req.Name,
		}
		sendJSONResponse(w, http.StatusCreated, response)
		
	default:
		// 未找到的路径
		sendErrorResponse(w, http.StatusNotFound, fmt.Sprintf("路径 %s %s 未找到", r.Method, r.URL.Path))
	}
}

func main() {
	// 创建新的ServeMux
	mux := http.NewServeMux()
	
	// 注册主处理器
	mux.HandleFunc("/", loggingMiddleware(mainHandler))
	
	// 配置服务器
	server := &http.Server{
		Addr:         ":8080",
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
	
	log.Println("Hello World API服务器启动中...")
	log.Println("监听地址: http://localhost:8080")
	log.Println("可用端点:")
	log.Println("  GET  /          - 欢迎信息")
	log.Println("  GET  /hello     - 基本问候")
	log.Println("  GET  /hello/{name} - 个性化问候")
	log.Println("  POST /hello     - JSON问候")
	
	// 启动服务器
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("服务器启动失败: %v", err)
	}
}