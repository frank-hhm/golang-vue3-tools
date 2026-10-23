<template>
    <div class="blank-layout">
        <!-- 自绘标题栏：无边框窗口下由前端提供标题、视图菜单与窗口按钮 -->
        <div class="win-titlebar">
            <div class="tb-title">{{ title }}</div>
            <div class="tb-actions">
                <a-space>
                    <a-switch size="small" v-model:checked="themeMode" @change="toggleTheme" />
                    <a-button size="small" @click="reload"><icon-refresh /></a-button>
                    <a-button size="small" v-if="!isSettingsPage" @click="toggleSettings"><icon-settings /></a-button>
                    <a-button size="small" v-if="!noFullscreen" @click="toggleFullscreen"><icon-fullscreen /></a-button>
                    <a-button size="small" @click="minimise"><icon-minus /></a-button>
                    <a-button size="small" @click="close"><icon-close /></a-button>
                </a-space>
            </div>
        </div>
        <div class="win-content" :class="{ 'no-pad': noPadding }">
            <router-view />
        </div>
    </div>
</template>

<script setup lang="ts">
import { computed, ref } from "vue";
import { useRoute } from "vue-router";
import { Window } from "@wailsio/runtime";
import { applyTheme, currentTheme } from "@/utils/theme";
import { useToolWindow } from "@/views/index/useToolWindow";

/**
 * 空白布局：用于 Wails 原生新窗口（/win 前缀）。
 * 采用无边框窗口，标题栏自绘：左侧标题，右侧“视图”菜单与最小化/最大化/关闭按钮。
 * 全屏时该标题栏随内容铺满显示器，菜单依旧可见。
 */
const route = useRoute();
const { openToolWindow } = useToolWindow();
const title = computed(() => (route.meta?.menu_name as string) || "go-tools");

// 当前页即设置页时，视图菜单不再重复展示“设置”入口
const isSettingsPage = computed(() => route.path.includes("/tools/settings"));
// 工具配置“禁用全屏”（URL query fs=0）：隐藏最大化按钮与“切换全屏”菜单
const noFullscreen = computed(() => String(route.query.fs ?? "") === "0");
// 内容区内边距：默认给；取色器等需要内容铺满边缘的工具页去掉 padding
const noPadding = computed(() => route.path.includes("/tools/colorpicker"));

// 当前主题模式：用于菜单项图标与文案（暗黑 <-> 浅色快速切换，“跟随系统”按当前实际深浅切换）
const themeMode = ref(currentTheme());

const reload = () => location.reload();
// 用最大化实现“铺满显示器”，避免排他全屏把自绘标题栏顶出屏幕
const toggleFullscreen = () => Window.ToggleMaximise();
const minimise = () => Window.Minimise();
const toggleMaximise = () => Window.ToggleMaximise();
const close = () => Window.Close();

const toggleTheme = () => {
    // auto 模式下按系统实际偏好落地为固定主题，再在浅色/暗黑间切换
    const dark = themeMode.value === "dark" || (themeMode.value === "auto" && window.matchMedia("(prefers-color-scheme: dark)").matches);
    const next = dark ? "light" : "dark";
    applyTheme(next);
    themeMode.value = next;
};

const onMenuSelect = (v: string | number | Record<string, any> | undefined) => {
    switch (v) {
        case "reload":
            reload();
            break;
        case "fullscreen":
            toggleFullscreen();
            break;
        case "theme":
            toggleTheme();
            break;
        case "settings":
            toggleSettings();
            break;
    }
};

// 打开设置：走单例工具窗口，不在当前窗口跳转；设置窗口已存在时由后端直接复用（显示+取消最小化+聚焦）
const toggleSettings = () => openToolWindow("/tools/settings", "设置", { single: true });

</script>

<style scoped>
.blank-layout {
    position: fixed;
    left: 0;
    top: 0;
    width: 100%;
    height: 100%;
    background: var(--color-bg-1);
}

.win-titlebar {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    height: 40px;
    z-index: 1000;
    display: flex;
    align-items: center;
    justify-content: space-between;
    background: var(--color-bg-1);
    color: var(--color-text-1);
    user-select: none;
    -webkit-app-region: drag;
    app-region: drag;
    border-bottom: 1px solid var(--color-border-1);
}

.tb-title {
    padding-left: 14px;
    font-size: var(--base-size-medium);
    font-weight: bold;
}

.tb-actions {
    display: flex;
    align-items: center;
    -webkit-app-region: no-drag;
    app-region: no-drag;
    margin-right: var(--base-padding);
}

.tb-menu {
    display: flex;
    align-items: center;
    gap: var(--base-space-xs);
    height: 40px;
    padding: 0 14px;
    cursor: pointer;
    font-size: var(--base-size-medium);
}

.tb-menu:hover {
    background: rgba(255, 255, 255, 0.12);
}

.tb-caret {
    font-size: var(--base-size-small);
}

.tb-btn {
    width: 44px;
    height: 40px;
    display: flex;
    align-items: center;
    justify-content: center;
    cursor: pointer;
    font-size: var(--base-size-default);
}

.tb-btn:hover {
    background: rgba(var(--primary-6), 0.12);
}

.tb-close:hover {
    background: rgb(var(--red-6));
    color: #fff;
}

.win-content {
    position: absolute;
    top: 40px;
    left: 0;
    right: 0;
    bottom: 0;
    padding: var(--base-padding);
    overflow-y: auto;
}

/* 取色器等铺满型工具页：去掉内容区留白 */
.win-content.no-pad {
    padding: 0;
}
</style>
