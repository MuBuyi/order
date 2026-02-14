import { createApp } from 'vue'
import App from './App.vue'
import ElementPlus from 'element-plus'
import 'element-plus/dist/index.css'
import axios from 'axios'

// 启动时如果本地有 token，则设置到 axios 默认头
const TOKEN_KEY = 'ordercount-token'
const savedToken = typeof window !== 'undefined' ? window.localStorage.getItem(TOKEN_KEY) : null
if (savedToken) {
	axios.defaults.headers.common.Authorization = `Bearer ${savedToken}`
}

const app = createApp(App)
app.use(ElementPlus)
app.mount('#app')
