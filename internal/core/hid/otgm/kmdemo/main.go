package main

import (
	"fmt"
	"kmdemo/otg"
	"log"
	"time"
)

func main() {
	// 创建OTG设备实例
	hidDev := otg.NewOTGDevice()
	defer hidDev.Close()

	// 打开OTG设备（实际使用时请根据需要选择正确的设备路径）
	// 通常OTG键盘设备路径为 /dev/hidg0
	// 通常OTG鼠标设备路径为 /dev/hidg1
	devicePath := "/dev/hidg0"
	fmt.Printf("Opening OTG device: %s\n", devicePath)
	if err := hidDev.Open(devicePath); err != nil {
		log.Fatal(err)
	}

	fmt.Println("OTG device opened successfully!")

	// 示例：模拟键盘操作
	fmt.Println("\n=== Keyboard Test ===")
	fmt.Println("Typing 'Hello World!'")
	hidDev.TypeString("Hello World!")
	time.Sleep(500 * time.Millisecond)

	fmt.Println("Pressing Enter")
	hidDev.PressKey(otg.KeyEnter)
	time.Sleep(50 * time.Millisecond)
	hidDev.ReleaseKey(otg.KeyEnter)
	time.Sleep(500 * time.Millisecond)

	fmt.Println("Testing Ctrl+L clear screen")
	hidDev.ClearScreen()
	time.Sleep(500 * time.Millisecond)

	hidDev.TypeString("ls -l")
	hidDev.PressKey(otg.KeyEnter)
	time.Sleep(500 * time.Millisecond)
	hidDev.ReleaseKey(otg.KeyEnter)

	// 注意：Ctrl+Alt+Delete会触发系统重启，测试时请谨慎
	fmt.Println("\n=== Three-Key Shortcut Test ===")
	fmt.Println("This demonstrates Ctrl+Alt+Delete functionality (not actually executed)")
	// hidDev.Reboot() // 取消注释以实际执行Ctrl+Alt+Delete
	time.Sleep(1000 * time.Millisecond)

	// 示例：测试符号输入
	fmt.Println("\n=== Symbol Input Test ===")
	fmt.Println("Typing '!@#$%^&*()_+-=[]{}|;':,.<>?")
	hidDev.TypeString("!@#$%^&*()_+-=[]{}|;:,.<>?")
	time.Sleep(500 * time.Millisecond)

	// 关闭键盘设备，打开鼠标设备
	fmt.Println("\n=== Mouse Test ===")
	hidDev.Close()
	mousePath := "/dev/hidg1"
	fmt.Printf("Opening OTG mouse device: %s\n", mousePath)
	if err := hidDev.Open(mousePath); err != nil {
		log.Fatal(err)
	}

	fmt.Println("Moving mouse right 50px")
	hidDev.MoveMouse(50, 0)
	time.Sleep(500 * time.Millisecond)

	fmt.Println("Moving mouse down 50px")
	hidDev.MoveMouse(0, 50)
	time.Sleep(500 * time.Millisecond)

	fmt.Println("Left click")
	hidDev.ClickMouse(otg.MouseLeft)
	time.Sleep(500 * time.Millisecond)

	fmt.Println("Right click")
	hidDev.ClickMouse(otg.MouseRight)
	time.Sleep(500 * time.Millisecond)

	fmt.Println("\nTest completed!")
}
