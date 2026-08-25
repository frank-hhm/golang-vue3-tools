import BasicLayout from "@/components/layouts/index.vue";

export default [{
    path: '/',
    name: '/',
    header: '/',
    redirect: {
        path: `/index`
    },
    children: [
        {
            path: '/index',
            name: '/index',
            meta: {
                auth: false,
                menu_name: "首页",
            },
            component: () => import('@/views/index/index.vue')
        },
        {
            path: '/tools/sniffer',
            name: '/toolsSniffer',
            meta: {
                auth: false,
                menu_name: "抓包工具",
            },
            component: () => import('@/views/tools/sniffer/index.vue')
        }
    ]
}]
