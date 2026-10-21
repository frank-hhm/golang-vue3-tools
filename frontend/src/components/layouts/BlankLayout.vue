<template>
    <div class="blank-layout">
        <router-view />
    </div>
</template>

<script setup lang="ts">
/**
 * 空白布局：用于 Wails 原生新窗口打开的工具路由（/win 前缀）。
 * 不含面包屑与返回栏，内容占满窗口。
 */
</script>

<style scoped>
.blank-layout {
    position: fixed;
    left: 0;
    top: 0;
    width: 100%;
    height: 100%;
    padding: var(--base-padding);
    background: var(--color-bg-body);
    overflow-y: auto;
}
</style>
