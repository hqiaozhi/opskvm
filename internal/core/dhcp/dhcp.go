package dhcp

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/insomniacslk/dhcp/dhcpv4"
	"github.com/insomniacslk/dhcp/dhcpv4/server4"
)

// DHCPServerConfig DHCP服务器配置结构体
type DHCPServerConfig struct {
	InnerIface string
	GatewayIP  net.IP
	SubnetMask net.IP
	DNS        []net.IP
	LeaseTime  time.Duration
	StartIP    net.IP
	NM         NetworkManager
}

// DHCPServer DHCP服务器接口
type DHCPServer interface {
	Start(ctx context.Context) error
	Stop() error
}

// DHCPPacketHandler DHCP包处理接口
type DHCPPacketHandler interface {
	HandlePacket(conn net.PacketConn, peer net.Addr, m *dhcpv4.DHCPv4) error
}

// -------------------------- 实现结构体 --------------------------

// linuxDHCPServer Linux平台DHCP服务器实现
type linuxDHCPServer struct {
	config     *DHCPServerConfig
	server     *server4.Server
	outerIface string
	ctx        context.Context
	cancel     context.CancelFunc
}

// dhcpPacketHandler DHCP包处理器实现
type dhcpPacketHandler struct {
	gatewayIP  net.IP
	subnetMask net.IP
	dns        []net.IP
	leaseTime  time.Duration
	startIP    net.IP
}

// -------------------------- DHCP包处理器实现 --------------------------

// NewDHCPPacketHandler 创建DHCP包处理器实例
func NewDHCPPacketHandler(gatewayIP, subnetMask net.IP, dns []net.IP, leaseTime time.Duration, startIP net.IP) DHCPPacketHandler {
	return &dhcpPacketHandler{
		gatewayIP:  gatewayIP,
		subnetMask: subnetMask,
		dns:        dns,
		leaseTime:  leaseTime,
		startIP:    startIP,
	}
}

// HandlePacket 处理DHCP数据包
func (h *dhcpPacketHandler) HandlePacket(conn net.PacketConn, peer net.Addr, m *dhcpv4.DHCPv4) error {
	// 解析请求类型
	msgType := m.MessageType()
	log.Printf("收到DHCP请求 - 客户端MAC: %s, 请求类型: %s", m.ClientHWAddr.String(), msgType.String())

	// 仅处理DISCOVER和REQUEST请求
	if msgType != dhcpv4.MessageTypeDiscover && msgType != dhcpv4.MessageTypeRequest {
		log.Printf("忽略非DISCOVER/REQUEST请求: %s", msgType.String())
		return nil
	}

	// 构建响应包
	resp, err := dhcpv4.NewReplyFromRequest(m)
	if err != nil {
		log.Printf("构建响应包失败: %v", err)
		return err
	}

	// 分配IP（简单实现：固定分配起始IP，可扩展为IP池管理）
	assignedIP := h.startIP
	resp.YourIPAddr = assignedIP
	resp.ServerIPAddr = h.gatewayIP

	// 设置DHCP选项
	resp.UpdateOption(dhcpv4.OptSubnetMask(net.IPMask(h.subnetMask)))
	resp.UpdateOption(dhcpv4.OptRouter(h.gatewayIP))
	resp.UpdateOption(dhcpv4.OptDNS(h.dns...))
	resp.UpdateOption(dhcpv4.OptIPAddressLeaseTime(h.leaseTime))
	resp.UpdateOption(dhcpv4.OptServerIdentifier(h.gatewayIP))

	// 设置响应类型（OFFER/ACK）
	if msgType == dhcpv4.MessageTypeDiscover {
		resp.UpdateOption(dhcpv4.OptMessageType(dhcpv4.MessageTypeOffer))
	} else {
		resp.UpdateOption(dhcpv4.OptMessageType(dhcpv4.MessageTypeAck))
	}

	// 发送响应
	if _, err := conn.WriteTo(resp.ToBytes(), peer); err != nil {
		log.Printf("发送DHCP响应失败: %v", err)
		return err
	} else {
		log.Printf("成功响应 - 分配IP: %s, 网关: %s", assignedIP, h.gatewayIP)
		return nil
	}
}

// ToServerHandler 转换为server4.Handler类型（适配外部库）
func (h *dhcpPacketHandler) ToServerHandler() server4.Handler {
	return func(conn net.PacketConn, peer net.Addr, m *dhcpv4.DHCPv4) {
		h.HandlePacket(conn, peer, m)
	}
}

// -------------------------- 遗留的DHCPHandler函数（保持向后兼容） --------------------------
// DHCPHandler DHCPv4请求处理器（适配insomniacslk/dhcp库）
func DHCPHandler(gatewayIP, subnetMask net.IP, dns []net.IP, leaseTime time.Duration, startIP net.IP) server4.Handler {
	handler := NewDHCPPacketHandler(gatewayIP, subnetMask, dns, leaseTime, startIP)
	return handler.(*dhcpPacketHandler).ToServerHandler()
}

// -------------------------- DHCP服务器实现 --------------------------

// NewDHCPServer 创建DHCP服务器实例
func NewDHCPServer(config *DHCPServerConfig) DHCPServer {
	ctx, cancel := context.WithCancel(context.Background())
	return &linuxDHCPServer{
		config: config,
		ctx:    ctx,
		cancel: cancel,
	}
}

// Start 启动DHCP服务器
func (s *linuxDHCPServer) Start(ctx context.Context) error {
	// 检查root权限
	if os.Geteuid() != 0 {
		return fmt.Errorf("必须以root权限运行（sudo），否则无法操作网卡/iptables/67端口")
	}

	// 1. 处理外网网卡（自动获取或指定）
	outerIfaceName, err := s.config.NM.GetDefaultRouteIface()
	if err != nil {
		return fmt.Errorf("获取默认路由网卡失败: %v", err)
	}
	s.outerIface = outerIfaceName

	// 2. 拉起内网网卡并配置IP
	if err := s.config.NM.UpIfaceWithIP(s.config.InnerIface, "172.168.8.1/24"); err != nil {
		return fmt.Errorf("配置内网网卡失败: %v", err)
	}

	// 3. 获取内网网卡IP（网关IP）
	gatewayIP, err := s.config.NM.GetIfaceIP(s.config.InnerIface)
	if err != nil {
		return fmt.Errorf("获取内网网卡IP失败: %v", err)
	}
	s.config.GatewayIP = gatewayIP

	// 4. 配置IP转发和NAT
	if err := s.config.NM.SetupIPForwardAndNAT(s.config.InnerIface, outerIfaceName); err != nil {
		return fmt.Errorf("配置网络转发失败: %v", err)
	}

	// 5. 配置DHCP参数（如果未指定）
	if s.config.SubnetMask == nil {
		s.config.SubnetMask = net.ParseIP("255.255.255.0")
	}
	if s.config.DNS == nil {
		s.config.DNS = []net.IP{
			net.ParseIP("223.5.5.5"),
			net.ParseIP("223.6.6.6"),
		}
	}
	if s.config.LeaseTime == 0 {
		s.config.LeaseTime = 12 * time.Hour
	}
	if s.config.StartIP == nil {
		s.config.StartIP = net.ParseIP("172.168.8.2")
	}

	// 6. 监听退出信号，清理资源
	go func() {
		<-ctx.Done()
		log.Println("收到退出信号，正在清理资源...")
		s.Stop()
	}()

	// 7. 启动DHCPv4服务器
	listenAddr := &net.UDPAddr{
		IP:   s.config.GatewayIP,
		Port: 67, // DHCP服务器默认端口
	}

	handler := NewDHCPPacketHandler(
		s.config.GatewayIP,
		s.config.SubnetMask,
		s.config.DNS,
		s.config.LeaseTime,
		s.config.StartIP,
	)

	server, err := server4.NewServer(s.config.InnerIface, listenAddr, handler.(*dhcpPacketHandler).ToServerHandler())
	if err != nil {
		return fmt.Errorf("创建DHCP服务器失败: %v", err)
	}
	s.server = server

	// 打印启动信息
	log.Printf(`DHCP服务器启动成功！
  内网网卡: %s (网关IP: %s)
  外网网卡: %s
  IP地址池起始: %s
  租期: %s`,
		s.config.InnerIface, s.config.GatewayIP,
		outerIfaceName,
		s.config.StartIP,
		s.config.LeaseTime)

	// 运行服务器
	go func() {
		if err := server.Serve(); err != nil && err != context.Canceled {
			log.Printf("DHCP服务器运行失败: %v", err)
		}
	}()

	return nil
}

// Stop 停止DHCP服务器并清理资源
func (s *linuxDHCPServer) Stop() error {
	if s.server != nil {
		s.server.Close()
	}

	if s.config.NM != nil && s.outerIface != "" {
		if err := s.config.NM.CleanupIPForwardAndNAT(s.config.InnerIface, s.outerIface); err != nil {
			log.Printf("清理网络转发失败: %v", err)
		}
	}

	if s.cancel != nil {
		s.cancel()
	}

	log.Println("DHCP服务器已停止，资源已清理")
	return nil
}

// -------------------------- 兼容原有API的Start函数 --------------------------

// Start 启动DHCP服务器（兼容原有API）
func Start(innerIface string) {
	// 检查root权限
	if os.Geteuid() != 0 {
		log.Fatal("错误：必须以root权限运行（sudo），否则无法操作网卡/iptables/67端口")
	}

	// 创建默认网络管理器
	nm := &LinuxNetworkManager{}

	// 创建配置
	config := &DHCPServerConfig{
		InnerIface: innerIface,
		NM:         nm,
	}

	// 创建并启动服务器
	server := NewDHCPServer(config)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 监听退出信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("收到退出信号，正在停止DHCP服务器...")
		cancel()
	}()

	// 启动服务器
	if err := server.Start(ctx); err != nil {
		log.Fatalf("启动DHCP服务器失败: %v", err)
	}

	// 阻塞主线程
	<-ctx.Done()
}
