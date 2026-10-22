<template>
    <layout-body title="设置" height="full">
        <div class="settings-wrap">
            <a-tabs v-model:active-key="activeTab" type="rounded" class="settings-tabs">
                <!-- ===================== 日志设置 ===================== -->
                <a-tab-pane key="log" title="日志">
                    <a-form :model="form" :style="{ maxWidth: '640px' }" layout="vertical" auto-focus>
                        <a-form-item field="enabled">
                            <template #label>启用文件日志</template>
                            <a-switch v-model="form.log.enabled" />
                            <span class="hint">关闭后不再写入日志文件（仅开发控制台可见）。</span>
                        </a-form-item>

                        <a-form-item field="level">
                            <template #label>日志级别</template>
                            <a-select v-model="form.log.level" style="width: 220px">
                                <a-option value="debug">debug（最详细）</a-option>
                                <a-option value="info">info（默认，含操作记录）</a-option>
                                <a-option value="warn">warn（含异常告警）</a-option>
                                <a-option value="error">error（仅错误）</a-option>
                            </a-select>
                        </a-form-item>

                        <a-form-item field="path">
                            <template #label>日志目录</template>
                            <div class="path-row">
                                <a-input v-model="form.log.path" placeholder="留空使用默认目录（程序目录下 logs）"
                                    allow-clear />
                                <a-button @click="pickDir">选择目录</a-button>
                            </div>
                            <div class="hint">只需配置一个目录，日志按“级别 + 当前日期”自动分文件，如
                                Error20261022.log、Info20261022.log（Warn/Debug 同理）。当前目录：{{
                                activePath || "（未加载）" }}</div>
                        </a-form-item>

                        <a-form-item>
                            <a-space :size="12">
                                <a-button type="primary" :loading="saving" @click="save">保存设置</a-button>
                                <a-button @click="resetLog">恢复默认</a-button>
                                <a-button v-if="activePath" @click="openLogDir">打开日志目录</a-button>
                            </a-space>
                        </a-form-item>
                    </a-form>
                </a-tab-pane>

                <!-- ===================== 通用设置 ===================== -->
                <a-tab-pane key="general" title="通用">
                    <a-form :model="form" :style="{ maxWidth: '640px' }" layout="vertical">
                        <a-form-item field="general.appName">
                            <template #label>应用名称</template>
                            <a-input v-model="form.general.appName" placeholder="Tools" :style="{ maxWidth: '280px' }" />
                            <span class="hint">窗口标题、托盘提示显示的名称，保存后立即生效。</span>
                        </a-form-item>
                        <a-form-item field="general.appLogo">
                            <template #label>应用 Logo</template>
                            <div class="logo-picker">
                                <div class="logo-preview">
                                    <img v-if="logoPreview" :src="logoPreview" class="logo-img" />
                                    <span v-else class="logo-placeholder">无</span>
                                </div>
                                <a-space>
                                    <a-button size="small" @click="pickLogo">选择图片</a-button>
                                    <a-button size="small" status="danger" v-if="form.general.appLogo" @click="form.general.appLogo = ''; logoPreview = ''">清除</a-button>
                                </a-space>
                            </div>
                            <span class="hint">建议 PNG 正方形（如 128x128），用于浮动球中心图标。留空使用内置默认。</span>
                        </a-form-item>
                        <a-form-item field="general.theme">
                            <template #label>主题模式</template>
                            <a-radio-group v-model="form.general.theme" type="button" size="small" @change="onThemeChange">
                                <a-radio value="light">浅色</a-radio>
                                <a-radio value="dark">暗黑</a-radio>
                                <a-radio value="auto">跟随系统</a-radio>
                            </a-radio-group>
                            <span class="hint">切换后立即生效，所有窗口同步。</span>
                        </a-form-item>
                        <a-form-item field="autoStart">
                            <template #label>开机默认启动</template>
                            <a-switch v-model="form.general.autoStart" />
                            <span class="hint">开启后 Windows 登录时自动启动应用。</span>
                        </a-form-item>
                        <a-form-item>
                            <a-space :size="12">
                                <a-button type="primary" :loading="saving" @click="save">保存设置</a-button>
                            </a-space>
                        </a-form-item>
                    </a-form>
                </a-tab-pane>

                <!-- ===================== 浮动球设置 ===================== -->
                <a-tab-pane key="ball" title="浮动球">
                    <a-form :model="form" :style="{ maxWidth: '640px' }" layout="vertical">
                        <a-form-item field="ballEnabled">
                            <template #label>显示桌面浮动球</template>
                            <a-switch v-model="form.ball.enabled" />
                            <span class="hint">开启后浮动球常驻屏幕右下角（独立原生小窗，置顶、可拖动，点击展开快捷功能）。保存后立即显隐，并在下次启动时自动恢复。</span>
                        </a-form-item>
                        <a-form-item>
                            <a-space :size="12">
                                <a-button type="primary" :loading="saving" @click="save">保存设置</a-button>
                            </a-space>
                        </a-form-item>
                    </a-form>
                </a-tab-pane>

                <!-- ===================== 工具列表编辑 ===================== -->
                <a-tab-pane key="tools" title="工具">
                    <div class="tools-pane">
                        <div class="tools-toolbar">
                            <a-button type="primary" size="small" @click="openAdd">
                                <template #icon><icon-plus /></template>新增功能
                            </a-button>
                            <a-button size="small" @click="resetTools">恢复默认列表</a-button>
                            <span class="hint">编辑后立即点“保存设置”持久化。勾选“快捷”的功能会出现在桌面浮动球菜单；快捷键为系统全局热键（如 Ctrl+Alt+S）。</span>
                        </div>

                        <div class="tools-table-wrap" ref="toolsTableRef">
                            <a-table :data="tools" :pagination="false" size="small" row-key="id" :column-resizable="true"
                                :scroll="tableScroll">
                                <template #columns>
                                    <a-table-column title="排序" :width="88">
                                        <template #cell="{ rowIndex }">
                                            <a-space :size="2">
                                                <a-button size="mini" :disabled="rowIndex === 0" @click="move(rowIndex, -1)">
                                                    <icon-up />
                                                </a-button>
                                                <a-button size="mini" :disabled="rowIndex === tools.length - 1"
                                                    @click="move(rowIndex, 1)">
                                                    <icon-down />
                                                </a-button>
                                            </a-space>
                                        </template>
                                    </a-table-column>
                                    <a-table-column title="标题" data-index="name" :width="150" ellipsis tooltip />
                                    <a-table-column title="内置功能" data-index="key" :width="130" ellipsis tooltip />
                                    <a-table-column title="图标" data-index="icon" :width="140">
                                        <template #cell="{ record }">
                                            <span v-if="record.icon" class="icon-cell">
                                                <i v-if="String(record.icon).indexOf('icon-') === 0" class="icon iconfont-preview"
                                                    :class="record.icon" />
                                                <component v-else :is="iconComp(record.icon)" class="arco-preview" />
                                                <span class="icon-text">{{ record.icon }}</span>
                                            </span>
                                            <span v-else class="off">—</span>
                                        </template>
                                    </a-table-column>
                                    <a-table-column title="快捷" :width="60">
                                        <template #cell="{ record }">
                                            <icon-check-circle v-if="record.quick" class="on" />
                                            <span v-else class="off">—</span>
                                        </template>
                                    </a-table-column>
                                    <a-table-column title="单例" :width="60">
                                        <template #cell="{ record }">
                                            <icon-check-circle v-if="record.single" class="on" />
                                            <span v-else class="off">—</span>
                                        </template>
                                    </a-table-column>
                                    <a-table-column title="全屏" :width="60">
                                        <template #cell="{ record }">
                                            <icon-check-circle v-if="record.fullscreen" class="on" />
                                            <span v-else class="off">—</span>
                                        </template>
                                    </a-table-column>
                                    <a-table-column title="快捷键" data-index="shortcut" :width="120" ellipsis tooltip>
                                        <template #cell="{ record }">
                                            <a-tag v-if="record.shortcut" size="small">{{ record.shortcut }}</a-tag>
                                            <span v-else class="off">—</span>
                                        </template>
                                    </a-table-column>
                                    <a-table-column title="操作" :width="120" fixed="right">
                                        <template #cell="{ record, rowIndex }">
                                            <a-space :size="6">
                                                <a-link @click="openEdit(rowIndex)">编辑</a-link>
                                                <a-popconfirm content="确定删除该功能？" @ok="del(rowIndex)">
                                                    <a-link status="danger">删除</a-link>
                                                </a-popconfirm>
                                            </a-space>
                                        </template>
                                    </a-table-column>
                                </template>
                            </a-table>
                        </div>

                        <!-- 统一的“保存设置”入口（工具与日志共用） -->
                        <div class="save-bar" ref="saveBarRef">
                            <a-button type="primary" :loading="saving" @click="save">保存设置</a-button>
                        </div>
                    </div>
                </a-tab-pane>
            </a-tabs>
        </div>

        <!-- ===================== 工具编辑弹窗 ===================== -->
        <a-modal v-model:visible="editVisible" :title="editIndex < 0 ? '新增功能' : '编辑功能'" :width="520" @ok="confirmEdit"
            @cancel="editVisible = false" ok-text="确定" cancel-text="取消">
            <a-form :model="editForm" layout="vertical">
                <a-form-item field="key" label="内置功能">
                    <a-select v-model="editForm.key" placeholder="选择要绑定的内置功能" allow-search @change="onBuiltinChange">
                        <a-option v-for="b in BUILTINS" :key="b.key" :value="b.key">{{ b.name }}</a-option>
                    </a-select>
                </a-form-item>
                <a-form-item field="name" label="标题">
                    <a-input v-model="editForm.name" placeholder="显示名称" />
                </a-form-item>
                <a-form-item field="icon" label="图标">
                    <a-input v-model="editForm.icon" placeholder="选择图标" allow-clear>
                        <template #prefix>
                            <component v-if="editForm.icon" :is="iconComp(editForm.icon)" />
                        </template>
                        <template #append>
                            <a-button @click="() => { proxy?.$refs['selectModalIconRef']?.open(); }">图标</a-button>
                            <select-icon-modal ref="selectModalIconRef" type="arco"
                                @change="iconChange"></select-icon-modal>
                        </template>
                    </a-input>
                </a-form-item>
                <a-row :gutter="12">
                    <a-col :span="8">
                        <a-form-item field="quick" label="浮动球快捷">
                            <a-switch v-model="editForm.quick" />
                        </a-form-item>
                    </a-col>
                    <a-col :span="8">
                        <a-form-item field="single" label="单例窗口">
                            <a-switch v-model="editForm.single" />
                        </a-form-item>
                    </a-col>
                    <a-col :span="8">
                        <a-form-item field="fullscreen" label="全屏启动">
                            <a-switch v-model="editForm.fullscreen" />
                        </a-form-item>
                    </a-col>
                </a-row>
                <a-row :gutter="12">
                    <a-col :span="12">
                        <a-form-item field="width" label="视口宽（0=默认）">
                            <a-input-number v-model="editForm.width" :min="0" :step="10" hide-button />
                        </a-form-item>
                    </a-col>
                    <a-col :span="12">
                        <a-form-item field="height" label="视口高（0=默认）">
                            <a-input-number v-model="editForm.height" :min="0" :step="10" hide-button />
                        </a-form-item>
                    </a-col>
                </a-row>
                <a-form-item field="shortcut" label="全局快捷键">
                    <a-input :model-value="editForm.shortcut" readonly :placeholder="recording ? '请按下组合键…（Esc 取消）' : '点击后直接按下组合键（如 Ctrl+Alt+S），留空不设置'"
                        @click="startRecord" @keydown="onShortcutKeydown" allow-clear @clear="editForm.shortcut = ''">
                        <template #suffix>
                            <span v-if="recording" class="rec-tip">录制中</span>
                        </template>
                    </a-input>
                    <div class="hint">点击输入框后直接按下你想要的组合键即可自动录入；按 Esc 退出录制。</div>
                </a-form-item>
            </a-form>
        </a-modal>
    </layout-body>
</template>

<script setup lang="ts">
import { computed, getCurrentInstance, nextTick, onBeforeUnmount, onMounted, reactive, ref, watch } from "vue";
import { Message } from "@arco-design/web-vue";
import { applyTheme, type ThemeMode } from "@/utils/theme";

const { proxy } = getCurrentInstance() as any;
const GoTools = (getCurrentInstance() as any)?.appContext.config.globalProperties.$GoTools;

/** 图标名（如 camera / clock-circle）→ Arco 全局组件名（IconCamera / IconClockCircle） */
function iconComp(icon?: string): string | null {
    if (!icon) return null;
    const pascal = String(icon)
        .split(/[-_\s]+/)
        .filter(Boolean)
        .map((s) => s[0].toUpperCase() + s.slice(1))
        .join("");
    return pascal ? "Icon" + pascal : null;
}

/** 内置功能目录：新增时选择后自动回填 path/event/icon 等默认值 */
const BUILTINS = [
    { key: "screenshot", name: "截图", event: "screenshot", icon: "camera" },
    { key: "colorpicker", name: "颜色取色器", path: "/tools/colorpicker", icon: "color-picker", single: true, width: 480, height: 560 },
    { key: "video", name: "视频录制", path: "/tools/video", icon: "video-camera", single: true, fullscreen: true },
    { key: "qrcode", name: "二维码生成", path: "/tools/qrcode", icon: "qrcode" },
    { key: "timestamp", name: "时间戳转换", path: "/tools/timestamp", icon: "clock-circle" },
    { key: "strrandom", name: "随机字符串生成", path: "/tools/strrandom", icon: "random" },
    { key: "encdec", name: "文字加密解密", path: "/tools/encdec", icon: "lock" },
    { key: "hexconvert", name: "进制转换", path: "/tools/hexconvert", icon: "swap" },
    { key: "imageBase64", name: "图片base64", path: "/tools/imageBase64", icon: "image" },
    { key: "sniffer", name: "抓包工具", path: "/tools/sniffer", icon: "wifi", single: true, fullscreen: true },
    { key: "websocket", name: "WebSocket调试", path: "/tools/websocket", icon: "connection", fullscreen: true },
    { key: "usbserial", name: "USB调试", path: "/tools/usbserial", icon: "usb" },
    { key: "redis", name: "Redis管理", path: "/tools/redis", icon: "storage", single: true, fullscreen: true },
    { key: "mysql", name: "MySQL管理", path: "/tools/mysql", icon: "code-block", single: true, fullscreen: true },
    { key: "settings", name: "设置", path: "/tools/settings", icon: "settings", single: true },
];

interface ToolRow {
    id: string; key: string; name: string; path: string; event: string; icon: string;
    single: boolean; fullscreen: boolean; width: number; height: number; shortcut: string; quick: boolean;
}

/** 完整配置（保存时整体回传） */
const fullConfig = ref<any>({ log: { enabled: true, level: "info", path: "" }, ball: { enabled: false }, general: { autoStart: false, appName: "Tools", appLogo: "", theme: "light" }, tools: [] });
const form = reactive<{
    log: { enabled: boolean; level: string; path: string };
    ball: { enabled: boolean };
    general: { autoStart: boolean; appName: string; appLogo: string; theme: string };
}>({
    log: { enabled: true, level: "info", path: "" },
    ball: { enabled: false },
    general: { autoStart: false, appName: "Tools", appLogo: "", theme: "light" },
});
const tools = ref<ToolRow[]>([]);

const saving = ref(false);

/** 主题切换：立即应用到 DOM（localStorage 同步广播到其他窗口），保存时再持久化到配置 */
function onThemeChange(val: any) {
    applyTheme(String(val) as ThemeMode);
}
const activePath = ref("");
const logoPreview = ref("");

async function pickLogo() {
    try {
        const path = await GoTools?.ChooseAppLogo();
        if (path) {
            form.general.appLogo = path;
            // 生成预览（读取文件转 base64）
            const dataUrl = await GoTools?.GetAppLogo();
            if (dataUrl) logoPreview.value = dataUrl;
        }
    } catch { /* cancelled */ }
}

function normTool(t: any): ToolRow {
    return {
        id: t.id || t.key || "",
        key: t.key || "",
        name: t.name || "",
        path: t.path || "",
        event: t.event || "",
        icon: t.icon || "",
        single: !!t.single,
        fullscreen: !!t.fullscreen,
        width: t.width || 0,
        height: t.height || 0,
        shortcut: t.shortcut || "",
        quick: !!t.quick,
    };
}

async function load() {
    try {
        const cfg = await GoTools?.GetConfig();
        if (cfg) {
            fullConfig.value = cfg;
            form.log = { enabled: !!cfg.log?.enabled, level: cfg.log?.level || "info", path: cfg.log?.path || "" };
            form.ball = { enabled: !!cfg.ball?.enabled };
            form.general = { autoStart: !!cfg.general?.autoStart, appName: cfg.general?.appName || "Tools", appLogo: cfg.general?.appLogo || "", theme: cfg.general?.theme || "light" };
            tools.value = (cfg.tools || []).map(normTool);
        }
        activePath.value = (await GoTools?.GetLogDir()) || "";
        // 加载 logo 预览
        try { const lp = await GoTools?.GetAppLogo(); if (lp) logoPreview.value = lp; } catch { /* */ }
    } catch {
        Message.error("读取配置失败");
    }
}

/* ---------------- 工具编辑 ---------------- */
const editVisible = ref(false);
const editIndex = ref(-1);
const editForm = reactive<ToolRow>(normTool({}));

function openAdd() {
    editIndex.value = -1;
    Object.assign(editForm, normTool({ id: "", key: "", name: "", path: "", event: "", icon: "", single: false, fullscreen: false, width: 0, height: 0, shortcut: "", quick: false }));
    editVisible.value = true;
}

function openEdit(i: number) {
    editIndex.value = i;
    Object.assign(editForm, normTool(tools.value[i]));
    editVisible.value = true;
}

/** 图标选择弹层回填（arco 模式 emit 图标名，如 camera / clock-circle） */
function iconChange(val: any) {
    editForm.icon = String(val ?? "");
}

function onBuiltinChange(val: any) {
    const key = String(val ?? "");
    const b = BUILTINS.find((x) => x.key === key);
    if (!b) return;
    editForm.key = b.key;
    if (!editForm.name || editIndex.value < 0) editForm.name = b.name;
    editForm.path = (b as any).path || "";
    editForm.event = (b as any).event || "";
    editForm.icon = b.icon || editForm.icon;
    if (editIndex.value < 0) {
        editForm.single = !!(b as any).single;
        editForm.fullscreen = !!(b as any).fullscreen;
        editForm.width = (b as any).width || 0;
        editForm.height = (b as any).height || 0;
    }
}

function genId() {
    return "t" + Date.now().toString(36) + Math.random().toString(36).slice(2, 6);
}

function confirmEdit() {
    if (!editForm.key && !editForm.path && !editForm.event) {
        Message.warning("请选择内置功能或填写路由/事件");
        return;
    }
    const row = normTool({ ...editForm, id: editForm.id || genId() });
    if (editIndex.value < 0) {
        tools.value.push(row);
    } else {
        tools.value[editIndex.value] = row;
    }
    editVisible.value = false;
}

function del(i: number) {
    tools.value.splice(i, 1);
}

function move(i: number, dir: number) {
    const j = i + dir;
    if (j < 0 || j >= tools.value.length) return;
    const arr = tools.value;
    [arr[i], arr[j]] = [arr[j], arr[i]];
}

function resetTools() {
    tools.value = DEFAULT_TOOLS.map(normTool);
}

/* ---------------- 日志操作 ---------------- */
async function pickDir() {
    try {
        const dir = await GoTools?.ChooseLogDir();
        if (dir) form.log.path = dir;
    } catch {
        /* 用户取消 */
    }
}

async function save() {
    saving.value = true;
    try {
        const payload = { log: { ...form.log }, ball: { ...form.ball }, general: { ...form.general }, tools: tools.value.map(normTool) };
        const path = await GoTools?.SaveConfig(payload);
        fullConfig.value = payload;
        activePath.value = path || "";
        // 应用名称变更立即生效到窗口标题
        try { await GoTools?.ApplyAppName(); } catch { /* */ }
        Message.success("设置已保存");
    } catch {
        Message.error("保存失败");
    } finally {
        saving.value = false;
    }
}

function resetLog() {
    form.log = { enabled: true, level: "info", path: "" };
}

async function openLogDir() {
    try {
        await GoTools?.OpenLogDir();
    } catch {
        if (activePath.value) Message.info(activePath.value);
    }
}

/** 默认工具列表（与后端 DefaultConfig 对齐，用于“恢复默认列表”） */
const DEFAULT_TOOLS: ToolRow[] = [
    { id: "screenshot", key: "screenshot", name: "截图", path: "", event: "screenshot", icon: "camera", single: false, fullscreen: false, width: 0, height: 0, shortcut: "", quick: true },
    { id: "colorpicker", key: "colorpicker", name: "颜色取色器", path: "/tools/colorpicker", event: "", icon: "color-picker", single: true, fullscreen: false, width: 480, height: 560, shortcut: "", quick: true },
    { id: "video", key: "video", name: "视频录制", path: "/tools/video", event: "", icon: "video-camera", single: true, fullscreen: true, width: 0, height: 0, shortcut: "", quick: true },
    { id: "qrcode", key: "qrcode", name: "二维码生成", path: "/tools/qrcode", event: "", icon: "qrcode", single: false, fullscreen: false, width: 0, height: 0, shortcut: "", quick: true },
    { id: "timestamp", key: "timestamp", name: "时间戳转换", path: "/tools/timestamp", event: "", icon: "clock-circle", single: false, fullscreen: false, width: 0, height: 0, shortcut: "", quick: true },
    { id: "strrandom", key: "strrandom", name: "随机字符串生成", path: "/tools/strrandom", event: "", icon: "random", single: false, fullscreen: false, width: 0, height: 0, shortcut: "", quick: false },
    { id: "encdec", key: "encdec", name: "文字加密解密", path: "/tools/encdec", event: "", icon: "lock", single: false, fullscreen: false, width: 0, height: 0, shortcut: "", quick: false },
    { id: "hexconvert", key: "hexconvert", name: "进制转换", path: "/tools/hexconvert", event: "", icon: "swap", single: false, fullscreen: false, width: 0, height: 0, shortcut: "", quick: false },
    { id: "imageBase64", key: "imageBase64", name: "图片base64", path: "/tools/imageBase64", event: "", icon: "image", single: false, fullscreen: false, width: 0, height: 0, shortcut: "", quick: false },
    { id: "sniffer", key: "sniffer", name: "抓包工具", path: "/tools/sniffer", event: "", icon: "wifi", single: true, fullscreen: true, width: 0, height: 0, shortcut: "", quick: false },
    { id: "websocket", key: "websocket", name: "WebSocket调试", path: "/tools/websocket", event: "", icon: "connection", single: false, fullscreen: true, width: 0, height: 0, shortcut: "", quick: false },
    { id: "usbserial", key: "usbserial", name: "USB调试", path: "/tools/usbserial", event: "", icon: "usb", single: false, fullscreen: false, width: 0, height: 0, shortcut: "", quick: false },
    { id: "redis", key: "redis", name: "Redis管理", path: "/tools/redis", event: "", icon: "storage", single: true, fullscreen: true, width: 0, height: 0, shortcut: "", quick: false },
    { id: "mysql", key: "mysql", name: "MySQL管理", path: "/tools/mysql", event: "", icon: "code-block", single: true, fullscreen: true, width: 0, height: 0, shortcut: "", quick: false },
    { id: "settings", key: "settings", name: "设置", path: "/tools/settings", event: "", icon: "settings", single: true, fullscreen: false, width: 0, height: 0, shortcut: "", quick: false },
];

/* ---------------- 表格高度自适应（继承窗口高度，表体内部滚动） ---------------- */
const toolsTableRef = ref<HTMLElement>();
const saveBarRef = ref<HTMLElement>();
const activeTab = ref<string>("log");
const tableScrollY = ref<number>(0);
const tableScroll = computed(() =>
    tableScrollY.value > 0
        ? { x: "100%", y: tableScrollY.value }
        : { x: "100%" }
);

/** 测量表格包裹容器可用高度，扣除表头，得到可滚动的表体 y */
function calcTableScroll() {
    if (activeTab.value !== "tools") return;
    const wrap = toolsTableRef.value;
    if (!wrap) return;
    const header = wrap.querySelector(".arco-table-header") as HTMLElement | null;
    const headerH = header?.offsetHeight ?? 34;
    // 优先取 flex 撑出的可用高度；若为 0（尚未约束）则按视口回退估算
    let avail = wrap.clientHeight;
    if (!avail) {
        const top = wrap.getBoundingClientRect().top;
        const saveH = saveBarRef.value?.offsetHeight ?? 48;
        avail = window.innerHeight - top - saveH - 32;
    }
    tableScrollY.value = Math.max(160, avail - headerH);
}

let resizeTimer = 0;
const onWinResize = () => {
    clearTimeout(resizeTimer);
    resizeTimer = window.setTimeout(calcTableScroll, 100);
};

let ro: ResizeObserver | null = null;
watch(tools, () => nextTick(calcTableScroll));

//  切换到“工具”页时重新观测包裹容器并计算高度
watch(activeTab, (k) => {
    nextTick(() => {
        if (ro) {
            ro.disconnect();
            if (toolsTableRef.value) ro.observe(toolsTableRef.value);
        }
        calcTableScroll();
    });
});

/* ---------------- 快捷键录制（按下组合键直接赋值） ---------------- */
const recording = ref(false);

function startRecord() {
    recording.value = true;
}

/** 将按键名规范化为热键文本（与后端全局热键解析一致） */
function normalizeKey(k: string): string {
    if (k === " ") return "Space";
    if (k.length === 1) return k.toUpperCase();
    const map: Record<string, string> = {
        ArrowUp: "Up", ArrowDown: "Down", ArrowLeft: "Left", ArrowRight: "Right",
        Escape: "Esc", Delete: "Del", Insert: "Ins", PageUp: "PgUp", PageDown: "PgDn",
    };
    return map[k] || k;
}

function onShortcutKeydown(e: KeyboardEvent) {
    e.preventDefault();
    if (e.key === "Escape") {
        recording.value = false;
        return;
    }
    if (e.key === "Backspace" || e.key === "Delete") {
        editForm.shortcut = "";
        return;
    }
    // 仅按下修饰键时忽略，等待主键
    if (["Control", "Alt", "Shift", "Meta"].includes(e.key)) return;
    const parts: string[] = [];
    if (e.ctrlKey) parts.push("Ctrl");
    if (e.altKey) parts.push("Alt");
    if (e.shiftKey) parts.push("Shift");
    if (e.metaKey) parts.push("Win");
    parts.push(normalizeKey(e.key));
    editForm.shortcut = parts.join("+");
    recording.value = false;
}

onMounted(() => {
    load();
    window.addEventListener("resize", onWinResize);
    if (typeof ResizeObserver !== "undefined") {
        ro = new ResizeObserver(() => calcTableScroll());
        if (toolsTableRef.value) ro.observe(toolsTableRef.value);
    }
});

onBeforeUnmount(() => {
    window.removeEventListener("resize", onWinResize);
    ro?.disconnect();
    ro = null;
    clearTimeout(resizeTimer);
});
</script>

<style scoped>
.settings-wrap {
    height: 100%;
    display: flex;
    flex-direction: column;
    padding: 16px 20px;
    box-sizing: border-box;
}

.settings-tabs {
    background: var(--color-bg-2);
    border-radius: 8px;
    padding: 8px 16px 16px;
    flex: 1;
    min-height: 0;
    display: flex;
    flex-direction: column;
}

/* 让 Tabs 内容区继承高度，以便工具表格包裹容器获得确定的可用高度 */
.settings-tabs :deep(.arco-tabs-nav) {
    flex: none;
}

.settings-tabs :deep(.arco-tabs-content) {
    flex: 1;
    min-height: 0;
    overflow: hidden;
}

.settings-tabs :deep(.arco-tabs-content-list),
.settings-tabs :deep(.arco-tabs-pane) {
    height: 100%;
}

.settings-tabs :deep(.arco-tabs-pane) {
    overflow: auto;
}

.path-row {
    display: flex;
    gap: 8px;
    width: 100%;
}

.path-row :deep(.arco-input-wrapper) {
    flex: 1;
}

.hint {
    color: var(--color-text-3);
    font-size: 12px;
    margin-left: 12px;
}

.tools-toolbar {
    display: flex;
    align-items: center;
    gap: 12px;
    margin-bottom: 12px;
    flex-wrap: wrap;
    flex: none;
}

.tools-pane {
    height: 100%;
    display: flex;
    flex-direction: column;
}

.tools-table-wrap {
    flex: 1;
    min-height: 0;
    overflow: hidden;
}

.tools-toolbar .hint {
    margin-left: 0;
}

.on {
    color: rgb(var(--green-6));
}

.off {
    color: var(--color-text-4);
}

.save-bar {
    flex: none;
    margin-top: 16px;
    text-align: right;
}

/* 图标列：图标预览 + 名称 */
.icon-cell {
    display: inline-flex;
    align-items: center;
    gap: 6px;
}

.icon-cell .arco-preview,
.icon-cell .iconfont-preview {
    font-size: 16px;
    color: var(--color-text-1);
}

.icon-cell .icon-text {
    color: var(--color-text-3);
    font-size: 12px;
}

/* 快捷键录制提示 */
.rec-tip {
    font-size: 12px;
    color: rgb(var(--primary-6));
}

/* Logo 选择器 */
.logo-picker {
    display: flex;
    align-items: center;
    gap: 16px;
}
.logo-preview {
    width: 48px;
    height: 48px;
    border-radius: 8px;
    border: 1px dashed var(--color-border-3);
    display: flex;
    align-items: center;
    justify-content: center;
    overflow: hidden;
    background: var(--color-fill-1);
}
.logo-img {
    width: 100%;
    height: 100%;
    object-fit: contain;
}
.logo-placeholder {
    font-size: 12px;
    color: var(--color-text-4);
}
</style>
