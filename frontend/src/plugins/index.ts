import { type App } from "vue"
import { loadStorePlus } from "./store"
import { loadRouterPlus } from "./router"
import { loadArcoPlus } from "./arco"
import { loadArcoIconPlus } from "./arco-icon"

export function loadPlugins(app: App) {
    loadStorePlus(app);
    loadRouterPlus(app)
    loadArcoPlus(app)
    loadArcoIconPlus(app)
}
