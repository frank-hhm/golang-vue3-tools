<template>
    <div class="basic-layout">
        <div class="layout-content">
            <router-view />
        </div>
    </div>
</template>
<script setup lang="ts">
/**
 * 主窗口基础布局。
 * 顶部 header（面包屑/设置/返回/刷新）已移除：
 * - 设置入口移到主页“快捷操作”卡片右上角
 * - 刷新入口移到主页“欢迎回来”卡片右上角
 */
</script>
<style scoped>
.basic-layout {
    min-height: 100vh;
    display: flex;
    flex-direction: column;
}

.layout-content {
    left: 0;
    top: 0;
    position: fixed;
    width: calc(100% - (2*var(--base-padding)));
    height: calc(100% - (2*var(--base-padding)));
    padding: var(--base-padding);
    background: var(--color-bg-body);
    overflow-y: scroll;
}
</style>
