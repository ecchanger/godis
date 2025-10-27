package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
)

// 响应结构体
type HelloResponse struct {
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
	Version   string    `json:"version"`
}

// Hello World API处理器
func helloHandler(w http.ResponseWriter, r *http.Request) {
	// 设置响应头
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	
	// 创建响应数据
	response := HelloResponse{
		Message:   "Hello, World! 你好世界！",
		Timestamp: time.Now(),
		Version:   "1.0.0",
	}
	
	// 编码为JSON并发送响应
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("编码响应失败: %v", err)
		http.Error(w, "内部服务器错误", http.StatusInternalServerError)
		return
	}
	
	log.Printf("Hello World API请求处理完成: %s", r.RemoteAddr)
}

// 健康检查处理器
func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	
	healthResponse := map[string]interface{}{
		"status":    "健康",
		"timestamp": time.Now(),
		"uptime":    "运行正常",
	}
	
	json.NewEncoder(w).Encode(healthResponse)
}

// 根路径处理器
func rootHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	
	html := `<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Hello World API</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; background: #f5f5f5; }
        .container { max-width: 600px; margin: 0 auto; background: white; padding: 20px; border-radius: 8px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); }
        h1 { color: #333; text-align: center; }
        .endpoint { margin: 20px 0; padding: 15px; background: #f9f9f9; border-left: 4px solid #007cba; }
        code { background: #e4e4e4; padding: 2px 4px; border-radius: 3px; }
    </style>
</head>
<body>
    <div class="container">
        <h1>🌍 Hello World API</h1>
        <p>欢迎使用简单的Hello World RESTful API服务！</p>
        
        <div class="endpoint">
            <h3>API 端点：</h3>
            <p><strong>GET /hello</strong> - 获取Hello World消息</p>
            <p><strong>GET /health</strong> - 健康检查</p>
        </div>
        
        <div class="endpoint">
            <h3>示例用法：</h3>
            <p>在终端中运行: <code>curl http://localhost:8888/hello</code></p>
            <p>或访问: <a href="/hello" target="_blank">/hello</a></p>
        </div>
    </div>
</body>
</html>`
	
	w.Write([]byte(html))
}

func main() {
	// 使用传统但可靠的路由方式
	mux := http.NewServeMux()
	
	// 注册路由处理器
	mux.HandleFunc("/", rootHandler)
	mux.HandleFunc("/hello", helloHandler)
	mux.HandleFunc("/health", healthHandler)
	
	// 服务器配置
	server := &http.Server{
		Addr:         ":8888",
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
	
	log.Println("🚀 Hello World API服务器启动中...")
	log.Println("🌐 服务器地址: http://localhost:8888")
	log.Println("📡 可用端点:")
	log.Println("   GET /        - 主页")
	log.Println("   GET /hello   - Hello World API")
	log.Println("   GET /health  - 健康检查")
	log.Println("")
	log.Println("按Ctrl+C停止服务器")
	
	// 启动服务器
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("服务器启动失败: %v", err)
	}
}