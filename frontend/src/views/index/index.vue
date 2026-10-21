<template>
  <layout-body>
    <div class="tools-list">
      <div v-for="(item, index) in toolsList" :key="index" class="tools-item" @click="onToolClick(item)">
        {{ item.name }}
      </div>
    </div>

    <a-modal
      v-model:visible="toolEvents.screenshot.visible"
      title="截图预览"
      :width="800"
      ok-text="下载"
      cancel-text="关闭"
      @ok="toolEvents.screenshot.download"
    >
      <div class="preview-wrap">
        <img v-if="toolEvents.screenshot.image" :src="toolEvents.screenshot.image" class="preview-img" alt="screenshot" />
        <div class="preview-tip">截图已自动复制到剪贴板，可直接 Ctrl+V 粘贴；点击「下载」保存为 PNG 文件。</div>
      </div>
    </a-modal>
  </layout-body>
</template>
<script setup lang="ts">
import router from "@/router";
import { reactive, ref } from "vue";
import { useToolEvents, type ToolItem } from "./events";

/** 工具事件处理器集合（reactive 包裹后，模板可深层访问内部 ref） */
const toolEvents = reactive(useToolEvents());

const toolsList = ref<ToolItem[]>([
  {
    name: '截图',
    event: 'screenshot',
  },
  {
    name: '时间戳转换',
    path: '/tools/timestamp',
  },
  {
    name: '随机字符串生成',
    path: '/tools/strrandom',
  },
  {
    name: '文字加密解密',
    path: '/tools/encdec',
  },
  {
    name: '进制转换',
    path: '/tools/hexconvert',
  },
  {
    name: '抓包工具',
    path: '/tools/sniffer',
  },
  {
    name: 'USB调试',
    path: '/tools/usbserial',
  },
  
]);

/** 统一处理工具卡片点击：有 path 则路由跳转，有 event 则派发对应事件 */
const onToolClick = (item: ToolItem) => {
  if (item.path) {
    router.push(item.path);
    return;
  }
  if (item.event) {
    toolEvents.runEvent(item.event);
  }
};
</script>
<style scoped>

.tools-list {
  padding: 20px;
  display: flex;
  items-align: center;
  flex-wrap: wrap;
}

.tools-item {
  width: 200px;
  padding: 30px 10px;
  margin: 10px;
  text-align: center;
  background-color: #f5f5f5;
  cursor: pointer;
  border-radius: var(--base-radius-default);
  user-select: none;
  font-size: var(--base-size-large);
  font-weight: bold;
}
.tools-item:hover {
  background-color: #e5e5e5;
}

.preview-wrap {
  display: flex;
  flex-direction: column;
  gap: 12px;
}
.preview-img {
  max-width: 100%;
  border: 1px solid #e5e6eb;
  border-radius: var(--base-radius-default);
}
.preview-tip {
  color: #86909c;
  font-size: 13px;
}
</style>
