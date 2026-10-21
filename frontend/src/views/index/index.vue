<template>
  <layout-body>
    <div class="tools-list">
      <div v-for="(item, index) in toolsList" :key="index" class="tools-item" @click="onToolClick(item)">
        {{ item.name }}
      </div>
    </div>
  </layout-body>
</template>
<script setup lang="ts">
import { reactive, ref } from "vue";
import { useToolEvents, type ToolItem } from "./events";
import { useToolWindow } from "./useToolWindow";

/** 工具事件处理器集合（reactive 包裹后，模板可深层访问内部 ref） */
const toolEvents = reactive(useToolEvents());

/** 统一的工具窗口打开入口（新窗口优先，失败回退 web 路由） */
const { openToolWindow } = useToolWindow();

const toolsList = ref<ToolItem[]>([
  {
    name: '截图',
    event: 'screenshot',
  },
  {
    name: '颜色取色器',
    path: '/tools/colorpicker',
    single: true,
    width: 480,
    height: 560,
  },
  {
    name: '视频录制',
    path: '/tools/video',
    single: true,
    fullscreen: true,
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
    name: '图片base64',
    path: '/tools/imageBase64',
  },
  {
    name: '二维码生成',
    path: '/tools/qrcode',
  },

  {
    name: '抓包工具',
    path: '/tools/sniffer',
    single: true,
    fullscreen: true,
  },
  {
    name: 'WebSocket调试',
    path: '/tools/websocket',
    single: false,
    fullscreen: true,
  },
  {
    name: 'USB调试',
    path: '/tools/usbserial',
  },
  {
    name: 'Redis管理',
    path: '/tools/redis',
    single: true,
    fullscreen: true,
  },
  {
    name: 'MySQL管理',
    path: '/tools/mysql',
    single: true,
    fullscreen: true,
  },
  {
    name: 'test',
    path: '/demo/gomessage',
  }
]);

/** 统一处理工具卡片点击：path 用原生新窗口打开（按 single/fullscreen/width/height 配置）；event 直接派发 */
const onToolClick = (item: ToolItem) => {
  if (item.path) {
    openToolWindow(item.path, item.name, {
      single: item.single,
      fullscreen: item.fullscreen,
      width: item.width,
      height: item.height,
    });
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
  padding: 20px 10px;
  margin: 10px;
  text-align: center;
  background-color: #f5f5f5;
  cursor: pointer;
  border-radius: var(--base-radius-default);
  user-select: none;
  font-size: var(--base-size-default);
  font-weight: bold;
}

.tools-item:hover {
  background-color: #e5e5e5;
}
</style>
