//go:build windows

package main

import (
	"bytes"
	"fmt"
	"image"
	"image/png"
	"math"
	"os"
	"runtime"
	"sync"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

// ============================================================
// 桌面浮动球 —— 纯 Go 原生 Win32 实现（不再用 WebView 小窗）
//
// 形态：一个 WS_EX_LAYERED 分层窗口，用 UpdateLayeredWindow 提交
// 「逐像素 alpha」的圆形位图 —— 真正的透明圆球（无灰底方框）。
// 交互：按住拖动用 SetCapture + SetWindowPos 原生移动（无 IPC 往返，不闪动）；
// 单击（未拖动）在球上弹出原生系统菜单（TrackPopupMenu），选择项经
// App.handleBallAction → "ball:action" 事件交回主窗口复用前端开窗/截图逻辑。
// 置顶 WS_EX_TOPMOST、不占任务栏 WS_EX_TOOLWINDOW，默认贴屏幕工作区右下角。
//
// 消息循环安全（沿用本仓库既有约定）：
//   - 专用 goroutine + runtime.LockOSThread，退出时不 UnlockOSThread，
//     线程随之销毁，其队列残留一并带走，绝不毒化复用线程；
//   - 全程不 PostQuitMessage，退出靠 WM_DESTROY 置 done 标志；
//   - 关闭由 stopNativeBall 从其它 goroutine PostMessage(WM_CLOSE) 触发。
// ============================================================

// ---------------- Win32 常量 ----------------

const (
	ballWS_POPUP      = 0x80000000
	ballWS_EX_LAYERED = 0x00080000
	ballWS_EX_TOPMOST = 0x00000008
	ballWS_EX_TOOLWIN = 0x00000080

	ballSW_SHOWNOACTIVATE = 4

	ballWM_DESTROY     = 0x0002
	ballWM_CLOSE       = 0x0010
	ballWM_MOUSEMOVE   = 0x0200
	ballWM_LBUTTONDOWN = 0x0201
	ballWM_LBUTTONUP   = 0x0202
	ballWM_RBUTTONUP   = 0x0205

	ballIDC_ARROW = 32512

	ballSWP_NOSIZE     = 0x0001
	ballSWP_NOZORDER   = 0x0004
	ballSWP_NOACTIVATE = 0x0010

	ballULW_ALPHA = 0x00000002
	ballAC_SRC_OVER_ALPHA

	ballMF_STRING       = 0x0000
	ballMF_SEPARATOR    = 0x0800
	ballTPM_RETURNCMD   = 0x0100
	ballTPM_RIGHTBUTTON = 0x0002

	ballSPI_GETWORKAREA = 0x0030
)

// ---------------- Win32 结构体 ----------------

type ballPoint struct{ X, Y int32 }
type ballSize struct{ CX, CY int32 }
type ballRect struct {
	Left, Top, Right, Bottom int32
}
type ballBlend struct {
	BlendOp             byte
	BlendFlags          byte
	SourceConstantAlpha byte
	AlphaFormat         byte
}
type ballMsg struct {
	Hwnd    uintptr
	Message uint32
	WParam  uintptr
	LParam  uintptr
	Time    uint32
	Pt      ballPoint
}
type ballWndClass struct {
	Style         uint32
	LpfnWndProc   uintptr
	CbClsExtra    int32
	CbWndExtra    int32
	HInstance     uintptr
	HIcon         uintptr
	HCursor       uintptr
	HbrBackground uintptr
	LpszMenuName  *uint16
	LpszClassName *uint16
}
type ballBmih struct {
	BiSize          uint32
	BiWidth         int32
	BiHeight        int32
	BiPlanes        uint16
	BiBitCount      uint16
	BiCompression   uint32
	BiSizeImage     uint32
	BiXPelsPerMeter int32
	BiYPelsPerMeter int32
	BiClrUsed       uint32
	BiClrImportant  uint32
}
type ballBitmapInfo struct {
	Header ballBmih
	Colors uintptr
	_      [0]byte
	pad    [0]byte
	// Colors 占位；实际颜色表紧随其后（32bpp BI_RGB 无需调色板）
	_unused [0]uint32
}

// ---------------- Win32 过程 ----------------

var (
	ballUser32   = windows.NewLazySystemDLL("user32.dll")
	ballGdi32    = windows.NewLazySystemDLL("gdi32.dll")
	ballKern32   = windows.NewLazySystemDLL("kernel32.dll")
	ballNewClass = ballUser32.NewProc("RegisterClassW")
	ballCreate   = ballUser32.NewProc("CreateWindowExW")
	ballShow     = ballUser32.NewProc("ShowWindow")
	ballGetMsg   = ballUser32.NewProc("GetMessageW")
	ballTrans    = ballUser32.NewProc("TranslateMessage")
	ballDispatch = ballUser32.NewProc("DispatchMessageW")
	ballDefProc  = ballUser32.NewProc("DefWindowProcW")
	ballDestroy  = ballUser32.NewProc("DestroyWindow")
	ballPostMsg  = ballUser32.NewProc("PostMessageW")
	ballSetCap   = ballUser32.NewProc("SetCapture")
	ballRelCap   = ballUser32.NewProc("ReleaseCapture")
	ballGetCur   = ballUser32.NewProc("GetCursorPos")
	ballGetRect  = ballUser32.NewProc("GetWindowRect")
	ballSetPos   = ballUser32.NewProc("SetWindowPos")
	ballLoadCur  = ballUser32.NewProc("LoadCursorW")
	ballSetFore  = ballUser32.NewProc("SetForegroundWindow")
	ballGetDC    = ballUser32.NewProc("GetDC")
	ballRelDC    = ballUser32.NewProc("ReleaseDC")
	ballSysInfo  = ballUser32.NewProc("SystemParametersInfoW")
	ballGetDpi   = ballUser32.NewProc("GetDpiForSystem")
	ballCreatePM = ballUser32.NewProc("CreatePopupMenu")
	ballAppendM  = ballUser32.NewProc("AppendMenuW")
	ballTrackPM  = ballUser32.NewProc("TrackPopupMenuEx")
	ballDestroyM = ballUser32.NewProc("DestroyMenu")
	ballCompatDC = ballGdi32.NewProc("CreateCompatibleDC")
	ballDeleteDC = ballGdi32.NewProc("DeleteDC")
	ballDIB      = ballGdi32.NewProc("CreateDIBSection")
	ballSelObj   = ballGdi32.NewProc("SelectObject")
	ballDelObj   = ballGdi32.NewProc("DeleteObject")
	ballULW      = ballUser32.NewProc("UpdateLayeredWindow")
	ballModHdr   = ballKern32.NewProc("GetModuleHandleW")
)

// ---------------- 运行期状态（仅在球线程访问的拖拽标志无需加锁） ----------------

var (
	ballMu        sync.Mutex
	ballRunning   bool
	ballApp       *App
	ballHwnd      uintptr
	ballDone      bool
	ballClassOnce sync.Once
	ballClassName *uint16
	ballClassAtom uint16

	ballSizePx     int // 球窗口边长（物理像素，含阴影留白）
	ballDiameterPx int // 圆球本体直径（物理像素）
	ballShadowPad  int // 四周为阴影预留的边距（物理像素）
	ballRadiusF    float64
	ballLogoImg    image.Image // 浮动球中心 logo 图像（nil 则回退闪电）
	ballCx, ballCy float64

	// 拖拽（仅球线程读写）
	ballCaptured bool
	ballDragging bool
	ballDownPt   ballPoint
	ballOrigLeft int32
	ballOrigTop  int32
)

func ballUtf16(s string) *uint16 {
	p, _ := syscall.UTF16PtrFromString(s)
	return p
}

// ballLogf 浮动球专用日志辅助（沿用 selector 的 selLogf 约定）：
// 关键生命周期/用户离散操作/异常统一走全局可配置日志（logger.go）。
func ballLogf(format string, args ...any) {
	appLogf(logInfo, "ball", format, args...)
}

// startNativeBall 启动桌面浮动球（幂等）。
func startNativeBall(a *App) error {
	if a == nil {
		return fmt.Errorf("app 为空")
	}
	ballMu.Lock()
	if ballRunning {
		ballMu.Unlock()
		return nil
	}
	ballApp = a
	ballRunning = true
	ballDone = false
	ballMu.Unlock()

	// 加载 logo 图片
	ballLoadLogo()

	go ballThread()
	return nil
}

// stopNativeBall 请求关闭浮动球（从任意 goroutine 调用）。
func stopNativeBall() {
	ballMu.Lock()
	h := ballHwnd
	ballMu.Unlock()
	if h != 0 {
		ballPostMsg.Call(h, ballWM_CLOSE, 0, 0)
	}
}

// ballLoadLogo 从配置加载 logo 图片到 ballLogoImg。
func ballLoadLogo() {
	ballLogoImg = nil
	cfg := CurrentConfig()
	var data []byte
	if cfg.General.AppLogo != "" {
		if b, err := os.ReadFile(cfg.General.AppLogo); err == nil {
			data = b
		}
	}
	if len(data) == 0 && ballApp != nil {
		data = ballApp.defaultLogo
	}
	if len(data) == 0 {
		return
	}
	img, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		ballLogf("logo PNG 解码失败: %v", err)
		return
	}
	ballLogoImg = img
}

// ballThread 在专用 OS 线程上建窗并泵消息，直到收到关闭。
func ballThread() {
	runtime.LockOSThread()
	// 故意不 UnlockOSThread：线程随 goroutine 退出而销毁，队列残留一并清除。
	defer func() {
		ballMu.Lock()
		ballRunning = false
		ballHwnd = 0
		ballMu.Unlock()
	}()

	ballClassOnce.Do(ballRegisterClass)
	if ballClassAtom == 0 {
		ballLogf("注册窗口类失败，放弃")
		return
	}

	// 计算尺寸与初始位置（物理像素；按系统 DPI 放大以在高分屏保持视觉大小）
	dpi := uintptr(96)
	if r, _, _ := ballGetDpi.Call(); r > 0 {
		dpi = r
	}
	scale := float64(dpi) / 96.0
	if scale < 1 {
		scale = 1
	}
	if ballLogoImg != nil {
		// 纯 logo 模式：以 logo 自然尺寸缩放到 48 DIP，无圆球背景无阴影
		bounds := ballLogoImg.Bounds()
		maxDim := 48.0
		var imgScale float64
		if float64(bounds.Dx()) > maxDim || float64(bounds.Dy()) > maxDim {
			imgScale = maxDim / math.Max(float64(bounds.Dx()), float64(bounds.Dy()))
		} else {
			imgScale = 1.0
		}
		ballDiameterPx = int(math.Max(float64(bounds.Dx())*imgScale, float64(bounds.Dy())*imgScale) * scale)
		ballShadowPad = 2 // 极小边距防裁切
		ballSizePx = ballDiameterPx + 2*ballShadowPad
	} else {
		// 回退圆球模式
		ballDiameterPx = int(float64(44) * scale)
		ballShadowPad = int(math.Max(6, float64(11)*scale))
		ballSizePx = ballDiameterPx + 2*ballShadowPad
	}
	ballCx = float64(ballSizePx-1) / 2
	ballCy = ballCx
	ballRadiusF = float64(ballDiameterPx)/2 - math.Max(1, float64(ballDiameterPx)/32)

	left, top := ballDefaultPos(ballSizePx)

	hInst, _, _ := ballModHdr.Call(0)
	hwnd, _, _ := ballCreate.Call(
		uintptr(ballWS_EX_LAYERED|ballWS_EX_TOPMOST|ballWS_EX_TOOLWIN),
		uintptr(unsafe.Pointer(ballClassName)),
		uintptr(unsafe.Pointer(ballUtf16("go-tools 浮动球"))),
		uintptr(ballWS_POPUP),
		uintptr(left), uintptr(top), uintptr(ballSizePx), uintptr(ballSizePx),
		0, 0, hInst, 0,
	)
	if hwnd == 0 {
		ballLogf("CreateWindowEx 失败: %v", windows.GetLastError())
		return
	}
	ballMu.Lock()
	ballHwnd = hwnd
	ballMu.Unlock()

	ballPaint(hwnd)
	ballShow.Call(hwnd, ballSW_SHOWNOACTIVATE)

	var msg ballMsg
	for {
		ret, _, _ := ballGetMsg.Call(uintptr(unsafe.Pointer(&msg)), 0, 0, 0)
		if int32(ret) == 0 {
			break // 意外 WM_QUIT：本线程即将销毁，安全退出
		}
		ballTrans.Call(uintptr(unsafe.Pointer(&msg)))
		ballDispatch.Call(uintptr(unsafe.Pointer(&msg)))
		ballMu.Lock()
		done := ballDone
		ballMu.Unlock()
		if done {
			break
		}
	}
}

// ballDefaultPos 工作区右下角（物理像素），避开任务栏。
func ballDefaultPos(size int) (int32, int32) {
	var rc ballRect
	ballSysInfo.Call(ballSPI_GETWORKAREA, 0, uintptr(unsafe.Pointer(&rc)), 0)
	margin := int32(16)
	right := rc.Right - int32(size) - margin
	bottom := rc.Bottom - int32(size) - margin
	if right < rc.Left {
		right = rc.Left
	}
	if bottom < rc.Top {
		bottom = rc.Top
	}
	return right, bottom
}

func ballRegisterClass() {
	ballClassName = ballUtf16("GoToolsFloatBallCls")
	hInst, _, _ := ballModHdr.Call(0)
	wc := ballWndClass{
		LpfnWndProc:   syscall.NewCallback(ballWndProc),
		HInstance:     hInst,
		HCursor:       ballLoadCurForArrow(),
		LpszClassName: ballClassName,
	}
	atom, _, err := ballNewClass.Call(uintptr(unsafe.Pointer(&wc)))
	if atom == 0 {
		ballLogf("RegisterClassW 失败: %v", err)
		return
	}
	ballClassAtom = uint16(atom)
}

func ballLoadCurForArrow() uintptr {
	r, _, _ := ballLoadCur.Call(0, uintptr(ballIDC_ARROW))
	return r
}

// ballWndProc 窗口过程（仅在球线程被调用）。
func ballWndProc(hwnd uintptr, msg uint32, wParam, lParam uintptr) uintptr {
	switch msg {
	case ballWM_LBUTTONDOWN:
		x := int32(int16(uint16(lParam & 0xFFFF)))
		y := int32(int16(uint16((lParam >> 16) & 0xFFFF)))
		if !ballInsideCircle(float64(x), float64(y)) {
			break // 圆外透明区不响应
		}
		ballCaptured = true
		ballDragging = false
		ballGetCur.Call(uintptr(unsafe.Pointer(&ballDownPt)))
		var rc ballRect
		ballGetRect.Call(hwnd, uintptr(unsafe.Pointer(&rc)))
		ballOrigLeft, ballOrigTop = rc.Left, rc.Top
		ballSetCap.Call(hwnd)
		return 0

	case ballWM_MOUSEMOVE:
		if !ballCaptured {
			break
		}
		var p ballPoint
		ballGetCur.Call(uintptr(unsafe.Pointer(&p)))
		dx := p.X - ballDownPt.X
		dy := p.Y - ballDownPt.Y
		if !ballDragging && (abs32(dx) > 4 || abs32(dy) > 4) {
			ballDragging = true
		}
		if ballDragging {
			nl, nt := ballOrigLeft+dx, ballOrigTop+dy
			nl, nt = ballClampToWork(nl, nt, int32(ballSizePx))
			ballSetPos.Call(hwnd, uintptr(nl), uintptr(nt), 0, 0,
				ballSWP_NOSIZE|ballSWP_NOZORDER|ballSWP_NOACTIVATE)
		}
		return 0

	case ballWM_LBUTTONUP:
		if ballCaptured {
			ballRelCap.Call()
			ballCaptured = false
			if !ballDragging {
				ballShowMenu(hwnd)
			}
		}
		return 0

	case ballWM_RBUTTONUP:
		ballShowMenu(hwnd)
		return 0

	case ballWM_CLOSE:
		ballDestroy.Call(hwnd)
		return 0

	case ballWM_DESTROY:
		ballMu.Lock()
		ballDone = true
		ballMu.Unlock()
		return 0
	}
	r, _, _ := ballDefProc.Call(hwnd, uintptr(msg), wParam, lParam)
	return r
}

func abs32(v int32) int32 {
	if v < 0 {
		return -v
	}
	return v
}

func ballInsideCircle(x, y float64) bool {
	if ballLogoImg != nil {
		// logo 模式：整个窗口都可点击
		return true
	}
	dx, dy := x-ballCx, y-ballCy
	return dx*dx+dy*dy <= ballRadiusF*ballRadiusF
}

// ballClampToWork 把窗口左上角钳制到工作区内。
func ballClampToWork(l, t int32, size int32) (int32, int32) {
	var rc ballRect
	ballSysInfo.Call(ballSPI_GETWORKAREA, 0, uintptr(unsafe.Pointer(&rc)), 0)
	if l < rc.Left {
		l = rc.Left
	}
	if t < rc.Top {
		t = rc.Top
	}
	if l+size > rc.Right {
		l = rc.Right - size
	}
	if t+size > rc.Bottom {
		t = rc.Bottom - size
	}
	return l, t
}

// ballQuickItems 从当前配置读取标记为快捷（Quick=true）的工具，作为原生菜单项（ID → 显示名）。
func ballQuickItems() []struct{ ID, Label string } {
	var out []struct{ ID, Label string }
	for _, t := range CurrentConfig().Tools {
		if !t.Quick {
			continue
		}
		id := t.ID
		if id == "" {
			id = t.Key
		}
		label := t.Name
		if label == "" {
			label = t.Key
		}
		out = append(out, struct{ ID, Label string }{id, label})
	}
	return out
}

// ballShowMenu 在光标处弹出原生快捷菜单（菜单项动态来自配置的 quick 工具）。
func ballShowMenu(hwnd uintptr) {
	items := ballQuickItems()
	menu, _, _ := ballCreatePM.Call()
	if menu == 0 {
		return
	}
	for i, it := range items {
		ballAppendM.Call(menu, ballMF_STRING, uintptr(i+1), uintptr(unsafe.Pointer(ballUtf16(it.Label))))
	}
	ballAppendM.Call(menu, ballMF_SEPARATOR, 0, 0)
	closeCmd := uintptr(len(items) + 1)
	ballAppendM.Call(menu, ballMF_STRING, closeCmd, uintptr(unsafe.Pointer(ballUtf16("关闭浮动球"))))

	var p ballPoint
	ballGetCur.Call(uintptr(unsafe.Pointer(&p)))
	ballSetFore.Call(hwnd) // 让菜单在点击别处时正确消失
	cmd, _, _ := ballTrackPM.Call(menu, ballTPM_RIGHTBUTTON|ballTPM_RETURNCMD,
		uintptr(p.X), uintptr(p.Y), hwnd, 0)
	ballDestroyM.Call(menu)
	ballPostMsg.Call(hwnd, 0, 0, 0) // WM_NULL：解除菜单与前台窗口的已知死锁

	if cmd == 0 {
		return
	}
	app := ballApp
	if app == nil {
		return
	}
	if cmd == closeCmd {
		go app.handleBallAction("close")
		return
	}
	if int(cmd-1) < len(items) {
		id := items[cmd-1].ID
		go app.handleBallAction(id)
	}
}

// ============================================================
// 位图渲染：Go 生成逐像素 alpha 的圆形球（预乘 BGRA，自上而下）
// ============================================================

// ballRender 返回 size*size 的预乘 BGRA 缓冲（top-down）
func ballRender(size int) []byte {
	buf := make([]byte, size*size*4)

	if ballLogoImg != nil {
		// 纯 logo 模式：将图像缩放到 size x size，无背景无阴影
		bounds := ballLogoImg.Bounds()
		imgW, imgH := bounds.Dx(), bounds.Dy()
		for y := 0; y < size; y++ {
			for x := 0; x < size; x++ {
				// 映射到图像坐标（居中显示）
				ix := int(float64(x) / float64(size) * float64(imgW))
				iy := int(float64(y) / float64(size) * float64(imgH))
				if ix >= imgW {
					ix = imgW - 1
				}
				if iy >= imgH {
					iy = imgH - 1
				}
				cr, cg, cb, ca := ballLogoImg.At(bounds.Min.X+ix, bounds.Min.Y+iy).RGBA()
				a := float64(ca) / 65535.0
				o := (y*size + x) * 4
				// 预乘 BGRA
				buf[o] = byte(float64(cr) / 257.0 * a)
				buf[o+1] = byte(float64(cg) / 257.0 * a)
				buf[o+2] = byte(float64(cb) / 257.0 * a)
				buf[o+3] = byte(ca >> 8)
			}
		}
		return buf
	}

	// 回退圆球模式
	c := float64(size-1) / 2
	rad := float64(ballDiameterPx)/2 - math.Max(1, float64(ballDiameterPx)/32)
	shadowOffY := math.Max(1.5, float64(ballShadowPad)*0.16)
	blur := math.Max(3, float64(ballShadowPad)*0.72)
	const shadowMax = 0.40

	bolt := [][2]float64{
		{0.10, -0.48}, {-0.28, 0.06}, {-0.02, 0.06},
		{-0.12, 0.48}, {0.28, -0.10}, {0.02, -0.10},
	}
	s := rad * 1.5
	boltPoly := make([][2]float64, len(bolt))
	for i, pt := range bolt {
		boltPoly[i] = [2]float64{c + pt[0]*s, c + pt[1]*s}
	}

	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			dx := float64(x) - c
			dyBall := float64(y) - c
			distBall := math.Sqrt(dx*dx + dyBall*dyBall)
			var cov float64
			switch {
			case distBall <= rad-1:
				cov = 1
			case distBall >= rad+1:
				cov = 0
			default:
				cov = rad + 1 - distBall
			}
			dySh := float64(y) - (c + shadowOffY)
			distSh := math.Sqrt(dx*dx + dySh*dySh)
			var shA float64
			if distSh <= rad {
				shA = shadowMax
			} else if distSh < rad+blur {
				f := (rad + blur - distSh) / blur
				shA = shadowMax * f * f
			}
			if cov <= 0 && shA <= 0.004 {
				continue
			}
			var r, g, b float64
			if cov > 0 {
				t := (float64(y) - (c - rad)) / (2 * rad)
				if t < 0 {
					t = 0
				} else if t > 1 {
					t = 1
				}
				r = lerp(62, 150, t)
				g = lerp(110, 60, t)
				b = lerp(255, 220, t)
				hl := math.Max(0, 1-distBall/(rad*1.4)) * 0.35
				r = clamp255(r + hl*120)
				g = clamp255(g + hl*120)
				b = clamp255(b + hl*120)
				if pointInPoly(float64(x), float64(y), boltPoly) {
					r, g, b = 250, 250, 255
				}
			}
			aTotal := cov + (1-cov)*shA
			o := (y*size + x) * 4
			buf[o] = byte(b * cov)
			buf[o+1] = byte(g * cov)
			buf[o+2] = byte(r * cov)
			buf[o+3] = byte(aTotal * 255)
		}
	}
	return buf
}

func lerp(a, b, t float64) float64 { return a + (b-a)*t }
func clamp255(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 255 {
		return 255
	}
	return v
}

func pointInPoly(px, py float64, poly [][2]float64) bool {
	inside := false
	n := len(poly)
	for i, j := 0, n-1; i < n; j, i = i, i+1 {
		xi, yi := poly[i][0], poly[i][1]
		xj, yj := poly[j][0], poly[j][1]
		if ((yi > py) != (yj > py)) &&
			(px < (xj-xi)*(py-yi)/(yj-yi)+xi) {
			inside = !inside
		}
	}
	return inside
}

// ballPaint 用 UpdateLayeredWindow 提交逐像素 alpha 的球位图。
func ballPaint(hwnd uintptr) {
	size := ballSizePx
	premul := ballRender(size)

	screenDC, _, _ := ballGetDC.Call(0)
	if screenDC == 0 {
		return
	}
	defer ballRelDC.Call(0, screenDC)

	memDC, _, _ := ballCompatDC.Call(screenDC)
	if memDC == 0 {
		return
	}
	defer ballDeleteDC.Call(memDC)

	var bmi ballBitmapInfo
	bmi.Header.BiSize = uint32(unsafe.Sizeof(bmi.Header))
	bmi.Header.BiWidth = int32(size)
	bmi.Header.BiHeight = -int32(size) // top-down
	bmi.Header.BiPlanes = 1
	bmi.Header.BiBitCount = 32
	bmi.Header.BiCompression = 0 // BI_RGB

	var bits *byte
	// CreateDIBSection(hdc, pbmi, iUsage, ppvBits, hSection, offset)
	// iUsage=DIB_RGB_COLORS(0)；ppvBits=&bits（第 3、4 参不可写反，否则返回 ERROR_INVALID_PARAMETER）
	bmp, _, _ := ballDIB.Call(memDC, uintptr(unsafe.Pointer(&bmi)),
		0, uintptr(unsafe.Pointer(&bits)), 0, 0)
	if bmp == 0 || bits == nil {
		return
	}
	defer ballDelObj.Call(bmp)

	n := size * size * 4
	copy(unsafe.Slice(bits, n), premul)

	old, _, _ := ballSelObj.Call(memDC, bmp)

	var ptDst ballPoint
	l, t := ballCurPos(hwnd)
	ptDst.X, ptDst.Y = l, t
	sz := ballSize{CX: int32(size), CY: int32(size)}
	ptSrc := ballPoint{0, 0}
	blend := ballBlend{BlendOp: 0, BlendFlags: 0, SourceConstantAlpha: 255, AlphaFormat: 1} // AC_SRC_ALPHA

	ballULW.Call(hwnd, screenDC, uintptr(unsafe.Pointer(&ptDst)), uintptr(unsafe.Pointer(&sz)),
		memDC, uintptr(unsafe.Pointer(&ptSrc)), 0, uintptr(unsafe.Pointer(&blend)), ballULW_ALPHA)

	if old != 0 {
		ballSelObj.Call(memDC, old)
	}
}

// ballCurPos 读取当前窗口左上角（物理像素）。
func ballCurPos(hwnd uintptr) (int32, int32) {
	var rc ballRect
	ballGetRect.Call(hwnd, uintptr(unsafe.Pointer(&rc)))
	return rc.Left, rc.Top
}

// 确保 image 包被引用（ballRender 以字节缓冲实现，未直接使用 image 类型时保留占位）
var _ = image.Point{}
