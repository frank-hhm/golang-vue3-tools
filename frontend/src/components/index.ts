import { App, Plugin } from 'vue';
import LayoutBody from '@/components/layouts/LayoutBody.vue'

export const ComponentsPlugin: Plugin = {
    install(app: App) {
        app.component('layout-body', LayoutBody);

    },
};
export {  LayoutBody};