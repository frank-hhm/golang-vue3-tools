<template>
    <layout-body>
        <div class="redis-page">
            <!-- 左侧：连接列表 -->
            <div class="conn-panel">
                <div class="conn-header">
                    <span class="conn-title">连接</span>
                    <a-button size="mini" type="text" @click="openAddModal()">
                        <template #icon><icon-plus /></template>
                    </a-button>
                </div>
                <div class="conn-list">
                    <div v-for="c in savedConns" :key="c.id" class="conn-item" :class="{ active: isOpen(c.id) }"
                        @click="openConnection(c)">
                        <span class="led" :class="{ on: isOpen(c.id) }"></span>
                        <span class="conn-name">{{ c.name || (c.host + ':' + c.port) }}</span>
                        <a-dropdown trigger="hover" @select="(v: any) => onConnAction(v, c)">
                            <icon-more class="conn-more" @click.stop />
                            <template #content>
                                <a-doption value="open"><icon-play-arrow /> 打开连接</a-doption>
                                <a-doption value="edit"><icon-edit /> 编辑</a-doption>
                                <a-doption value="del"><icon-delete /> 删除</a-doption>
                            </template>
                        </a-dropdown>
                    </div>
                    <div v-if="savedConns.length === 0" class="conn-empty">暂无连接</div>
                </div>
            </div>

            <!-- 右侧：主区域 -->
            <div class="main-panel">
                <!-- 空状态 -->
                <div v-if="openTabs.length === 0" class="empty-state">
                    <icon-empty :size="56" class="empty-icon" />
                    <div class="empty-text">可以从左边选择并打开连接</div>
                    <a-button type="outline" @click="openAddModal()">
                        <template #icon><icon-plus /></template>添加新连接
                    </a-button>
                </div>

                <template v-else>
                    <!-- 已打开连接标签 -->
                    <div class="tab-bar">
                        <div v-for="t in openTabs" :key="t.id" class="tab-item"
                            :class="{ active: t.id === st.activeTab }" @click="switchTab(t.id)">
                            <icon-storage />
                            <span class="tab-name">{{ t.name }}</span>
                            <icon-close class="tab-close" @click.stop="closeTab(t.id)" />
                        </div>
                    </div>

                    <!-- 子标签 -->
                    <div class="sub-tab-bar">
                        <a-radio-group v-model="st.activeSubTab" type="button" size="small" @change="onSubTabChange">
                            <a-radio value="status"><icon-dashboard /> 状态</a-radio>
                            <a-radio value="keys"><icon-list /> 键详情</a-radio>
                            <a-radio value="slowlog"><icon-clock-circle /> 慢日志</a-radio>
                            <a-radio value="monitor"><icon-scan /> 监控命令</a-radio>
                        </a-radio-group>
                    </div>

                    <div class="tab-content">
                        <!-- 状态 -->
                        <div v-show="st.activeSubTab === 'status'" class="status-view">
                            <a-button size="mini" type="text" @click="loadStatus()"><template
                                    #icon><icon-refresh /></template>刷新</a-button>
                            <div class="status-head">
                                <span class="status-addr">{{ st.status?.host || activeTabName }}</span>
                                <a-tag size="small" color="red">v{{ st.status?.version || '?' }}</a-tag>
                                <a-tag size="small" color="orange">{{ st.status?.mode || 'standalone' }}</a-tag>
                                <a-tag size="small" color="purple">{{ st.status?.role || 'master' }}</a-tag>
                            </div>
                            <div class="status-grid">
                                <div class="stat-card">
                                    <div class="stat-label">运行时间</div>
                                    <div class="stat-value">{{ formatUptime(st.status?.uptimeSeconds || 0) }}</div>
                                </div>
                                <div class="stat-card">
                                    <div class="stat-label">已连客户端</div>
                                    <div class="stat-value">{{ st.status?.connectedClients ?? 0 }}</div>
                                </div>
                                <div class="stat-card">
                                    <div class="stat-label">键总数</div>
                                    <div class="stat-value">{{ st.status?.totalKeys ?? 0 }}</div>
                                </div>
                                <div class="stat-card">
                                    <div class="stat-label">内存使用</div>
                                    <div class="stat-value">{{ st.status?.usedMemoryHuman || '0' }}</div>
                                </div>
                            </div>
                        </div>

                        <!-- 键详情 -->
                        <div v-show="st.activeSubTab === 'keys'" class="keys-view">
                            <div class="key-list-panel">
                                <div class="key-toolbar">
                                    <a-input v-model="st.keyFilter" size="mini" placeholder="筛选键 (支持 * 通配)" allow-clear
                                        @press-enter="reloadKeys()">
                                        <template #prefix><icon-search /></template>
                                    </a-input>
                                    <a-button size="mini" @click="reloadKeys()"><template
                                            #icon><icon-refresh /></template></a-button>
                                    <a-button size="mini" type="primary" @click="openAddKey()"><template
                                            #icon><icon-plus /></template></a-button>
                                </div>
                                <div class="key-list">
                                    <div v-for="k in st.keys" :key="k.key" class="key-item"
                                        :class="{ active: k.key === st.selectedKey }" @click="selectKey(k.key)">
                                        <span class="type-badge" :class="'t-' + k.type">{{ (k.type[0] ||
                                            '?').toUpperCase() }}</span>
                                        <span class="key-name" :title="k.key">{{ k.key }}</span>
                                        <icon-delete class="key-del" @click.stop="delKey(k.key)" />
                                    </div>
                                    <div v-if="st.keys.length === 0" class="conn-empty">无键</div>
                                </div>
                            </div>

                            <div class="key-detail-panel">
                                <div v-if="!st.selectedKey" class="conn-empty" style="margin-top: 40px">请选择左侧键查看</div>
                                <template v-else>
                                    <div class="kd-toolbar">
                                        <span class="type-badge big" :class="'t-' + st.keyValue.type">{{
                                            st.keyValue.type }}</span>
                                        <span class="kd-key">{{ st.selectedKey }}</span>
                                        <a-button size="mini" @click="selectKey(st.selectedKey)"><template
                                                #icon><icon-refresh /></template></a-button>
                                        <a-tag size="small" color="arcoblue">TTL {{ formatTTL(st.keyValue.ttl)
                                        }}</a-tag>
                                        <a-button v-if="st.keyValue.type === 'string'" size="mini"
                                            @click="toggleEdit()">
                                            <template #icon><icon-edit /></template>{{ st.editing ? '取消' : '编辑' }}
                                        </a-button>
                                        <a-button v-if="st.editing" size="mini" type="primary" status="success"
                                            @click="saveKey()">
                                            <template #icon><icon-save /></template>保存
                                        </a-button>
                                        <a-button size="mini" @click="copyValue()"><template
                                                #icon><icon-copy /></template>复制值</a-button>
                                        <a-button size="mini" status="danger" @click="delKey(st.selectedKey)"><template
                                                #icon><icon-delete /></template></a-button>
                                    </div>

                                    <div class="kd-body">
                                        <!-- string -->
                                        <template v-if="st.keyValue.type === 'string'">
                                            <a-textarea v-if="st.editing" v-model="st.editStr"
                                                :auto-size="{ minRows: 6 }" />
                                            <div v-else class="kv-lines">
                                                <div v-for="(ln, i) in strLines" :key="i" class="kv-line">
                                                    <span class="ln-no">{{ i + 1 }}</span>
                                                    <span class="ln-txt">{{ ln }}</span>
                                                </div>
                                            </div>
                                        </template>
                                        <!-- hash -->
                                        <a-table v-else-if="st.keyValue.type === 'hash'" :columns="hashCols"
                                            :data="hashData" :pagination="false" size="small" />
                                        <!-- list -->
                                        <a-table v-else-if="st.keyValue.type === 'list'" :columns="listCols"
                                            :data="listData" :pagination="false" size="small" />
                                        <!-- set -->
                                        <a-table v-else-if="st.keyValue.type === 'set'" :columns="setCols"
                                            :data="setData" :pagination="false" size="small" />
                                        <!-- zset -->
                                        <a-table v-else-if="st.keyValue.type === 'zset'" :columns="zsetCols"
                                            :data="zsetData" :pagination="false" size="small" />
                                        <div v-else class="conn-empty">不支持的类型: {{ st.keyValue.type }}</div>
                                    </div>
                                </template>
                            </div>
                        </div>

                        <!-- 慢日志 -->
                        <div v-show="st.activeSubTab === 'slowlog'" class="slowlog-view">
                            <a-button size="mini" type="text" @click="loadSlowlog()"><template
                                    #icon><icon-refresh /></template>刷新</a-button>
                            <a-table :columns="slowCols" :data="slowData" :pagination="false" size="small"
                                :scroll="{ y: 460 }" />
                        </div>

                        <!-- 监控命令 -->
                        <div v-show="st.activeSubTab === 'monitor'" class="monitor-view">
                            <div class="mon-toolbar">
                                <a-button v-if="!st.monitoring" size="mini" type="primary"
                                    @click="startMonitor()"><template
                                        #icon><icon-play-arrow /></template>开始监控</a-button>
                                <a-button v-else size="mini" type="primary" status="warning"
                                    @click="stopMonitor()"><template #icon><icon-pause /></template>停止</a-button>
                                <a-button size="mini" @click="st.monitorLines = []"><template
                                        #icon><icon-delete /></template>清空</a-button>
                                <span class="mon-count">共 {{ st.monitorLines.length }} 条</span>
                            </div>
                            <div class="mon-list">
                                <div v-for="(m, i) in st.monitorLines" :key="i" class="mon-line">
                                    <span class="mon-time">{{ m.time }}</span>
                                    <span class="mon-meta" v-if="m.db || m.client">[{{ m.db }} {{ m.client }}]</span>
                                    <span class="mon-cmd">{{ m.cmd }}</span>
                                </div>
                                <div v-if="st.monitorLines.length === 0" class="conn-empty" style="margin-top: 30px">
                                    点击「开始监控」实时查看命令</div>
                            </div>
                        </div>
                    </div>
                </template>
            </div>
        </div>

        <!-- 新增/编辑连接 -->
        <a-modal v-model:visible="st.connModal" :title="st.connEditing ? '编辑连接' : '添加新连接'" :width="460"
            @ok="saveConnection">
            <a-form :model="connForm" layout="vertical">
                <a-form-item field="name" label="连接名称">
                    <a-input v-model="connForm.name" placeholder="例如 127.0.0.1" />
                </a-form-item>
                <a-row :gutter="12">
                    <a-col :span="16">
                        <a-form-item field="host" label="主机">
                            <a-input v-model="connForm.host" placeholder="127.0.0.1" />
                        </a-form-item>
                    </a-col>
                    <a-col :span="8">
                        <a-form-item field="port" label="端口">
                            <a-input-number v-model="connForm.port" :min="1" :max="65535" />
                        </a-form-item>
                    </a-col>
                </a-row>
                <a-row :gutter="12">
                    <a-col :span="16">
                        <a-form-item field="password" label="密码">
                            <a-input-password v-model="connForm.password" placeholder="无密码留空" />
                        </a-form-item>
                    </a-col>
                    <a-col :span="8">
                        <a-form-item field="db" label="数据库">
                            <a-input-number v-model="connForm.db" :min="0" :max="255" />
                        </a-form-item>
                    </a-col>
                </a-row>
                <a-form-item label="保存位置">
                    <a-radio-group v-model="st.saveTarget" type="button" size="small">
                        <a-radio value="file"><icon-save /> 文件</a-radio>
                        <a-radio value="cache"><icon-storage /> 缓存</a-radio>
                    </a-radio-group>
                    <div style="color: #86909c; font-size: 12px; margin-top: 4px">
                        文件：保存到用户配置目录，跨会话持久；缓存：保存到本地 localStorage。
                    </div>
                </a-form-item>
            </a-form>
        </a-modal>

        <!-- 新增键 -->
        <a-modal v-model:visible="st.addKeyModal" title="新增键 (String)" :width="460" @ok="doAddKey">
            <a-form :model="keyForm" layout="vertical">
                <a-form-item field="key" label="键名"><a-input v-model="keyForm.key" placeholder="key" /></a-form-item>
                <a-form-item field="value" label="值"><a-textarea v-model="keyForm.value"
                        :auto-size="{ minRows: 3 }" /></a-form-item>
                <a-form-item field="ttl" label="过期时间(秒, 0=不过期)"><a-input-number v-model="keyForm.ttl"
                        :min="0" /></a-form-item>
            </a-form>
        </a-modal>
    </layout-body>
</template>

<script setup lang="ts">
import { computed, getCurrentInstance, onMounted, onUnmounted, reactive, ref, watch } from "vue";
import { Message, Modal } from "@arco-design/web-vue";
import { Events } from "@wailsio/runtime";

const GoTools = (getCurrentInstance() as any)?.appContext.config.globalProperties.$GoTools;

const CACHE_KEY = "redis_connections_cache";

interface ConnConfig { id: string; name: string; host: string; port: number; password: string; db: number }
interface OpenTab { id: string; name: string; cfg: ConnConfig }

/** 已保存的连接列表 */
const savedConns = ref<ConnConfig[]>([]);
/** 已打开的连接标签 */
const openTabs = ref<OpenTab[]>([]);

const st = reactive({
    activeTab: "",
    activeSubTab: "status",
    // 状态
    status: null as any,
    // 键
    keys: [] as any[],
    keyFilter: "",
    selectedKey: "",
    keyValue: { type: "", ttl: -1, str: "", hash: {}, list: [], set: [], zset: [] } as any,
    editing: false,
    editStr: "",
    // 慢日志
    slowlog: [] as any[],
    // 监控
    monitoring: false,
    monitorLines: [] as any[],
    // 弹窗
    connModal: false,
    connEditing: false,
    saveTarget: "file",
    addKeyModal: false,
});

const connForm = reactive<ConnConfig>({ id: "", name: "", host: "127.0.0.1", port: 6379, password: "", db: 0 });
const keyForm = reactive({ key: "", value: "", ttl: 0 });

const activeTabName = computed(() => openTabs.value.find((t) => t.id === st.activeTab)?.name || "");
const activeCfg = computed(() => openTabs.value.find((t) => t.id === st.activeTab)?.cfg);

const isOpen = (id: string) => openTabs.value.some((t) => t.id === id);

/* ============================ 连接管理 ============================ */

async function loadConnections() {
    // 优先读文件，其次读缓存
    let list: ConnConfig[] = [];
    try {
        list = (await GoTools.RedisLoadConnections()) || [];
    } catch (e) {
        list = [];
    }
    if (!list || list.length === 0) {
        try {
            const cache = localStorage.getItem(CACHE_KEY);
            if (cache) list = JSON.parse(cache) || [];
        } catch (e) {
            list = [];
        }
    }
    savedConns.value = (list || []).map((c) => ({ ...c }));
}

async function persistConnections() {
    const data = JSON.parse(JSON.stringify(savedConns.value));
    // 缓存始终镜像一份，避免丢失
    localStorage.setItem(CACHE_KEY, JSON.stringify(data));
    if (st.saveTarget === "file") {
        try {
            await GoTools.RedisSaveConnections(data);
        } catch (e) {
            /* ignore */
        }
    }
}

function openAddModal() {
    st.connEditing = false;
    Object.assign(connForm, { id: "", name: "", host: "127.0.0.1", port: 6379, password: "", db: 0 });
    st.connModal = true;
}

function openEditModal(c: ConnConfig) {
    st.connEditing = true;
    Object.assign(connForm, { ...c });
    st.connModal = true;
}

async function saveConnection() {
    if (!connForm.host) {
        Message.warning("请填写主机");
        return;
    }
    const item: ConnConfig = { ...connForm };
    if (!item.name) item.name = `${item.host}:${item.port}`;
    if (st.connEditing) {
        const i = savedConns.value.findIndex((c) => c.id === item.id);
        if (i >= 0) savedConns.value[i] = item;
    } else {
        item.id = "c" + Date.now().toString(36);
        savedConns.value.push(item);
    }
    await persistConnections();
    st.connModal = false;
    Message.success("已保存到" + (st.saveTarget === "file" ? "文件" : "缓存"));
}

function onConnAction(v: string, c: ConnConfig) {
    if (v === "open") openConnection(c);
    else if (v === "edit") openEditModal(c);
    else if (v === "del") confirmDeleteConn(c);
}

function confirmDeleteConn(c: ConnConfig) {
    Modal.warning({
        title: "删除连接",
        content: `确定删除连接「${c.name}」？`,
        hideCancel: false,
        onOk: async () => {
            if (isOpen(c.id)) closeTab(c.id);
            savedConns.value = savedConns.value.filter((x) => x.id !== c.id);
            await persistConnections();
        },
    });
}

async function openConnection(c: ConnConfig) {
    if (isOpen(c.id)) {
        switchTab(c.id);
        return;
    }
    try {
        await GoTools.RedisOpen({ ...c });
    } catch (e: any) {
        Message.error("连接失败: " + (e?.message || e));
        return;
    }
    openTabs.value.push({ id: c.id, name: c.name || `${c.host}:${c.port}`, cfg: { ...c } });
    switchTab(c.id);
}

function switchTab(id: string) {
    if (st.monitoring) stopMonitor();
    st.activeTab = id;
    st.status = null;
    st.keys = [];
    st.selectedKey = "";
    st.slowlog = [];
    st.monitorLines = [];
    loadSubTab();
}

function closeTab(id: string) {
    if (activeTabName.value && st.activeTab === id && st.monitoring) stopMonitor();
    GoTools.RedisClose(id);
    openTabs.value = openTabs.value.filter((t) => t.id !== id);
    if (st.activeTab === id) {
        if (openTabs.value.length > 0) switchTab(openTabs.value[openTabs.value.length - 1].id);
        else st.activeTab = "";
    }
}

/* ============================ 子标签数据 ============================ */

function onSubTabChange() {
    if (st.monitoring && st.activeSubTab !== "monitor") stopMonitor();
    loadSubTab();
}

function loadSubTab() {
    if (!st.activeTab) return;
    if (st.activeSubTab === "status") loadStatus();
    else if (st.activeSubTab === "keys") reloadKeys();
    else if (st.activeSubTab === "slowlog") loadSlowlog();
}

async function loadStatus() {
    try {
        const s = await GoTools.RedisStatus(st.activeTab);
        s.host = activeCfg.value ? `${activeCfg.value.host}:${activeCfg.value.port}` : "";
        st.status = s;
    } catch (e: any) {
        Message.error("获取状态失败: " + (e?.message || e));
    }
}

async function reloadKeys() {
    try {
        let pattern = st.keyFilter.trim();
        if (pattern && !pattern.includes("*")) pattern = "*" + pattern + "*";
        const res = await GoTools.RedisScan(st.activeTab, pattern || "*", 200, 2000);
        st.keys = res.keys || [];
    } catch (e: any) {
        Message.error("扫描键失败: " + (e?.message || e));
    }
}

async function selectKey(key: string) {
    st.selectedKey = key;
    st.editing = false;
    try {
        st.keyValue = await GoTools.RedisGetKey(st.activeTab, key);
    } catch (e: any) {
        Message.error("读取键失败: " + (e?.message || e));
    }
}

function toggleEdit() {
    st.editing = !st.editing;
    if (st.editing) st.editStr = st.keyValue.str || "";
}

async function saveKey() {
    try {
        await GoTools.RedisSetString(st.activeTab, st.selectedKey, st.editStr, st.keyValue.ttl > 0 ? st.keyValue.ttl : 0);
        Message.success("已保存");
        st.editing = false;
        selectKey(st.selectedKey);
    } catch (e: any) {
        Message.error("保存失败: " + (e?.message || e));
    }
}

function delKey(key: string) {
    Modal.warning({
        title: "删除键",
        content: `确定删除「${key}」？`,
        hideCancel: false,
        onOk: async () => {
            await GoTools.RedisDeleteKeys(st.activeTab, [key]);
            if (st.selectedKey === key) {
                st.selectedKey = "";
                st.keyValue = { type: "", ttl: -1, str: "", hash: {}, list: [], set: [], zset: [] };
            }
            reloadKeys();
        },
    });
}

function openAddKey() {
    Object.assign(keyForm, { key: "", value: "", ttl: 0 });
    st.addKeyModal = true;
}

async function doAddKey() {
    if (!keyForm.key) {
        Message.warning("请填写键名");
        return;
    }
    try {
        await GoTools.RedisSetString(st.activeTab, keyForm.key, keyForm.value, keyForm.ttl);
        Message.success("已创建");
        st.addKeyModal = false;
        reloadKeys();
        selectKey(keyForm.key);
    } catch (e: any) {
        Message.error("创建失败: " + (e?.message || e));
    }
}

/* ============================ 慢日志 ============================ */

async function loadSlowlog() {
    try {
        st.slowlog = (await GoTools.RedisSlowlog(st.activeTab, 100)) || [];
    } catch (e: any) {
        Message.error("获取慢日志失败: " + (e?.message || e));
    }
}

const slowCols = [
    { title: "序号", dataIndex: "id", width: 80 },
    { title: "耗时(μs)", dataIndex: "duration", width: 110 },
    { title: "时间", slotName: "time", width: 170 },
    { title: "命令", dataIndex: "command" },
    { title: "客户端", dataIndex: "clientAddr", width: 160 },
];
const slowData = computed(() =>
    st.slowlog.map((e: any) => ({ ...e, time: e.time ? new Date(e.time * 1000).toLocaleString() : "" })),
);

/* ============================ 监控命令 ============================ */

let offMonitor: any = null;

async function startMonitor() {
    if (!activeCfg.value) return;
    try {
        await GoTools.RedisStopMonitor();
        st.monitorLines = [];
        await GoTools.RedisStartMonitor({ ...activeCfg.value });
        st.monitoring = true;
    } catch (e: any) {
        Message.error("启动监控失败: " + (e?.message || e));
    }
}

async function stopMonitor() {
    try {
        await GoTools.RedisStopMonitor();
    } catch (e) {
        /* ignore */
    }
    st.monitoring = false;
}

/* ============================ 键详情表格列 ============================ */

const hashCols = [
    { title: "Field", dataIndex: "field", width: 240, ellipsis: true, tooltip: true },
    { title: "Value", dataIndex: "value" },
];
const hashData = computed(() => Object.entries(st.keyValue.hash || {}).map(([field, value]) => ({ field, value })));

const listCols = [
    { title: "#", dataIndex: "idx", width: 70 },
    { title: "Value", dataIndex: "value" },
];
const listData = computed(() => (st.keyValue.list || []).map((value: string, i: number) => ({ idx: i, value })));

const setCols = [{ title: "Member", dataIndex: "value" }];
const setData = computed(() => (st.keyValue.set || []).map((value: string) => ({ value })));

const zsetCols = [
    { title: "Member", dataIndex: "member" },
    { title: "Score", dataIndex: "score", width: 140 },
];
const zsetData = computed(() => st.keyValue.zset || []);

const strLines = computed<string[]>(() => String(st.keyValue.str ?? "").split("\n"));

/* ============================ 工具函数 ============================ */

function formatTTL(ttl: number): string {
    if (ttl < 0) return "持久";
    const h = Math.floor(ttl / 3600);
    const m = Math.floor((ttl % 3600) / 60);
    const s = ttl % 60;
    const p = (n: number) => n.toString().padStart(2, "0");
    return `${p(h)}:${p(m)}:${p(s)}`;
}

function formatUptime(sec: number): string {
    const d = Math.floor(sec / 86400);
    const h = Math.floor((sec % 86400) / 3600);
    if (d > 0) return `${d} 天 ${h} 小时`;
    if (h > 0) return `${h} 小时`;
    return `${Math.floor(sec / 60)} 分钟`;
}

function copyValue() {
    const v = st.keyValue.type === "string" ? st.keyValue.str : JSON.stringify(st.keyValue[st.keyValue.type] ?? {}, null, 2);
    const input = document.createElement("textarea");
    input.value = v || "";
    document.body.appendChild(input);
    input.select();
    document.execCommand("copy");
    document.body.removeChild(input);
    Message.success("已复制");
}

/* TTL 本地倒计时（仅展示，不改变真实值） */
watch(
    () => st.keyValue.ttl,
    (v) => {
        if (typeof v === "number" && v > 0) {
            // 简单每秒递减，切换键时重置
        }
    },
);

/* ============================ 生命周期 ============================ */

onMounted(async () => {
    await loadConnections();
    offMonitor = Events.On("redis:monitor", (ev: any) => {
        const line = ev?.data ?? ev;
        st.monitorLines.push(line);
        if (st.monitorLines.length > 3000) st.monitorLines.splice(0, st.monitorLines.length - 3000);
    });
});

onUnmounted(() => {
    if (offMonitor) offMonitor();
    if (st.monitoring) stopMonitor();
    openTabs.value.forEach((t) => GoTools.RedisClose(t.id));
});
</script>

<style scoped>
.redis-page {
    display: flex;
    height: 100%;
    min-height: 0;
    background: var(--color-fill-1);
}

/* 左侧连接面板 */
.conn-panel {
    width: 220px;
    flex-shrink: 0;
    border-right: 1px solid var(--color-border-2);
    background: var(--color-bg-2);
    display: flex;
    flex-direction: column;
}

.conn-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 8px 12px;
    border-bottom: 1px solid var(--color-border-2);
}

.conn-title {
    font-weight: 600;
    font-size: 14px;
}

.conn-list {
    flex: 1;
    overflow-y: auto;
    padding: 4px;
}

.conn-item {
    display: flex;
    align-items: center;
    gap: 6px;
    padding: 8px 10px;
    border-radius: 4px;
    cursor: pointer;
    font-size: 13px;
}

.conn-item:hover {
    background: var(--color-fill-3);
}

.conn-item.active {
    background: var(--color-primary-light-1);
}

.conn-name {
    flex: 1;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
}

.conn-more {
    color: var(--color-text-3);
}

.conn-empty {
    color: var(--color-text-3);
    text-align: center;
    font-size: 12px;
    padding: 12px;
}

.led {
    width: 8px;
    height: 8px;
    border-radius: 50%;
    background: #c9cdd4;
    flex-shrink: 0;
}

.led.on {
    background: #00b42a;
    box-shadow: 0 0 4px #00b42a;
}

/* 主区域 */
.main-panel {
    flex: 1;
    min-width: 0;
    display: flex;
    flex-direction: column;
    background: var(--color-bg-1);
}

.empty-state {
    flex: 1;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 12px;
    color: var(--color-text-3);
}

.empty-icon {
    color: var(--color-text-4);
}

.empty-text {
    font-size: 14px;
}

.tab-bar {
    display: flex;
    gap: 2px;
    background: var(--color-fill-2);
    padding: 4px 4px 0;
    border-bottom: 1px solid var(--color-border-2);
}

.tab-item {
    display: flex;
    align-items: center;
    gap: 6px;
    padding: 6px 12px;
    font-size: 13px;
    border-radius: 4px 4px 0 0;
    background: var(--color-fill-3);
    cursor: pointer;
}

.tab-item.active {
    background: var(--color-bg-1);
    color: rgb(var(--red-6));
    font-weight: 600;
}

.tab-close {
    font-size: 12px;
    color: var(--color-text-3);
}

.tab-name {
    max-width: 160px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
}

.sub-tab-bar {
    padding: 8px 12px;
    border-bottom: 1px solid var(--color-border-2);
}

.tab-content {
    flex: 1;
    overflow: auto;
    padding: 12px;
}

/* 状态 */
.status-head {
    display: flex;
    align-items: center;
    gap: 8px;
    margin: 12px 0;
}

.status-addr {
    font-size: 20px;
    font-weight: 700;
}

.status-grid {
    display: grid;
    grid-template-columns: repeat(4, 1fr);
    gap: 16px;
}

.stat-card {
    background: var(--color-fill-1);
    border: 1px solid var(--color-border-2);
    border-radius: 6px;
    padding: 16px;
}

.stat-label {
    color: var(--color-text-3);
    font-size: 13px;
}

.stat-value {
    font-size: 26px;
    font-weight: 700;
    margin-top: 8px;
}

/* 键详情 */
.keys-view {
    display: flex;
    gap: 12px;
    height: 100%;
    min-height: 480px;
}

.key-list-panel {
    width: 300px;
    flex-shrink: 0;
    border: 1px solid var(--color-border-2);
    border-radius: 6px;
    display: flex;
    flex-direction: column;
    overflow: hidden;
}

.key-toolbar {
    display: flex;
    gap: 4px;
    padding: 6px;
    border-bottom: 1px solid var(--color-border-2);
}

.key-list {
    flex: 1;
    overflow-y: auto;
}

.key-item {
    display: flex;
    align-items: center;
    gap: 6px;
    padding: 6px 8px;
    font-size: 13px;
    cursor: pointer;
    border-bottom: 1px solid var(--color-fill-1);
}

.key-item:hover {
    background: var(--color-fill-2);
}

.key-item.active {
    background: var(--color-primary-light-1);
}

.key-name {
    flex: 1;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
}

.key-del {
    color: var(--color-text-4);
    opacity: 0;
}

.key-item:hover .key-del {
    opacity: 1;
}

.type-badge {
    display: inline-block;
    min-width: 18px;
    text-align: center;
    font-size: 11px;
    font-weight: 700;
    color: #fff;
    border-radius: 3px;
    padding: 1px 4px;
    text-transform: uppercase;
}

.type-badge.big {
    font-size: 12px;
}

.t-string {
    background: #722ed1;
}

.t-hash {
    background: #f53f3f;
}

.t-list {
    background: #ff7d00;
}

.t-set {
    background: #00b42a;
}

.t-zset {
    background: #165dff;
}

.t-none {
    background: #86909c;
}

.key-detail-panel {
    flex: 1;
    min-width: 0;
    border: 1px solid var(--color-border-2);
    border-radius: 6px;
    display: flex;
    flex-direction: column;
    overflow: hidden;
}

.kd-toolbar {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 8px;
    border-bottom: 1px solid var(--color-border-2);
    flex-wrap: wrap;
}

.kd-key {
    font-weight: 600;
    max-width: 40%;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
}

.kd-body {
    flex: 1;
    overflow: auto;
    padding: 8px;
}

.kv-lines {
    font-family: Consolas, monospace;
    font-size: 13px;
    line-height: 1.7;
}

.kv-line {
    display: flex;
}

.ln-no {
    width: 36px;
    text-align: right;
    padding-right: 10px;
    color: var(--color-text-4);
    user-select: none;
    flex-shrink: 0;
}

.ln-txt {
    white-space: pre-wrap;
    word-break: break-all;
}

/* 慢日志 */
.slowlog-view {
    padding: 4px;
}

/* 监控 */
.mon-toolbar {
    display: flex;
    align-items: center;
    gap: 8px;
    margin-bottom: 8px;
}

.mon-count {
    color: var(--color-text-3);
    font-size: 12px;
}

.mon-list {
    border: 1px solid var(--color-border-2);
    border-radius: 6px;
    background: #1d1e1f;
    color: #e5e6eb;
    font-family: Consolas, monospace;
    font-size: 12px;
    height: 480px;
    overflow-y: auto;
    padding: 8px;
}

.mon-line {
    display: flex;
    gap: 8px;
    line-height: 1.6;
    white-space: pre-wrap;
    word-break: break-all;
}

.mon-time {
    color: #86909c;
    flex-shrink: 0;
}

.mon-meta {
    color: #ffb400;
    flex-shrink: 0;
}

.mon-cmd {
    color: #4ac7ff;
}
</style>
