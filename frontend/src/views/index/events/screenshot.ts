import { getCurrentInstance } from "vue";
import { Message } from "@arco-design/web-vue";
import { useToolWindow } from "../useToolWindow";

/** 与截图预览页约定的共享存储键（同源多窗口共享 localStorage） */
const STORAGE_KEY = "tool:screenshot";

/**
 * 截图工具事件（单个 event 的封装）
 *
 * 调用后端区域选区截图：确认后把结果写入共享 localStorage，
 * 再用 Wails 原生新窗口打开 /tools/screenshot 展示预览（回调在新窗口处理）。
 * 用户取消（ESC / 右键 / 无有效选区）时后端返回空字符串，静默处理。
 */
export function useScreenshotEvent() {
  const {
    proxy: { $GoTools },
  } = getCurrentInstance() as any;
  const { openToolWindow } = useToolWindow();

  /** 触发区域选区截图，并在新窗口中展示结果 */
  const capture = async () => {
    try {
      const data = await $GoTools.CaptureSelectedRegion("png");
      if (!data) {
        return;
      }
      try {
        localStorage.setItem(STORAGE_KEY, data);
      } catch {
        /* localStorage 不可用时忽略，仅影响新窗口预览 */
      }
      // 截图预览窗口单例：重复触发时复用并聚焦已有窗口
      await openToolWindow("/tools/screenshot", "截图预览", { single: true });
      Message.success("截图完成，已复制到剪贴板");
    } catch (err) {
      console.error("截图失败:", err);
      Message.error("截图失败: " + err);
    }
  };

  return { capture };
}
