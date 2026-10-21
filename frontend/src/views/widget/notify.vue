<template>
  <div class="notify-wrap">
    <div class="notify-card" :class="level">
      <!-- 头部：应用图标 + 名称 + ⋯ +  -->
      <div class="nc-head">
        <span class="nc-app-ic">G</span>
        <span class="nc-app">go-tools</span>
        <span class="nc-more btn-no-drag" title="更多">⋯</span>
        <span class="nc-close btn-no-drag" title="关闭" @click="close">✕</span>
      </div>
      <!-- 正文：标题 + 内容 -->
      <div class="nc-title">{{ title }}</div>
      <div class="nc-content">{{ content }}</div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, getCurrentInstance } from "vue";
import { useRoute } from "vue-router";

const route = useRoute();
const GoTools = (getCurrentInstance() as any)?.appContext.config.globalProperties.$GoTools;

const title = computed(() => String(route.query.title ?? "通知"));
const content = computed(() => String(route.query.content ?? ""));
// 预留等级样式（info 默认，success/warning/error 可后续扩展）
const level = computed(() => String(route.query.level ?? "info"));

function close() {
  try {
    GoTools?.NotifyClose();
  } catch {
    window.close();
  }
}
</script>

<style scoped>
.notify-wrap {
  width: 100vw;
  height: 100vh;
  box-sizing: border-box;
  padding: 6px;
  background: transparent;
  /* 背景可拖动通知卡 */
  app-region: drag;
  -webkit-app-region: drag;
  user-select: none;
  overflow: hidden;
}
.notify-card {
  width: 100%;
  height: 100%;
  box-sizing: border-box;
  background: #fff;
  border: 1px solid #e5e6eb;
  border-radius: 10px;
  box-shadow: 0 6px 24px rgba(0, 0, 0, 0.14);
  padding: 12px 14px;
  display: flex;
  flex-direction: column;
}
.nc-head {
  display: flex;
  align-items: center;
  gap: 7px;
  margin-bottom: 8px;
}
.nc-app-ic {
  width: 18px;
  height: 18px;
  border-radius: 5px;
  background: linear-gradient(135deg, #165dff, #4080ff);
  color: #fff;
  font-size: 11px;
  font-weight: 700;
  display: flex;
  align-items: center;
  justify-content: center;
  flex: none;
}
.nc-app {
  font-size: 12px;
  color: #4e5969;
  flex: 1;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.nc-more,
.nc-close {
  flex: none;
  padding: 2px 6px;
  border-radius: 4px;
  font-size: 13px;
  color: #4e5969;
  cursor: pointer;
}
.nc-more:hover,
.nc-close:hover {
  background: #f2f3f5;
  color: #1d2129;
}
.nc-title {
  font-size: 14px;
  font-weight: 600;
  color: #1d2129;
  margin-bottom: 4px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.nc-content {
  font-size: 13px;
  color: #4e5969;
  line-height: 1.5;
  flex: 1;
  display: -webkit-box;
  -webkit-line-clamp: 3;
  -webkit-box-orient: vertical;
  overflow: hidden;
}
.btn-no-drag {
  app-region: no-drag;
  -webkit-app-region: no-drag;
}
</style>
