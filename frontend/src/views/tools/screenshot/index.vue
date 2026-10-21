<template>
    <layout-body title="截图预览">
        <a-card class="tool-card">
            <a-row justify="space-between" align="center" class="form-row">
                <a-col>
                    <span class="field-label" style="margin: 0">截图结果</span>
                </a-col>
                <a-col>
                    <a-space :size="12">
                        <a-button type="primary" status="success" @click="recapture">
                            <template #icon><icon-camera /></template>重新截图
                        </a-button>
                        <a-button :disabled="!image" @click="copyImage">
                            <template #icon><icon-copy /></template>复制图片
                        </a-button>
                        <a-button :disabled="!image" @click="download">
                            <template #icon><icon-download /></template>下载 PNG
                        </a-button>
                    </a-space>
                </a-col>
            </a-row>

            <div class="preview-wrap">
                <img v-if="image" :src="image" class="preview-img" alt="screenshot" />
                <a-empty v-else description="暂无截图，点击「重新截图」进行区域捕获" />
            </div>
            <div class="tip">截图会先复制到系统剪贴板，可直接 Ctrl+V 粘贴；也可「复制图片」或「下载 PNG」。</div>
        </a-card>
    </layout-body>
</template>

<script setup lang="ts">
import { getCurrentInstance, onMounted, ref } from "vue";
import { Message } from "@arco-design/web-vue";

const {
    proxy: { $GoTools },
} = getCurrentInstance() as any;

/** 与首页事件约定的共享存储键（同源多窗口共享 localStorage） */
const STORAGE_KEY = "tool:screenshot";

const image = ref("");

onMounted(() => {
    try {
        image.value = localStorage.getItem(STORAGE_KEY) || "";
    } catch {
        /* localStorage 不可用时忽略 */
    }
});

/** 重新触发区域截图并刷新预览 */
const recapture = async () => {
    try {
        const data = await $GoTools.CaptureSelectedRegion("png");
        if (!data) return; // 用户取消
        image.value = data;
        try {
            localStorage.setItem(STORAGE_KEY, data);
        } catch {
            /* ignore */
        }
        Message.success("截图完成");
    } catch (e: any) {
        Message.error("截图失败：" + (e?.message || e));
    }
};

/** 下载为 PNG 文件 */
const download = () => {
    if (!image.value) return;
    const a = document.createElement("a");
    a.href = image.value;
    a.download = `screenshot-${Date.now()}.png`;
    a.click();
};

/** 将图片本身写入剪贴板 */
const copyImage = async () => {
    if (!image.value) return;
    try {
        const blob = await (await fetch(image.value)).blob();
        await navigator.clipboard.write([new ClipboardItem({ [blob.type]: blob })]);
        Message.success("图片已复制到剪贴板");
    } catch (e: any) {
        Message.error("复制失败：" + (e?.message || e));
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
}

.preview-wrap {
    border: 1px solid #e5e6eb;
    border-radius: 8px;
    padding: 12px;
    min-height: 240px;
    display: flex;
    align-items: center;
    justify-content: center;
    background: var(--color-fill-1);
}

.preview-img {
    max-width: 100%;
    border-radius: 4px;
}

.tip {
    margin-top: 10px;
    color: #86909c;
    font-size: 13px;
}
</style>
