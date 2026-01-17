package ch9329

import (
	"fmt"
	"log"
	"testing"
	"time"

	"go.bug.st/serial"
)

func Test_KM(t *testing.T) {
	// 列出所有可用串口
	ports, err := serial.GetPortsList()
	if err != nil {
		log.Fatal(err)
	}
	if len(ports) == 0 {
		log.Fatal("No serial ports found!")
	}

	fmt.Println("Available serial ports:")
	for i, port := range ports {
		fmt.Printf("%d: %s\n", i+1, port)
	}

	// 创建CH9329设备实例
	hidDev := NewCH9329()
	defer hidDev.Close()

	// 打开第一个可用串口（实际使用时请根据需要选择正确的串口）
	portName := ports[0]
	fmt.Printf("\nOpening port: %s\n", portName)
	if err := hidDev.Open(); err != nil {
		log.Fatal(err)
	}

	fmt.Println("CH9329 device opened successfully!")

	// 示例：模拟鼠标操作
	fmt.Println("\n=== Mouse Test ===")
	fmt.Println("Moving mouse right 50px")
	hidDev.MoveMouse(50, 0)
	time.Sleep(500 * time.Millisecond)

	fmt.Println("Moving mouse down 50px")
	hidDev.MoveMouse(0, 50)
	time.Sleep(500 * time.Millisecond)

	fmt.Println("Left click")
	hidDev.ClickMouse(MouseLeft)
	time.Sleep(500 * time.Millisecond)

	fmt.Println("Right click")
	hidDev.ClickMouse(MouseRight)
	time.Sleep(500 * time.Millisecond)

	// 示例：模拟键盘操作
	fmt.Println("\n=== Keyboard Test ===")
	fmt.Println("Typing 'Hello World!'")
	hidDev.TypeString("Hello World!")
	time.Sleep(500 * time.Millisecond)

	fmt.Println("Pressing Enter")
	hidDev.PressKey(KeyEnter)
	time.Sleep(50 * time.Millisecond)
	hidDev.ReleaseKey(KeyEnter)
	time.Sleep(500 * time.Millisecond)

	fmt.Println("Testing Ctrl+L clear screen")
	hidDev.ClearScreen()
	time.Sleep(500 * time.Millisecond)

	hidDev.TypeString("ls -l")
	hidDev.PressKey(KeyEnter)
	time.Sleep(500 * time.Millisecond)
	hidDev.ReleaseKey(KeyEnter)

	// 注意：Ctrl+Alt+Delete会触发系统重启，测试时请谨慎
	fmt.Println("\n=== Three-Key Shortcut Test ===")
	fmt.Println("This demonstrates Ctrl+Alt+Delete functionality (not actually executed)")
	// hidDev.Reboot() // 取消注释以实际执行Ctrl+Alt+Delete
	time.Sleep(1000 * time.Millisecond)
	fmt.Println("\nTest completed!")
}
