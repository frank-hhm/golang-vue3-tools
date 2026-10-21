<template>
    <!-- 空白布局（原生新窗口）：不渲染卡片/标题，直接输出内容 -->
    <slot v-if="blank"></slot>
    <div v-else class="layout-body-main" :style="styles">

        <div class="layout-body-main-header" v-if="title">
            <div class="layout-body-main-header-title">{{ title }}</div>
            <div>
                <slot name="right"></slot>
            </div>
        </div>
        <div class="layout-body-main-content">
            <slot></slot>
        </div>
    </div>
</template>
<script lang="ts">
export default {
    name: "LayoutBody",
};
</script>
<script lang="ts" setup>
import { computed, StyleValue } from "vue";
import { useRoute } from "vue-router";

const route = useRoute();

/** 当前路由是否为“空白布局”（原生新窗口），是则不渲染 LayoutBody 卡片 */
const blank = computed(() => !!route.meta?.blank);

const props = withDefaults(
    defineProps<{
        height?: string | boolean;
        title?: string;
    }>(),
    {
        height: false,
        title: "",
    }
);

const styles = computed<StyleValue>(() => {
    return [
        props.height ? `height:calc(${props.height} - 2px)` : "",
    ];
});
</script>
<style scoped>
.layout-body-main {
    position: relative;
    background: var(--color-bg-2);
    border-radius: var(--base-radius-default);
}

.layout-body-main-header {
    min-height: 20px;
    border-bottom: 1px solid var(--color-border-1);
    padding: var(--base-padding);
    display: flex;
    align-items: center;
    justify-content: space-between;
}

.layout-body-main-header-title {
    /* font-size: var(--s); */
}

.layout-body-main-content {
    padding: var(--base-padding);
}
</style>