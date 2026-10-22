<template>
    <layout-body title="进制转换">
        <!-- 输入进制选择 -->
        <a-card class="tool-card">
            <a-row>
                <a-col :span="24" class="mb12">
                    <a-space :size="8">
                        <a-radio-group v-model="inputBase" type="radio">
                            <a-radio v-for="b in COMMON_BASES" :key="b" :value="b">{{ b }}进制</a-radio>
                        </a-radio-group>
                        <a-divider direction="vertical" />
                        <span class="field-label" style="margin: 0">更多：</span>
                        <a-select v-model="inputBase" style="width: 96px">
                            <a-option v-for="b in ALL_BASES" :key="b" :value="b">{{ b }}</a-option>
                        </a-select>
                    </a-space>
                </a-col>
                <a-col :span="24">
                    <a-row :gutter="12">
                        <a-col :span="18">
                            <a-input v-model="inputStr" :placeholder="`请输入 ${inputBase} 进制数`" allow-clear
                                @press-enter="convert" />
                        </a-col>
                        <a-col :span="6">
                            <a-space>
                                <a-button type="primary" status="success" long @click="convert">转换</a-button>
                            </a-space>
                        </a-col>
                    </a-row>
                </a-col>
            </a-row>
        </a-card>

        <!-- 转换结果表 -->
        <a-card class="tool-card">
            <a-table :data="rows" :pagination="false" :loading="false" row-key="base"
                :row-class="(record) => (record.base === inputBase ? 'row-active' : '')">
                <template #columns>
                    <a-table-column title="进制" data-index="base" :width="80" />
                    <a-table-column title="结果">
                        <template #cell="{ record }">
                            <a-input :model-value="record.value" readonly @click="selectAll">
                                <template #append>
                                    <a-button :disabled="!record.value" @click="copyText(record.value)">
                                        <template #icon><icon-copy /></template>
                                    </a-button>
                                </template>
                            </a-input>
                            <!-- 二进制专属：补位 + 字节序 -->
                            <div v-if="record.base === 2" class="bin-opts">
                                <a-space wrap :size="12">
                                    <a-checkbox v-model="padEnabled">补0至</a-checkbox>
                                    <a-select v-model="padBits" :disabled="!padEnabled" style="width: 96px"
                                        size="small">
                                        <a-option :value="8">8位</a-option>
                                        <a-option :value="16">16位</a-option>
                                        <a-option :value="32">32位</a-option>
                                        <a-option :value="64">64位</a-option>
                                    </a-select>
                                    <a-radio-group v-model="endian" size="small">
                                        <a-radio value="big">大端序</a-radio>
                                        <a-radio value="little">小端序</a-radio>
                                    </a-radio-group>
                                </a-space>
                            </div>
                        </template>
                    </a-table-column>
                    <a-table-column title="解释" data-index="explain" :width="260" />
                </template>
            </a-table>
        </a-card>

        <!-- 简介 -->
        <a-card class="tool-card" title="简介">
            <p class="intro">
                进制其实是一种记数的方法，所以也称为进位记数法/值计数法，可以用有限的数字符号代表所有的数值。
                可使用数字符号的数目称为基数（英文：radix）或底数，基数为 n，即可称 n 进制，简称 n 进制。
                例如平常生活中我们经常用到的十进制，就是使用 10 个阿拉伯数字 0-9 进行记数，所以它的基数就是 10，称为十进制。
            </p>
        </a-card>
    </layout-body>
</template>

<script setup lang="ts">
import { computed, ref } from "vue";
import { Message } from "@arco-design/web-vue";

/** 常用输入进制（单选项） */
const COMMON_BASES = [2, 8, 10, 16, 32, 36, 58, 62];
/** 全部可选进制（更多下拉：2 ~ 62） */
const ALL_BASES = Array.from({ length: 61 }, (_, i) => i + 2);
/** 结果表展示的目标进制 */
const TARGET_BASES = [2, 8, 10, 16, 26, 32, 36, 52, 58, 62];

/** 各进制的解释 */
const EXPLAIN: Record<number, string> = {
    26: "小写字母",
    32: "数字 + 大写字母，不包含 ILOU 字符",
    36: "数字 + 小写字母",
    52: "大小写字母",
    58: "数字 + 大小写字母，不包含 0O1I 字符",
    62: "数字 + 大小写字母",
};

/* ============ 状态 ============ */
const inputBase = ref(62);
const inputStr = ref("");
const resultValue = ref<bigint | null>(null);

/** 二进制专属选项 */
const padEnabled = ref(false);
const padBits = ref(32);
const endian = ref<"big" | "little">("big");

/* ============ 字符集（不同进制采用业界常用字母表） ============ */
function getAlphabet(base: number): string {
    switch (base) {
        case 16:
            return "0123456789abcdef";
        case 26:
            return "abcdefghijklmnopqrstuvwxyz";
        case 32:
            return "0123456789ABCDEFGHJKMNPQRSTVWXYZ"; // Crockford Base32
        case 36:
            return "0123456789abcdefghijklmnopqrstuvwxyz";
        case 52:
            return "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ";
        case 58:
            return "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"; // Bitcoin Base58
        case 62:
            return "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz";
        default:
            return "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ".slice(0, base);
    }
}

/** 字符串 -> BigInt（按输入进制解析） */
function parseToBigInt(str: string, base: number): bigint | null {
    const alpha = getAlphabet(base);
    const map = new Map<string, number>();
    for (let i = 0; i < alpha.length; i++) map.set(alpha[i], i);
    // 小写字母表进制（如 16/36）额外接受大写输入
    for (const [k, v] of Array.from(map)) {
        if (k >= "a" && k <= "z") map.set(k.toUpperCase(), v);
    }
    let s = str.trim();
    if (!s) return null;
    let neg = false;
    if (s[0] === "-") {
        neg = true;
        s = s.slice(1);
    }
    if (!s) return null;
    const b = BigInt(base);
    let result = 0n;
    for (const ch of s) {
        const v = map.get(ch);
        if (v === undefined) return null;
        result = result * b + BigInt(v);
    }
    return neg ? -result : result;
}

/** BigInt -> 指定进制字符串 */
function formatBigInt(value: bigint, base: number): string {
    const alpha = getAlphabet(base);
    if (value === 0n) return alpha[0];
    const neg = value < 0n;
    let v = neg ? -value : value;
    const b = BigInt(base);
    let out = "";
    while (v > 0n) {
        out = alpha[Number(v % b)] + out;
        v /= b;
    }
    return (neg ? "-" : "") + out;
}

/** 二进制结果（含补位与字节序处理） */
function formatBinary(value: bigint): string {
    let bin = formatBigInt(value, 2);
    if (padEnabled.value && bin.length < padBits.value) {
        bin = bin.padStart(padBits.value, "0");
    }
    if (endian.value === "little") {
        let b = bin;
        const rem = b.length % 8;
        if (rem) b = b.padStart(b.length + (8 - rem), "0");
        const bytes = b.match(/.{8}/g) || [];
        bin = bytes.reverse().join("");
    }
    return bin;
}

/* ============ 表格数据（响应式） ============ */
const rows = computed(() =>
    TARGET_BASES.map((base) => {
        let value = "";
        if (resultValue.value !== null) {
            value = base === 2 ? formatBinary(resultValue.value) : formatBigInt(resultValue.value, base);
        }
        return { base, value, explain: EXPLAIN[base] ?? "" };
    })
);

/* ============ 交互 ============ */
const convert = () => {
    const v = parseToBigInt(inputStr.value, inputBase.value);
    if (v === null) {
        resultValue.value = null;
        Message.error(`请输入合法的 ${inputBase.value} 进制数`);
        return;
    }
    resultValue.value = v;
};

const selectAll = (e: Event) => (e.target as HTMLInputElement)?.select?.();

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
    margin-bottom: var(--base-space-lg)$3
}

.field-label {
    color: var(--color-text-2);
    font-size: var(--base-size-medium);
}

.bin-opts {
    margin-top: var(--base-space-md)$3
}

.intro {
    color: var(--color-text-2);
    font-size: var(--base-size-medium);
    line-height: 1.8;
    margin: 0;
}

:deep(.row-active) {
    background-color: rgba(var(--cyan-6), 0.1);
}
</style>
