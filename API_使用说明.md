# Hello World API 使用说明

## 概述
这是一个使用 Go 1.22+ 标准库 `net/http` 和新 ServeMux 特性构建的 Hello World HTTP API 服务器。

## 启动服务器
```bash
go run helloworld.go
```
服务器将在 `http://localhost:8080` 启动。

## API 端点

### 1. GET /hello
获取基本的 Hello World 消息。

**请求示例：**
```bash
curl http://localhost:8080/hello
```

**响应示例：**
```json
{
  "message": "Hello, World! 欢迎使用 Go HTTP API!",
  "timestamp": "2025-09-01T02:58:25Z"
}
```

### 2. POST /hello
发送个性化问候。

**请求示例：**
```bash
curl -X POST http://localhost:8080/hello \
  -H "Content-Type: application/json" \
  -d '{"name":"你的名字"}'
```

**响应示例：**
```json
{
  "message": "Hello, 你的名字! 欢迎使用我们的 API 服务!",
  "timestamp": "2025-09-01T02:58:25Z"
}
```

### 3. GET /health
健康检查端点。

**请求示例：**
```bash
curl http://localhost:8080/health
```

**响应示例：**
```json
{
  "service": "Hello World API",
  "status": "healthy",
  "timestamp": "2025-09-01T02:58:25Z",
  "uptime": "运行正常"
}
```

## 快速测试
运行包含的测试脚本：
```bash
./test_api.sh
```

## 特性
- ✅ 使用 Go 1.22 新 ServeMux 路由特性
- ✅ 完整的错误处理和状态码
- ✅ JSON 请求/响应格式
- ✅ 输入验证
- ✅ 日志中间件
- ✅ 健康检查端点
- ✅ 超时配置
- ✅ 优雅的错误响应