<template>
    <layout-body title="抓包工具">
        <!-- 顶部工具栏 -->
        <a-card class="tool-card toolbar">
            <a-row justify="space-between" align="center">
                <a-col>
                    <a-space :size="12">
                        <a-input-number v-model="port" :min="1" :max="65535" size="small" style="width: 110px"
                            :disabled="running" hide-button />
                        <a-tag v-if="!running" color="gray">未开始</a-tag>
                        <a-tag v-else color="green">
                            <template #icon><icon-record /></template>捕获中 127.0.0.1:{{ port }}
                        </a-tag>
                        <a-button v-if="!running" type="primary" status="success" @click="onStart">
                            <template #icon><icon-play-arrow /></template>开始抓包
                        </a-button>
                        <a-button v-else status="danger" @click="onStop">
                            <template #icon><icon-pause /></template>停止
                        </a-button>
                        <a-button :disabled="!flows.length" @click="clearAll">
                            <template #icon><icon-delete /></template>清空
                        </a-button>
                    </a-space>
                </a-col>
                <a-col>
                    <span class="count-tip">共 {{ flows.length }} 条 / 显示 {{ filteredFlows.length }} 条</span>
                </a-col>
            </a-row>
            <div class="ca-tip">
                <icon-info-circle /> 开始后会安装本地根证书并临时接管系统代理（停止时自动还原），用于解密 HTTPS 流量。
            </div>
        </a-card>

        <div class="main-area">
            <!-- 左侧：应用 / 域名 筛选 -->
            <div class="sidebar">
                <a-input-search v-model="sideSearch" placeholder="筛选" size="small" allow-clear class="side-search" />

                <div class="side-group">
                    <div class="side-title" @click="showApps = !showApps">
                        <icon-down :class="{ collapsed: !showApps }" /> 应用程序
                    </div>
                    <ul v-show="showApps" class="side-list">
                        <li :class="{ active: !selectedProcess, item: true }" @click="selectedProcess = ''">
                            <span class="name">全部</span><span class="cnt">{{ flows.length }}</span>
                        </li>
                        <li v-for="p in processList" :key="p.name"
                            :class="{ active: selectedProcess === p.name, item: true }" @click="toggleProcess(p.name)">
                            <span class="name" :title="p.name">{{ p.name }}</span><span class="cnt">{{ p.count }}</span>
                        </li>
                    </ul>
                </div>

                <a-divider margin="8" />

                <div class="side-group">
                    <div class="side-title" @click="showDomains = !showDomains">
                        <icon-down :class="{ collapsed: !showDomains }" /> 域名
                    </div>
                    <ul v-show="showDomains" class="side-list">
                        <li v-for="d in domainList" :key="d.name"
                            :class="{ active: selectedDomain === d.name, item: true }" @click="toggleDomain(d.name)">
                            <span class="name" :title="d.name">{{ d.name }}</span><span class="cnt">{{ d.count }}</span>
                        </li>
                    </ul>
                </div>
            </div>

            <!-- 中间：协议 Tab + 请求列表 -->
            <div class="center">
                <div class="proto-tabs">
                    <span v-for="t in PROTO_TABS" :key="t.key" class="proto-tab"
                        :class="{ active: selectedProto === t.key }" @click="selectedProto = t.key">
                        {{ t.label }}<span v-if="protoCount(t.key) && t.key !== 'All'" class="pcnt">{{ protoCount(t.key)
                            }}</span>
                    </span>
                    <a-input-search v-model="keyword" placeholder="搜索 URL / 域名" size="mini" allow-clear
                        style="width: 180px; margin-left: auto" />
                </div>

                <a-table :data="pagedFlows" row-key="id" size="small" :pagination="false"
                    :scroll="{ y: '100%', x: 720 }" :column-resizable="true"
                    :row-class="(r) => (r.id === selectedId ? 'row-selected' : '')" @row-click="onRowClick"
                    class="flow-table">
                    <template #columns>
                        <a-table-column title="ID" data-index="id" :width="60" />
                        <a-table-column title="协议" data-index="protocol" :width="70">
                            <template #cell="{ record }">
                                <a-tag size="small" :color="protoColor(record.protocol)">{{ record.protocol }}</a-tag>
                            </template>
                        </a-table-column>
                        <a-table-column title="方法" data-index="method" :width="70" />
                        <a-table-column title="URL" data-index="url" :ellipsis="true" :tooltip="true" />
                        <a-table-column title="应用程序" data-index="process" :width="120" :ellipsis="true"
                            :tooltip="true" />
                        <a-table-column title="状态" data-index="statusCode" :width="70">
                            <template #cell="{ record }">
                                <span :style="{ color: statusColor(record.statusCode) }">{{ record.statusCode }}</span>
                            </template>
                        </a-table-column>
                        <a-table-column title="大小" :width="90">
                            <template #cell="{ record }">{{ formatSize(record.respSize) }}</template>
                        </a-table-column>
                        <a-table-column title="耗时" :width="80">
                            <template #cell="{ record }">{{ record.duration }}ms</template>
                        </a-table-column>
                    </template>
                </a-table>

                <div class="pager">
                    <a-pagination v-model:current="page" :total="filteredFlows.length" :page-size="pageSize" size="mini"
                        :page-count="pageSize" show-total />
                </div>
            </div>

            <!-- 右侧：详情 -->
            <div class="detail">
                <template v-if="selectedFlow">
                    <div class="detail-head">
                        <a-tag :color="protoColor(selectedFlow.protocol)">{{ selectedFlow.method }}</a-tag>
                        <span class="detail-url" :title="selectedFlow.url">{{ selectedFlow.url }}</span>
                        <a-tag :style="{ color: statusColor(selectedFlow.statusCode) }">{{ selectedFlow.statusCode }}
                        </a-tag>
                    </div>
                    <a-tabs default-active-key="overview" size="small" type="text">
                        <a-tab-pane key="overview" title="概览">
                            <a-descriptions :column="1" size="small" bordered>
                                <a-descriptions-item label="时间">{{ selectedFlow.time }}</a-descriptions-item>
                                <a-descriptions-item label="协议">{{ selectedFlow.protocol }}</a-descriptions-item>
                                <a-descriptions-item label="方法">{{ selectedFlow.method }}</a-descriptions-item>
                                <a-descriptions-item label="域名">{{ selectedFlow.host }}</a-descriptions-item>
                                <a-descriptions-item label="路径">{{ selectedFlow.path }}</a-descriptions-item>
                                <a-descriptions-item label="状态">{{ selectedFlow.statusCode }}
                                    {{ selectedFlow.status }}</a-descriptions-item>
                                <a-descriptions-item label="应用程序">{{ selectedFlow.process || '-' }} (PID
                                    {{ selectedFlow.pid }})</a-descriptions-item>
                                <a-descriptions-item label="响应类型">{{ selectedFlow.contentType || '-'
                                }}</a-descriptions-item>
                                <a-descriptions-item label="请求/响应大小">{{ formatSize(selectedFlow.reqSize) }} /
                                    {{ formatSize(selectedFlow.respSize) }}</a-descriptions-item>
                                <a-descriptions-item label="耗时">{{ selectedFlow.duration }} ms</a-descriptions-item>
                            </a-descriptions>
                        </a-tab-pane>
                        <a-tab-pane key="reqHeaders" :title="`请求头(${headerCount(selectedFlow.requestHeaders)})`">
                            <a-descriptions v-if="headerEntries(selectedFlow.requestHeaders).length" :column="1"
                                size="small" bordered>
                                <a-descriptions-item v-for="hd in headerEntries(selectedFlow.requestHeaders)"
                                    :key="hd.k" :label="hd.k">
                                    <span class="hdr-val">{{ hd.v }}</span>
                                </a-descriptions-item>
                            </a-descriptions>
                            <a-empty v-else description="（无请求头）" />
                        </a-tab-pane>
                        <a-tab-pane key="reqBody" title="请求体">
                            <div class="body-bar">
                                <a-tag size="small" color="gray">{{ bodyKind(selectedFlow.requestBody) }}</a-tag>
                                <a-button v-if="selectedFlow.requestBody" size="mini" type="text"
                                    @click="copyText(selectedFlow.requestBody)">
                                    <template #icon><icon-copy /></template>复制
                                </a-button>
                            </div>
                            <pre class="body-pre">{{ formatBody(selectedFlow.requestBody) }}</pre>
                        </a-tab-pane>
                        <a-tab-pane key="respHeaders" :title="`响应头(${headerCount(selectedFlow.responseHeaders)})`">
                            <a-descriptions v-if="headerEntries(selectedFlow.responseHeaders).length" :column="1"
                                size="small" bordered>
                                <a-descriptions-item v-for="hd in headerEntries(selectedFlow.responseHeaders)"
                                    :key="hd.k" :label="hd.k">
                                    <span class="hdr-val">{{ hd.v }}</span>
                                </a-descriptions-item>
                            </a-descriptions>
                            <a-empty v-else description="（无响应头）" />
                        </a-tab-pane>
                        <a-tab-pane key="respBody" title="响应体">
                            <div class="body-bar">
                                <a-tag size="small" color="gray">{{ bodyKind(selectedFlow.responseBody) }}</a-tag>
                                <a-button v-if="selectedFlow.responseBody" size="mini" type="text"
                                    @click="copyText(selectedFlow.responseBody)">
                                    <template #icon><icon-copy /></template>复制
                                </a-button>
                            </div>
                            <pre class="body-pre">{{ formatBody(selectedFlow.responseBody) }}</pre>
                        </a-tab-pane>
                    </a-tabs>
                </template>
                <a-empty v-else description="选择左侧请求查看详情" style="margin-top: 60px" />
            </div>
        </div>
    </layout-body>
</template>

<script setup lang="ts">
import { computed, getCurrentInstance, onMounted, onUnmounted, ref } from "vue";
import { Message } from "@arco-design/web-vue";
import type { TableData } from "@arco-design/web-vue";
import { Events } from "@wailsio/runtime";
import type { Flow } from "bindings/go-tools/tools/sniffer/models";

const {
    proxy: { $GoTools },
} = getCurrentInstance() as any;

/** 协议 Tab */
const PROTO_TABS = [
    { key: "All", label: "All" },
    { key: "HTTP", label: "HTTP" },
    { key: "HTTPS", label: "HTTPS" },
    { key: "WS", label: "WebSocket" },
];

/* ============ 状态 ============ */
const port = ref(8888);
const running = ref(false);
const flows = ref<Flow[]>([]);
const MAX_FLOWS = 3000;

const selectedProto = ref("All");
const selectedProcess = ref("");
const selectedDomain = ref("");
const keyword = ref("");
const sideSearch = ref("");
const showApps = ref(true);
const showDomains = ref(true);

const selectedId = ref<number | null>(null);
const page = ref(1);
const pageSize = 200;

let offEvent: (() => void) | null = null;

/* ============ 捕获控制 ============ */
const onStart = async () => {
    try {
        await $GoTools.StartSnifferProxy(port.value);
        running.value = true;
        Message.success("抓包已启动，系统代理已接管");
    } catch (e: any) {
        Message.error("启动失败：" + (e?.message || e));
    }
};

const onStop = async () => {
    try {
        await $GoTools.StopSnifferProxy();
        running.value = false;
        Message.success("抓包已停止，系统代理已还原");
    } catch (e: any) {
        Message.error("停止失败：" + (e?.message || e));
    }
};

const clearAll = () => {
    flows.value = [];
    selectedId.value = null;
};

const onFlow = (ev: any) => {
    const f = ev.data as Flow;
    if (!f) return;
    flows.value.push(f);
    if (flows.value.length > MAX_FLOWS) {
        flows.value.splice(0, flows.value.length - MAX_FLOWS);
    }
};

onMounted(async () => {
    offEvent = Events.On("sniffer:flow", onFlow);
    try {
        const st = await $GoTools.SnifferProxyStatus();
        if (Array.isArray(st)) {
            running.value = !!st[0];
            if (st[1]) port.value = st[1];
        }
    } catch {
        /* 忽略状态查询失败 */
    }
});

onUnmounted(() => {
    if (offEvent) offEvent();
});

/* ============ 筛选与统计 ============ */
const matchKeyword = (f: Flow) => {
    if (!keyword.value) return true;
    const k = keyword.value.toLowerCase();
    return f.url?.toLowerCase().includes(k) || f.host?.toLowerCase().includes(k);
};

const filteredFlows = computed(() =>
    flows.value.filter(
        (f) =>
            (selectedProto.value === "All" || f.protocol === selectedProto.value) &&
            (!selectedProcess.value || f.process === selectedProcess.value) &&
            (!selectedDomain.value || f.host === selectedDomain.value) &&
            matchKeyword(f)
    )
);

const pagedFlows = computed(() => {
    const start = (page.value - 1) * pageSize;
    return filteredFlows.value.slice(start, start + pageSize);
});

const processList = computed(() => buildCount("process"));
const domainList = computed(() => {
    const kw = sideSearch.value.toLowerCase();
    return buildCount("host").filter((d) => !kw || d.name.toLowerCase().includes(kw));
});

function buildCount(field: "process" | "host") {
    const map = new Map<string, number>();
    for (const f of flows.value) {
        const key = (f[field] || "").trim() || "(未知)";
        map.set(key, (map.get(key) || 0) + 1);
    }
    const kw = field === "process" ? sideSearch.value.toLowerCase() : "";
    let arr = Array.from(map, ([name, count]) => ({ name, count }));
    if (kw) arr = arr.filter((x) => x.name.toLowerCase().includes(kw));
    return arr.sort((a, b) => b.count - a.count);
}

const protoCount = (key: string) => {
    if (key === "All") return flows.value.length;
    return flows.value.filter((f) => f.protocol === key).length;
};

const selectedFlow = computed(() => flows.value.find((f) => f.id === selectedId.value) || null);

const onRowClick = (record: TableData) => {
    selectedId.value = (record as Flow).id;
};

const toggleProcess = (name: string) => (selectedProcess.value = selectedProcess.value === name ? "" : name);
const toggleDomain = (name: string) => (selectedDomain.value = selectedDomain.value === name ? "" : name);

/* ============ 展示辅助 ============ */
const headerCount = (hObj: Hdrs) => (hObj ? Object.keys(hObj).length : 0);

const formatSize = (n: number) => {
    if (!n) return "0 B";
    if (n < 1024) return `${n} B`;
    if (n < 1024 * 1024) return `${(n / 1024).toFixed(1)} KB`;
    return `${(n / 1024 / 1024).toFixed(1)} MB`;
};

const protoColor = (p: string) => (p === "HTTPS" ? "green" : p === "WS" ? "purple" : "arcoblue");

const statusColor = (code: number) => {
    if (!code) return "var(--color-text-3)";
    if (code < 300) return "rgb(var(--green-6))";
    if (code < 400) return "rgb(var(--arcoblue-6))";
    if (code < 500) return "rgb(var(--orange-6))";
    return "rgb(var(--red-6))";
};

/* ============ 详情美化辅助 ============ */
type Hdrs = Record<string, string[] | null | undefined> | null | undefined;

/** 把请求/响应头映射转为 [{ 名, 值 }] 列表 */
const headerEntries = (hObj: Hdrs) =>
    Object.entries(hObj || {}).map(([k, v]) => ({ k, v: (v || []).join(", ") }));

/** 尝试将 JSON 美化缩进，失败则原样返回 */
const formatBody = (s?: string) => {
    if (!s) return "(空)";
    try {
        return JSON.stringify(JSON.parse(s), null, 2);
    } catch {
        return s;
    }
};

/** 粗略判断请求/响应体类型 */
const bodyKind = (s?: string) => {
    if (!s) return "空";
    try {
        JSON.parse(s);
        return "JSON";
    } catch {
        /* 非 JSON */
    }
    if (s.trimStart().startsWith("<")) return "HTML/XML";
    return "文本";
};

const copyText = async (text: string) => {
    try {
        await navigator.clipboard.writeText(text);
        Message.success("已复制到剪贴板");
    } catch {
        Message.error("复制失败");
    }
};
</script>

<style scoped>
.tool-card {
    margin-bottom: var(--base-margin);
}

.toolbar .count-tip {
    color: var(--color-text-3);
    font-size: var(--base-size-small);
}

.ca-tip {
    margin-top: var(--base-padding);
    color: var(--color-text-3);
    font-size: var(--base-size-small);
}

.main-area {
    display: flex;
    gap: 12px;
    height: calc(100vh - 220px);
    min-height: 420px;
}

/* 左侧 */
.sidebar {
    width: 200px;
    flex: 0 0 200px;
    border: 1px solid var(--color-border-2);
    border-radius: var(--base-radius-large);
    padding: var(--base-padding);
    overflow-y: auto;
    background: var(--color-bg-2);
}

.side-search {
    margin-bottom: var(--base-padding);
}

.side-title {
    display: flex;
    align-items: center;
    gap: 4px;
    font-weight: 600;
    font-size: var(--base-size-small);
    color: var(--color-text-1);
    cursor: pointer;
    margin: var(--base-radius-default) 0;
}

.side-title .collapsed {
    transform: rotate(-90deg);
}

.side-list {
    list-style: none;
    margin: 0;
    padding: 0;
}

.side-list .item {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: var(--base-radius-default) var(--base-padding);
    border-radius: var(--base-radius-default);
    cursor: pointer;
    font-size: var(--base-size-small);
    color: var(--color-text-2);
}

.side-list .item:hover {
    background: var(--color-fill-2);
}

.side-list .item.active {
    background: var(--color-primary-light-1);
    color: rgb(var(--primary-6));
}

.side-list .name {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    max-width: 130px;
}

.side-list .cnt {
    color: var(--color-text-3);
    font-size: var(--base-size-small);
}

/* 中间 */
.center {
    flex: 1 1 46%;
    display: flex;
    flex-direction: column;
    border: 1px solid var(--color-border-2);
    border-radius: var(--base-radius-large);
    overflow: hidden;
}

.proto-tabs {
    display: flex;
    align-items: center;
    gap: 4px;
    padding: var(--base-radius-default) var(--base-radius-large);
    border-bottom: 1px solid var(--color-border-2);
    flex-wrap: wrap;
}

.proto-tab {
    padding: var(--base-radius-small) var(--base-radius-large);
    border-radius: var(--base-radius-default);
    cursor: pointer;
    font-size: var(--base-size-small);
    color: var(--color-text-2);
}

.proto-tab:hover {
    background: var(--color-fill-2);
}

.proto-tab.active {
    background: rgb(var(--primary-6));
    color: #fff;
}

.proto-tab .pcnt {
    margin-left: var(--base-radius-default);
    opacity: 0.85;
    font-size: var(--base-size-small);
}

.flow-table {
    flex: 1;
    overflow: hidden;
}

.pager {
    padding: var(--base-radius-default) var(--base-radius-large);
    border-top: 1px solid var(--color-border-2);
    display: flex;
    justify-content: flex-end;
}

:deep(.row-selected) {
    background-color: var(--color-primary-light-1) !important;
}

/* 右侧详情 */
.detail {
    flex: 1 1 40%;
    border: 1px solid var(--color-border-2);
    border-radius: var(--base-radius-large);
    padding: var(--base-radius-large) var(--base-padding);
    overflow-y: auto;
    background: var(--color-bg-2);
}

.detail-head {
    display: flex;
    align-items: center;
    gap: 8px;
    margin-bottom: var(--base-padding);
}

.detail-url {
    flex: 1;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    font-size: var(--base-size-small);
    color: var(--color-text-1);
}

.body-pre {
    background: #1d2129;
    color: #e6e6e6;
    padding: var(--base-radius-large) var(--base-padding);
    border-radius: var(--base-radius-default);
    font-size: var(--base-size-small);
    white-space: pre-wrap;
    word-break: break-all;
    max-height: 360px;
    overflow: auto;
}

.body-bar {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: var(--base-radius-default);
}

.hdr-val {
    word-break: break-all;
    white-space: pre-wrap;
}
</style>
