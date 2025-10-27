#!/bin/bash

echo "🚀 测试Hello World API服务器"
echo "================================"

echo ""
echo "1. 测试根路径 (GET /)"
curl -s http://localhost:8080/

echo ""
echo "2. 测试基本问候 (GET /hello)"
curl -s http://localhost:8080/hello

echo ""
echo "3. 测试个性化问候 (GET /hello/{name})"
curl -s "http://localhost:8080/hello/小明"

echo ""
echo "4. 测试POST问候"
curl -s -X POST -H "Content-Type: application/json" -d '{"name":"开发者"}' http://localhost:8080/hello

echo ""
echo "5. 测试错误处理 - 404"
curl -s http://localhost:8080/notfound

echo ""
echo "6. 测试错误处理 - 无效JSON"
curl -s -X POST -H "Content-Type: application/json" -d 'invalid' http://localhost:8080/hello

echo ""
echo "7. 测试错误处理 - 空名字"
curl -s -X POST -H "Content-Type: application/json" -d '{"name":""}' http://localhost:8080/hello

echo ""
echo "✅ 测试完成！"