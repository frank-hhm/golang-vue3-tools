import { getCurrentInstance } from "vue";
import router from "@/router";

/**
 * 统一的“工具窗口”打开入口。
 *
 * - 优先调用 Wails 后端 OpenToolWindow，在原生新窗口打开：走 /win 前缀的空白布局路由
 *   （无面包屑、LayoutBody 裸渲染）。
 * - 若调用失败（非 Wails 环境 / 方法不可用 / 后端报错），回退到 web 路由 router.push，
 *   该路由保留原本的 LayoutBody 与面包屑。
 */
export function useToolWindow() {
    const {
        proxy: { $GoTools },
    } = getCurrentInstance() as any;

    /**
     * @param path 常规路由路径，如 "/tools/timestamp"
     * @param title 新窗口标题
     * @param single 是否单例窗口（true：已存在则复用/聚焦，不重复开）
     * @param fullscreen 新窗口是否全屏启动（true：全屏；否则沿用常规尺寸）
     */
    const openToolWindow = async (path: string, title: string, single = false, fullscreen = false) => {
        try {
            if (typeof $GoTools?.OpenToolWindow !== "function") {
                throw new Error("OpenToolWindow 不可用");
            }
            // 新窗口统一用 /win 前缀的空白布局路由（菜单由原生窗口菜单提供）
            await $GoTools.OpenToolWindow("/win" + path, title, !!single, !!fullscreen);
        } catch (e) {
            console.warn("新窗口打开失败，回退到当前窗口路由:", e);
            await router.push(path);
        }
    };

    return { openToolWindow };
}
