import { getCurrentInstance, ref } from "vue";
import { Message } from "@arco-design/web-vue";

/**
 * 截图工具事件（单个 event 的封装）
 *
 * 调用后端区域选区截图：确认后展示预览弹窗，图片已同步写入系统剪贴板；
 * 用户取消（ESC / 右键 / 无有效选区）时后端返回空字符串，静默处理。
 */
export function useScreenshotEvent() {
  const {
    proxy: { $GoTools },
  } = getCurrentInstance() as any;

  /** 预览弹窗显隐 */
  const visible = ref(false);
  /** 截图结果（data URL） */
  const image = ref<string>("");

  /** 触发区域选区截图 */
  const capture = async () => {
    try {
      const data = await $GoTools.CaptureSelectedRegion("png");
      if (!data) {
        return;
      }
      image.value = data;
      visible.value = true;
      Message.success("截图完成，已复制到剪贴板");
    } catch (err) {
      console.error("截图失败:", err);
      Message.error("截图失败: " + err);
    }
  };

  /** 将截图（data URL）下载为本地 PNG 文件 */
  const download = () => {
    if (!image.value) return;
    const a = document.createElement("a");
    a.href = image.value;
    a.download = `screenshot-${Date.now()}.png`;
    a.click();
  };

  return { visible, image, capture, download };
}
