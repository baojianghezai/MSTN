import { createApp } from 'vue'
import ElementPlus from 'element-plus'
import 'element-plus/dist/index.css'
import 'uno.css' // UnoCSS 生成的原子类样式
import { createPinia } from 'pinia'

import App from './App.vue'
import router from './router'
import '@/router/guards' // 路由守卫（副作用注册）
import './styles/index.css'

const app = createApp(App)

app.use(createPinia())
app.use(router)
app.use(ElementPlus)

app.mount('#app')
