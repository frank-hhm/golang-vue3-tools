package usbserial

import (
	"context"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
	"go.bug.st/serial"
	"go.bug.st/serial/enumerator"
)

// 1. 定义用于管理多串口的状态结构体
type PortState struct {
	port       serial.Port
	isReading  bool
	readStopCh chan bool
}

var (
	portMap       sync.Map // 使用 sync.Map 并发安全地保存所有已连接的串口
	wailsCtx      context.Context
	currentPort   string     // ⭐ 新增：记录当前操作的串口
	currentPortMu sync.Mutex // ⭐ 新增：保护 currentPort 的并发安全
)

// SetWailsContext 保存 Wails 上下文
func SetWailsContext(ctx context.Context) {
	wailsCtx = ctx
}

// SerialDeviceInfo 串口设备信息
type SerialDeviceInfo struct {
	PortName     string `json:"port_name"`
	IsUSB        bool   `json:"is_usb"`
	SerialNumber string `json:"serial_number"`
	Product      string `json:"product"`
	VendorID     string `json:"vendor_id"`
	ProductID    string `json:"product_id"`
	IsUSBDevice  bool   `json:"is_usb_device"`
}

// ListSerialPorts 获取所有串口列表
func ListSerialPorts() ([]SerialDeviceInfo, error) {
	ports, err := enumerator.GetDetailedPortsList()
	if err != nil {
		return nil, err
	}
	var devices []SerialDeviceInfo
	for _, port := range ports {
		if port.PID != "" {
			devices = append(devices, SerialDeviceInfo{
				PortName:     port.Name,
				IsUSB:        port.IsUSB,
				SerialNumber: port.SerialNumber,
				Product:      port.Product,
				VendorID:     port.VID,
				ProductID:    port.PID,
				IsUSBDevice:  port.IsUSB,
			})
		}
	}
	return devices, nil
}

// ConnectSerial 连接指定串口
func ConnectSerial(portId string) error {
	portId = strings.ToUpper(portId)

	// 检查是否已经连接
	if _, loaded := portMap.Load(portId); loaded {
		return fmt.Errorf("串口 %s 已经处于连接状态", portId)
	}

	mode := &serial.Mode{
		BaudRate: 9600,
		DataBits: 8,
		StopBits: serial.OneStopBit,
		Parity:   serial.NoParity,
	}

	port, err := serial.Open(portId, mode)
	if err != nil {
		return fmt.Errorf("打开串口 %s 失败: %v", portId, err)
	}

	if err = port.SetReadTimeout(100 * time.Millisecond); err != nil {
		port.Close()
		return fmt.Errorf("设置读超时失败: %v", err)
	}

	// 存入 Map
	portMap.Store(portId, &PortState{port: port})

	// ⭐ 新增：连接成功后，自动将其设为当前串口
	currentPortMu.Lock()
	currentPort = portId
	currentPortMu.Unlock()

	fmt.Printf("✅ 成功连接串口: %s (已设为当前串口)\n", portId)
	return nil
}

// StartReading 启动指定串口的监听
func StartReading(portId string) error {
	currentPortMu.Lock()
	if portId == "" {
		portId = currentPort
	}
	currentPortMu.Unlock()

	portId = strings.ToUpper(portId)
	if portId == "" {
		return fmt.Errorf("未指定串口，且当前没有活动的串口")
	}

	val, ok := portMap.Load(portId)
	if !ok {
		return fmt.Errorf("串口 %s 未连接", portId)
	}

	state := val.(*PortState)
	if state.isReading {
		return fmt.Errorf("串口 %s 正在读取中", portId)
	}

	state.isReading = true
	state.readStopCh = make(chan bool)

	go func() {
		buf := make([]byte, 1024)

		fmt.Printf("🚀 开始监听串口 %s 数据...\n", portId)
		for {
			select {
			case <-state.readStopCh:
				fmt.Printf("🛑 串口 %s 监听已停止\n", portId)
				return
			default:
				n, err := state.port.Read(buf)
				if err != nil {
					if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
						continue
					}
					fmt.Printf("❌ 串口 %s 读取错误: %v\n", portId, err)
					state.isReading = false
					return
				}

				if n > 0 {
					fmt.Printf("📩 串口 %s 收到完整 HEX 数据: %s\n", portId, buf)
					if wailsCtx != nil {
						runtime.EventsEmit(wailsCtx, "serial-data", portId, buf)
					}
				}
			}
		}
	}()
	return nil
}

// StopReading 停止指定串口的数据监听，不传参数则停止当前串口
func StopReading(portId string) error {
	currentPortMu.Lock()
	if portId == "" {
		portId = currentPort
	}
	currentPortMu.Unlock()

	portId = strings.ToUpper(portId)
	if portId == "" {
		return fmt.Errorf("未指定串口，且当前没有活动的串口")
	}

	val, ok := portMap.Load(portId)
	if !ok {
		return fmt.Errorf("串口 %s 未连接", portId)
	}

	state := val.(*PortState)
	if state.isReading && state.readStopCh != nil {
		close(state.readStopCh)
		state.isReading = false
		fmt.Printf("⏸️ 已暂停串口 %s 的数据监听\n", portId)
	}
	return nil
}

// DisconnectSerial 断开指定串口，如果传入 "all" 则断开所有
func DisconnectSerial(portId string) error {
	portId = strings.ToUpper(portId)

	// 1. 处理断开所有的情况
	if portId == "ALL" {
		var lastErr error
		portMap.Range(func(key, value interface{}) bool {
			pId := key.(string)
			if err := closeSinglePort(pId, value.(*PortState)); err != nil {
				lastErr = err
			}
			return true // 继续遍历
		})
		if lastErr != nil {
			return fmt.Errorf("断开所有串口时部分失败: %v", lastErr)
		}
		fmt.Println("✅ 已成功断开所有串口")
		return nil
	}

	// 2. 处理断开指定串口的情况
	val, ok := portMap.Load(portId)
	if !ok {
		return fmt.Errorf("串口 %s 未连接或不存在", portId)
	}

	if err := closeSinglePort(portId, val.(*PortState)); err != nil {
		return err
	}
	fmt.Printf("✅ 成功断开串口: %s\n", portId)
	return nil
}

// closeSinglePort 内部方法：安全关闭单个串口
func closeSinglePort(portId string, state *PortState) error {
	// 停止读取协程
	if state.isReading && state.readStopCh != nil {
		close(state.readStopCh)
		state.isReading = false
	}
	// 关闭底层端口
	err := state.port.Close()
	// 无论关闭是否报错，都从 Map 中移除
	portMap.Delete(portId)
	return err
}
