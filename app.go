package main

import (
	"context"
	"fmt"
	"go-tools/tools/redis"
	"go-tools/tools/sniffer"
	"go-tools/tools/usbserial"
	"strings"
	"sync"

	"github.com/wailsapp/wails/v3/pkg/application"
)

// App struct
type App struct {
	app      *application.App
	ctx      context.Context
	engine   *sniffer.Engine
	packetCh chan sniffer.PacketInfo

	snifferMu      sync.Mutex
	mitm           *sniffer.Mitm
	snifferRunning bool
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// SetApp 注入 Wails v3 应用实例，供事件推送使用
func (a *App) SetApp(app *application.App) {
	a.app = app
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// ApplicationStartup 标记 App 为 Wails v3 服务，并在应用启动时注入上下文
func (a *App) ApplicationStartup(ctx context.Context) {
	a.startup(ctx)
}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}

// OpenToolWindow 在新的 Wails 原生窗口中打开指定前端路由（hash 路由）。
// path 形如 "/tools/timestamp"；title 为新窗口标题；
// single 为 true 时为该路由登记窗口名，已存在则直接复用（展示+取消最小化+聚焦），不重复创建；
// fullscreen 为 true 时新窗口最大化铺满显示器（非排他全屏，避免顶部自绘标题栏被遮挡）；
// width/height 为新窗口视口尺寸，<=0 时继承默认尺寸（1000x718）。
func (a *App) OpenToolWindow(path string, title string, single bool, fullscreen bool, width int, height int) error {
	if a.app == nil {
		return fmt.Errorf("应用尚未初始化")
	}
	if path == "" {
		return fmt.Errorf("path 不能为空")
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	if title == "" {
		title = "go-tools"
	}
	startState := application.WindowStateNormal
	if fullscreen {
		// 用最大化铺满显示器（而非排他全屏），保证自绘标题栏/菜单不被遮挡
		startState = application.WindowStateMaximised
	}
	// 视口尺寸：未设定（<=0）则继承默认 1000x718
	if width <= 0 {
		width = 1000
	}
	if height <= 0 {
		height = 718
	}
	opts := application.WebviewWindowOptions{
		Title:            title,
		Width:            width,
		Height:           height,
		BackgroundColour: application.NewRGB(6, 7, 15),
		StartState:       startState,
		// 无边框：标题栏与“视图”菜单均由前端 BlankLayout 自绘（菜单在标题右侧）
		Frameless: true,
		Windows: application.WindowsWindow{
			// 启用 WebView2 非客户区支持，使 app-region: drag 自绘标题栏可拖动
			NonClientRegionSupport: true,
		},
		URL: "/#" + path,
	}
	// 单例窗口：以路由路径作为唯一名称，已存在则聚焦复用
	if single {
		name := "win:" + path
		if w, ok := a.app.Window.GetByName(name); ok {
			w.Show()
			w.UnMinimise()
			w.Focus()
			return nil
		}
		opts.Name = name
	}
	a.app.Window.NewWithOptions(opts)
	return nil
}

/* ============================ 串口 / 网络调试 ============================ */

// GetSerialPorts 获取系统串口列表
func (a *App) GetSerialPorts() ([]usbserial.SerialDeviceInfo, error) {
	return usbserial.ListSerialPorts()
}

// GetLocalIPs 获取本机可用 IPv4 地址列表（供 TCP/UDP 本地地址选择）
func (a *App) GetLocalIPs() []string {
	return usbserial.LocalIPs()
}

// OpenSerial 打开串口
func (a *App) OpenSerial(cfg usbserial.ConnConfig) error {
	return usbserial.OpenSerial(cfg)
}

// OpenTCPClient 作为 TCP 客户端连接远端
func (a *App) OpenTCPClient(cfg usbserial.ConnConfig) error {
	return usbserial.OpenTCPClient(cfg)
}

// ListenTCPServer 作为 TCP 服务端侦听
func (a *App) ListenTCPServer(cfg usbserial.ConnConfig) error {
	return usbserial.ListenTCPServer(cfg)
}

// OpenUDP 打开 UDP 连接
func (a *App) OpenUDP(cfg usbserial.ConnConfig) error {
	return usbserial.OpenUDP(cfg)
}

// CloseConnection 关闭当前连接（串口/TCP/UDP 通用）
func (a *App) CloseConnection() error {
	return usbserial.Close()
}

// SendSerialHex 以 hex 字符串发送数据到当前连接对端，返回发送字节数
func (a *App) SendSerialHex(hexData string) (int, error) {
	return usbserial.SendHex(hexData)
}

// SetSerialControlBits 实时切换串口 RTS / DTR 电平
func (a *App) SetSerialControlBits(rts, dtr bool) error {
	return usbserial.SetControlBits(rts, dtr)
}

// ConnectionStatus 返回当前连接状态
func (a *App) ConnectionStatus() usbserial.ConnStatus {
	return usbserial.Status()
}

// GetDevices 暴露给前端：获取网卡列表
//
//	func (a *App) GetDevicesSniffer() ([]sniffer.DeviceInfo, error) {
//		return sniffer.GetDevices()
//	}
func (a *App) GetDevicesSniffer() ([]sniffer.PcapDeviceInfo, error) {
	return sniffer.GetPcapDevices()
}

// StartCaptureSniffer 暴露给前端：开始抓包
// 注意：devices 参数现在直接接收 Npcap 的真实设备名 (pcapName)
// 新增 ports 参数，例如传入 ["80"] 或 ["80", "443"]
func (a *App) StartCaptureSniffer(devices []string, protocols []string, ports []string) error {
	fmt.Printf("[Sniffer] 收到抓包请求 - 网卡: %v, 协议: %v, 端口: %v\n", devices, protocols, ports)
	a.app.Event.Emit("packet:update", "收到抓包请求")

	// 将 ports 传递给底层
	packetCh, err := sniffer.StartCapture(a.ctx, devices, protocols, ports)
	if err != nil {
		fmt.Printf("[Sniffer] 抓包启动失败: %v\n", err)
		return err
	}

	fmt.Println("[Sniffer] 抓包已成功启动，开始监听数据包...")

	// 启动后台协程，将抓到的包实时推送到前端
	go func() {
		for pkt := range packetCh {
			if pkt.Request != nil {
				fmt.Printf("[Sniffer] %s %s%s -> %s:%d\n", pkt.Request.Method, pkt.Request.Host, pkt.Request.URL, pkt.DstIP, pkt.DstPort)
			} else if pkt.Response != nil {
				fmt.Printf("[Sniffer] HTTP Response %s from %s:%d\n", pkt.Response.Status, pkt.SrcIP, pkt.SrcPort)
			} else {
				fmt.Printf("[Sniffer] %s %s:%d -> %s:%d [%s]\n", pkt.Protocol, pkt.SrcIP, pkt.SrcPort, pkt.DstIP, pkt.DstPort, pkt.Flags)
			}
			a.app.Event.Emit("packet:update", pkt)
		}
		fmt.Println("[Sniffer] 数据包通道已关闭，抓包协程退出。")
	}()

	return nil
}

// StopCapture 暴露给前端：停止抓包
func (a *App) StopCaptureSniffer() {
	// 3. 调用 sniffer 包的 StopCapture 函数
	sniffer.StopCapture()
}

/* ================= MITM 代理抓包（HTTP/HTTPS 解密） ================= */

// StartSnifferProxy 启动 MITM 代理：安装根证书、设置系统代理、开始捕获。
// 每捕获一条请求/响应，通过 "sniffer:flow" 事件推送给前端。
func (a *App) StartSnifferProxy(port int) error {
	a.snifferMu.Lock()
	defer a.snifferMu.Unlock()
	if a.snifferRunning {
		return fmt.Errorf("抓包代理已在运行")
	}
	if port <= 0 || port > 65535 {
		port = 8888
	}

	m, err := sniffer.NewMitm()
	if err != nil {
		return fmt.Errorf("初始化代理失败: %w", err)
	}
	if err := m.Start(port, func(f sniffer.Flow) {
		if a.app != nil {
			a.app.Event.Emit("sniffer:flow", f)
		}
	}); err != nil {
		return err
	}

	// 安装根证书 + 设置系统代理（失败仅告警，不影响代理本身）
	if err := sniffer.InstallCA(m.CertPath()); err != nil {
		fmt.Printf("[Sniffer] 安装根证书失败: %v\n", err)
	}
	if err := sniffer.SetSystemProxy(port); err != nil {
		fmt.Printf("[Sniffer] 设置系统代理失败: %v\n", err)
	}

	a.mitm = m
	a.snifferRunning = true
	fmt.Printf("[Sniffer] MITM 代理已启动，监听 127.0.0.1:%d\n", port)
	return nil
}

// StopSnifferProxy 停止 MITM 代理：还原系统代理并卸载根证书
func (a *App) StopSnifferProxy() error {
	a.snifferMu.Lock()
	defer a.snifferMu.Unlock()
	if !a.snifferRunning || a.mitm == nil {
		return nil
	}
	sniffer.ClearSystemProxy()
	a.mitm.Stop()
	sniffer.UninstallCA()
	a.mitm = nil
	a.snifferRunning = false
	fmt.Println("[Sniffer] MITM 代理已停止，系统代理已还原")
	return nil
}

// SnifferProxyStatus 返回代理运行状态与端口
func (a *App) SnifferProxyStatus() (bool, int) {
	a.snifferMu.Lock()
	defer a.snifferMu.Unlock()
	if a.mitm != nil {
		return a.snifferRunning, a.mitm.Port()
	}
	return a.snifferRunning, 0
}

/* ============================ Redis 管理 ============================ */

// RedisOpen 建立并登记一个 Redis 连接
func (a *App) RedisOpen(cfg redis.ConnConfig) error {
	return redis.Open(cfg)
}

// RedisClose 关闭指定 Redis 连接
func (a *App) RedisClose(id string) error {
	return redis.Close(id)
}

// RedisStatus 获取服务器状态概览
func (a *App) RedisStatus(id string) (redis.ServerStatus, error) {
	return redis.Status(id)
}

// RedisScan 增量扫描键（pattern 为空默认 *，limit 限制返回数量）
func (a *App) RedisScan(id, pattern string, count, limit int) (redis.ScanResult, error) {
	return redis.Scan(id, pattern, count, limit)
}

// RedisGetKey 读取键详情（按类型返回结构）
func (a *App) RedisGetKey(id, key string) (redis.KeyValue, error) {
	return redis.GetKey(id, key)
}

// RedisSetString 写入/更新 string 类型键
func (a *App) RedisSetString(id, key, value string, ttl int) error {
	return redis.SetString(id, key, value, ttl)
}

// RedisDeleteKeys 删除若干键，返回删除数量
func (a *App) RedisDeleteKeys(id string, keys []string) (int64, error) {
	return redis.DeleteKeys(id, keys)
}

// RedisSetTTL 设置/清除键过期时间
func (a *App) RedisSetTTL(id, key string, ttl int) error {
	return redis.SetTTL(id, key, ttl)
}

// RedisSlowlog 拉取慢查询日志
func (a *App) RedisSlowlog(id string, count int) ([]redis.SlowlogEntry, error) {
	return redis.Slowlog(id, count)
}

// RedisStartMonitor 开始监控命令（逐条通过 redis:monitor 事件推送）
func (a *App) RedisStartMonitor(cfg redis.ConnConfig) error {
	return redis.StartMonitor(cfg)
}

// RedisStopMonitor 停止监控命令
func (a *App) RedisStopMonitor() error {
	return redis.StopMonitor()
}

// RedisSaveConnections 将连接配置保存到用户配置文件，返回文件路径
func (a *App) RedisSaveConnections(configs []redis.ConnConfig) (string, error) {
	return redis.SaveConnections(configs)
}

// RedisLoadConnections 从配置文件读取连接列表
func (a *App) RedisLoadConnections() ([]redis.ConnConfig, error) {
	return redis.LoadConnections()
}

// RedisConnectionsFilePath 返回配置文件路径
func (a *App) RedisConnectionsFilePath() string {
	return redis.ConnectionsFilePath()
}
