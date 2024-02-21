import { createApp } from 'vue'
import App from './App.vue'
import router from './router'

import '@/assets/main.css'
import '@/assets/tailwind.css'

//引入css
import 'xterm/css/xterm.css'
import 'element-plus/dist/index.css'

//引入组件
import ElementPlus from 'element-plus'
import zhCn from 'element-plus/dist/locale/zh-cn.mjs'
import * as ElementPlusIconsVue from '@element-plus/icons-vue'

//悬浮球
import FloatingBall from 'vue3-floating-ball';

const app = createApp(App)

//注册组件
for (const [key, component] of (<any>Object).entries(ElementPlusIconsVue)) {
    app.component(key, component)
}

app.use(FloatingBall)
app.use(ElementPlus, { locale: zhCn })
app.use(router)

app.mount('#app')
