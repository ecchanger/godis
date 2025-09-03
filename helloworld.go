// Hello World 示例程序
// 这是一个基本的Go语言程序，用于演示Go的基础语法
package main

// 导入所需的包
import (
	"fmt" // fmt包提供了格式化输入输出的功能
)

// main函数是程序的入口点
// 当程序运行时，会自动执行这个函数
func main() {
	// 使用fmt.Println输出英文的"Hello, World!"
	// Println函数会在输出内容后自动添加换行符
	fmt.Println("Hello, World!")
	
	// 输出中文的"你好，世界！"
	// Go语言原生支持UTF-8编码，可以直接输出中文
	fmt.Println("你好，世界！")
}