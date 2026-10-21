package main

import (
	"context"
	"fmt"
	"go-tools/tools/sniffer"
	"go-tools/tools/usbserial"

	"github.com/wailsapp/wails/v3/pkg/application"
)

// App struct
type App struct {
	app      *application.App
	ctx      context.Context
	engine   *sniffer.Engine
	packetCh chan sniffer.PacketInfo
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

// GetSerialPorts 暴露给 Vue 前端的获取串口方法
func (a *App) GetSerialPorts() ([]usbserial.SerialDeviceInfo, error) {
	return usbserial.ListSerialPorts()
}
func (a *App) ConnectSerial(portId string) error {
	return usbserial.ConnectSerial(portId)
}
func (a *App) StartReadingSerial(portId string) error {
	return usbserial.StartReading(portId)
}
func (a *App) StopReadingSerial(portId string) error {
	return usbserial.StopReading(portId)
}
func (a *App) DisconnectSerial(portId string) error {
	return usbserial.DisconnectSerial(portId)
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
