package dhcp

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"strings"
)

// -------------------------- 接口定义 --------------------------
// NetworkManager 网络管理接口，封装网卡、路由、转发相关操作
type NetworkManager interface {
	GetDefaultRouteIface() (string, error)                      // 获取默认路由网卡名
	UpIfaceWithIP(ifaceName, ipWithCIDR string) error           // 拉起网卡并设置IP
	SetupIPForwardAndNAT(innerIface, outerIface string) error   // 配置IP转发和NAT
	CleanupIPForwardAndNAT(innerIface, outerIface string) error // 清理转发和NAT规则
	GetIfaceIP(ifaceName string) (net.IP, error)                // 获取网卡IPv4地址
}

// -------------------------- 接口实现 --------------------------
// LinuxNetworkManager Linux系统的网络管理实现
type LinuxNetworkManager struct{}

// GetDefaultRouteIface 获取系统默认路由对应的网卡名
func (l *LinuxNetworkManager) GetDefaultRouteIface() (string, error) {
	// 优先解析/proc/net/route
	file, err := os.Open("/proc/net/route")
	if err != nil {
		return "", fmt.Errorf("打开路由文件失败: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	if scanner.Scan() {
	} // 跳过表头

	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}
		// 00000000 表示默认路由（0.0.0.0）
		if fields[1] == "00000000" {
			return fields[0], nil
		}
	}

	// 备用方案：调用ip route命令
	cmd := exec.Command("ip", "route", "show", "default")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("获取默认路由失败: %v, 输出: %s", err, string(output))
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "dev") {
			parts := strings.Fields(line)
			for i, part := range parts {
				if part == "dev" && i+1 < len(parts) {
					return parts[i+1], nil
				}
			}
		}
	}

	return "", fmt.Errorf("未找到系统默认路由网卡")
}

// UpIfaceWithIP 拉起网卡并配置IP（格式：192.168.7.1/24）
func (l *LinuxNetworkManager) UpIfaceWithIP(ifaceName, ipWithCIDR string) error {
	// 1. 拉起网卡
	cmd := exec.Command("ip", "link", "set", ifaceName, "up")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("拉起网卡%s失败: %v, 输出: %s", ifaceName, err, string(output))
	}
	log.Printf("已拉起网卡: %s", ifaceName)

	// 2. 配置IP地址
	cmd = exec.Command("ip", "addr", "add", ipWithCIDR, "dev", ifaceName)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("配置网卡%s IP失败: %v, 输出: %s", ifaceName, err, string(output))
	}
	log.Printf("已配置网卡%s IP: %s", ifaceName, ipWithCIDR)

	return nil
}

// SetupIPForwardAndNAT 配置IP转发和NAT规则
func (l *LinuxNetworkManager) SetupIPForwardAndNAT(innerIface, outerIface string) error {
	// 启用IP转发
	cmd := exec.Command("sysctl", "-w", "net.ipv4.ip_forward=1")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("启用IP转发失败: %v, 输出: %s", err, string(output))
	}
	log.Println("已启用系统IP转发")

	// 清除旧NAT规则
	cmd = exec.Command("iptables", "-t", "nat", "-F")
	if output, err := cmd.CombinedOutput(); err != nil {
		log.Printf("清除旧NAT规则警告: %v, 输出: %s", err, string(output))
	}

	// 添加MASQUERADE规则（动态NAT）
	cmd = exec.Command("iptables", "-t", "nat", "-A", "POSTROUTING", "-o", outerIface, "-j", "MASQUERADE")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("配置NAT规则失败: %v, 输出: %s", err, string(output))
	}

	// 允许转发规则
	cmd = exec.Command("iptables", "-A", "FORWARD", "-i", innerIface, "-o", outerIface, "-j", "ACCEPT")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("配置转发规则1失败: %v, 输出: %s", err, string(output))
	}

	cmd = exec.Command("iptables", "-A", "FORWARD", "-i", outerIface, "-o", innerIface, "-m", "state", "--state", "RELATED,ESTABLISHED", "-j", "ACCEPT")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("配置转发规则2失败: %v, 输出: %s", err, string(output))
	}

	log.Printf("已配置NAT规则：%s → %s", innerIface, outerIface)
	return nil
}

// CleanupIPForwardAndNAT 清理IP转发和NAT规则
func (l *LinuxNetworkManager) CleanupIPForwardAndNAT(innerIface, outerIface string) error {
	// 清除NAT规则
	cmd := exec.Command("iptables", "-t", "nat", "-D", "POSTROUTING", "-o", outerIface, "-j", "MASQUERADE")
	if output, err := cmd.CombinedOutput(); err != nil {
		log.Printf("清除NAT规则警告: %v, 输出: %s", err, string(output))
	}

	// 清除转发规则
	cmd = exec.Command("iptables", "-D", "FORWARD", "-i", innerIface, "-o", outerIface, "-j", "ACCEPT")
	cmd.Run()
	cmd = exec.Command("iptables", "-D", "FORWARD", "-i", outerIface, "-o", innerIface, "-m", "state", "--state", "RELATED,ESTABLISHED", "-j", "ACCEPT")
	cmd.Run()

	log.Println("已清理网络转发规则")
	return nil
}

// GetIfaceIP 获取指定网卡的IPv4地址
func (l *LinuxNetworkManager) GetIfaceIP(ifaceName string) (net.IP, error) {
	iface, err := net.InterfaceByName(ifaceName)
	if err != nil {
		return nil, fmt.Errorf("获取网卡%s信息失败: %v", ifaceName, err)
	}

	addrs, err := iface.Addrs()
	if err != nil {
		return nil, fmt.Errorf("获取网卡%s地址失败: %v", ifaceName, err)
	}

	for _, addr := range addrs {
		ipNet, ok := addr.(*net.IPNet)
		if !ok || ipNet.IP.IsLoopback() {
			continue
		}
		ipv4 := ipNet.IP.To4()
		if ipv4 != nil {
			return ipv4, nil
		}
	}

	return nil, fmt.Errorf("网卡%s未配置IPv4地址", ifaceName)
}
