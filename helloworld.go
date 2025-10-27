package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

// HelloResponse 定义标准的响应结构
type HelloResponse struct {
	Message   string `json:"message"`
	Timestamp string `json:"timestamp"`
}

// HelloRequest 定义 POST 请求的结构
type HelloRequest struct {
	Name string `json:"name"`
}

// ErrorResponse 定义错误响应结构
type ErrorResponse struct {
	Error   string `json:"error"`
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// 简单的日志中间件
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		log.Printf("开始处理请求: %s %s", r.Method, r.URL.Path)
		
		next.ServeHTTP(w, r)
		
		duration := time.Since(start)
		log.Printf("请求处理完成: %s %s - 耗时: %v", r.Method, r.URL.Path, duration)
	})
}

// 设置 JSON 响应头
func setJSONHeaders(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Powered-By", "Go-Hello-World")
}

// 发送 JSON 错误响应
func sendErrorResponse(w http.ResponseWriter, statusCode int, errorMsg string) {
	setJSONHeaders(w)
	w.WriteHeader(statusCode)
	
	response := ErrorResponse{
		Error:   errorMsg,
		Code:    statusCode,
		Message: "请求处理失败",
	}
	
	json.NewEncoder(w).Encode(response)
}

// GET /hello 处理器 - 返回简单的 Hello World
func handleGetHello(w http.ResponseWriter, r *http.Request) {
	setJSONHeaders(w)
	
	response := HelloResponse{
		Message:   "Hello, World! 欢迎使用 Go HTTP API!",
		Timestamp: time.Now().Format(time.RFC3339),
	}
	
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("编码响应失败: %v", err)
		sendErrorResponse(w, http.StatusInternalServerError, "响应编码失败")
		return
	}
	
	log.Println("成功响应 GET /hello 请求")
}

// POST /hello 处理器 - 接收用户名并返回个性化问候
func handlePostHello(w http.ResponseWriter, r *http.Request) {
	// 验证 Content-Type
	if r.Header.Get("Content-Type") != "application/json" {
		sendErrorResponse(w, http.StatusBadRequest, "Content-Type 必须是 application/json")
		return
	}
	
	var req HelloRequest
	
	// 解析 JSON 请求体
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("解析请求体失败: %v", err)
		sendErrorResponse(w, http.StatusBadRequest, "无效的 JSON 格式")
		return
	}
	defer r.Body.Close()
	
	// 验证输入
	if req.Name == "" {
		sendErrorResponse(w, http.StatusBadRequest, "姓名字段不能为空")
		return
	}
	
	setJSONHeaders(w)
	
	response := HelloResponse{
		Message:   fmt.Sprintf("Hello, %s! 欢迎使用我们的 API 服务!", req.Name),
		Timestamp: time.Now().Format(time.RFC3339),
	}
	
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("编码响应失败: %v", err)
		sendErrorResponse(w, http.StatusInternalServerError, "响应编码失败")
		return
	}
	
	log.Printf("成功响应 POST /hello 请求，用户: %s", req.Name)
}

// 健康检查端点
func handleHealthCheck(w http.ResponseWriter, r *http.Request) {
	setJSONHeaders(w)
	w.WriteHeader(http.StatusOK)
	
	response := map[string]interface{}{
		"status":    "healthy",
		"service":   "Hello World API",
		"timestamp": time.Now().Format(time.RFC3339),
		"uptime":    "运行正常",
	}
	
	json.NewEncoder(w).Encode(response)
}

// 404 处理器
func handleNotFound(w http.ResponseWriter, r *http.Request) {
	log.Printf("404 - 未找到路径: %s %s", r.Method, r.URL.Path)
	sendErrorResponse(w, http.StatusNotFound, "请求的端点不存在")
}

func main() {
	// 创建新的 ServeMux (Go 1.22 新特性)
	mux := http.NewServeMux()
	
	// 注册路由处理器，使用 Go 1.22 的新路由语法
	mux.HandleFunc("GET /hello", handleGetHello)
	mux.HandleFunc("POST /hello", handlePostHello)
	mux.HandleFunc("GET /health", handleHealthCheck)
	
	// 为了兼容性，也添加不带方法前缀的路由
	mux.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handleGetHello(w, r)
		case http.MethodPost:
			handlePostHello(w, r)
		default:
			sendErrorResponse(w, http.StatusMethodNotAllowed, "不支持的 HTTP 方法")
		}
	})
	
	mux.HandleFunc("/health", handleHealthCheck)
	
	// 处理所有其他请求 (404)
	mux.HandleFunc("/", handleNotFound)
	
	// 应用日志中间件
	handler := loggingMiddleware(mux)
	
	// 配置服务器
	server := &http.Server{
		Addr:         ":8080",
		Handler:      handler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
	
	log.Println("🚀 Hello World API 服务器启动中...")
	log.Println("📡 监听端口: :8080")
	log.Println("🛠️ 可用端点:")
	log.Println("   GET  /hello  - 获取 Hello World 消息")
	log.Println("   POST /hello  - 发送个性化问候 (需要 JSON: {\"name\": \"your_name\"})")
	log.Println("   GET  /health - 健康检查")
	log.Println("✨ 按 Ctrl+C 停止服务器")
	
	// 启动服务器
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("❌ 服务器启动失败: %v", err)
	}
}