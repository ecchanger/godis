package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	// 创建新的ServeMux路由器 (Go 1.22+)
	mux := http.NewServeMux()

	// 定义Hello World处理器
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		fmt.Fprintf(w, "Hello, World! 你好，世界！\n")
	})

	// 定义JSON格式的Hello World处理器
	mux.HandleFunc("/json", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		fmt.Fprintf(w, `{"message": "Hello, World!", "chinese": "你好，世界！"}`)
	})

	// 健康检查端点
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status": "ok", "message": "服务运行正常"}`)
	})

	// 启动服务器
	port := ":3333"
	log.Printf("Hello World服务器启动在端口 %s", port)
	log.Printf("访问 http://localhost%s 查看Hello World", port)
	log.Printf("访问 http://localhost%s/json 查看JSON格式", port)
	log.Printf("访问 http://localhost%s/health 进行健康检查", port)

	if err := http.ListenAndServe(port, mux); err != nil {
		log.Fatal("服务器启动失败:", err)
	}
}