<template>
    <layout-body title="串口调试">
        <!-- 连接配置区 -->
        <a-card class="tool-card conn-card">
            <a-row :gutter="12" align="center" class="conn-row">
                <a-col :span="24">
                    <a-space :size="10" wrap>
                        <span class="led" :class="{ on: st.opened }" :title="st.opened ? '已连接' : '未连接'"></span>
                        <span class="field-label">端口号</span>
                        <a-select v-model="st.selected" style="width: 220px" :disabled="st.opened" size="small"
                            @change="onSelectChange">
                            <a-option v-for="p in ports" :key="p.value" :value="p.value">{{ p.label }}</a-option>
                            <a-option value="TCPClient">TCPClient</a-option>
                            <a-option value="TCPServer">TCPServer</a-option>
                            <a-option value="UDP">UDP</a-option>
                        </a-select>

                        <!-- 打开 / 关闭 -->
                        <a-button v-if="!st.opened" type="primary" status="success" size="small" @click="openConn">
                            <template #icon><icon-play-arrow /></template>{{ openLabel }}
                        </a-button>
                        <a-button v-else status="danger" size="small" @click="closeConn">
                            <template #icon><icon-poweroff /></template>{{ closeLabel }}
                        </a-button>
                        <a-button v-if="mode === 'serial'" size="small" @click="loadPorts" :disabled="st.opened">
                            <template #icon><icon-refresh /></template>刷新
                        </a-button>
                        <a-button v-if="mode === 'serial'" size="small" @click="settingVisible = true"
                            :disabled="st.opened">
                            <template #icon><icon-settings /></template>串口设置
                        </a-button>
                    </a-space>
                </a-col>

                <!-- 串口参数行 -->
                <a-col v-if="mode === 'serial'" :span="24">
                    <a-space :size="12" wrap class="sub-row">
                        <a-checkbox v-model="st.rts" @change="applyControlBits">RTS</a-checkbox>
                        <a-checkbox v-model="st.dtr" @change="applyControlBits">DTR</a-checkbox>
                        <span class="field-label">波特率：</span>
                        <a-select v-model="st.baud" size="small" style="width: 120px" :disabled="st.opened">
                            <a-option v-for="b in BAUDS" :key="b" :value="b">{{ b }}</a-option>
                        </a-select>
                        <a-tag size="small" color="gray">{{ st.dataBits }}-{{ st.parity }}-{{ st.stopBits }}</a-tag>
                    </a-space>
                </a-col>

                <!-- 网络参数行 -->
                <a-col v-else :span="24">
                    <a-space :size="10" wrap class="sub-row">
                        <template v-if="mode !== 'tcpserver'">
                            <span class="field-label">远程</span>
                            <a-input v-model="st.remoteAddr" size="small" style="width: 160px" :disabled="st.opened"
                                placeholder="192.168.1.20" />
                            <a-input-number v-model="st.remotePort" size="small" style="width: 90px" :min="1"
                                :max="65535" :disabled="st.opened" hide-button />
                        </template>
                        <span class="field-label">本地</span>
                        <a-select v-model="st.localAddr" size="small" style="width: 150px" :disabled="st.opened"
                            allow-search placeholder="本地IP">
                            <a-option v-for="ip in localIPs" :key="ip" :value="ip">{{ ip }}</a-option>
                        </a-select>
                        <a-input-number v-model="st.localPort" size="small" style="width: 90px" :min="0" :max="65535"
                            :disabled="st.opened" hide-button />
                    </a-space>
                </a-col>
            </a-row>
        </a-card>

        <!-- 收发选项 -->
        <a-card class="tool-card opt-card">
            <a-space :size="14" wrap>
                <a-checkbox v-model="st.hexDisplay">HEX显示</a-checkbox>
                <a-checkbox v-model="st.hexSend">HEX发送</a-checkbox>
                <a-checkbox v-model="st.addCRLF">加回车换行</a-checkbox>
                <a-checkbox v-model="st.timestamp">加时间戳和分包显示</a-checkbox>
                <a-checkbox v-model="st.timerSend">定时发送</a-checkbox>
                <a-input-number v-model="st.timerInterval" size="mini" :min="50" :step="100" style="width: 90px"
                    :disabled="!st.timerSend" hide-button />
                <span class="unit">ms/次</span>
                <span class="field-label">加校验</span>
                <a-select v-model="st.checksum" size="mini" style="width: 140px">
                    <a-option v-for="c in CHECKSUM_OPTIONS" :key="c.value" :value="c.value">{{ c.label }}</a-option>
                </a-select>
                <a-button size="mini" type="outline" @click="clearRecv">
                    <template #icon><icon-delete /></template>清空接收
                </a-button>
                <a-button size="mini" type="outline" :disabled="!recvLines.length" @click="exportRecv">
                    <template #icon><icon-download /></template>保存数据
                </a-button>
            </a-space>
        </a-card>

        <!-- 接收区 -->
        <a-card class="tool-card recv-card">
            <div class="recv-head">
                <a-space :size="16">
                    <a-tag color="arcoblue" size="small">接收 RX: {{ st.rx }}</a-tag>
                    <a-tag color="green" size="small">发送 TX: {{ st.tx }}</a-tag>
                    <span v-if="st.label" class="conn-tip">{{ st.label }}</span>
                </a-space>
            </div>
            <div ref="recvRef" class="recv-body" :class="{ hex: st.hexDisplay }">{{ recvText }}</div>
        </a-card>

        <!-- 发送区 -->
        <a-card class="tool-card send-card">
            <div class="field-label">发送数据</div>
            <a-textarea v-model="sendText" :auto-size="{ minRows: 2, maxRows: 6 }"
                :placeholder="st.hexSend ? '以 HEX 发送，例如：AA AA 1E AA A6 00 …' : '输入要发送的文本，Ctrl+Enter 发送'"
                allow-clear @press-enter.ctrl="doSend" />
            <a-row justify="space-between" align="center" class="send-bar">
                <a-col>
                    <a-space :size="10">
                        <a-button size="small" @click="sendText = ''" :disabled="!sendText">清空</a-button>
                        <a-button size="small" @click="loadHistory" :disabled="!history.length">
                            <template #icon><icon-history /></template>历史
                        </a-button>
                    </a-space>
                </a-col>
                <a-col>
                    <a-button type="primary" status="success" :disabled="!st.opened || !sendText" @click="doSend">
                        <template #icon><icon-send /></template>发 送
                    </a-button>
                </a-col>
            </a-row>
        </a-card>

        <!-- 串口设置弹窗 -->
        <a-modal v-model:visible="settingVisible" title="串口设置" :width="360" :footer="false">
            <a-form :model="st" layout="vertical" auto-width>
                <a-form-item field="dataBits" label="数据位">
                    <a-select v-model="st.dataBits">
                        <a-option v-for="d in [5, 6, 7, 8]" :key="d" :value="d">{{ d }}</a-option>
                    </a-select>
                </a-form-item>
                <a-form-item field="parity" label="校验位">
                    <a-select v-model="st.parity">
                        <a-option value="none">None</a-option>
                        <a-option value="odd">Odd</a-option>
                        <a-option value="even">Even</a-option>
                        <a-option value="mark">Mark</a-option>
                        <a-option value="space">Space</a-option>
                    </a-select>
                </a-form-item>
                <a-form-item field="stopBits" label="停止位">
                    <a-select v-model="st.stopBits">
                        <a-option :value="1">1</a-option>
                        <a-option :value="2">2</a-option>
                    </a-select>
                </a-form-item>
            </a-form>
        </a-modal>

        <!-- 发送历史下拉 -->
        <a-drawer v-model:visible="historyVisible" title="发送历史" :width="360" :footer="false" placement="right">
            <div v-for="(h, i) in history" :key="i" class="history-item" @click="pickHistory(h)">
                {{ h }}
            </div>
            <a-empty v-if="!history.length" description="暂无历史" />
        </a-drawer>
    </layout-body>
</template>

<script setup lang="ts">
import { computed, getCurrentInstance, nextTick, onMounted, onUnmounted, reactive, ref, watch } from "vue";
import { Message } from "@arco-design/web-vue";
import { Events } from "@wailsio/runtime";
import {
    appendChecksum,
    bytesToHex,
    bytesToText,
    CHECKSUM_OPTIONS,
    type ChecksumKind,
    hexToBytes,
    nowStamp,
    textToBytes,
} from "./serialUtils";

const {
    proxy: { $GoTools },
} = getCurrentInstance() as any;

const BAUDS = [1200, 2400, 4800, 9600, 19200, 38400, 57600, 115200, 230400, 460800, 921600];

/** 统一状态对象，便于模板深层访问 */
const st = reactive({
    selected: "",
    opened: false,
    label: "",
    // 串口
    baud: 115200,
    dataBits: 8,
    parity: "none",
    stopBits: 1,
    rts: false,
    dtr: true,
    // 网络
    remoteAddr: "192.168.1.20",
    remotePort: 8234,
    localAddr: "",
    localPort: 777,
    // 选项
    hexDisplay: true,
    hexSend: false,
    addCRLF: false,
    timestamp: true,
    timerSend: false,
    timerInterval: 2000,
    checksum: "none" as ChecksumKind,
    // 计数
    rx: 0,
    tx: 0,
});

const ports = ref<{ value: string; label: string }[]>([]);
const localIPs = ref<string[]>([]);
const sendText = ref("");
const settingVisible = ref(false);
const historyVisible = ref(false);
const history = ref<string[]>([]);
const recvRef = ref<HTMLElement | null>(null);

interface RecvLine {
    time: string;
    bytes: number[];
    dir: "in" | "out";
}
const recvLines = ref<RecvLine[]>([]);
const MAX_LINES = 2000;

/** 当前连接模式：serial / tcpclient / tcpserver / udp */
const mode = computed(() => {
    const s = st.selected;
    if (s === "TCPClient") return "tcpclient";
    if (s === "TCPServer") return "tcpserver";
    if (s === "UDP") return "udp";
    return "serial";
});

const openLabel = computed(
    () =>
        ({ serial: "打开串口", tcpclient: "连接", tcpserver: "侦听", udp: "打开端口" }[mode.value] || "打开")
);
const closeLabel = computed(() => (mode.value === "tcpserver" ? "停止侦听" : "关闭") + (mode.value === "serial" ? "串口" : ""));

/** 接收区文本渲染 */
const recvText = computed(() =>
    recvLines.value
        .map((l) => {
            const body = st.hexDisplay ? bytesToHex(l.bytes) : bytesToText(l.bytes);
            const head = (st.timestamp ? `[${l.time}] ` : "") + (l.dir === "out" ? "→ " : "");
            return head + body;
        })
        .join("\n")
);

/* ============ 加载串口列表 / 本机 IP ============ */
async function loadPorts() {
    try {
        const res = await $GoTools.GetSerialPorts();
        ports.value = (res || []).map((p: any) => ({
            value: p.port_name,
            label: `${p.port_name} ${p.description}`,
        }));
        if (!st.selected && ports.value.length) st.selected = ports.value[0].value;
    } catch (e: any) {
        Message.error("获取串口列表失败：" + (e?.message || e));
    }
}

async function loadLocalIPs() {
    try {
        const ips = await $GoTools.GetLocalIPs();
        localIPs.value = ["0.0.0.0", ...(ips || [])];
        if (!st.localAddr) st.localAddr = "0.0.0.0";
    } catch {
        localIPs.value = ["0.0.0.0"];
        st.localAddr = "0.0.0.0";
    }
}

function onSelectChange() {
    // 切换端口/模式时，若已连接先关闭
    if (st.opened) closeConn();
}

/* ============ 连接控制 ============ */
function buildCfg() {
    return {
        port: st.selected,
        baud: st.baud,
        dataBits: st.dataBits,
        parity: st.parity,
        stopBits: st.stopBits,
        rts: st.rts,
        dtr: st.dtr,
        localAddr: st.localAddr,
        localPort: st.localPort,
        remoteAddr: st.remoteAddr,
        remotePort: st.remotePort,
    };
}

async function openConn() {
    const cfg = buildCfg();
    try {
        if (mode.value === "serial") await $GoTools.OpenSerial(cfg);
        else if (mode.value === "tcpclient") await $GoTools.OpenTCPClient(cfg);
        else if (mode.value === "tcpserver") await $GoTools.ListenTCPServer(cfg);
        else if (mode.value === "udp") await $GoTools.OpenUDP(cfg);
        st.opened = true;
        st.label = cfgLabel();
        Message.success(openLabel.value + "成功");
        if (st.timerSend) startTimer();
    } catch (e: any) {
        st.opened = false;
        Message.error("打开失败：" + (e?.message || e));
    }
}

function cfgLabel() {
    if (mode.value === "serial") return `${st.selected} @${st.baud}`;
    if (mode.value === "tcpclient") return `→ ${st.remoteAddr}:${st.remotePort}`;
    if (mode.value === "tcpserver") return `侦听 ${st.localAddr || "0.0.0.0"}:${st.localPort}`;
    return `UDP ${st.localAddr || "0.0.0.0"}:${st.localPort}`;
}

async function closeConn() {
    stopTimer();
    try {
        await $GoTools.CloseConnection();
    } catch {
        /* 忽略 */
    }
    st.opened = false;
    st.label = "";
}

async function applyControlBits() {
    if (mode.value !== "serial" || !st.opened) return;
    try {
        await $GoTools.SetSerialControlBits(st.rts, st.dtr);
    } catch (e: any) {
        Message.error("控制位设置失败：" + (e?.message || e));
    }
}

/* ============ 发送 ============ */
function buildSendBytes(): number[] {
    let bytes = st.hexSend ? hexToBytes(sendText.value) : textToBytes(sendText.value);
    if (st.addCRLF) bytes = [...bytes, 0x0d, 0x0a];
    bytes = appendChecksum(bytes, st.checksum);
    return bytes;
}

async function doSend() {
    if (!st.opened) {
        Message.warning("请先打开连接");
        return;
    }
    const bytes = buildSendBytes();
    if (!bytes.length) return;
    try {
        await $GoTools.SendSerialHex(bytesToHex(bytes).replace(/ /g, ""));
        st.tx += bytes.length;
        pushLine({ time: nowStamp(), bytes, dir: "out" });
        rememberHistory();
    } catch (e: any) {
        Message.error("发送失败：" + (e?.message || e));
    }
}

function rememberHistory() {
    const t = sendText.value.trim();
    if (t) history.value = [t, ...history.value.filter((x) => x !== t)].slice(0, 30);
}

function loadHistory() {
    historyVisible.value = true;
}
function pickHistory(h: string) {
    sendText.value = h;
    historyVisible.value = false;
}

/* ============ 定时发送 ============ */
let timer: number | null = null;
function startTimer() {
    stopTimer();
    if (!st.timerSend || !st.opened) return;
    timer = window.setInterval(() => doSend(), Math.max(50, st.timerInterval));
}
function stopTimer() {
    if (timer) {
        clearInterval(timer);
        timer = null;
    }
}
watch(
    () => st.timerSend,
    (v) => (v ? startTimer() : stopTimer())
);
watch(() => st.timerInterval, () => st.timerSend && startTimer());

/* ============ 接收 ============ */
function pushLine(line: RecvLine) {
    recvLines.value.push(line);
    if (recvLines.value.length > MAX_LINES) {
        recvLines.value.splice(0, recvLines.value.length - MAX_LINES);
    }
    scrollToBottom();
}

async function scrollToBottom() {
    await nextTick();
    if (recvRef.value) recvRef.value.scrollTop = recvRef.value.scrollHeight;
}

const onData = (ev: any) => {
    const d = ev.data;
    if (!d || !d.hex) return;
    const bytes = hexToBytes(d.hex);
    st.rx += bytes.length;
    pushLine({ time: d.time || nowStamp(), bytes, dir: "in" });
};

const onStatus = (ev: any) => {
    const s = ev.data;
    if (!s) return;
    if (s.note) Message.info(s.note);
    if (!s.opened && st.opened) {
        st.opened = false;
        st.label = "";
        stopTimer();
    }
};

function clearRecv() {
    recvLines.value = [];
    st.rx = 0;
    st.tx = 0;
}

/* ============ 导出数据 ============ */
function exportRecv() {
    const content = recvLines.value
        .map((l) => `${l.time}  ${l.dir === "out" ? "→" : "←"}  ${st.hexDisplay ? bytesToHex(l.bytes) : bytesToText(l.bytes)}`)
        .join("\n");
    const blob = new Blob([content], { type: "text/plain;charset=utf-8" });
    const url = URL.createObjectURL(blob);
    const a = document.createElement("a");
    a.href = url;
    a.download = `serial-recv-${Date.now()}.txt`;
    a.click();
    URL.revokeObjectURL(url);
}

/* ============ 生命周期 ============ */
let offData: (() => void) | null = null;
let offStatus: (() => void) | null = null;

onMounted(async () => {
    offData = Events.On("serial:data", onData);
    offStatus = Events.On("serial:status", onStatus);
    await Promise.all([loadPorts(), loadLocalIPs()]);
    try {
        const s = await $GoTools.ConnectionStatus();
        if (s && s.opened) {
            st.opened = true;
            st.label = s.label;
        }
    } catch {
        /* 忽略 */
    }
});

onUnmounted(() => {
    stopTimer();
    if (offData) offData();
    if (offStatus) offStatus();
    if (st.opened) $GoTools.CloseConnection().catch(() => {});
});
</script>

<style scoped>
.tool-card {
    margin-bottom: 12px;
}

.led {
    display: inline-block;
    width: 12px;
    height: 12px;
    border-radius: 50%;
    background: #f53f3f;
    box-shadow: 0 0 4px rgba(245, 63, 63, 0.6);
    vertical-align: middle;
}

.led.on {
    background: #00b42a;
    box-shadow: 0 0 6px rgba(0, 180, 42, 0.7);
}

.field-label {
    color: #4e5969;
    font-size: 13px;
    font-weight: 500;
}

.sub-row {
    margin-top: 6px;
}

.unit {
    color: #86909c;
    font-size: 12px;
}

.conn-tip {
    color: #86909c;
    font-size: 12px;
}

.recv-card :deep(.arco-card-body) {
    padding: 8px 12px;
}

.recv-head {
    margin-bottom: 6px;
}

.recv-body {
    height: 260px;
    overflow-y: auto;
    background: #1d2129;
    color: #e6e6e6;
    border-radius: 6px;
    padding: 8px 10px;
    font-size: 13px;
    line-height: 1.6;
    white-space: pre-wrap;
    word-break: break-all;
    font-family: Consolas, Monaco, "Courier New", monospace;
}

.recv-body.hex {
    color: #ffd591;
}

.send-bar {
    margin-top: 10px;
}

.history-item {
    padding: 8px 10px;
    border-bottom: 1px solid #f2f3f5;
    cursor: pointer;
    font-family: Consolas, Monaco, monospace;
    font-size: 13px;
    word-break: break-all;
}

.history-item:hover {
    background: var(--color-fill-2);
}
</style>
