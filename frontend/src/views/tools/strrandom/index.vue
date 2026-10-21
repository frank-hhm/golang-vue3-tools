<template>
    <layout-body title="随机字符串生成">
        <a-card class="tool-card">
            <!-- 长度：支持固定长度或区间随机 -->
            <a-row :gutter="12" align="center" class="form-row">
                <a-col :span="3"><span class="lbl">长度</span></a-col>
                <a-col :span="21">
                    <a-space :size="12">
                        <a-radio-group v-model="lenMode" type="button" size="small">
                            <a-radio value="fixed">固定</a-radio>
                            <a-radio value="range">区间</a-radio>
                        </a-radio-group>
                        <template v-if="lenMode === 'fixed'">
                            <a-input-number v-model="fixedLen" :min="1" :max="1000" :precision="0"
                                style="width: 130px" />
                        </template>
                        <template v-else>
                            <a-input-number v-model="minLen" :min="1" :max="1000" :precision="0"
                                style="width: 130px" />
                            <span class="tilde">~</span>
                            <a-input-number v-model="maxLen" :min="1" :max="1000" :precision="0"
                                style="width: 130px" />
                        </template>
                    </a-space>
                </a-col>
            </a-row>

            <!-- 字符集 -->
            <a-row :gutter="12" align="center" class="form-row">
                <a-col :span="3"><span class="lbl">字符集</span></a-col>
                <a-col :span="21">
                    <a-input v-model="charset" placeholder="生成时使用的字符集，可手动编辑" allow-clear />
                </a-col>
            </a-row>

            <!-- 字符类型勾选 -->
            <a-row class="form-row">
                <a-col :span="3"></a-col>
                <a-col :span="21">
                    <a-space direction="vertical" :size="10">
                        <a-checkbox v-model="useDigits">数字 0123456789</a-checkbox>
                        <a-checkbox v-model="useLower">小写字母 abcdefghijklmnopqrstuvwxyz</a-checkbox>
                        <a-checkbox v-model="useUpper">大写字母 ABCDEFGHIJKLMNOPQRSTUVWXYZ</a-checkbox>
                        <a-checkbox v-model="useSymbols">常用符号 ~!@#$%^&amp;*()_+</a-checkbox>
                    </a-space>
                </a-col>
            </a-row>

            <!-- 生成数量 -->
            <a-row :gutter="12" align="center" class="form-row">
                <a-col :span="3"><span class="lbl">生成数量</span></a-col>
                <a-col :span="21">
                    <a-space :size="12">
                        <a-input-number v-model="count" :min="1" :max="1000" :precision="0" />
                        <a-button type="primary" status="success" @click="generate">
                            <template #icon><icon-refresh /></template>生成
                        </a-button>
                        <a-button :disabled="!result" @click="copyText(result)">
                            <template #icon><icon-copy /></template>复制
                        </a-button>
                    </a-space>
                </a-col>
            </a-row>

            <!-- 结果 -->
            <div class="field-label">结果</div>
            <a-textarea v-model="result" :auto-size="{ minRows: 6, maxRows: 16 }" placeholder="结果" readonly />
        </a-card>
    </layout-body>
</template>

<script setup lang="ts">
import { ref, watch } from "vue";
import { Message } from "@arco-design/web-vue";

/** 各类字符集常量 */
const DIGITS = "0123456789";
const LOWER = "abcdefghijklmnopqrstuvwxyz";
const UPPER = "ABCDEFGHIJKLMNOPQRSTUVWXYZ";
const SYMBOLS = "~!@#$%^&*()_+";

/** 表单状态 */
const lenMode = ref<"fixed" | "range">("fixed");
const fixedLen = ref(16);
const minLen = ref(16);
const maxLen = ref(20);
const count = ref(5);
const charset = ref("");
const result = ref("");

const useDigits = ref(true);
const useLower = ref(true);
const useUpper = ref(true);
const useSymbols = ref(true);

/** 根据勾选项重建字符集（手动编辑字符集后再改动勾选会覆盖，属预期行为） */
const rebuildCharset = () => {
    let s = "";
    if (useDigits.value) s += DIGITS;
    if (useLower.value) s += LOWER;
    if (useUpper.value) s += UPPER;
    if (useSymbols.value) s += SYMBOLS;
    charset.value = s;
};

watch([useDigits, useLower, useUpper, useSymbols], rebuildCharset);
rebuildCharset();

/** 生成 [0, maxExclusive) 区间内的安全随机整数（优先使用 crypto） */
function randomInt(maxExclusive: number): number {
    if (maxExclusive <= 0) return 0;
    if (typeof crypto !== "undefined" && crypto.getRandomValues) {
        const range = 0x100000000;
        const limit = range - (range % maxExclusive);
        const arr = new Uint32Array(1);
        let x = 0;
        do {
            crypto.getRandomValues(arr);
            x = arr[0];
        } while (x >= limit);
        return x % maxExclusive;
    }
    return Math.floor(Math.random() * maxExclusive);
}

/** 生成随机字符串 */
const generate = () => {
    const pool = Array.from(charset.value);
    if (pool.length === 0) {
        result.value = "";
        Message.warning("字符集不能为空");
        return;
    }
    let min: number;
    let max: number;
    if (lenMode.value === "fixed") {
        min = Math.floor(fixedLen.value) || 1;
        if (min < 1) min = 1;
        max = min;
    } else {
        min = Math.floor(minLen.value) || 1;
        max = Math.floor(maxLen.value) || 1;
        if (min < 1) min = 1;
        if (max < min) {
            Message.warning("最大长度不能小于最小长度");
            return;
        }
    }
    const n = Math.min(Math.max(1, Math.floor(count.value) || 1), 1000);

    const out: string[] = [];
    for (let i = 0; i < n; i++) {
        const len = min + randomInt(max - min + 1);
        let s = "";
        for (let j = 0; j < len; j++) {
            s += pool[randomInt(pool.length)];
        }
        out.push(s);
    }
    result.value = out.join("\n");
};

/** 复制到剪贴板 */
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
</script>

<style scoped>
.tool-card {
    margin-bottom: 16px;
}

.form-row {
    margin-bottom: 12px;
}

.lbl {
    color: #4e5969;
    font-size: 14px;
}

.tilde {
    text-align: center;
    color: #86909c;
}

.field-label {
    color: #4e5969;
    font-size: 13px;
    margin-bottom: 6px;
}
</style>
