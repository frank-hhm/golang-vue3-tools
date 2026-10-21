<template>
    <layout-body title="图片base64编码">
        <!-- 上传区 -->
        <a-card class="tool-card">
            <div class="dropzone" :class="{ 'is-over': dragOver }" @click="pickFile" @dragover.prevent="dragOver = true"
                @dragleave.prevent="dragOver = false" @drop.prevent="onDrop">
                <div class="drop-text">
                    拖放图片到这里上传，或者
                    <span class="drop-link">选择图片</span>
                </div>
                <input ref="fileInput" type="file" accept="image/*" style="display: none" @change="onFileChange" />
            </div>
        </a-card>

        <!-- 结果区 -->
        <a-card v-if="base64" class="tool-card">
            <a-row :gutter="16">
                <a-col :span="6">
                    <div class="field-label">预览</div>
                    <img :src="dataUri" class="preview" alt="preview" />
                    <div class="meta">
                        <div><span class="meta-k">文件名</span>{{ fileName }}</div>
                        <div><span class="meta-k">类型</span>{{ mimeType }}</div>
                        <div><span class="meta-k">大小</span>{{ fileSize }}</div>
                    </div>
                </a-col>
                <a-col :span="18">
                    <div class="flex between align-center">
                        <div class="field-label" style="margin: 0">Base64 编码</div>
                        <a-space :size="12">
                            <a-switch v-model="withPrefix" type="round">
                                <template #checked>含 data 前缀</template>
                                <template #unchecked>仅纯编码</template>
                            </a-switch>
                            <a-button type="primary" size="small" status="success" @click="copyText(output)">
                                <template #icon><icon-copy /></template>复制
                            </a-button>
                            <a-button @click="reset" size="small">
                                <template #icon><icon-delete /></template>清空
                            </a-button>
                        </a-space>
                    </div>
                    <a-textarea class="mt20" :model-value="output" :auto-size="{ minRows: 8, maxRows: 18 }" readonly />
                </a-col>
            </a-row>
        </a-card>
    </layout-body>
</template>

<script setup lang="ts">
import { computed, ref } from "vue";
import { Message } from "@arco-design/web-vue";

const fileInput = ref<HTMLInputElement | null>(null);
const dragOver = ref(false);
const withPrefix = ref(true);

const fileName = ref("");
const mimeType = ref("");
const fileSize = ref("");
const base64 = ref("");

/** 带前缀的完整 data URI（用于预览） */
const dataUri = computed(() =>
    base64.value ? `data:${mimeType.value};base64,${base64.value}` : ""
);

/** 实际输出（按开关决定是否含前缀） */
const output = computed(() =>
    base64.value
        ? withPrefix.value
            ? `data:${mimeType.value};base64,${base64.value}`
            : base64.value
        : ""
);

function formatSize(bytes: number): string {
    if (bytes < 1024) return `${bytes} B`;
    if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(2)} KB`;
    return `${(bytes / 1024 / 1024).toFixed(2)} MB`;
}

const handleFile = (file: File | undefined) => {
    if (!file) return;
    if (!file.type.startsWith("image/")) {
        Message.warning("请选择图片文件");
        return;
    }
    fileName.value = file.name;
    mimeType.value = file.type;
    fileSize.value = formatSize(file.size);
    const reader = new FileReader();
    reader.onload = () => {
        const result = String(reader.result || "");
        // readAsDataURL 形如 data:image/png;base64,xxxx，取逗号后的纯编码
        const comma = result.indexOf(",");
        base64.value = comma >= 0 ? result.slice(comma + 1) : result;
    };
    reader.onerror = () => Message.error("读取文件失败");
    reader.readAsDataURL(file);
};

const pickFile = () => fileInput.value?.click();

const onFileChange = (e: Event) => {
    const input = e.target as HTMLInputElement;
    handleFile(input.files?.[0]);
    input.value = ""; // 允许重复选择同一文件
};

const onDrop = (e: DragEvent) => {
    dragOver.value = false;
    const file = e.dataTransfer?.files?.[0];
    handleFile(file);
};

const reset = () => {
    fileName.value = "";
    mimeType.value = "";
    fileSize.value = "";
    base64.value = "";
};

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

.field-label {
    color: #4e5969;
    font-size: 13px;
    margin-bottom: 6px;
}

.dropzone {
    border: 1px dashed #c9cdd4;
    border-radius: 8px;
    padding: 60px 20px;
    text-align: center;
    cursor: pointer;
    transition: border-color 0.2s, background-color 0.2s;
}

.dropzone:hover,
.dropzone.is-over {
    border-color: rgb(var(--green-6));
    background-color: var(--color-fill-2);
}

.drop-text {
    color: #4e5969;
    font-size: 14px;
}

.drop-link {
    color: rgb(var(--green-6));
}

.preview {
    max-width: 100%;
    border: 1px solid #e5e6eb;
    border-radius: 8px;
}

.meta {
    margin-top: 12px;
    color: #4e5969;
    font-size: 13px;
    line-height: 1.9;
}

.meta-k {
    display: inline-block;
    width: 56px;
    color: #86909c;
}
</style>
