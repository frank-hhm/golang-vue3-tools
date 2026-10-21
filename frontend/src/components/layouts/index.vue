<template>
    <div class="basic-layout">
        <div class="layout-header">
            <a-breadcrumb class="layout-breadcrumb">
                <a-breadcrumb-item v-for="(item, index) in breadcrumbs" :key="index"
                    :class="{ clickable: index < breadcrumbs.length - 1 }"
                    @click="index < breadcrumbs.length - 1 && router.push(item.path === '/' ? '/index' : item.path)">
                    {{ item.title }}
                </a-breadcrumb-item>
            </a-breadcrumb>
            <a-button v-if="showBack" size="small" class="ml10" @click="goBack">
                返回
            </a-button>
            <a-button v-else size="small" class="ml10" @click="onRefresh">
                刷新
            </a-button>
        </div>
        <div class="layout-content">
            <router-view />
        </div>
    </div>
</template>
<script setup lang="ts">
import router from "@/router";
import { computed } from "vue";
import { useRoute } from "vue-router";

const route = useRoute();

/** 自行维护访问来源记录，替代不可靠的浏览器历史 */
const historyStack: string[] = [];
// 返回时跳过来源记录
let isGoingBack = false;

router.beforeEach((to, from) => {
    if (isGoingBack) return;
    // 忽略初始导航（from 为根占位路由），并对来源去重，避免重复回退同一页
    if (from.path && from.name !== "/" && historyStack[historyStack.length - 1] !== from.path) {
        historyStack.push(from.path);
    }
});

/** 从路由 matched 中提取面包屑（取带 menu_name 的层级，连续同名去重） */
const breadcrumbs = computed(() => {
    const list: { title: string; path: string }[] = [];
    route.matched
        .filter((item) => item.meta?.menu_name)
        .forEach((item) => {
            const title = item.meta.menu_name as string;
            if (list[list.length - 1]?.title !== title) {
                list.push({ title, path: item.path });
            }
        });
    return list;
});

/** 首页不显示返回按钮 */
const showBack = computed(() => route.path !== "/index");

const goBack = () => {
    // 弹出来源页面，没有记录时回首页
    const prev = historyStack.pop();
    const target = prev && prev !== route.path ? prev : "/index";
    // 屏蔽本次返回自身产生的来源记录
    isGoingBack = true;
    router.push(target).finally(() => {
        isGoingBack = false;
    });
};

/** 刷新当前页面 */
const onRefresh = () => {
    router.replace({ path: "/index", query: { t: Date.now() } });
};

</script>
<style scoped>
.basic-layout {
    min-height: 100vh;
    display: flex;
    flex-direction: column;
}

.layout-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    height: 50px;
    padding: 0 20px;
    background-color: #fff;
    border-bottom: 1px solid #e5e6eb;
}

.layout-breadcrumb :deep(.clickable) {
    cursor: pointer;
}

.layout-content {

    left: 0;
    top: 50px;
    position: fixed;
    width: calc(100% - (2*var(--base-padding)));
    height: calc(100% - (2*var(--base-padding)) - 51px);
    padding: var(--base-padding);
    background: var(--color-bg-body);
    overflow-y: scroll;
}
</style>
