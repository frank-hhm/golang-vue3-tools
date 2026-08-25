import { createApp } from 'vue'
import App from "@/App.vue"
import router from '@/router'
import "@/router/permission"

// load
import { loadPlugins } from "@/plugins"
//css
import "@/styles/index.scss"

import * as Utils from '@/utils'

import * as Wailsjs from '@/wailsjs/go/main/App'

import * as WailsjsRun from '@/wailsjs/runtime/runtime'

const app = createApp(App);

/** 加载插件 */
loadPlugins(app)

app.config.globalProperties.$utils = Utils;
app.config.globalProperties.$wailsjs = Wailsjs;
app.config.globalProperties.$wailsRun = WailsjsRun;

router.isReady().then(() => {
    app.mount("#app")
}).catch(err => {
    console.log(err)
})

