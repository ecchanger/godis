#!/bin/bash

echo "🚀 测试 Hello World API"
echo "========================"
echo

echo "1. 测试 GET /hello 端点："
echo "------------------------"
curl -s http://localhost:8080/hello
echo

echo "2. 测试 POST /hello 端点（个性化问候）："
echo "--------------------------------"
curl -s -X POST http://localhost:8080/hello \
  -H "Content-Type: application/json" \
  -d '{"name":"张三"}'
echo

echo "3. 测试健康检查端点："
echo "------------------"
curl -s http://localhost:8080/health
echo

echo "4. 测试 404 错误处理："
echo "-------------------"
curl -s http://localhost:8080/nonexistent
echo

echo "5. 测试方法不允许错误："
echo "-------------------"
curl -s -X DELETE http://localhost:8080/hello
echo

echo "✅ API 测试完成！"