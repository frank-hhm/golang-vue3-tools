<template>
    <a-row class="shot-page" :gutter="6">
        <!-- 左侧约 25%：截图操作 -->
        <a-col :span="6" class="side-col">
            <a-card class="side-card" size="small" :body-style="{ padding: 0 }">
                <template #title>截图操作</template>
                <template #extra>
                    <a-tag size="small" :color="image ? 'green' : 'gray'">
                        {{ image ? "已有截图" : "暂无截图" }}
                    </a-tag>
                </template>
                <div class="side-body">
                    <a-button type="primary" status="success" long @click="recapture">
                        <template #icon><icon-camera /></template>重新截图
                    </a-button>
                    <a-button long :disabled="!image" @click="copyImage">
                        <template #icon><icon-copy /></template>复制图片
                    </a-button>
                    <a-button long :disabled="!image" :loading="saving" @click="download">
                        <template #icon><icon-download /></template>保存 PNG
                    </a-button>
                    <div class="tip">
                        截图会先复制到系统剪贴板，可直接 Ctrl+V 粘贴；「保存 PNG」会存到程序目录 data/screenshots 下。
                    </div>
                </div>
            </a-card>
        </a-col>

        <!-- 右侧：预览 -->
        <a-col :span="18" class="main-col">
            <a-card class="main-panel" :body-style="{ padding: 0 }">
                <template #title>
                    <a-space :size="8">
                        <span>截图预览</span>
                        <a-tag size="small">PNG</a-tag>
                    </a-space>
                </template>
                <div class="preview-body">
                    <div class="preview-wrap">
                        <img v-if="image" :src="image" class="preview-img" alt="screenshot" />
                        <a-empty v-else description="暂无截图，点击左侧「重新截图」进行区域捕获" />
                    </div>
                </div>
            </a-card>
        </a-col>
    </a-row>
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
/* 页面根容器：a-row 铺满工具窗口内容区 */
.shot-page {
    height: 100%;
    min-height: 0;
    flex-wrap: nowrap;
    background: var(--color-fill-1);
}

/* 左侧操作卡片 */
.side-col {
    height: 100%;
    min-width: 0;
    display: flex;
    flex-direction: column;
}

.side-card {
    flex: 1;
    min-height: 0;
    display: flex;
    flex-direction: column;
}

.side-card :deep(.arco-card-body) {
    flex: 1;
    min-height: 0;
    display: flex;
    flex-direction: column;
    overflow: hidden;
}

.side-body {
    flex: 1;
    min-height: 0;
    display: flex;
    flex-direction: column;
    gap: var(--base-padding);
    padding: var(--base-padding);
    overflow-y: auto;
}

.tip {
    margin-top: auto;
    color: var(--color-text-3);
    font-size: var(--base-size-small);
    line-height: 1.8;
}

/* 右侧主列 */
.main-col {
    height: 100%;
    min-width: 0;
    display: flex;
    flex-direction: column;
}

.main-panel {
    width: 100%;
    height: 100%;
    flex: 1;
    min-width: 0;
    display: flex;
    flex-direction: column;
    background: var(--color-bg-1);
}

.main-panel :deep(.arco-card-body) {
    flex: 1;
    min-height: 0;
    display: flex;
    flex-direction: column;
    overflow: hidden;
}

.preview-body {
    flex: 1;
    min-height: 0;
    display: flex;
    padding: var(--base-padding);
}

/* 预览区随窗口铺满，图片等比缩放 */
.preview-wrap {
    flex: 1;
    min-height: 0;
    display: flex;
    align-items: center;
    justify-content: center;
    padding: var(--base-padding);
    border: 1px solid var(--color-border-2);
    border-radius: var(--base-radius-large);
    background: var(--color-fill-1);
    overflow: hidden;
}

.preview-img {
    max-width: 100%;
    max-height: 100%;
    object-fit: contain;
    border-radius: var(--base-radius-default);
}
</style>
