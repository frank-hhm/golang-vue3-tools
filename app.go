package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"go-tools/tools/mysql"
	"go-tools/tools/redis"
	"go-tools/tools/sniffer"
	"go-tools/tools/usbserial"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"
	"unsafe"

	"github.com/wailsapp/wails/v3/pkg/application"
	"golang.org/x/sys/windows"
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

	recMu         sync.Mutex
	recWidget     application.Window // 录制悬浮小窗
	recToolWin    application.Window // 发起录制的视频工具窗口（录制期间隐藏）
	videoSessions map[string]*videoSaveSession

	toastMu  sync.Mutex
	toastWin application.Window // 原生消息通知小窗（屏幕右下角，自动消失）

	notifyMu  sync.Mutex
	notifyWin application.Window // 系统风格通知卡片（屏幕右下角，驻留至手动关闭）
}

// videoSaveSession 录制完成后分块上传的视频缓冲（用于原生“另存为”写盘）
type videoSaveSession struct {
	name string
	data bytes.Buffer
}

// VideoStatus 录制悬浮小窗展示用的状态（由录制工具窗经 Go 转发）
type VideoStatus struct {
	State   string `json:"state"`   // recording|paused|saving|cancelsave|savefail|saved|closed
	Elapsed int    `json:"elapsed"` // 已录制秒数
	Note    string `json:"note"`    // 进度/路径/错误信息
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{videoSessions: map[string]*videoSaveSession{}}
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

/* ============================ MySQL 管理 ============================ */

// MySQLOpen 建立并登记一个 MySQL 连接
func (a *App) MySQLOpen(cfg mysql.ConnConfig) error { return mysql.Open(cfg) }

// MySQLClose 关闭指定连接
func (a *App) MySQLClose(id string) error { return mysql.Close(id) }

// MySQLListDatabases 列出数据库
func (a *App) MySQLListDatabases(id string) ([]string, error) { return mysql.ListDatabases(id) }

// MySQLListTables 列出基表
func (a *App) MySQLListTables(id, database string) ([]mysql.TableInfo, error) {
	return mysql.ListTables(id, database)
}

// MySQLListViews 列出视图
func (a *App) MySQLListViews(id, database string) ([]mysql.ViewInfo, error) {
	return mysql.ListViews(id, database)
}

// MySQLListRoutines 列出函数/存储过程
func (a *App) MySQLListRoutines(id, database string) ([]mysql.RoutineInfo, error) {
	return mysql.ListRoutines(id, database)
}

// MySQLTableSchema 返回表结构（列/索引/DDL）
func (a *App) MySQLTableSchema(id, database, table string) (mysql.TableSchema, error) {
	return mysql.GetTableSchema(id, database, table)
}

// MySQLAddColumn 新增列
func (a *App) MySQLAddColumn(id, database, table string, col mysql.ColumnDef) error {
	return mysql.AddColumn(id, database, table, col)
}

// MySQLModifyColumn 修改/重命名列
func (a *App) MySQLModifyColumn(id, database, table, oldName string, col mysql.ColumnDef) error {
	return mysql.ModifyColumn(id, database, table, oldName, col)
}

// MySQLDropColumn 删除列
func (a *App) MySQLDropColumn(id, database, table, col string) error {
	return mysql.DropColumn(id, database, table, col)
}

// MySQLTableData 分页浏览表数据
func (a *App) MySQLTableData(id, database, table, filter, orderBy string, desc bool, page, size int) (mysql.QueryResult, error) {
	return mysql.TableData(id, database, table, filter, orderBy, desc, page, size)
}

// MySQLUpdateCell 按主键更新单元格
func (a *App) MySQLUpdateCell(id, database, table, pkCol, pkVal, col, val string) error {
	return mysql.UpdateCell(id, database, table, pkCol, pkVal, col, val)
}

// MySQLInsertRow 插入一行
func (a *App) MySQLInsertRow(id, database, table string, data map[string]string) (int64, error) {
	return mysql.InsertRow(id, database, table, data)
}

// MySQLDeleteRow 按主键删除一行
func (a *App) MySQLDeleteRow(id, database, table, pkCol, pkVal string) error {
	return mysql.DeleteRow(id, database, table, pkCol, pkVal)
}

// MySQLExecute 执行 SQL（查询返回结果集，其余返回受影响行）
func (a *App) MySQLExecute(id, database, sqlText string) (mysql.QueryResult, error) {
	return mysql.Execute(id, database, sqlText)
}

// MySQLBeginTx 开启事务，返回事务 ID
func (a *App) MySQLBeginTx(id, database string) (string, error) { return mysql.BeginTx(id, database) }

// MySQLTxExecute 在事务中执行语句
func (a *App) MySQLTxExecute(txID, sqlText string) (mysql.QueryResult, error) {
	return mysql.TxExecute(txID, sqlText)
}

// MySQLCommitTx 提交事务
func (a *App) MySQLCommitTx(txID string) error { return mysql.CommitTx(txID) }

// MySQLRollbackTx 回滚事务
func (a *App) MySQLRollbackTx(txID string) error { return mysql.RollbackTx(txID) }

// MySQLBackup 生成 SQL 备份文本
func (a *App) MySQLBackup(id, database string, tables []string) (string, error) {
	return mysql.Backup(id, database, tables)
}

// MySQLSaveConnections 保存连接配置到文件
func (a *App) MySQLSaveConnections(configs []mysql.ConnConfig) (string, error) {
	return mysql.SaveConnections(configs)
}

// MySQLLoadConnections 从文件读取连接列表
func (a *App) MySQLLoadConnections() ([]mysql.ConnConfig, error) { return mysql.LoadConnections() }

/* ============================ 视频录制 ============================ */

// VideoOpenWidget 打开录制悬浮小窗（由视频录制工具页在用户确认共享、录制器已启动后调用）。
// 录制期间会隐藏视频录制工具窗口，避免其入镜形成递归镜像（“无限缩小”）。
// mode: screen|region；fx/fy/fw/fh 为区域相对整屏的裁剪比例（screen 模式忽略）；
// anchorX/anchorY 为小窗锚定的屏幕右下角坐标（DIP）；filename 为建议保存文件名。
func (a *App) VideoOpenWidget(mode string, fps int, fx, fy, fw, fh float64, anchorX, anchorY int, filename string) error {
	if a.app == nil {
		return fmt.Errorf("应用尚未初始化")
	}
	if fps <= 0 {
		fps = 30
	}
	const widgetW, widgetH = 340, 76
	x := anchorX - widgetW - 12
	y := anchorY - widgetH - 12
	if x < 0 {
		x = 0
	}
	if y < 0 {
		y = 0
	}

	// 隐藏（而非关闭）发起录制的工具窗口：录制完成后由 VideoWidgetDone 恢复显示
	if w, ok := a.app.Window.GetByName("win:/win/tools/video"); ok {
		a.recToolWin = w
		w.Hide()
	}

	// 小窗单例：已存在则先关闭重建
	a.recMu.Lock()
	if a.recWidget != nil {
		a.recWidget.Close()
		a.recWidget = nil
	}
	a.recMu.Unlock()

	url := fmt.Sprintf("/#/widget/recorder?mode=%s&fps=%d&fx=%g&fy=%g&fw=%g&fh=%g&name=%s",
		mode, fps, fx, fy, fw, fh, filename)
	w := a.app.Window.NewWithOptions(application.WebviewWindowOptions{
		Name:             "win:recwidget",
		Title:            "屏幕录制",
		Width:            widgetW,
		Height:           widgetH,
		X:                x,
		Y:                y,
		Frameless:        true,
		AlwaysOnTop:      true,
		DisableResize:    true,
		StartState:       application.WindowStateNormal,
		BackgroundColour: application.NewRGB(24, 24, 28),
		URL:              url,
		Windows: application.WindowsWindow{
			// 无边框小窗支持前端 app-region: drag 拖动
			NonClientRegionSupport: true,
		},
	})
	a.recMu.Lock()
	a.recWidget = w
	a.recMu.Unlock()
	return nil
}

// VideoWidgetDone 小窗退出：重新显示工具窗口并推送退出事件（保留 recToolWin 引用供另存对话框挂父窗）
func (a *App) VideoWidgetDone(savePath string) {
	if a.app == nil {
		return
	}
	a.recMu.Lock()
	if a.recWidget != nil {
		a.recWidget.Close()
		a.recWidget = nil
	}
	a.recMu.Unlock()
	if a.recToolWin != nil {
		a.recToolWin.Show()
		a.recToolWin.UnMinimise()
		a.recToolWin.Focus()
	}
	a.app.Event.Emit("video:saved", savePath)
}

// VideoCmd 悬浮小窗 → 录制工具窗的命令转发（start|pause|resume|stop|cancel）
func (a *App) VideoCmd(command string) {
	if a.app == nil {
		return
	}
	a.app.Event.Emit("video:cmd", command)
}

// VideoReport 录制工具窗 → 悬浮小窗的状态上报
func (a *App) VideoReport(state string, elapsed int, note string) {
	if a.app == nil {
		return
	}
	a.app.Event.Emit("video:status", VideoStatus{State: state, Elapsed: elapsed, Note: note})
}

/* ============================ Go 端消息提示（原生桌面通知小窗） ============================ */

// UIMessage 通知载荷（预留给 WebView 内桥接展示；当前 ShowMessage 走原生小窗，不再使用此事件）
type UIMessage struct {
	Level    string `json:"level"`    // info|success|warning|error
	Text     string `json:"text"`     // 提示内容
	Duration int    `json:"duration"` // 毫秒，0 = 前端默认
}

// ShowMessage 在 WebView 之外的桌面（屏幕右上角）弹出一个原生置顶通知小窗，
// 自动消失；同一时间仅保留最新一条通知。
func (a *App) ShowMessage(level string, text string) {
	if a.app == nil {
		return
	}
	if level == "" {
		level = "info"
	}
	const toastW, toastH = 360, 64
	// 屏幕右下角（与系统通知位置一致，基于含 DPI 换算的工作区，避开任务栏）
	_, _, waRight, waBottom := workAreaDIP()
	x := waRight - toastW - 16
	y := waBottom - toastH - 16
	if x < 16 {
		x = 16
	}
	if y < 16 {
		y = 16
	}

	// 新通知替换旧通知
	a.toastMu.Lock()
	if a.toastWin != nil {
		a.toastWin.Close()
		a.toastWin = nil
	}
	a.toastMu.Unlock()

	w := a.app.Window.NewWithOptions(application.WebviewWindowOptions{
		Name:             fmt.Sprintf("win:toast-%d", time.Now().UnixNano()),
		Title:            "通知",
		Width:            toastW,
		Height:           toastH,
		X:                x,
		Y:                y,
		Frameless:        true,
		AlwaysOnTop:      true,
		DisableResize:    true,
		StartState:       application.WindowStateNormal,
		BackgroundColour: application.NewRGB(255, 255, 255),
		URL: fmt.Sprintf("/#/widget/toast?level=%s&text=%s",
			url.QueryEscape(level), url.QueryEscape(text)),
		Windows: application.WindowsWindow{
			// 无边框通知窗支持前端 app-region: drag 拖动
			NonClientRegionSupport: true,
		},
	})
	a.toastMu.Lock()
	a.toastWin = w
	a.toastMu.Unlock()
}

// ToastClose 关闭通知小窗（通知页倒计时结束或用户点关闭时调用）
func (a *App) ToastClose() {
	if a.app == nil {
		return
	}
	a.toastMu.Lock()
	if a.toastWin != nil {
		a.toastWin.Close()
		a.toastWin = nil
	}
	a.toastMu.Unlock()
}

// workAreaDIP 返回主显示器工作区（已排除任务栏）的 DIP 坐标（left, top, right, bottom），
// 按系统 DPI 将物理像素换算为 Wails 窗口使用的 DIP
func workAreaDIP() (int, int, int, int) {
	user32 := windows.NewLazySystemDLL("user32.dll")
	dpi := uintptr(96)
	if r, _, _ := user32.NewProc("GetDpiForSystem").Call(); r > 0 {
		dpi = r
	}
	scale := float64(dpi) / 96.0
	var rc struct {
		Left, Top, Right, Bottom int32
	}
	const spiGetWorkArea = 0x0030
	user32.NewProc("SystemParametersInfoW").Call(spiGetWorkArea, 0, uintptr(unsafe.Pointer(&rc)), 0)
	return int(float64(rc.Left) / scale), int(float64(rc.Top) / scale),
		int(float64(rc.Right) / scale), int(float64(rc.Bottom) / scale)
}

// ShowNotification 在屏幕右下角弹出一张系统风格的通知卡片（仿 Windows 通知：
// 应用名 + 标题 + 内容 + 关闭钮），驻留直到用户点 ✕ 或代码调用 NotifyClose。
func (a *App) ShowNotification(title string, content string) {
	if a.app == nil {
		return
	}
	const notifyW, notifyH = 380, 132
	_, _, waRight, waBottom := workAreaDIP()
	x := waRight - notifyW - 16
	y := waBottom - notifyH - 16
	if x < 16 {
		x = 16
	}
	if y < 16 {
		y = 16
	}

	// 新通知替换旧通知
	a.notifyMu.Lock()
	if a.notifyWin != nil {
		a.notifyWin.Close()
		a.notifyWin = nil
	}
	a.notifyMu.Unlock()

	w := a.app.Window.NewWithOptions(application.WebviewWindowOptions{
		Name:             fmt.Sprintf("win:notify-%d", time.Now().UnixNano()),
		Title:            "通知",
		Width:            notifyW,
		Height:           notifyH,
		X:                x,
		Y:                y,
		Frameless:        true,
		AlwaysOnTop:      true,
		DisableResize:    true,
		StartState:       application.WindowStateNormal,
		BackgroundColour: application.NewRGB(243, 243, 243),
		URL: fmt.Sprintf("/#/widget/notify?title=%s&content=%s",
			url.QueryEscape(title), url.QueryEscape(content)),
		Windows: application.WindowsWindow{
			// 无边框通知卡支持前端 app-region: drag 拖动
			NonClientRegionSupport: true,
		},
	})
	a.notifyMu.Lock()
	a.notifyWin = w
	a.notifyMu.Unlock()
}

// NotifyClose 关闭通知卡片
func (a *App) NotifyClose() {
	if a.app == nil {
		return
	}
	a.notifyMu.Lock()
	if a.notifyWin != nil {
		a.notifyWin.Close()
		a.notifyWin = nil
	}
	a.notifyMu.Unlock()
}

/* ---------------- 系统通知中心（Windows 原生 Toast） ---------------- */

// toastPSScript 通过环境变量 TOAST_TITLE / TOAST_CONTENT 发送 WinRT Toast 的
// PowerShell 脚本（Windows 10+ 内建能力，无需任何依赖；标题+正文两行模板）
const toastPSScript = `
$t = $env:TOAST_TITLE
$c = $env:TOAST_CONTENT
[void][Windows.UI.Notifications.ToastNotificationManager,Windows.UI.Notifications,ContentType=WindowsRuntime]
[void][Windows.Data.Xml.Dom.XmlDocument,Windows.Data.Xml.Dom.XmlDocument,ContentType=WindowsRuntime]
$tpl = [Windows.UI.Notifications.ToastNotificationManager]::GetTemplateContent([Windows.UI.Notifications.ToastTemplateType]::ToastText02)
$txts = $tpl.SelectNodes("//text")
$txts.Item(0).AppendChild($tpl.CreateTextNode($t)) > $null
$txts.Item(1).AppendChild($tpl.CreateTextNode($c)) > $null
$notifier = [Windows.UI.Notifications.ToastNotificationManager]::CreateToastNotifier("go-tools")
$toast = [Windows.UI.Notifications.ToastNotification]::new($tpl)
$notifier.Show($toast)
`

// ShowSystemNotification 直接调用 Windows 系统通知中心弹原生 Toast
// （系统右下角渲染、自动消失、可进入操作中心/专注助手），异步执行不阻塞调用方。
func (a *App) ShowSystemNotification(title string, content string) error {
	if title == "" {
		title = "go-tools"
	}
	go func() {
		// 脚本经 stdin 传入（-Command -），避免命令行引号转义问题
		cmd := exec.Command("powershell.exe", "-NoProfile", "-NonInteractive",
			"-ExecutionPolicy", "Bypass", "-Command", "-")
		cmd.Stdin = strings.NewReader(toastPSScript)
		// 内容经环境变量传递，避免命令行注入/转义问题
		cmd.Env = append(os.Environ(), "TOAST_TITLE="+title, "TOAST_CONTENT="+content)
		cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true} // 不闪现 PowerShell 控制台窗口
		_ = cmd.Run()
	}()
	return nil
}

// VideoSaveBegin 开始一次视频分块上传会话，返回会话 ID
func (a *App) VideoSaveBegin(filename string) (string, error) {
	if filename == "" {
		filename = fmt.Sprintf("recording-%s.mp4", time.Now().Format("20060102-150405"))
	}
	id := fmt.Sprintf("vss-%d", time.Now().UnixNano())
	a.recMu.Lock()
	defer a.recMu.Unlock()
	if a.videoSessions == nil {
		a.videoSessions = map[string]*videoSaveSession{}
	}
	a.videoSessions[id] = &videoSaveSession{name: filename}
	return id, nil
}

// VideoSaveChunk 追加一块 base64 编码的视频数据
func (a *App) VideoSaveChunk(session, base64Chunk string) error {
	a.recMu.Lock()
	defer a.recMu.Unlock()
	s, ok := a.videoSessions[session]
	if !ok {
		return fmt.Errorf("会话不存在或已结束")
	}
	raw, err := base64.StdEncoding.DecodeString(base64Chunk)
	if err != nil {
		return fmt.Errorf("解码分块失败: %w", err)
	}
	if s.data.Len()+len(raw) > 4<<30 { // 4GB 上限保护
		return fmt.Errorf("视频数据超出上限")
	}
	_, _ = s.data.Write(raw)
	return nil
}

// VideoSavePrompt 弹出原生“另存为”对话框选择目录并写盘；用户取消时返回空字符串
func (a *App) VideoSavePrompt(session string) (string, error) {
	a.recMu.Lock()
	s, ok := a.videoSessions[session]
	a.recMu.Unlock()
	if !ok {
		return "", fmt.Errorf("会话不存在或已结束")
	}
	if a.app == nil {
		return "", fmt.Errorf("应用尚未初始化")
	}
	ext := strings.TrimPrefix(strings.ToLower(filepath.Ext(s.name)), ".")
	if ext == "" {
		ext = "mp4"
	}
	dlg := a.app.Dialog.SaveFile().
		SetFilename(s.name).
		SetMessage("另存录制视频").
		AddFilter("视频文件 (*."+ext+")", "*."+ext).
		AddFilter("所有文件 (*.*)", "*.*")
	a.recMu.Lock()
	parent := a.recToolWin // 小窗停止后已关闭，另存对话框挂在工具窗上
	a.recMu.Unlock()
	if parent != nil {
		dlg = dlg.AttachToWindow(parent)
	}
	path, err := dlg.PromptForSingleSelection()
	if err != nil {
		// 用户取消对话框：不算错误（Wails 内部返回 "cancelled by user"），保留会话以便重试另存
		if strings.Contains(strings.ToLower(err.Error()), "cancel") {
			return "", nil
		}
		return "", err
	}
	if path == "" {
		// 用户取消，保留会话以便重试另存
		return "", nil
	}
	if !strings.EqualFold(filepath.Ext(path), "."+ext) {
		path += "." + ext
	}
	if err := os.WriteFile(path, s.data.Bytes(), 0644); err != nil {
		return "", fmt.Errorf("写入文件失败: %w", err)
	}
	a.recMu.Lock()
	delete(a.videoSessions, session)
	a.recMu.Unlock()
	return path, nil
}

// VideoSaveAbort 放弃一次上传会话（不写盘）
func (a *App) VideoSaveAbort(session string) error {
	a.recMu.Lock()
	defer a.recMu.Unlock()
	delete(a.videoSessions, session)
	return nil
}
