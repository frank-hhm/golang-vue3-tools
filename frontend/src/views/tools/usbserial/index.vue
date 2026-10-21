<template>
  <layout-body>
    <a-button @click="_onConnectSerial">连接串口</a-button>
    <a-button @click="_onStartReading">监听</a-button>
    <a-button @click="_onStopReading">取消监听</a-button>
    <a-button @click="_onDisconnectSerial">断开</a-button>
  </layout-body>
</template>
<script setup lang="ts">
import { getCurrentInstance, ref, onMounted, watch } from "vue";

const {
  proxy,
  proxy: { $utils, $GoTools },
} = getCurrentInstance() as any;

const ports = ref<any[]>([]);
// 当前已连接的串口号，供监听/断开等操作使用
const connectedPort = ref<string>("");

const _onConnectSerial = () => {
  $GoTools.GetSerialPorts().then(res => {
    ports.value = res;
    if (!res || res.length === 0) {
      console.warn("未发现可用串口");
      return;
    }
    const portName = res[0].port_name;
    $GoTools.ConnectSerial(portName).then(() => {
      connectedPort.value = portName;
      console.log('连接成功', portName);
    }).catch((err) => {
      console.error('连接失败', err);
    });
  }).catch((err) => {
    console.error("获取串口列表失败:", err);
  });
}

const _onStartReading = () => {
  if (!connectedPort.value) {
    console.warn("请先连接串口");
    return;
  }
  $GoTools.StartReadingSerial(connectedPort.value).then((res) => {
    console.log('监听成功', res);
  }).catch((err) => {
    console.error('监听失败', err);
  });
}

const _onStopReading = () => {
  if (!connectedPort.value) {
    console.warn("请先连接串口");
    return;
  }
  $GoTools.StopReadingSerial(connectedPort.value).then((res) => {
    console.log('取消成功', res);
  }).catch((err) => {
    console.error('取消监听失败', err);
  });
}

const _onDisconnectSerial = () => {
  if (!connectedPort.value) {
    console.warn("请先连接串口");
    return;
  }
  const portName = connectedPort.value;
  $GoTools.DisconnectSerial(portName).then(() => {
    connectedPort.value = "";
    console.log('已断开', portName);
  }).catch((err) => {
    console.error('断开失败', err);
  });
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
</style>
