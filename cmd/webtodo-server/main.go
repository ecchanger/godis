package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/hdt3213/godis/config"
	"github.com/hdt3213/godis/lib/logger"
	"github.com/hdt3213/godis/webtodo"
)

// Main entry point for the combined Godis Redis server with Web Todo feature
func main() {
	// Parse command line arguments
	args := os.Args[1:]
	var configFile string
	var webTodoPort int = 8080
	var enableWebTodo bool = false

	// Simple argument parsing
	for i, arg := range args {
		switch arg {
		case "--config", "-c":
			if i+1 < len(args) {
				configFile = args[i+1]
			}
		case "--web-todo-port":
			if i+1 < len(args) {
				fmt.Sscanf(args[i+1], "%d", &webTodoPort)
			}
		case "--enable-web-todo":
			enableWebTodo = true
		case "--help", "-h":
			printUsage()
			return
		}
	}

	// Initialize configuration
	if configFile != "" {
		config.SetupConfig(configFile)
	} else {
		// Use default configuration or environment-based config
		if fileExists("redis.conf") {
			config.SetupConfig("redis.conf")
		} else {
			// Set default properties if no config file
			config.Properties = &config.ServerProperties{
				Bind:           "0.0.0.0",
				Port:           6399,
				AppendOnly:     false,
				AppendFilename: "appendonly.aof",
				MaxClients:     1000,
			}
		}
	}

	// Setup logger
	logger.Setup(&logger.Settings{
		Path:       "logs",
		Name:       "godis",
		Ext:        "log",
		TimeFormat: "2006-01-02",
	})

	// Start Web Todo server if enabled
	var webTodoServer *webtodo.WebTodoServer
	if enableWebTodo || os.Getenv("ENABLE_WEB_TODO") == "true" {
		webTodoConfig := &webtodo.Config{
			Port:      webTodoPort,
			RedisAddr: fmt.Sprintf("%s:%d", config.Properties.Bind, config.Properties.Port),
			StaticDir: "./webtodo/static",
		}

		// Override from environment variables
		if port := os.Getenv("WEB_TODO_PORT"); port != "" {
			fmt.Sscanf(port, "%d", &webTodoConfig.Port)
		}
		if addr := os.Getenv("WEB_TODO_REDIS_ADDR"); addr != "" {
			webTodoConfig.RedisAddr = addr
		}
		if dir := os.Getenv("WEB_TODO_STATIC_DIR"); dir != "" {
			webTodoConfig.StaticDir = dir
		}

		webTodoServer = webtodo.NewWebTodoServer(webTodoConfig)
		if err := webTodoServer.Start(); err != nil {
			logger.Errorf("Failed to start Web Todo server: %v", err)
			log.Fatalf("Web Todo server startup failed: %v", err)
		}
		
		logger.Info("Web Todo server started successfully")
	}

	// Setup graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Wait for shutdown signal
	sig := <-sigChan
	logger.Info(fmt.Sprintf("Received signal %v, shutting down...", sig))

	// Shutdown Web Todo server if running
	if webTodoServer != nil {
		if err := webTodoServer.Stop(); err != nil {
			logger.Errorf("Error stopping Web Todo server: %v", err)
		}
	}

	logger.Info("Shutdown complete")
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	return err == nil && !info.IsDir()
}

func printUsage() {
	fmt.Println("Godis Redis Server with Web Todo Feature")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  godis [options]")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  -c, --config FILE          Use configuration file")
	fmt.Println("  --enable-web-todo          Enable Web Todo feature")
	fmt.Println("  --web-todo-port PORT       Web Todo server port (default: 8080)")
	fmt.Println("  -h, --help                 Show this help message")
	fmt.Println()
	fmt.Println("Environment Variables:")
	fmt.Println("  ENABLE_WEB_TODO=true       Enable Web Todo feature")
	fmt.Println("  WEB_TODO_PORT=8080         Web Todo server port")
	fmt.Println("  WEB_TODO_REDIS_ADDR=...    Redis server address for Web Todo")
	fmt.Println("  WEB_TODO_STATIC_DIR=...    Static files directory")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  # Start with Web Todo feature")
	fmt.Println("  godis --enable-web-todo")
	fmt.Println()
	fmt.Println("  # Start with custom ports")
	fmt.Println("  godis --enable-web-todo --web-todo-port 9000")
	fmt.Println()
	fmt.Println("  # Start with config file")
	fmt.Println("  godis --config my-redis.conf --enable-web-todo")
}