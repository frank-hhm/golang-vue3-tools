import BasicLayout from "@/components/layouts/index.vue";
import BlankLayout from "@/components/layouts/BlankLayout.vue";

/** 工具路由定义（相对路径，供两种父布局复用） */
const TOOLS = [
    { key: "Sniffer", menu_name: "抓包工具", path: "tools/sniffer", component: () => import('@/views/tools/sniffer/index.vue') },
    { key: "Usbserial", menu_name: "USB调试", path: "tools/usbserial", component: () => import('@/views/tools/usbserial/index.vue') },
    { key: "Timestamp", menu_name: "时间戳转换", path: "tools/timestamp", component: () => import('@/views/tools/timestamp/index.vue') },
    { key: "Strrandom", menu_name: "随机字符串生成", path: "tools/strrandom", component: () => import('@/views/tools/strrandom/index.vue') },
    { key: "Encdec", menu_name: "文字加密解密", path: "tools/encdec", component: () => import('@/views/tools/encdec/index.vue') },
    { key: "Hexconvert", menu_name: "进制转换", path: "tools/hexconvert", component: () => import('@/views/tools/hexconvert/index.vue') },
    { key: "ImageBase64", menu_name: "图片base64编码", path: "tools/imageBase64", component: () => import('@/views/tools/imageBase64/index.vue') },
    { key: "Qrcode", menu_name: "二维码生成", path: "tools/qrcode", component: () => import('@/views/tools/qrcode/index.vue') },
    { key: "Websocket", menu_name: "WebSocket调试", path: "tools/websocket", component: () => import('@/views/tools/websocket/index.vue') },
    { key: "Screenshot", menu_name: "截图预览", path: "tools/screenshot", component: () => import('@/views/tools/screenshot/index.vue') },
];

/**
 * 生成子路由。
 * @param win true 时用于原生新窗口的空白布局（meta.blank，无面包屑、LayoutBody 裸渲染）
 */
const buildChildren = (win: boolean) =>
    TOOLS.map((t) => ({
        path: t.path,
        name: (win ? "/win/tools" : "/tools") + t.key,
        meta: {
            auth: false,
            menu_name: t.menu_name,
            ...(win ? { blank: true } : {}),
        },
        component: t.component,
    }));

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
        ...buildChildren(false)
    ]
}, {
    // 原生新窗口专用：空白布局，无面包屑与返回栏
    path: '/win',
    name: '/win',
    meta: {
        blank: true,
    },
    component: BlankLayout,
    children: buildChildren(true)
}]
