import BasicLayout from "@/components/layouts/index.vue";

export default [{
    path: '/',
    name: '/',
    header: '/',
    meta: {
        menu_name: "首页",
    },
    redirect: {
        path: `/index`
    },
    component: BasicLayout,
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
        },
        {
            path: '/tools/usbserial',
            name: '/toolsUsbserial',
            meta: {
                auth: false,
                menu_name: "USB调试",
            },
            component: () => import('@/views/tools/usbserial/index.vue')
        },
        {
            path: '/tools/timestamp',
            name: '/toolsTimestamp',
            meta: {
                auth: false,
                menu_name: "时间戳转换",
            },
            component: () => import('@/views/tools/timestamp/index.vue')
        },
        {
            path: '/tools/strrandom',
            name: '/toolsStrrandom',
            meta: {
                auth: false,
                menu_name: "随机字符串生成",
            },
            component: () => import('@/views/tools/strrandom/index.vue')
        },
        {
            path: '/tools/encdec',
            name: '/toolsEncdec',
            meta: {
                auth: false,
                menu_name: "文字加密解密",
            },
            component: () => import('@/views/tools/encdec/index.vue')
        }
    ]
}]
