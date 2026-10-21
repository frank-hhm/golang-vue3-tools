<template>
    <div class="blank-layout">
        <!-- 自绘标题栏：无边框窗口下由前端提供标题、视图菜单与窗口按钮 -->
        <div class="win-titlebar">
            <div class="tb-title">{{ title }}</div>
            <div class="tb-actions">
                <a-dropdown trigger="click" position="br" @select="onMenuSelect">
                    <div class="tb-menu">视图<icon-down class="tb-caret" /></div>
                    <template #content>
                        <a-doption value="reload">
                            <template #icon><icon-refresh /></template>重新加载
                        </a-doption>
                        <a-doption value="fullscreen">
                            <template #icon><icon-fullscreen /></template>切换全屏
                        </a-doption>
                    </template>
                </a-dropdown>
                <div class="tb-btn" title="最小化" @click="minimise"><icon-minus /></div>
                <div class="tb-btn" title="最大化 / 还原" @click="toggleMaximise"><icon-expand /></div>
                <div class="tb-btn tb-close" title="关闭" @click="close"><icon-close /></div>
            </div>
        </div>
        <div class="win-content">
            <router-view />
        </div>
    </div>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { useRoute } from "vue-router";
import { Window } from "@wailsio/runtime";

/**
 * 空白布局：用于 Wails 原生新窗口（/win 前缀）。
 * 采用无边框窗口，标题栏自绘：左侧标题，右侧“视图”菜单与最小化/最大化/关闭按钮。
 * 全屏时该标题栏随内容铺满显示器，菜单依旧可见。
 */
const route = useRoute();
const title = computed(() => (route.meta?.menu_name as string) || "go-tools");

const reload = () => location.reload();
// 用最大化实现“铺满显示器”，避免排他全屏把自绘标题栏顶出屏幕
const toggleFullscreen = () => Window.ToggleMaximise();
const minimise = () => Window.Minimise();
const toggleMaximise = () => Window.ToggleMaximise();
const close = () => Window.Close();

const onMenuSelect = (v: string | number | Record<string, any> | undefined) => {
    switch (v) {
        case "reload":
            reload();
            break;
        case "fullscreen":
            toggleFullscreen();
            break;
    }
};
</script>

<style scoped>
.blank-layout {
    position: fixed;
    left: 0;
    top: 0;
    width: 100%;
    height: 100%;
    background: var(--color-bg-body);
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
    background: #1d2129;
    color: #fff;
    user-select: none;
    -webkit-app-region: drag;
    app-region: drag;
}

.tb-title {
    padding-left: 14px;
    font-size: 13px;
    font-weight: bold;
}

.tb-actions {
    display: flex;
    align-items: center;
    -webkit-app-region: no-drag;
    app-region: no-drag;
}

.tb-menu {
    display: flex;
    align-items: center;
    gap: 4px;
    height: 40px;
    padding: 0 14px;
    cursor: pointer;
    font-size: 13px;
}
.tb-menu:hover {
    background: rgba(255, 255, 255, 0.12);
}

.tb-caret {
    font-size: 12px;
}

.tb-btn {
    width: 44px;
    height: 40px;
    display: flex;
    align-items: center;
    justify-content: center;
    cursor: pointer;
    font-size: 14px;
}
.tb-btn:hover {
    background: rgba(255, 255, 255, 0.12);
}
.tb-close:hover {
    background: #f53f3f;
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
</style>
