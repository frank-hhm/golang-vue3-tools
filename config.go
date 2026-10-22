package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

// ============================================================
// 全局应用配置（Go 侧 JSON 文件持久化，与 redis/mysql 连接配置同目录）
//
// 位置：os.UserConfigDir()/go-tools/config.json
// 内容：日志设置（开关/级别/路径）+ 可编辑工具列表（含快捷标记、全局热键等）
// 读取方：全局日志器、浮动球原生菜单、全局热键、前端设置页/首页
// ============================================================

// LogConfig 日志设置
type LogConfig struct {
	Enabled bool   `json:"enabled"` // 是否写文件日志
	Level   string `json:"level"`   // debug | info | warn | error（最低记录级别）
	Path    string `json:"path"`    // 日志目录；空 = 默认（配置目录下 go-tools/logs）。按级别+日期自动分文件：Error20261022.log / Info20261022.log 等
}

// BallConfig 桌面浮动球设置
type BallConfig struct {
	Enabled bool `json:"enabled"` // 是否在启动时显示浮动球（设置页可即时显隐）
}

// GeneralConfig 应用通用设置
type GeneralConfig struct {
	AutoStart bool   `json:"autoStart"` // 是否注册开机（Windows 登录时）自动启动，写入 HKCU\...\Run 注册表键
	AppName   string `json:"appName"`   // 应用显示名称（窗口标题/托盘提示），默认 "Tools"
	AppLogo   string `json:"appLogo"`   // 自定义 logo 图片绝对路径，空表示使用内置默认
	Theme     string `json:"theme"`     // 主题模式：light / dark / auto，默认 light（前端消费，后端仅持久化）
}

// ToolConfig 单个工具项（首页列表 / 浮动球快捷 / 全局热键 共用）
type ToolConfig struct {
	ID         string `json:"id"`         // 稳定唯一标识（编辑/删除用）
	Key        string `json:"key"`        // 内置功能标识（对应工具类型）
	Name       string `json:"name"`       // 显示标题
	Path       string `json:"path"`       // 前端路由，如 /tools/timestamp（与 Event 二选一）
	Event      string `json:"event"`      // 直接触发的事件（如 screenshot），空表示走 Path
	Icon       string `json:"icon"`       // 图标名（Arco 图标，去掉 icon- 前缀亦可）
	Single     bool   `json:"single"`     // 窗口单例
	Fullscreen bool   `json:"fullscreen"` // 新窗口最大化启动
	Width      int    `json:"width"`      // 视口宽（<=0 继承默认）
	Height     int    `json:"height"`     // 视口高（<=0 继承默认）
	Shortcut   string `json:"shortcut"`   // 全局热键，如 "Ctrl+Alt+S"，空表示未设置
	Quick      bool   `json:"quick"`      // 是否出现在浮动球快捷菜单
}

// AppConfig 应用总配置
type AppConfig struct {
	Log     LogConfig     `json:"log"`
	Ball    BallConfig    `json:"ball"`
	General GeneralConfig `json:"general"`
	Tools   []ToolConfig  `json:"tools"`
}

var (
	appCfgMu  sync.RWMutex
	appCfg    AppConfig
	appCfgSet bool
)

// DefaultConfig 返回带默认日志设置与默认工具列表的配置。
func DefaultConfig() AppConfig {
	return AppConfig{
		Log:     LogConfig{Enabled: true, Level: "info", Path: ""},
		Ball:    BallConfig{Enabled: false},
		General: GeneralConfig{AutoStart: false, AppName: "Tools", AppLogo: "", Theme: "light"},
		Tools: []ToolConfig{
			{ID: "screenshot", Key: "screenshot", Name: "截图", Event: "screenshot", Icon: "camera", Quick: true},
			{ID: "colorpicker", Key: "colorpicker", Name: "颜色取色器", Path: "/tools/colorpicker", Icon: "color-picker", Single: true, Width: 480, Height: 560, Quick: true},
			{ID: "video", Key: "video", Name: "视频录制", Path: "/tools/video", Icon: "video-camera", Single: true, Fullscreen: true, Quick: true},
			{ID: "qrcode", Key: "qrcode", Name: "二维码生成", Path: "/tools/qrcode", Icon: "qrcode", Quick: true},
			{ID: "timestamp", Key: "timestamp", Name: "时间戳转换", Path: "/tools/timestamp", Icon: "clock-circle", Quick: true},
			{ID: "strrandom", Key: "strrandom", Name: "随机字符串生成", Path: "/tools/strrandom", Icon: "random"},
			{ID: "encdec", Key: "encdec", Name: "文字加密解密", Path: "/tools/encdec", Icon: "lock"},
			{ID: "hexconvert", Key: "hexconvert", Name: "进制转换", Path: "/tools/hexconvert", Icon: "swap"},
			{ID: "imageBase64", Key: "imageBase64", Name: "图片base64", Path: "/tools/imageBase64", Icon: "image"},
			{ID: "sniffer", Key: "sniffer", Name: "抓包工具", Path: "/tools/sniffer", Icon: "wifi", Single: true, Fullscreen: true},
			{ID: "websocket", Key: "websocket", Name: "WebSocket调试", Path: "/tools/websocket", Icon: "connection", Fullscreen: true},
			{ID: "usbserial", Key: "usbserial", Name: "USB调试", Path: "/tools/usbserial", Icon: "usb"},
			{ID: "redis", Key: "redis", Name: "Redis管理", Path: "/tools/redis", Icon: "storage", Single: true, Fullscreen: true},
			{ID: "mysql", Key: "mysql", Name: "MySQL管理", Path: "/tools/mysql", Icon: "code-block", Single: true, Fullscreen: true},
			{ID: "settings", Key: "settings", Name: "设置", Path: "/tools/settings", Icon: "settings", Single: true},
		},
	}
}

// configFilePath 返回配置文件路径，必要时创建目录。
func configFilePath() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		home, herr := os.UserHomeDir()
		if herr != nil {
			return "", herr
		}
		base = filepath.Join(home, ".config")
	}
	dir := filepath.Join(base, "go-tools")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.json"), nil
}

// defaultLogDir 默认日志目录：程序目录下 logs（取不到 exe 路径时退回用户配置目录 go-tools/logs）。
func defaultLogDir() string {
	if exe, err := os.Executable(); err == nil {
		return filepath.Join(filepath.Dir(exe), "logs")
	}
	base, err := os.UserConfigDir()
	if err != nil {
		home, herr := os.UserHomeDir()
		if herr != nil {
			return "logs"
		}
		base = filepath.Join(home, ".config")
	}
	return filepath.Join(base, "go-tools", "logs")
}

// normalizeConfig 补齐缺省值（级别非法回退 info；工具 ID 缺失用 Key 兜底；AppName 空回退 "Tools"）。
func normalizeConfig(cfg *AppConfig) {
	switch cfg.Log.Level {
	case "debug", "info", "warn", "error":
	default:
		cfg.Log.Level = "info"
	}
	if cfg.General.AppName == "" {
		cfg.General.AppName = "Tools"
	}
	switch cfg.General.Theme {
	case "light", "dark", "auto":
	default:
		cfg.General.Theme = "light"
	}
	for i := range cfg.Tools {
		if cfg.Tools[i].ID == "" {
			cfg.Tools[i].ID = cfg.Tools[i].Key
		}
	}
}

// LoadAppConfig 读取配置；文件不存在/解析失败时回落默认值。结果缓存为当前配置。
func LoadAppConfig() AppConfig {
	cfg := DefaultConfig()
	if path, err := configFilePath(); err == nil {
		if data, err := os.ReadFile(path); err == nil && len(data) > 0 {
			var loaded AppConfig
			if err := json.Unmarshal(data, &loaded); err == nil {
				if loaded.Log.Level != "" || loaded.Log.Path != "" {
					cfg.Log = loaded.Log
				}
				cfg.Ball = loaded.Ball
				cfg.General = loaded.General
				if len(loaded.Tools) > 0 {
					cfg.Tools = loaded.Tools
				}
			}
		}
	}
	normalizeConfig(&cfg)
	appCfgMu.Lock()
	appCfg = cfg
	appCfgSet = true
	appCfgMu.Unlock()
	return cfg
}

// CurrentConfig 返回当前缓存配置（惰性加载）。
func CurrentConfig() AppConfig {
	appCfgMu.RLock()
	c, set := appCfg, appCfgSet
	appCfgMu.RUnlock()
	if !set {
		return LoadAppConfig()
	}
	return c
}

// SaveAppConfig 规范化并持久化配置，同时更新当前缓存；返回文件路径。
func SaveAppConfig(cfg AppConfig) (string, error) {
	normalizeConfig(&cfg)
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return "", err
	}
	path, err := configFilePath()
	if err != nil {
		return "", err
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return "", err
	}
	appCfgMu.Lock()
	appCfg = cfg
	appCfgSet = true
	appCfgMu.Unlock()
	return path, nil
}
