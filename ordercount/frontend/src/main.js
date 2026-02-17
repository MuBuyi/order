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

// 全局响应拦截：遇到 401 自动退出登录并回到登录页
axios.interceptors.response.use(
	(response) => response,
	(error) => {
		const { response, config } = error || {}
		// 登录接口本身的 401 交给页面自己处理
		if (response && response.status === 401 && !(config && config.url && config.url.includes('/api/login'))) {
			const USER_KEY = 'ordercount-user'
			if (typeof window !== 'undefined') {
				window.localStorage.removeItem(TOKEN_KEY)
				window.localStorage.removeItem(USER_KEY)
			}
			delete axios.defaults.headers.common.Authorization
			// 简单粗暴刷新页面，回到登录界面
			if (typeof window !== 'undefined') {
				window.location.reload()
			}
		}
		return Promise.reject(error)
	},
)

const app = createApp(App)
app.use(ElementPlus)
app.mount('#app')
