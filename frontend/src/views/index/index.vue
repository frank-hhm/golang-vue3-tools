<template>
  <div class="main-body">
    <div class="tools-list">
      <div v-for="(item, index) in toolsList" class="tools-item" @click="toPath(item)">
        {{ item.name }}
      </div>
    </div>
  </div>
</template>
<script setup lang="ts">
import router from "@/router";
import { getCurrentInstance, ref, onMounted, watch } from "vue";

const {
  proxy,
  proxy: { $utils },
} = getCurrentInstance() as any;

const toolsList = ref<any[]>([
  {
    name: '抓包工具',
    path: '/tools/sniffer',
  },
  {
    name: 'USB调试',
    path: '/tools/usbserial',
  },
]);


const toPath = (item: any) => {
  router.push(item.path);
};
</script>
<style scoped>
.main-body {
  min-height: 100vh;
  min-width: 100vw;
  background-color: #fff;
}

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
</style>
