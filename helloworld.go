package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

// HelloResponse 定义hello接口的响应结构
type HelloResponse struct {
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
	Status    string    `json:"status"`
}

// HealthResponse 定义健康检查接口的响应结构
type HealthResponse struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
}

// ErrorResponse 定义错误响应结构
type ErrorResponse struct {
	Error     string    `json:"error"`
	Timestamp time.Time `json:"timestamp"`
}

// 日志中间件
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		
		// 记录请求开始
		log.Printf("开始处理请求: %s %s 来自 %s", r.Method, r.URL.Path, r.RemoteAddr)
		
		// 调用下一个处理器
		next.ServeHTTP(w, r)
		
		// 记录请求完成
		duration := time.Since(start)
		log.Printf("请求完成: %s %s 耗时 %v", r.Method, r.URL.Path, duration)
	})
}

// CORS中间件
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		
		// 处理预检请求
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		
		next.ServeHTTP(w, r)
	})
}

// 设置JSON响应头
func setJSONHeader(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
}

// 发送JSON响应
func sendJSONResponse(w http.ResponseWriter, data interface{}, statusCode int) {
	setJSONHeader(w)
	w.WriteHeader(statusCode)
	
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("编码JSON响应失败: %v", err)
	}
}

// 发送错误响应
func sendErrorResponse(w http.ResponseWriter, message string, statusCode int) {
	response := ErrorResponse{
		Error:     message,
		Timestamp: time.Now(),
	}
	sendJSONResponse(w, response, statusCode)
}

// hello处理器 - 处理GET /hello请求
func helloHandler(w http.ResponseWriter, r *http.Request) {
	// 只允许GET方法
	if r.Method != http.MethodGet {
		sendErrorResponse(w, "方法不被允许", http.StatusMethodNotAllowed)
		return
	}
	
	// 获取查询参数name（可选）
	name := r.URL.Query().Get("name")
	if name == "" {
		name = "世界"
	}
	
	response := HelloResponse{
		Message:   fmt.Sprintf("你好, %s!", name),
		Timestamp: time.Now(),
		Status:    "success",
	}
	
	sendJSONResponse(w, response, http.StatusOK)
}

// 健康检查处理器 - 处理GET /health请求
func healthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		sendErrorResponse(w, "方法不被允许", http.StatusMethodNotAllowed)
		return
	}
	
	response := HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now(),
	}
	
	sendJSONResponse(w, response, http.StatusOK)
}

// 根路径处理器 - 处理GET /请求
func rootHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		sendErrorResponse(w, "方法不被允许", http.StatusMethodNotAllowed)
		return
	}
	
	response := map[string]interface{}{
		"message": "欢迎使用Hello World API！",
		"endpoints": map[string]string{
			"GET /":                "API信息",
			"GET /hello":           "返回hello world消息",
			"GET /hello?name=张三": "返回个性化hello消息",
			"GET /health":          "健康检查",
		},
		"timestamp": time.Now(),
	}
	
	sendJSONResponse(w, response, http.StatusOK)
}

// 404处理器
func notFoundHandler(w http.ResponseWriter, r *http.Request) {
	sendErrorResponse(w, "路径未找到", http.StatusNotFound)
}

func main() {
	// 创建新的ServeMux（Go 1.22特性）
	mux := http.NewServeMux()
	
	// 注册路由处理器
	mux.HandleFunc("/hello", helloHandler)
	mux.HandleFunc("/health", healthHandler)
	mux.HandleFunc("/", rootHandler)
	
	// 应用中间件
	handler := loggingMiddleware(corsMiddleware(mux))
	
	// 服务器配置
	server := &http.Server{
		Addr:         ":8080",
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
	
	// 启动消息
	fmt.Println("🚀 Hello World API 服务器启动中...")
	fmt.Println("📍 监听地址: http://localhost:8080")
	fmt.Println("📚 可用端点:")
	fmt.Println("   GET /        - API信息")
	fmt.Println("   GET /hello   - Hello World消息")
	fmt.Println("   GET /health  - 健康检查")
	fmt.Println("\n💡 示例请求:")
	fmt.Println("   curl http://localhost:8080/hello")
	fmt.Println("   curl http://localhost:8080/hello?name=张三")
	fmt.Println("\n按Ctrl+C停止服务器")
	
	// 启动服务器
	log.Printf("服务器在端口8080启动")
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("服务器启动失败: %v", err)
	}
}