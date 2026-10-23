<template>
    <a-row class="npm-page" :gutter="6">
        <!-- 左侧：静态站点实例 + 功能入口 -->
        <a-col :span="7" class="side-col">
            <a-card class="side-card" size="small" :body-style="{ padding: 0 }">
                <template #title>静态站点服务</template>
                <template #extra>
                    <div class="srv-head-actions">
                        <a-button size="small" title="新建站点" @click="openInstModal()">新建站点</a-button>
                        <a-button size="small" title="Node 版本管理" @click="openVersionsModal()">版本管理</a-button>
                    </div>
                </template>
                <a-list :bordered="false" size="small">
                    <a-list-item v-for="inst in instances" :key="inst.id"
                        :class="{ 'inst-active': selectedId === inst.id }" @click="selectInst(inst.id)">
                        <a-space>
                            <a-badge :status="states[inst.id]?.running ? 'success' : 'normal'" />
                            <span class="conn-name" :title="inst.name + ':' + inst.port">{{ inst.name }}</span>
                            <a-tag color="arcoblue" size="small">node {{ inst.version }}</a-tag>
                        </a-space>
                        <template #actions>
                            <a-dropdown trigger="hover" @select="(v: any) => onInstAction(v, inst)">
                                <icon-more class="conn-more" @click.stop />
                                <template #content>
                                    <a-doption v-if="!states[inst.id]?.running" value="start"><icon-play-arrow />
                                        启动</a-doption>
                                    <a-doption v-else value="stop"><icon-pause /> 停止</a-doption>
                                    <a-doption value="restart"><icon-refresh /> 重启</a-doption>
                                    <a-doption value="open"><icon-launch /> 打开站点</a-doption>
                                    <a-doption value="logs"><icon-face-folder-open /> 日志</a-doption>
                                    <a-doption value="edit"><icon-edit /> 编辑</a-doption>
                                    <a-doption value="del"><icon-delete /> 删除</a-doption>
                                </template>
                            </a-dropdown>
                        </template>
                    </a-list-item>
                    <a-list-item v-if="instances.length === 0">
                        <div class="conn-empty">暂无站点，点击右上"新建站点"</div>
                    </a-list-item>
                </a-list>
            </a-card>
            <a-card class="side-card tools-card" size="small" :body-style="{ padding: '10px' }">
                <template #title>npm 工具箱</template>
                <div class="toolbox">
                    <a-button long @click="openNpmConfig()">
                        <template #icon><icon-settings /></template>npm 配置 / 镜像源
                    </a-button>
                    <a-button long @click="openPkgs()">
                        <template #icon><icon-storage /></template>全局包管理
                    </a-button>
                </div>
                <div class="reg-hint">当前 registry：{{ registryVal || '-' }}</div>
            </a-card>
        </a-col>

        <!-- 右侧：选中站点详情 + 日志 -->
        <a-col :span="17" class="main-col">
            <a-card v-if="!selectedId" class="main-panel empty-card">
                <a-empty description="从左侧选择一个静态站点查看状态与访问日志" />
            </a-card>
            <a-card v-else class="main-panel" :body-style="{ padding: 0 }">
                <template #title>
                    <a-space>
                        <a-badge :status="selState?.running ? 'success' : 'normal'" />
                        <span>{{ selCfg?.name }}</span>
                        <a-tag color="arcoblue" size="small">:{{ selCfg?.port }}</a-tag>
                        <a-tag v-if="selCfg?.spa" size="small">SPA</a-tag>
                        <a-tag v-if="selCfg?.directory" size="small">列目录</a-tag>
                    </a-space>
                </template>
                <template #extra>
                    <a-space>
                        <a-button size="small" type="primary" :disabled="selState?.running"
                            @click="doStart">启动</a-button>
                        <a-button size="small" status="danger" :disabled="!selState?.running"
                            @click="doStop">停止</a-button>
                        <a-button size="small" @click="doRestart">重启</a-button>
                        <a-button size="small" :disabled="!selState?.running" @click="openSite(selCfg?.port)">在新窗口打开</a-button>
                    </a-space>
                </template>
                <div class="detail-grid">
                    <div class="ls-item"><span>运行状态</span>
                        <b :class="selState?.running ? 'ok' : 'no'">{{ selState?.running ? '运行中' : '已停止' }}</b>
                    </div>
                    <div class="ls-item"><span>运行时长</span><b>{{ fmtUptimeS(selState?.uptimeSeconds) }}</b></div>
                    <div class="ls-item"><span>Node 版本</span><b>{{ selState?.version || selCfg?.version || '-' }}</b>
                    </div>
                    <div class="ls-item"><span>npm 版本</span><b>{{ selState?.npmVer || '-' }}</b></div>
                    <div class="ls-item"><span>累计请求</span><b>{{ selState?.requests ?? '-' }}</b></div>
                    <div class="ls-item"><span>req/s</span><b>{{ selState?.requestsPerSec ? selState.requestsPerSec.toFixed(2) : '-' }}</b></div>
                    <div class="ls-item"><span>累计发送</span><b>{{ fmtBytes(selState?.bytes) }}</b></div>
                    <div class="ls-item"><span>站点文件数</span><b>{{ selState?.files ?? '-' }}</b></div>
                    <div class="ls-item"><span>registry</span><b class="ellipsis">{{ selState?.registry || '-' }}</b></div>
                    <div class="ls-item" title="站点根目录"><span>站点目录</span><b class="ellipsis">{{
                        selState?.docRoot || '-' }}</b></div>
                    <div class="ls-item" title="node.exe 路径"><span>node.exe</span><b class="ellipsis">{{
                        selState?.nodePath || '-' }}</b></div>
                </div>
                <div class="log-head">
                    <span>访问日志（实时）</span>
                    <a-button size="mini" @click="logLines = []">
                        <template #icon><icon-delete /></template>清空
                    </a-button>
                </div>
                <div ref="logBoxRef" class="ls-log">
                    <div v-for="(l, i) in logLines" :key="i" class="ls-line">{{ l }}</div>
                    <div v-if="logLines.length === 0" class="conn-empty">暂无日志</div>
                </div>
            </a-card>
        </a-col>
    </a-row>

    <!-- 新建/编辑站点实例 -->
    <a-modal v-model:visible="instVisible" :title="instForm.id ? '编辑站点' : '新建站点'" :width="520"
        @before-ok="saveInstance">
        <a-form :model="instForm" layout="vertical">
            <a-form-item field="name" label="站点名称" required>
                <a-input v-model="instForm.name" placeholder="如 docs-site" />
            </a-form-item>
            <a-form-item field="version" label="Node.js 版本（用于展示与配套 npm）" required>
                <a-select v-model="instForm.version" placeholder="选择已安装版本" :options="versionOptions" />
            </a-form-item>
            <a-form-item field="port" label="端口">
                <a-input-number v-model="instForm.port" :min="1" :max="65535" @blur="checkPortTaken" />
            </a-form-item>
            <a-form-item field="docRoot" label="站点根目录（空 = 实例目录 www/）">
                <a-input-search v-model="instForm.docRoot" placeholder="如 D:\www\dist" search-button
                    button-text="选择目录" @click-btn="chooseDocRoot" />
            </a-form-item>
            <a-form-item field="spa" label="SPA 模式（找不到文件回退 index.html）">
                <a-switch v-model="instForm.spa" />
            </a-form-item>
            <a-form-item field="directory" label="允许列目录（无 index.html 时）">
                <a-switch v-model="instForm.directory" />
            </a-form-item>
            <a-form-item field="openBrowser" label="启动成功后自动打开浏览器">
                <a-switch v-model="instForm.openBrowser" />
            </a-form-item>
        </a-form>
    </a-modal>

    <!-- Node 版本管理 -->
    <a-modal v-model:visible="versionsVisible" title="Node.js 版本管理" :width="780" :footer="false"
        unmount-on-close>
        <div class="ver-block">
            <div class="ver-title">本地可用版本</div>
            <a-table :data="localVersions" size="small" :pagination="false" row-key="version"
                :columns="[
                    { title: '版本', dataIndex: 'version', width: 110 },
                    { title: 'npm', dataIndex: 'npmVer', width: 90 },
                    { title: '来源', dataIndex: 'source', width: 90, slotName: 'source' },
                    { title: '路径', dataIndex: 'path', ellipsis: true, tooltip: true },
                    { title: '操作', dataIndex: 'ops', width: 90, slotName: 'ops' },
                ]">
                <template #source="{ record }">
                    <a-tag size="small" :color="record.source === 'bundled' ? 'purple' : record.source === 'downloaded' ? 'green' : 'orange'">{{ srcLabel(record.source) }}</a-tag>
                </template>
                <template #ops="{ record }">
                    <a-button size="mini" status="danger" :disabled="record.source === 'bundled'"
                        @click="removeVersion(record)">删除</a-button>
                </template>
            </a-table>
            <a-row class="add-local" :gutter="8">
                <a-col :span="16">
                    <a-input v-model="localDirInput" placeholder="注册本机已有 Node.js：填写 node.exe 所在目录" />
                </a-col>
                <a-col :span="4">
                    <a-button block @click="chooseLocalDir">浏览…</a-button>
                </a-col>
                <a-col :span="4">
                    <a-button type="primary" block @click="addLocalDir">添加</a-button>
                </a-col>
            </a-row>
        </div>
        <a-divider style="margin: 12px 0" />
        <div class="ver-block">
            <div class="ver-title">
                线上版本
                <a-button size="mini" :loading="remoteLoading" @click="loadRemoteVersions">刷新</a-button>
            </div>
            <a-spin :loading="remoteLoading" style="width:100%">
                <a-table :data="remoteVersions" size="small" :pagination="{ pageSize: 8 }" row-key="version"
                    :columns="[
                        { title: '版本', dataIndex: 'version', width: 110 },
                        { title: 'LTS', dataIndex: 'lts', width: 90, slotName: 'lts' },
                        { title: '包大小', dataIndex: 'size', width: 100, slotName: 'size' },
                        { title: '状态', dataIndex: 'st', width: 150, slotName: 'st' },
                        { title: '操作', dataIndex: 'ops', width: 100, slotName: 'ops' },
                    ]">
                    <template #lts="{ record }">
                        <a-tag v-if="record.lts" color="green" size="small">{{ record.lts }}</a-tag>
                    </template>
                    <template #size="{ record }">{{ fmtBytes(record.size) }}</template>
                    <template #st="{ record }">
                        <a-tag v-if="installedSet.has(record.version)" color="green" size="small">已安装</a-tag>
                        <a-tag v-else-if="installProgress[record.version]" :color="installProgress[record.version].stage === 'error' ? 'red' : 'arcoblue'" size="small">
                            {{ stageText(installProgress[record.version]) }}
                        </a-tag>
                        <a-tag v-else size="small" color="gray">未安装</a-tag>
                    </template>
                    <template #ops="{ record }">
                        <a-button v-if="!installedSet.has(record.version)" size="mini" type="primary"
                            :disabled="!!installProgress[record.version]" @click="installVersion(record)">安装</a-button>
                        <a-button v-else size="mini" @click="loadLocalVersions">已安装</a-button>
                    </template>
                </a-table>
            </a-spin>
            <div class="offline-tip">离线安装：将 node zip 放入 <a-link copyable :content="pluginsDirVal">{{ pluginsDirVal || '…' }}</a-link> 后刷新本地版本</div>
        </div>
        <div v-if="installLogs.length" class="ver-block install-log">
            <div class="ver-title">任务日志</div>
            <div class="ls-log small">
                <div v-for="(l, i) in installLogs" :key="i" class="ls-line">{{ l }}</div>
            </div>
        </div>
    </a-modal>

    <!-- npm 配置 -->
    <a-modal v-model:visible="cfgVisible" title="npm 配置（托管 .npmrc）" :width="720" :footer="false"
        unmount-on-close>
        <div class="cfg-registry">
            <div class="ver-title">镜像源快捷切换</div>
            <a-space wrap>
                <a-button v-for="r in commonRegistries" :key="r.value" size="small"
                    :type="registryVal === r.value ? 'primary' : 'outline'" @click="applyRegistry(r.value)">
                    {{ r.key }}
                </a-button>
            </a-space>
            <a-input-search v-model="registryInput" placeholder="自定义 registry URL" size="small"
                style="margin-top:8px;width:420px" search-button button-text="应用" @search="applyRegistry(registryInput)" />
        </div>
        <a-divider style="margin: 12px 0" />
        <div class="ver-title">
            配置项
            <a-button size="mini" @click="loadConfigItems">刷新</a-button>
        </div>
        <a-table :data="configItems" size="small" :pagination="false" row-key="key"
            :columns="[
                { title: 'Key', dataIndex: 'key', width: 220 },
                { title: 'Value', dataIndex: 'value', slotName: 'value' },
            ]">
            <template #value="{ record }">
                <a-input v-model="record.value" size="mini" @press-enter="setItem(record)"
                    @blur="setItem(record)" />
            </template>
        </a-table>
        <a-row class="add-local" :gutter="8">
            <a-col :span="8"><a-input v-model="newKey" size="small" placeholder="新增 Key" /></a-col>
            <a-col :span="12"><a-input v-model="newVal" size="small" placeholder="Value" /></a-col>
            <a-col :span="4"><a-button size="small" type="primary" block @click="addConfigItem">新增</a-button></a-col>
        </a-row>
        <a-divider style="margin: 12px 0" />
        <a-collapse>
            <a-collapse-item header=".npmrc 原文编辑" key="raw">
                <a-textarea v-model="npmrcText" :auto-size="{ minRows: 8, maxRows: 14 }" spellcheck="false" />
                <a-space style="margin-top:8px">
                    <a-button size="small" @click="loadNpmrcText">重新读取</a-button>
                    <a-button size="small" type="primary" @click="saveNpmrcText">保存 .npmrc</a-button>
                </a-space>
            </a-collapse-item>
        </a-collapse>
        <div class="user-npmrc">
            <div class="ver-title">用户全局 ~/.npmrc</div>
            <div class="offline-tip">{{ userNpmrcPath || '（未创建）' }}
                <a-button size="mini" style="margin-left:8px" @click="applyToUser">把托管配置同步过去（自动备份）</a-button>
            </div>
        </div>
    </a-modal>

    <!-- 全局包管理 -->
    <a-modal v-model:visible="pkgsVisible" title="全局包管理" :width="760" :footer="false" unmount-on-close>
        <a-row :gutter="8" align="center" style="margin-bottom:10px">
            <a-col :span="7">
                <a-select v-model="pkgNodeVer" :options="versionOptions" size="small" placeholder="Node 版本"
                    @change="loadPkgs" />
            </a-col>
            <a-col :span="12">
                <a-input v-model="pkgInstallName" size="small" placeholder="要全局安装的包名，如 axios / pm2"
                    @press-enter="installPkg" />
            </a-col>
            <a-col :span="5">
                <a-space>
                    <a-button size="small" type="primary" :disabled="!!pkgBusy" @click="installPkg">安装</a-button>
                    <a-button size="small" :loading="pkgsLoading" @click="loadPkgs">刷新</a-button>
                </a-space>
            </a-col>
        </a-row>
        <a-spin :loading="pkgsLoading" style="width:100%">
            <a-table :data="packages" size="small" :pagination="false" row-key="name"
                :columns="[
                    { title: '包名', dataIndex: 'name', width: 220 },
                    { title: '版本', dataIndex: 'version', width: 120 },
                    { title: '说明', dataIndex: 'description', ellipsis: true, tooltip: true },
                    { title: '操作', dataIndex: 'ops', width: 80, slotName: 'ops' },
                ]">
                <template #ops="{ record }">
                    <a-button size="mini" status="danger" :disabled="!!pkgBusy"
                        @click="uninstallPkg(record)">卸载</a-button>
                </template>
            </a-table>
            <div v-if="!pkgsLoading && packages.length === 0" class="conn-empty">暂无全局包（或先选择一个 Node
                版本）</div>
        </a-spin>
        <div v-if="pkgBusy" class="offline-tip">{{ pkgBusy }} …（进度见任务日志）</div>
        <div v-if="installLogs.length" class="ver-block install-log">
            <div class="ver-title">任务日志</div>
            <div class="ls-log small">
                <div v-for="(l, i) in installLogs" :key="i" class="ls-line">{{ l }}</div>
            </div>
        </div>
    </a-modal>
</template>

<script setup lang="ts">
import { computed, getCurrentInstance, nextTick, onMounted, onUnmounted, reactive, ref } from "vue";
import { Message, Modal } from "@arco-design/web-vue";
import { Events } from "@wailsio/runtime";

const GoTools = (getCurrentInstance() as any)?.appContext.config.globalProperties.$GoTools;

interface InstConfig { id: string; name: string; version: string; port: number; docRoot: string; spa: boolean; directory: boolean; openBrowser: boolean }

/* ============================ 状态 ============================ */
const instances = ref<InstConfig[]>([]);
const states = reactive<Record<string, any>>({});
const selectedId = ref("");
const logLines = ref<string[]>([]);
const logBoxRef = ref<HTMLElement | null>(null);
let logTimer: any = null;
let srvTimer: any = null;
let offInstall: any = null;

const selCfg = computed(() => instances.value.find((i) => i.id === selectedId.value) || null);
const selState = computed(() => (selectedId.value ? states[selectedId.value] : null));

const instForm = reactive<InstConfig>({
    id: "", name: "", version: "", port: 3000, docRoot: "", spa: true, directory: false, openBrowser: false,
});
const instVisible = ref(false);

const versionsVisible = ref(false);
const localVersions = ref<any[]>([]);
const remoteVersions = ref<any[]>([]);
const remoteLoading = ref(false);
const installProgress = reactive<Record<string, any>>({});
const installLogs = ref<string[]>([]);
const localDirInput = ref("");
const pluginsDirVal = ref("");

const cfgVisible = ref(false);
const configItems = ref<any[]>([]);
const commonRegistries = ref<any[]>([]);
const registryVal = ref("");
const registryInput = ref("");
const newKey = ref("");
const newVal = ref("");
const npmrcText = ref("");
const userNpmrcPath = ref("");

const pkgsVisible = ref(false);
const pkgNodeVer = ref("");
const packages = ref<any[]>([]);
const pkgsLoading = ref(false);
const pkgInstallName = ref("");
const pkgBusy = ref("");

const versionOptions = computed(() => localVersions.value.map((v: any) => ({ label: `${v.version}（${srcLabel(v.source)}）`, value: v.version })));
const installedSet = computed(() => new Set<string>(localVersions.value.map((v: any) => v.version)));

/* ============================ 小工具 ============================ */
function srcLabel(s: string): string {
    return s === "bundled" ? "内置" : s === "downloaded" ? "下载" : "本机";
}
function stageText(p: any): string {
    if (!p) return "";
    if (p.stage === "downloading") return `下载中 ${p.percent || 0}%`;
    if (p.stage === "extracting") return "解压中";
    if (p.stage === "done") return "完成";
    if (p.stage === "info") return "执行中";
    return "失败";
}
function fmtBytes(b?: number): string {
    if (!b) return "-";
    const units = ["B", "KB", "MB", "GB"]; let i = 0; let v = b;
    while (v >= 1024 && i < units.length - 1) { v /= 1024; i++; }
    return v.toFixed(i > 0 ? 1 : 0) + " " + units[i];
}
function fmtUptimeS(sec?: number): string {
    if (!sec || sec <= 0) return "-";
    const d = Math.floor(sec / 86400); const h2 = Math.floor((sec % 86400) / 3600);
    const m = Math.floor((sec % 3600) / 60);
    if (d > 0) return `${d}d ${h2}h ${m}m`; if (h2 > 0) return `${h2}h ${m}m`; return `${m}m`;
}
function pushInstallLog(txt: string) {
    if (!txt) return;
    const t = new Date();
    const pad = (n: number) => n.toString().padStart(2, "0");
    installLogs.value.push(`[${pad(t.getHours())}:${pad(t.getMinutes())}:${pad(t.getSeconds())}] ${txt}`);
    if (installLogs.value.length > 200) installLogs.value.splice(0, installLogs.value.length - 200);
}
function openSite(port?: number) {
    if (!port) return;
    window.open(`http://127.0.0.1:${port}/`, "_blank");
}

/* ============================ 站点实例 ============================ */
async function loadLocalVersions() {
    try { localVersions.value = (await GoTools.NpmSrvListVersions()) || []; }
    catch (e) { localVersions.value = []; }
}
async function loadInstances() {
    try { instances.value = (await GoTools.NpmSrvListInstances()) || []; }
    catch (e) { instances.value = []; }
    if (!selectedId.value && instances.value.length) selectInst(instances.value[0].id);
    refreshStates();
}
async function refreshStates() {
    for (const inst of instances.value) {
        try {
            states[inst.id] = await GoTools.NpmSrvStatus(inst.id);
            if (states[inst.id]?.registry) registryVal.value = states[inst.id].registry;
        } catch (e) { /* ignore */ }
    }
}
function selectInst(id: string) {
    selectedId.value = id;
    logLines.value = [];
    startLogPolling();
    refreshStates();
}
async function pollLogs() {
    if (!selectedId.value) return;
    try {
        const lines = (await GoTools.NpmSrvLogs(selectedId.value, 300)) || [];
        if (lines.join("\n") !== logLines.value.join("\n")) {
            logLines.value = lines;
            nextTick(() => { if (logBoxRef.value) logBoxRef.value.scrollTop = logBoxRef.value.scrollHeight; });
        }
    } catch (e) { /* ignore */ }
}
function startLogPolling() {
    if (logTimer) clearInterval(logTimer);
    pollLogs();
    logTimer = setInterval(pollLogs, 2000);
}
async function doStart() {
    const c = selCfg.value; if (!c) return;
    try { await GoTools.NpmSrvStart(c.id); Message.success(`站点 ${c.name} 已启动`); }
    catch (e: any) { Message.error("启动失败: " + (e?.message || e)); }
    refreshStates();
}
async function doStop() {
    const c = selCfg.value; if (!c) return;
    try { await GoTools.NpmSrvStop(c.id); Message.success(`站点 ${c.name} 已停止`); }
    catch (e: any) { Message.error("停止失败: " + (e?.message || e)); }
    refreshStates();
}
async function doRestart() {
    const c = selCfg.value; if (!c) return;
    try { await GoTools.NpmSrvRestart(c.id); Message.success(`站点 ${c.name} 已重启`); }
    catch (e: any) { Message.error("重启失败: " + (e?.message || e)); }
    refreshStates();
}
async function onInstAction(v: string, inst: InstConfig) {
    if (v === "start") {
        try { await GoTools.NpmSrvStart(inst.id); Message.success(`站点 ${inst.name} 已启动`); }
        catch (e: any) { Message.error("启动失败: " + (e?.message || e)); }
        refreshStates();
    } else if (v === "stop") {
        try { await GoTools.NpmSrvStop(inst.id); Message.success(`站点 ${inst.name} 已停止`); }
        catch (e: any) { Message.error("停止失败: " + (e?.message || e)); }
        refreshStates();
    } else if (v === "restart") {
        try { await GoTools.NpmSrvRestart(inst.id); Message.success(`站点 ${inst.name} 已重启`); }
        catch (e: any) { Message.error("重启失败: " + (e?.message || e)); }
        refreshStates();
    } else if (v === "open") {
        openSite(inst.port);
    } else if (v === "logs") {
        selectInst(inst.id);
    } else if (v === "edit") {
        openInstModal(inst);
    } else if (v === "del") {
        Modal.warning({
            title: "删除站点", content: `确定删除站点「${inst.name}」？运行中会先停止；站点目录保留。`, hideCancel: false,
            onOk: async () => {
                try {
                    await GoTools.NpmSrvRemoveInstance(inst.id);
                    delete states[inst.id];
                    if (selectedId.value === inst.id) selectedId.value = "";
                    await loadInstances();
                } catch (e: any) { Message.error("删除失败: " + (e?.message || e)); }
            },
        });
    }
}
function openInstModal(inst?: InstConfig) {
    if (inst) { Object.assign(instForm, { ...inst }); }
    else {
        Object.assign(instForm, { id: "", name: "", version: localVersions.value[0]?.version || "", port: 3000, docRoot: "", spa: true, directory: false, openBrowser: false });
    }
    if (localVersions.value.length === 0) loadLocalVersions();
    instVisible.value = true;
}
async function checkPortTaken() {
    if (!instForm.port) return;
    try {
        const suggest = await GoTools.NpmSrvSuggestPort(instForm.port);
        if (suggest && suggest !== instForm.port) Message.warning(`端口 ${instForm.port} 已被占用，建议改为 ${suggest}`);
    } catch (e) { /* ignore */ }
}
async function chooseDocRoot() {
    try { const dir = await GoTools.NpmSrvChooseDir(); if (dir) instForm.docRoot = dir; }
    catch (e) { /* cancel */ }
}
async function saveInstance(): Promise<boolean> {
    if (!instForm.name) { Message.warning("请填写站点名称"); return false; }
    if (!instForm.version) { Message.warning("请选择 Node.js 版本"); return false; }
    try {
        await GoTools.NpmSrvSaveInstance({ ...instForm });
        Message.success("已保存"); await loadInstances(); return true;
    } catch (e: any) { Message.error("保存失败: " + (e?.message || e)); return false; }
}

/* ============================ Node 版本管理 ============================ */
async function openVersionsModal() {
    versionsVisible.value = true;
    loadLocalVersions();
    if (!pluginsDirVal.value) {
        try { pluginsDirVal.value = (await GoTools.NpmSrvPluginsDir()) || ""; } catch (e) { /* ignore */ }
    }
    if (remoteVersions.value.length === 0) loadRemoteVersions();
}
async function loadRemoteVersions() {
    remoteLoading.value = true;
    try { remoteVersions.value = (await GoTools.NpmSrvRemoteVersions()) || []; }
    catch (e: any) { Message.error("获取远程版本失败: " + (e?.message || e)); }
    remoteLoading.value = false;
}
async function installVersion(r: any) {
    try {
        pushInstallLog(`提交安装请求：Node.js ${r.version}`);
        await GoTools.NpmSrvInstall(r.version);
        installProgress[r.version] = { stage: "downloading", percent: 0 };
    } catch (e: any) {
        pushInstallLog(`安装请求被拒绝：${e?.message || e}`);
        Message.error("安装失败: " + (e?.message || e));
    }
}
function removeVersion(v: any) {
    Modal.warning({
        title: "删除版本", content: `确定删除 Node.js ${v.version}？`, hideCancel: false,
        onOk: async () => {
            try { await GoTools.NpmSrvRemoveVersion(v.version); Message.success("已删除"); loadLocalVersions(); }
            catch (e: any) { Message.error("删除失败: " + (e?.message || e)); }
        },
    });
}
async function chooseLocalDir() {
    try { const dir = await GoTools.NpmSrvChooseDir(); if (dir) localDirInput.value = dir; }
    catch (e) { /* cancel */ }
}
async function addLocalDir() {
    const dir = localDirInput.value.trim();
    if (!dir) { Message.warning("请先选择或填写目录"); return; }
    try { await GoTools.NpmSrvAddLocal(dir); Message.success("已添加本机目录"); localDirInput.value = ""; loadLocalVersions(); }
    catch (e: any) { Message.error("添加失败: " + (e?.message || e)); }
}

/* ============================ npm 配置 ============================ */
async function openNpmConfig() {
    cfgVisible.value = true;
    loadConfigItems();
    loadNpmrcText();
    try { commonRegistries.value = (await GoTools.NpmCommonRegistries()) || []; } catch (e) { /* ignore */ }
    try { registryVal.value = (await GoTools.NpmGetRegistry()) || ""; } catch (e) { /* ignore */ }
    registryInput.value = registryVal.value;
    try {
        const res = await GoTools.NpmUserNpmrc();
        // bindings 多返回值：可能是数组或对象，兼容取第一个
        const p = Array.isArray(res) ? res[0] : (res?.path || res?.result1 || "");
        userNpmrcPath.value = p || "";
    } catch (e) { /* ignore */ }
}
async function loadConfigItems() {
    try { configItems.value = ((await GoTools.NpmConfigItems()) || []).map((x: any) => ({ ...x })); }
    catch (e: any) { Message.error("读取 npm 配置失败: " + (e?.message || e)); }
}
async function loadNpmrcText() {
    try { npmrcText.value = (await GoTools.NpmGetConfigText()) || ""; } catch (e) { npmrcText.value = ""; }
}
async function saveNpmrcText() {
    try { await GoTools.NpmSaveConfigText(npmrcText.value); Message.success("托管 .npmrc 已保存"); loadConfigItems(); }
    catch (e: any) { Message.error("保存失败: " + (e?.message || e)); }
}
async function setItem(rec: any) {
    if (!rec.key) return;
    try {
        await GoTools.NpmSetConfigItem(rec.key, rec.value);
        if (rec.key === "registry") registryVal.value = rec.value;
    } catch (e: any) { Message.error("设置失败: " + (e?.message || e)); }
}
async function addConfigItem() {
    const k = newKey.value.trim();
    if (!k) { Message.warning("请填写 Key"); return; }
    await setItem({ key: k, value: newVal.value.trim() });
    newKey.value = ""; newVal.value = "";
    loadConfigItems();
}
async function applyRegistry(url: string) {
    url = String(url || "").trim();
    if (!url) return;
    try {
        await GoTools.NpmSetRegistry(url);
        registryVal.value = url;
        Message.success("registry 已切换：" + url);
        loadConfigItems();
    } catch (e: any) { Message.error("切换失败: " + (e?.message || e)); }
}
async function applyToUser() {
    Modal.warning({
        title: "同步到用户全局", hideCancel: false,
        content: "将托管 .npmrc 的 registry 等配置写入用户全局 ~/.npmrc（写入前自动备份原文件）。确定继续？",
        onOk: async () => {
            try {
                const bak = await GoTools.NpmApplyToUserNpmrc();
                Message.success("已同步" + (bak ? "，原文件备份至 " + bak : ""));
            } catch (e: any) { Message.error("同步失败: " + (e?.message || e)); }
        },
    });
}

/* ============================ 全局包管理 ============================ */
async function openPkgs() {
    pkgsVisible.value = true;
    if (!pkgNodeVer.value) pkgNodeVer.value = localVersions.value[0]?.version || "";
    if (localVersions.value.length === 0) loadLocalVersions().then(() => loadPkgs());
    else loadPkgs();
}
async function loadPkgs() {
    if (!pkgNodeVer.value) { packages.value = []; return; }
    pkgsLoading.value = true;
    try { packages.value = (await GoTools.NpmGlobalPackages(pkgNodeVer.value)) || []; }
    catch (e: any) { packages.value = []; Message.error("读取全局包失败: " + String(e?.message || e).slice(0, 200)); }
    pkgsLoading.value = false;
}
async function installPkg() {
    const pkg = pkgInstallName.value.trim();
    if (!pkg) { Message.warning("请填写包名"); return; }
    if (!pkgNodeVer.value) { Message.warning("请选择 Node 版本"); return; }
    try {
        pkgBusy.value = `npm install -g ${pkg}`;
        await GoTools.NpmInstallGlobal(pkgNodeVer.value, pkg);
        pkgInstallName.value = "";
    } catch (e: any) {
        pkgBusy.value = "";
        Message.error("安装失败: " + (e?.message || e));
    }
}
function uninstallPkg(rec: any) {
    Modal.warning({
        title: "卸载全局包", hideCancel: false,
        content: `确定全局卸载 ${rec.name}@${rec.version}？`,
        onOk: async () => {
            try {
                pkgBusy.value = `npm uninstall -g ${rec.name}`;
                await GoTools.NpmUninstallGlobal(pkgNodeVer.value, rec.name);
            } catch (e: any) {
                pkgBusy.value = "";
                Message.error("卸载失败: " + (e?.message || e));
            }
        },
    });
}

/* ============================ 生命周期 ============================ */
onMounted(async () => {
    await loadInstances();
    loadLocalVersions();
    srvTimer = setInterval(refreshStates, 3000);
    offInstall = Events.On("npm:install", (ev: any) => {
        const p = ev?.data ?? ev;
        if (!p || !p.version) return;
        if (p.stage === "info") {
            pushInstallLog(p.note || "");
        } else if (p.stage === "done") {
            delete installProgress[p.version];
            pkgBusy.value = "";
            pushInstallLog(`完成：${p.note || p.version}`);
            Message.success(`${p.version} 任务完成`);
            loadLocalVersions();
            if (pkgsVisible.value) loadPkgs();
        } else if (p.stage === "error") {
            if (installProgress[p.version]) installProgress[p.version] = p;
            pkgBusy.value = "";
            pushInstallLog(`失败 ${p.version}：${p.note || "未知错误"}`);
            Modal.error({ title: "任务失败", content: `${p.version}：${p.note || "未知错误"}` });
        } else {
            installProgress[p.version] = p;
        }
    });
});
onUnmounted(() => {
    if (offInstall) offInstall();
    if (srvTimer) clearInterval(srvTimer);
    if (logTimer) clearInterval(logTimer);
});
</script>

<style scoped>
.npm-page {
    height: 100%;
    min-height: 0;
    flex-wrap: nowrap;
    background: var(--color-fill-1);
}

.side-col,
.main-col {
    height: 100%;
    min-width: 0;
    display: flex;
    flex-direction: column;
    gap: 6px;
}

.side-card {
    display: flex;
    flex-direction: column;
}

.side-card:first-child {
    flex: 1;
    min-height: 0;
}

.side-card:first-child :deep(.arco-card-body) {
    flex: 1;
    overflow: auto;
}

.tools-card {
    flex-shrink: 0;
}

.toolbox {
    display: flex;
    flex-direction: column;
    gap: 8px;
}

.reg-hint {
    margin-top: 8px;
    font-size: 12px;
    color: var(--color-text-3);
    word-break: break-all;
}

.main-panel {
    height: 100%;
    display: flex;
    flex-direction: column;
}

.main-panel :deep(.arco-card-body) {
    flex: 1;
    overflow: auto;
    display: flex;
    flex-direction: column;
}

.empty-card :deep(.arco-card-body) {
    align-items: center;
    justify-content: center;
}

.srv-head-actions {
    display: flex;
    gap: 6px;
}

.conn-name {
    max-width: 130px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
}

.conn-more {
    cursor: pointer;
    color: var(--color-text-3);
}

.conn-empty {
    color: var(--color-text-3);
    font-size: 12px;
    padding: 6px 0;
}

.inst-active {
    background: var(--color-primary-light-1);
    cursor: pointer;
}

:deep(.arco-list-item) {
    cursor: pointer;
}

.detail-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(240px, 1fr));
    gap: 6px 16px;
    padding: 10px 14px;
    border-bottom: 1px solid var(--color-border-2);
}

.ls-item {
    display: flex;
    justify-content: space-between;
    font-size: 12px;
    gap: 8px;
}

.ls-item span {
    color: var(--color-text-3);
    flex-shrink: 0;
}

.ls-item b {
    font-weight: 500;
    text-align: right;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
}

.ls-item b.ok {
    color: rgb(var(--green-6));
}

.ls-item b.no {
    color: var(--color-text-3);
}

.log-head {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 8px 14px 4px;
    font-size: 12px;
    color: var(--color-text-3);
}

.ls-log {
    flex: 1;
    min-height: 160px;
    overflow: auto;
    background: var(--color-fill-2);
    margin: 0 14px 12px;
    padding: 8px;
    border-radius: 4px;
    font-family: Consolas, Menlo, monospace;
    font-size: 12px;
}

.ls-log.small {
    max-height: 140px;
    min-height: 60px;
    margin: 0;
}

.ls-line {
    white-space: pre-wrap;
    word-break: break-all;
    line-height: 1.6;
}

.ver-block .ver-title,
.ver-title {
    display: flex;
    align-items: center;
    gap: 8px;
    font-weight: 600;
    font-size: 13px;
    margin-bottom: 8px;
}

.add-local {
    margin-top: 10px;
}

.offline-tip {
    margin-top: 8px;
    font-size: 12px;
    color: var(--color-text-3);
    word-break: break-all;
}

.install-log {
    margin-top: 12px;
}

.cfg-registry {
    margin-bottom: 4px;
}

.user-npmrc {
    margin-top: 12px;
}
</style>
