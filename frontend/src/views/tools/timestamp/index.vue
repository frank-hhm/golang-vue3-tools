<template>
    <layout-body title="时间戳转换">
        <!-- 当前时间戳 -->
        <a-card class="tool-card">
            <a-row>
                <a-col :span="12">
                    <div class="now-value">
                        <span class="now-num">{{ displayNow }}</span>
                        <a-tag color="green" size="small">{{ nowUnit === "ms" ? "毫秒" : "秒" }}</a-tag>
                    </div>
                </a-col>
                <a-col :span="12">
                    <a-space :size="12" align="end">
                        <a-button @click="toggleUnit">
                            <template #icon><icon-swap /></template>切换单位
                        </a-button>
                        <a-button @click="copyNow">
                            <template #icon><icon-copy /></template>复制
                        </a-button>
                        <a-button type="primary" :status="paused ? 'success' : 'danger'" @click="togglePause">
                            <template #icon>
                                <icon-play-circle v-if="paused" />
                                <icon-pause-circle v-else />
                            </template>
                            {{ paused ? "继续" : "暂停" }}
                        </a-button>
                    </a-space>
                </a-col>
            </a-row>
        </a-card>

        <a-card class="tool-card">
            <a-tabs default-active-key="single" type="rounded">
                <!-- 单个转换 -->
                <a-tab-pane key="single" title="单个转换">
                    <!-- 时间戳 -> 日期时间 -->
                    <div class="section-title"><icon-clock-circle /> 时间戳转日期时间</div>
                    <a-row :gutter="12">
                        <a-col :span="6">
                            <a-input v-model="tsInput" placeholder="请输入时间戳" allow-clear @press-enter="tsToDate" />
                        </a-col>
                        <a-col :span="12">
                            <a-space :size="12">
                                <a-select v-model="tsUnit">
                                    <a-option value="s">秒 (s)</a-option>
                                    <a-option value="ms">毫秒 (ms)</a-option>
                                </a-select>
                                <a-select v-model="tsTz">
                                    <a-option v-for="tz in TZ_LIST" :key="tz" :value="tz">{{ tz }}</a-option>
                                </a-select>
                                <a-button @click="fillCurrent">
                                    填入当前
                                </a-button>
                                <a-button type="primary" status="success" @click="tsToDate">
                                    转换
                                </a-button>
                                <a-input :model-value="tsResult" placeholder="转换结果" readonly style="width: 260px;">
                                    <template #append> <a-button :disabled="!tsResult" @click="copyText(tsResult)">
                                            <template #icon><icon-copy /></template>
                                        </a-button> </template>
                                </a-input>
                            </a-space>
                        </a-col>
                    </a-row>

                    <a-divider />

                    <!-- 日期时间 -> 时间戳 -->
                    <div class="section-title"><icon-calendar /> 日期时间转时间戳</div>
                    <a-row :gutter="12">
                        <a-col :span="6">
                            <a-input v-model="dateInput" placeholder="YYYY-MM-DD HH:mm:ss" allow-clear
                                @press-enter="dateToTs" />
                        </a-col>
                        <a-col :span="12">
                            <a-space :size="12">
                                <a-select v-model="dateUnit" @change="dateToTs">
                                    <a-option value="s">秒 (s)</a-option>
                                    <a-option value="ms">毫秒 (ms)</a-option>
                                </a-select>
                                <a-select v-model="dateTz">
                                    <a-option v-for="tz in TZ_LIST" :key="tz" :value="tz">{{ tz }}</a-option>
                                </a-select>
                                <a-button type="primary" status="success" @click="dateToTs">
                                    转换
                                </a-button>
                                <a-input :model-value="dateResult" placeholder="转换结果" readonly style="width: 260px;">
                                    <template #append> <a-button :disabled="!dateResult" @click="copyText(dateResult)">
                                            <template #icon><icon-copy /></template>
                                        </a-button></template>
                                </a-input>
                            </a-space>
                        </a-col>
                    </a-row>
                </a-tab-pane>

                <!-- 批量转换 -->
                <a-tab-pane key="batch" title="批量转换">
                    <div class="field-label">批量时间戳（每行一个）</div>
                    <a-textarea v-model="batchInput" :auto-size="{ minRows: 6, maxRows: 12 }"
                        placeholder="1792556334&#10;1792556416638" allow-clear class="form-row" />
                    <a-space :size="12">
                        <a-select v-model="batchUnit">
                            <a-option value="s">秒 (s)</a-option>
                            <a-option value="ms">毫秒 (ms)</a-option>
                        </a-select>
                        <a-select v-model="batchTz">
                            <a-option v-for="tz in TZ_LIST" :key="tz" :value="tz">{{ tz }}</a-option>
                        </a-select>
                        <a-button type="primary" status="success" @click="batchConvert">
                            <template #icon><icon-swap /></template>批量转换
                        </a-button>
                        <a-button :disabled="!batchOutput" @click="copyText(batchOutput)">
                            <template #icon><icon-copy /></template>复制
                        </a-button>
                    </a-space>
                    <div class="field-label">转换结果</div>
                    <a-textarea v-model="batchOutput" :auto-size="{ minRows: 6, maxRows: 16 }" placeholder="转换结果"
                        readonly />
                </a-tab-pane>
            </a-tabs>
        </a-card>
    </layout-body>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from "vue";
import { Message } from "@arco-design/web-vue";

/** 常用时区列表 */
const TZ_LIST = [
    "Asia/Shanghai",
    "UTC",
    "Asia/Singapore",
    "Asia/Tokyo",
    "Asia/Kolkata",
    "Asia/Dubai",
    "Europe/London",
    "Europe/Paris",
    "Europe/Moscow",
    "America/New_York",
    "America/Chicago",
    "America/Los_Angeles",
    "Australia/Sydney",
];

/* ============ 时区转换核心（基于 Intl，纯前端实现） ============ */

/** 给定 UTC 毫秒，返回该时刻在指定时区相对 UTC 的偏移（分钟） */
function tzOffsetMinutes(utcMs: number, tz: string): number {
    const dtf = new Intl.DateTimeFormat("en-US", {
        timeZone: tz,
        hourCycle: "h23",
        year: "numeric",
        month: "2-digit",
        day: "2-digit",
        hour: "2-digit",
        minute: "2-digit",
        second: "2-digit",
    });
    const p: Record<string, string> = {};
    for (const part of dtf.formatToParts(new Date(utcMs))) {
        if (part.type !== "literal") p[part.type] = part.value;
    }
    const asUTC = Date.UTC(+p.year, +p.month - 1, +p.day, +p.hour, +p.minute, +p.second);
    return Math.round((asUTC - utcMs) / 60000);
}

/** 将 UTC 毫秒格式化为指定时区的 "YYYY-MM-DD HH:mm:ss" */
function formatInTz(utcMs: number, tz: string): string {
    const dtf = new Intl.DateTimeFormat("zh-CN", {
        timeZone: tz,
        hourCycle: "h23",
        year: "numeric",
        month: "2-digit",
        day: "2-digit",
        hour: "2-digit",
        minute: "2-digit",
        second: "2-digit",
    });
    const p: Record<string, string> = {};
    for (const part of dtf.formatToParts(new Date(utcMs))) {
        if (part.type !== "literal") p[part.type] = part.value;
    }
    return `${p.year}-${p.month}-${p.day} ${p.hour}:${p.minute}:${p.second}`;
}

/** 将指定时区的墙上时间 "YYYY-MM-DD HH:mm:ss" 解析为 UTC 毫秒 */
function dateTimeToTsMs(wall: string, tz: string): number | null {
    const m = wall.trim().match(/^(\d{4})-(\d{1,2})-(\d{1,2})[ T](\d{1,2}):(\d{1,2}):(\d{1,2})$/);
    if (!m) return null;
    const y = +m[1], mo = +m[2], d = +m[3], hh = +m[4], mm = +m[5], ss = +m[6];
    const wallAsUTC = Date.UTC(y, mo - 1, d, hh, mm, ss);
    // 两次校正以兼容夏令时边界
    let off = tzOffsetMinutes(wallAsUTC, tz);
    let utc = wallAsUTC - off * 60000;
    off = tzOffsetMinutes(utc, tz);
    utc = wallAsUTC - off * 60000;
    return utc;
}

/* ============ 复制到剪贴板 ============ */
const copyText = async (text: string) => {
    if (!text) return;
    try {
        await navigator.clipboard.writeText(text);
        Message.success("已复制到剪贴板");
    } catch {
        const ta = document.createElement("textarea");
        ta.value = text;
        ta.style.position = "fixed";
        ta.style.opacity = "0";
        document.body.appendChild(ta);
        ta.select();
        document.execCommand("copy");
        document.body.removeChild(ta);
        Message.success("已复制到剪贴板");
    }
};

/* ============ 当前时间戳 ============ */
const nowMs = ref(Date.now());
const nowUnit = ref<"ms" | "s">("ms");
const paused = ref(false);
let timer: ReturnType<typeof setInterval> | null = null;

const displayNow = computed(() =>
    nowUnit.value === "ms" ? String(nowMs.value) : String(Math.floor(nowMs.value / 1000))
);

const startTick = () => {
    stopTick();
    timer = setInterval(() => (nowMs.value = Date.now()), 100);
};
const stopTick = () => {
    if (timer) {
        clearInterval(timer);
        timer = null;
    }
};
const toggleUnit = () => (nowUnit.value = nowUnit.value === "ms" ? "s" : "ms");
const togglePause = () => {
    paused.value = !paused.value;
    paused.value ? stopTick() : startTick();
};
const copyNow = () => copyText(displayNow.value);

onMounted(startTick);
onUnmounted(stopTick);

/* ============ 时间戳 -> 日期时间 ============ */
const tsInput = ref("");
const tsUnit = ref<"s" | "ms">("s");
const tsTz = ref("Asia/Shanghai");
const tsResult = ref("");

const tsToDate = () => {
    const raw = tsInput.value.trim();
    if (!/^\d+$/.test(raw)) {
        tsResult.value = "";
        Message.warning("请输入合法的时间戳（纯数字）");
        return;
    }
    const ms = tsUnit.value === "s" ? Number(raw) * 1000 : Number(raw);
    tsResult.value = formatInTz(ms, tsTz.value);
};

const fillCurrent = () => {
    tsInput.value = nowUnit.value === "ms" ? String(nowMs.value) : String(Math.floor(nowMs.value / 1000));
    tsUnit.value = nowUnit.value;
    tsToDate();
};

/* ============ 日期时间 -> 时间戳 ============ */
const dateInput = ref("");
const dateTz = ref("Asia/Shanghai");
const dateUnit = ref<"s" | "ms">("s");
const dateResult = ref("");

const dateToTs = () => {
    const ms = dateTimeToTsMs(dateInput.value, dateTz.value);
    if (ms === null) {
        dateResult.value = "";
        Message.warning("请输入合法日期时间，格式 YYYY-MM-DD HH:mm:ss");
        return;
    }
    dateResult.value = dateUnit.value === "s" ? String(Math.floor(ms / 1000)) : String(ms);
};

/* ============ 批量转换 ============ */
const batchInput = ref("");
const batchUnit = ref<"s" | "ms">("s");
const batchTz = ref("Asia/Shanghai");
const batchOutput = ref("");

const batchConvert = () => {
    const lines = batchInput.value.split(/\r?\n/).map((s) => s.trim()).filter(Boolean);
    if (lines.length === 0) {
        batchOutput.value = "";
        Message.warning("请输入至少一个时间戳");
        return;
    }
    batchOutput.value = lines
        .map((line) => {
            if (!/^\d+$/.test(line)) return `${line}\t->\t无效的时间戳`;
            const ms = batchUnit.value === "s" ? Number(line) * 1000 : Number(line);
            return `${line}\t->\t${formatInTz(ms, batchTz.value)}`;
        })
        .join("\n");
};
</script>

<style scoped>
.page-title {
    margin: 0 0 var(--base-padding);
    font-size: var(--base-size-large-x);
    font-weight: 600;
    color: #1d2129;
}

.tool-card {
    margin-bottom: var(--base-padding);
}

.form-row {
    margin-bottom: var(--base-margin);
}

.now-label {
    color: #86909c;
    font-size: var(--base-size-small);
    margin-bottom: var(--base-radius-default);
}

.now-value {
    display: flex;
    align-items: center;
    gap: 10px;
}

.now-num {
    font-size: var(--base-size-large-xx);
    font-weight: 700;
    color: #1d2129;
    font-variant-numeric: tabular-nums;
}

.section-title {
    display: flex;
    align-items: center;
    gap: 8px;
    font-size: var(--base-size-default);
    font-weight: 600;
    color: #1d2129;
    margin: var(--base-radius-default) 0 var(--base-padding);
}

.field-label {
    color: #4e5969;
    font-size: var(--base-size-small);
    margin-bottom: var(--base-radius-default);
}
</style>
