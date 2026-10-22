import { App, Plugin } from 'vue';
import LayoutBody from '@/components/layouts/LayoutBody.vue'
import selectIconModal from '@/components/select-modal/icon/index.vue'

export const ComponentsPlugin: Plugin = {
    install(app: App) {
        // select-icon-modal：图标选择弹层，通过 type 属性切换图标源
        //  - type="iconfont"（默认）：本地字体图标，emit `icon-xxx`
        //  - type="arco"：Arco Design 图标库，emit 图标名（如 camera / clock-circle）
        app.component('select-icon-modal', selectIconModal);
        app.component('layout-body', LayoutBody);

    },
};
export {  LayoutBody , selectIconModal};