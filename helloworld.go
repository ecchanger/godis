// Hello World API - 使用Go标准库net/http包和Go 1.22特性
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

// HelloResponse 表示Hello API的响应结构
type HelloResponse struct {
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
	Path      string    `json:"path"`
}

// ErrorResponse 表示错误响应结构
type ErrorResponse struct {
	Error     string    `json:"error"`
	Timestamp time.Time `json:"timestamp"`
	Path      string    `json:"path"`
}

// helloHandler 处理所有hello相关请求
func helloHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		sendErrorResponse(w, r, "方法不允许", http.StatusMethodNotAllowed)
		return
	}

	// 解析路径以提取名字
	path := strings.TrimPrefix(r.URL.Path, "/hello")
	path = strings.TrimPrefix(path, "/")
	
	var message string
	if path == "" {
		message = "Hello, World! 你好，世界！"
	} else {
		message = fmt.Sprintf("Hello, %s! 你好，%s！", path, path)
	}

	response := HelloResponse{
		Message:   message,
		Timestamp: time.Now(),
		Path:      r.URL.Path,
	}

	sendJSONResponse(w, response, http.StatusOK)
	log.Printf("响应发送: %s", message)
}

// healthHandler 处理健康检查请求
func healthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		sendErrorResponse(w, r, "方法不允许", http.StatusMethodNotAllowed)
		return
	}

	response := map[string]interface{}{
		"status":    "ok",
		"timestamp": time.Now(),
		"service":   "Hello World API",
	}

	sendJSONResponse(w, response, http.StatusOK)
}

// sendJSONResponse 发送JSON格式的响应
func sendJSONResponse(w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)
	
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("JSON编码错误: %v", err)
		http.Error(w, "内部服务器错误", http.StatusInternalServerError)
	}
}

// sendErrorResponse 发送错误响应
func sendErrorResponse(w http.ResponseWriter, r *http.Request, message string, statusCode int) {
	errorResp := ErrorResponse{
		Error:     message,
		Timestamp: time.Now(),
		Path:      r.URL.Path,
	}
	
	sendJSONResponse(w, errorResp, statusCode)
}

func main() {
	// 创建新的ServeMux
	mux := http.NewServeMux()

	// 注册路由处理器
	mux.HandleFunc("/hello", helloHandler)     // 基本hello
	mux.HandleFunc("/hello/", helloHandler)    // 带参数的hello (处理 /hello/xxx)
	mux.HandleFunc("/health", healthHandler)   // 健康检查

	// 服务器配置
	server := &http.Server{
		Addr:    ":3000",
		Handler: mux,
		// 设置合理的超时时间
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Println("🚀 Hello World API服务器启动中...")
	log.Println("📡 监听端口: :3000")
	log.Println("📋 可用端点:")
	log.Println("  GET /hello          - 基本问候")
	log.Println("  GET /hello/姓名      - 个性化问候")
	log.Println("  GET /health         - 健康检查")

	// 启动服务器
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("❌ 服务器启动失败: %v", err)
	}
}