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
                        <a-button :disabled="!image" :loading="saving" @click="download">
                            <template #icon><icon-download /></template>保存 PNG
                        </a-button>
                    </a-space>
                </a-col>
            </a-row>

            <div class="preview-wrap">
                <img v-if="image" :src="image" class="preview-img" alt="screenshot" />
                <a-empty v-else description="暂无截图，点击「重新截图」进行区域捕获" />
            </div>
            <div class="tip">截图会先复制到系统剪贴板，可直接 Ctrl+V 粘贴；「保存 PNG」会存到程序目录 data/screenshots 下。</div>
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

/** 保存 PNG 到程序目录下 data/screenshots */
const saving = ref(false);
const download = async () => {
    if (!image.value) return;
    saving.value = true;
    try {
        const dir = await $GoTools.AppDataDir("screenshots");
        const path = dir.replace(/[\\/]+$/, "") + "\\screenshot-" + Date.now() + ".png";
        await $GoTools.SaveImage(path, image.value);
        Message.success("已保存：" + path);
    } catch (e: any) {
        Message.error("保存失败：" + (e?.message || e));
    }
    saving.value = false;
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
    margin-bottom: var(--base-padding);
}

.form-row {
    margin-bottom: var(--base-margin);
}

.field-label {
    color: var(--color-text-2);
    font-size: var(--base-size-small);
}

.preview-wrap {
    border: 1px solid var(--color-border-2);
    border-radius: var(--base-radius-large);
    padding: var(--base-padding);
    min-height: 240px;
    display: flex;
    align-items: center;
    justify-content: center;
    background: var(--color-fill-1);
}

.preview-img {
    max-width: 100%;
    border-radius: var(--base-radius-default);
}

.tip {
    margin-top: var(--base-radius-large);
    color: var(--color-text-3);
    font-size: var(--base-size-small);
}
</style>
