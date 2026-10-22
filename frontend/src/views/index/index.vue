<template>
  <layout-body height="full" :isBg="false">
    <a-row :gutter="12">
      <a-col :span="16">
        <!-- ========== 左侧：系统信息 ========== -->

        <a-row :gutter="[12, 12]">
          <a-col :span="24">
            <a-card>

              <template #title>
                <!-- 欢迎栏 -->
                <span class="right-card-title">欢迎回来！</span>
              </template>
              <template #extra>
                <a-button size="mini" type="text" :loading="sysLoading" @click="refreshSysInfo">
                  <template #icon><icon-refresh /></template>
                </a-button>
              </template>

              <!-- 统计卡片行（圆形图标+标题+大数字） -->
              <div class="stat-cards">
                <div class="stat-card">
                  <div class="stat-circle stat-circle-cpu"><icon-thunderbolt /></div>
                  <div class="stat-info">
                    <div class="stat-label">CPU</div>
                    <div class="stat-value">{{ sysInfo ? sysInfo.cpuUsage.toFixed(1) : '--' }}<span
                        class="stat-unit">%</span>
                    </div>
                  </div>
                </div>
                <div class="stat-card">
                  <div class="stat-circle stat-circle-mem"><icon-storage /></div>
                  <div class="stat-info">
                    <div class="stat-head">
                      <div class="stat-label">内存</div>
                      <div class="stat-sub">{{ sysInfo ? fmtBytes(sysInfo.memUsed) + ' / ' + fmtBytes(sysInfo.memTotal) :
                        ''
                      }}
                      </div>
                    </div>
                    <div class="stat-value">{{ sysInfo ? sysInfo.memPercent.toFixed(1) : '--' }}<span
                        class="stat-unit">%</span>
                    </div>
                  </div>
                </div>
                <div class="stat-card">
                  <div class="stat-circle stat-circle-up"><icon-clock-circle /></div>
                  <div class="stat-info">
                    <div class="stat-label">运行时间</div>
                    <div class="stat-value">{{ sysInfo ? fmtUptime(sysInfo.uptime) : '--' }}</div>
                  </div>
                </div>
                <div class="stat-card">
                  <div class="stat-circle stat-circle-go"><icon-settings /></div>
                  <div class="stat-info">
                    <div class="stat-head">
                      <div class="stat-label">Go 运行时</div>
                      <div class="stat-sub">{{ sysInfo ? '堆 ' + fmtBytes(sysInfo.appMemAlloc) : '' }}</div>
                    </div>
                    <div class="stat-value">{{ sysInfo ? sysInfo.goroutines : '--' }}<span class="stat-unit"> 协程</span>
                    </div>
                  </div>
                </div>
              </div>
            </a-card>
          </a-col>
          <a-col :span="24">
            <a-card>
              <!-- 本地服务实例卡片 -->
              <div class="svc-section">
                <div class="svc-header">
                  <span class="svc-title">本地服务实例</span>
                  <a-radio-group v-model="svcFilter" type="button" size="mini">
                    <a-radio value="all">全部</a-radio>
                    <a-radio value="redis">Redis</a-radio>
                    <a-radio value="mysql">MySQL</a-radio>
                  </a-radio-group>
                </div>
                <div class="svc-cards" v-if="filteredServices.length">
                  <div v-for="(s, i) in filteredServices" :key="i" class="svc-card"
                    :class="{ 'svc-card-active': s.running }">
                    <div class="svc-card-top">
                      <div class="svc-logo" :class="s.type === 'redis' ? 'logo-redis' : 'logo-mysql'">
                        <icon-one-to-one v-if="s.type === 'redis'" />
                        <icon-code v-else />
                      </div>
                      <div class="svc-meta">
                        <div class="svc-name">{{ s.name }}</div>
                        <div class="svc-desc">{{ s.type === 'redis' ? 'Redis' : 'MySQL' }} · v{{ s.version }} · :{{
                          s.port
                        }}
                        </div>
                      </div>
                    </div>
                    <div class="svc-card-bottom">
                      <a-tag :color="s.running ? 'green' : 'gray'" size="small">
                        <template #icon>
                          <icon-check-circle-fill v-if="s.running" />
                          <icon-exclamation-circle-fill v-else />
                        </template>
                        {{ s.running ? '已启用' : '已停止' }}
                      </a-tag>
                      <a-switch :model-value="s.running" :loading="svcToggling === s.id" size="small"
                        @change="toggleService(s)" />
                    </div>
                  </div>
                </div>
                <a-empty v-else description="暂无服务实例" />
              </div>
            </a-card>
          </a-col>
          <a-col :span="24">
            <a-card>
              <!-- 端口列表 -->
              <div class="svc-section" v-if="sysInfo?.listeners?.length">
                <div class="svc-header">
                  <span class="svc-title">TCP 监听端口 ({{ sysInfo.listeners.length }})</span>
                </div>
                <a-list v-if="sysInfo?.listeners?.length" class="port-list" :border="false" :split="false">
                  <a-list-item v-for="(p, i) in sysInfo.listeners.slice(0, 30)" :key="i" class="port-row">
                    <span class="port-num">{{ p.port }}</span>
                    <span class="port-proc">{{ p.process || '-' }}</span>
                    <span class="port-pid">PID {{ p.pid }}</span>
                  </a-list-item>
                </a-list>
              </div>
            </a-card>
          </a-col>
        </a-row>
      </a-col>

      <a-col :span="8">

        <!-- ========== 右侧：快捷操作 ========== -->
        <a-card>
          <template #title>
            <span class="right-card-title">快捷操作</span>
          </template>
          <template #extra>
            <a-tooltip content="设置">
              <a-button size="mini" type="text" @click="openSettings">
                <template #icon><icon-settings /></template>
              </a-button>
            </a-tooltip>
          </template>
          <div class="tools-grid">
            <a-tooltip v-for="(item, index) in toolsList" :key="item.id || index"
              :content="item.shortcut ? item.name + ' (' + item.shortcut + ')' : item.name">
              <div class="tool-cell" @click="onToolClick(item)">
                <div class="tool-icon-wrap">
                  <i v-if="isIconfont(item.icon)" class="icon tool-icon" :class="item.icon" />
                  <component v-else :is="iconComp(item.icon) || 'IconApps'" class="tool-icon" />
                </div>
              </div>
            </a-tooltip>
          </div>
        </a-card>
      </a-col>
    </a-row>

  </layout-body>
</template>

<script setup lang="ts">
import { computed, getCurrentInstance, onMounted, onUnmounted, reactive, ref } from "vue";
import { Events } from "@wailsio/runtime";
import { IconCheckCircleFill, IconExclamationCircleFill } from "@arco-design/web-vue/es/icon";
import { useToolEvents, type ToolItem } from "./events";
import { useToolWindow } from "./useToolWindow";

interface ToolRow extends ToolItem {
  id?: string;
  key?: string;
  icon?: string;
  shortcut?: string;
}

const toolEvents = reactive(useToolEvents());
const { openToolWindow } = useToolWindow();
const GoTools = (getCurrentInstance() as any)?.appContext.config.globalProperties.$GoTools;

/* ==================== 工具列表 ==================== */
const toolsList = ref<ToolRow[]>([]);

function iconComp(icon?: string): string | null {
  if (!icon) return null;
  const pascal = String(icon).split(/[-_\s]+/).filter(Boolean).map((s) => s[0].toUpperCase() + s.slice(1)).join("");
  return pascal ? "Icon" + pascal : null;
}
function isIconfont(icon?: string): boolean {
  return !!icon && String(icon).indexOf("icon-") === 0;
}

async function loadTools() {
  try {
    const cfg = await GoTools?.GetConfig();
    const list: any[] = cfg?.tools || [];
    toolsList.value = list.map((t) => ({
      id: t.id, key: t.key, name: t.name, path: t.path || undefined,
      event: t.event || undefined, single: !!t.single, fullscreen: !!t.fullscreen,
      width: t.width || undefined, height: t.height || undefined, icon: t.icon, shortcut: t.shortcut,
    }));
  } catch { /* */ }
}

const onToolClick = (item: ToolRow) => {
  if (item.event) { toolEvents.runEvent(item.event as any); return; }
  if (item.path) {
    openToolWindow(item.path, item.name, { single: item.single, fullscreen: item.fullscreen, width: item.width, height: item.height });
  }
};

/** 设置入口（从顶部 header 移到快捷操作卡片右上角） */
function openSettings() {
  openToolWindow("/tools/settings", "设置", { single: true });
}
function activateById(id: string) {
  const item = toolsList.value.find((t) => t.id === id || t.key === id);
  if (item) onToolClick(item);
}

/* ==================== 系统信息 ==================== */
const sysInfo = ref<any>(null);
const sysLoading = ref(false);
const svcFilter = ref<"all" | "redis" | "mysql">("all");
const svcToggling = ref<string | null>(null);

const filteredServices = computed(() => {
  const list = sysInfo.value?.services || [];
  if (svcFilter.value === "all") return list;
  return list.filter((s: any) => s.type === svcFilter.value);
});

async function refreshSysInfo() {
  sysLoading.value = true;
  try { sysInfo.value = await GoTools?.GetSystemInfo(); } catch { /* */ }
  sysLoading.value = false;
}

async function toggleService(s: any) {
  svcToggling.value = s.id;
  try {
    if (s.running) {
      if (s.type === "redis") await GoTools?.RedisSrvStop(s.id);
      else await GoTools?.MySQLSrvStop(s.id);
    } else {
      if (s.type === "redis") await GoTools?.RedisSrvStart(s.id);
      else await GoTools?.MySQLSrvStart(s.id);
    }
    await refreshSysInfo();
  } catch (e) { /* */ }
  svcToggling.value = null;
}

function fmtBytes(b: number): string {
  if (!b) return "0 B";
  const units = ["B", "KB", "MB", "GB", "TB"]; let i = 0; let v = b;
  while (v >= 1024 && i < units.length - 1) { v /= 1024; i++; }
  return v.toFixed(i > 0 ? 1 : 0) + " " + units[i];
}
function fmtUptime(sec: number): string {
  if (!sec) return "-";
  const d = Math.floor(sec / 86400); const h = Math.floor((sec % 86400) / 3600);
  const m = Math.floor((sec % 3600) / 60);
  if (d > 0) return `${d}天${h}时`; if (h > 0) return `${h}时${m}分`; return `${m}分`;
}

/* ==================== 生命周期 ==================== */
const cancels: Array<() => void> = [];
function listen(name: string, cb: (id: string) => void) {
  try {
    const ret = Events.On(name, (ev: any) => cb(String(ev?.data ?? ev ?? "")));
    if (typeof ret === "function") cancels.push(ret as () => void);
  } catch { /* */ }
}

let sysTimer: ReturnType<typeof setInterval> | null = null;

onMounted(async () => {
  await loadTools();
  refreshSysInfo();
  sysTimer = setInterval(refreshSysInfo, 5000);
  listen("ball:action", activateById);
  listen("tool:activate", activateById);
});
onUnmounted(() => {
  if (sysTimer) clearInterval(sysTimer);
  cancels.forEach((f) => { try { f(); } catch { /* */ } });
  cancels.length = 0;
});
</script>

<style scoped>
/* ---- 左侧卡片 ---- */
.home-welcome {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: var(--base-padding);
}

.home-welcome-title {
  font-size: var(--base-size-large);
  font-weight: 600;
}

/* 统计卡片 - 浅色圆形图标风格 */
.stat-cards {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
  gap: var(--base-padding);
  margin-bottom:var(--base-padding);
}

.stat-card {
  display: flex;
  align-items: center;
  gap: var(--base-padding);
  padding: var(--base-padding);
  background: var(--color-bg-2);
  border-radius: var(--base-radius-large);
  border: 1px solid var(--color-border-1);
  transition: box-shadow 0.2s;
}

.stat-card:hover {
  box-shadow: var(--shadow-2-bottom);
}

.stat-circle {
  width: 44px;
  height: 44px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: var(--base-size-large-x);
  flex-shrink: 0;
}

.stat-circle-cpu {
  background: rgba(var(--arcoblue-6), 0.12);
  color: rgb(var(--arcoblue-6));
}

.stat-circle-mem {
  background: rgba(var(--cyan-6), 0.12);
  color: rgb(var(--cyan-6));
}

.stat-circle-up {
  background: rgba(var(--orange-6), 0.12);
  color: rgb(var(--orange-6));
}

.stat-circle-go {
  background: rgba(var(--purple-6), 0.12);
  color: rgb(var(--purple-6));
}

.stat-info {
  flex: 1;
  min-width: 0;
}

.stat-head {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  gap: 8px;
}

.stat-label {
  font-size: var(--base-size-small);
  color: var(--color-text-3);
  margin-bottom: var(--base-radius-small);
}

.stat-value {
  font-weight: bold;
  font-size: var(--base-size-default);
  line-height: 1.2;
  color: var(--color-text-1);
  margin-top: var(--base-padding);
}

.stat-unit {
  font-size: var(--base-size-small);
  font-weight: 400;
  color: var(--color-text-3);
  margin-left: var(--base-radius-small);
}

.stat-sub {
  font-size: var(--base-size-small);
  color: var(--color-text-4);
  white-space: nowrap;
}

/* 服务实例卡片 */
.svc-section {
  margin-bottom: var(--base-size-large-x);
}

.svc-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: var(--base-padding);
}

.svc-title {
  font-size: var(--base-size-default);
  font-weight: 600;
}

.svc-cards {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(240px, 1fr));
  gap: var(--base-margin);
}

.svc-card {
  background: var(--color-bg-2);
  border: 1px solid var(--color-border-2);
  border-radius: var(--base-radius-large);
  padding: var(--base-padding);
  transition: all 0.2s;
  cursor: default;
}

.svc-card:hover {
  border-color: rgb(var(--primary-3));
  box-shadow: 0 4px 16px rgba(var(--primary-6), 0.08);
  transform: translateY(-1px);
}

.svc-card-active {
  border-left: 3px solid rgb(var(--green-6));
}

.svc-card-top {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: var(--base-margin);
}

.svc-logo {
  width: 36px;
  height: 36px;
  border-radius: var(--base-radius-default);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: var(--base-size-large);
  flex-shrink: 0;
}

.logo-redis {
  background: rgba(var(--red-6), 0.12);
  color: rgb(var(--red-6));
}

.logo-mysql {
  background: rgba(var(--blue-6), 0.12);
  color: rgb(var(--blue-6));
}

.svc-meta {
  flex: 1;
  min-width: 0;
}

.svc-name {
  font-weight: 600;
  font-size: var(--base-size-default);
}

.svc-desc {
  font-size: var(--base-size-small);
  color: var(--color-text-3);
  margin-top: var(--base-radius-small);
}

.svc-card-bottom {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

/* 端口列表（a-list） */
.port-list {
  max-height: 200px;
  overflow-y: auto;
  border: 1px solid var(--color-border-2);
  border-radius: var(--base-radius-default);
  background: var(--color-bg-2);
}

.port-list :deep(.arco-list-item) {
  display: flex;
  align-items: center;
  gap: 12px;
  font-size: var(--base-size-small);
  font-family: Consolas, monospace;
  padding: var(--base-radius-small) var(--base-padding);
  border-bottom: 1px solid var(--color-border-1);
  transition: background 0.1s;
}

.port-list :deep(.arco-list-item:last-child) {
  border-bottom: none;
}

.port-list :deep(.arco-list-item:hover) {
  background: var(--color-fill-2);
}

.port-num {
  width: 50px;
  font-weight: 600;
  color: rgb(var(--primary-6));
}

.port-proc {
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.port-pid {
  color: var(--color-text-3);
}

/* ---- 右侧卡片 ---- */

.tools-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 8px;
}

.tool-cell {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: var(--base-padding);
  border-radius: var(--base-radius-large);
  cursor: pointer;
  transition: all 0.15s;
  background: transparent;
  border: none;
}

.tool-cell:hover {
  background: var(--color-fill-2);
}

.tool-cell:active {
  transform: scale(0.95);
}

.tool-icon-wrap {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(var(--primary-6), 0.12);
  color: rgb(var(--primary-6));
  font-size: var(--base-size-large);
}

.tool-icon {
  font-size: var(--base-size-large);
  color: inherit;
}
</style>
