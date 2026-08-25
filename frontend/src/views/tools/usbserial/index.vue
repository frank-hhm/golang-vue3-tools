<template>
  <div class="main-body">
    <a-button @click="_onConnectSerial">连接串口</a-button>
    <a-button @click="_onStartReading">监听</a-button>
    <a-button @click="_onStopReading">取消监听</a-button>
    <a-button @click="_onStartReading">断开</a-button>
  </div>
</template>
<script setup lang="ts">
import { getCurrentInstance, ref, onMounted, watch } from "vue";

const {
  proxy,
  proxy: { $utils, $wailsjs },
} = getCurrentInstance() as any;

const ports = ref<any[]>([]);

const _onConnectSerial = () => {
  $wailsjs.GetSerialPorts().then(res => {
    console.log(res)
    $wailsjs.ConnectSerial(res[0].port_name).then((res) => {
      console.log('连接成功', res)
    })

  })
}
const _onStartReading = () => {

  $wailsjs.StartReading().then((res) => {
    console.log('监听成功', res)
  })
}

const _onStopReading = () => {
  $wailsjs.StopReading().then((res) => {
    console.log('取消成功', res)
  })
}


onMounted(() => {
  console.log("res:", 11);
  try {

  } catch (err) {
    console.error("获取串口列表失败:", err);
  }
});
</script>
<style scoped>
.main-body {
  min-height: 100vh;
  min-width: 100vw;
  background-color: #fff;
}
</style>
