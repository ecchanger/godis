// helloworld.go
// 这是一个简单的Go语言HelloWorld程序
// 用于演示Go语言的基础语法和程序结构

package main // 声明这是main包，表示这是一个可执行程序

import "fmt" // 导入fmt包，用于格式化输入输出操作

// main函数是程序的入口点
// 当程序运行时，会自动执行main函数中的代码
func main() {
	// 使用fmt.Println()函数输出文本到控制台
	// Println会在输出后自动添加换行符
	fmt.Println("Hello, World!")        // 输出英文问候语
	fmt.Println("你好，世界！")              // 输出中文问候语  
	fmt.Println("这是一个Go语言的HelloWorld程序") // 输出程序说明
}