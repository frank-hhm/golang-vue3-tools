package main

import (
	"embed"
	"log"
	"time"

	"go-tools/tools/mysql"
	"go-tools/tools/redis"
	"go-tools/tools/screenshot"
	"go-tools/tools/sniffer"
	"go-tools/tools/usbserial"

	"github.com/wailsapp/wails/v3/pkg/application"
)

// Wails uses Go's `embed` package to embed the frontend files into the binary.
// Any files in the frontend/dist folder will be embedded into the binary and
// made available to the frontend.
// See https://pkg.go.dev/embed for more information.

//go:embed all:frontend/dist
var assets embed.FS

// 默认应用 logo
//
//go:embed logo.png
var defaultLogoPNG []byte

// 托盘图标（Windows 支持 PNG/ICO 字节）
//
//go:embed build/windows/icon.ico
var trayIconData []byte

func init() {
	// Register a custom event whose associated data type is string.
	// This is not required, but the binding generator will pick up registered events
	// and provide a strongly typed JS/TS API for them.
	application.RegisterEvent[string]("time")
	application.RegisterEvent[sniffer.Flow]("sniffer:flow")
	application.RegisterEvent[usbserial.SerialData]("serial:data")
	application.RegisterEvent[usbserial.ConnStatusEvent]("serial:status")
	application.RegisterEvent[redis.MonitorLine]("redis:monitor")
	application.RegisterEvent[redis.InstallProgress]("redis:install")
	application.RegisterEvent[redis.ServerLogLine]("redis:serverlog")
	application.RegisterEvent[string]("video:saved")
	application.RegisterEvent[string]("video:cmd")
	application.RegisterEvent[VideoStatus]("video:status")
	application.RegisterEvent[UIMessage]("ui:message")
	application.RegisterEvent[string]("ball:action")
	application.RegisterEvent[string]("tool:activate")
	application.RegisterEvent[mysql.InstallProgress]("mysql:install")
}

func main() {
	// 加载全局配置（用户配置目录 config.json）并初始化可配置文件日志
	startupCfg := LoadAppConfig()
	logDir := InitLogger(startupCfg.Log)
	opLogf("app", "=== go-tools 启动 === 日志目录: %s (enabled=%v level=%s)", logDir, startupCfg.Log.Enabled, startupCfg.Log.Level)

	// Create an instance of the app structure
	appInstance := NewApp()
	// Create a new Wails application by providing the necessary options.
	// Variables 'Name' and 'Description' are for application metadata.
	// 'Assets' configures the asset server with the 'FS' variable pointing to the frontend files.
	// 'Bind' is a list of Go struct instances. The frontend has access to the methods of these instances.
	// 'Mac' options tailor the application when running an macOS.
	app := application.New(application.Options{
		Name:        "go-tools",
		Description: "A demo of using raw HTML & CSS",
		Services: []application.Service{
			application.NewService(&GreetService{}),
			application.NewService(appInstance),
			application.NewService(screenshot.NewScreenshotService()),
		},
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
		Windows: application.WindowsOptions{
			// 录制期间工具窗口处于隐藏状态，保持其 JS 定时器/捕获回调不被后台节流
			AdditionalBrowserArgs: []string{
				"--disable-background-timer-throttling",
				"--disable-backgrounding-occluded-windows",
				"--disable-renderer-backgrounding",
			},
		},
		Mac: application.MacOptions{
			ApplicationShouldTerminateAfterLastWindowClosed: true,
		},
	})

	// 注入应用实例，使 App 可以向前端推送事件
	appInstance.SetApp(app)

	// 注入事件发射器，使 usbserial 包可将收到/状态数据实时推送给前端
	usbserial.SetEmitter(func(name string, data ...any) {
		app.Event.Emit(name, data...)
	})

	// 注入事件发射器，使 redis 包可将监控命令实时推送给前端
	redis.SetEmitter(func(name string, data ...any) {
		app.Event.Emit(name, data...)
	})

	// 注入事件发射器，使 mysql 包可将安装进度实时推送给前端
	mysql.SetEmitter(func(name string, data ...any) {
		app.Event.Emit(name, data...)
	})

	// Create a new window with the necessary options.
	// 'Title' is the title of the window.
	// 'Mac' options tailor the window when running on macOS.
	// 'BackgroundColour' is the background colour of the window.
	// 'URL' is the URL that will be loaded into the webview.
	mainWindow := app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title: startupCfg.General.AppName,
		// Window sized to the golden ratio (1000 / 618 ≈ 1.618).
		Width:  1000,
		Height: 718,
		Mac: application.MacWindow{
			InvisibleTitleBarHeight: 50,
			Backdrop:                application.MacBackdropTranslucent,
			TitleBar:                application.MacTitleBarHiddenInset,
		},
		BackgroundColour: application.NewRGB(6, 7, 15),
		URL:              "/",
	})

	// 注入主窗口与托盘图标：关闭主窗口时隐藏到托盘而非退出，托盘右键提供“设置/退出”
	appInstance.SetMainWindow(mainWindow)
	appInstance.SetTrayIcon(trayIconData)
	appInstance.SetDefaultLogo(defaultLogoPNG)

	// Create a goroutine that emits an event containing the current time every second.
	// The frontend can listen to this event and update the UI accordingly.
	go func() {
		for {
			now := time.Now().Format(time.RFC1123)
			app.Event.Emit("time", now)
			time.Sleep(time.Second)
		}
	}()

	// Run the application. This blocks until the application has been exited.
	err := app.Run()

	// If an error occurred while running the application, log it and exit.
	if err != nil {
		log.Fatal(err)
	}
}
