import { createApp } from 'vue'
import ElementPlus from 'element-plus'
import 'element-plus/dist/index.css'
import App from './App.vue'
import router from './router'
import * as ElementPlusIconsVue from '@element-plus/icons-vue'
import axios from 'axios'

// 配置 axios 默认值
axios.defaults.baseURL = 'http://192.168.0.118:8080'
axios.defaults.headers.common['Content-Type'] = 'application/json'

// 添加响应拦截器
axios.interceptors.response.use(
  response => response,
  error => {
    ElMessage.error(error.response?.data?.message || '操作失败')
    return Promise.reject(error)
  }
)

const app = createApp(App)

for (const [key, component] of Object.entries(ElementPlusIconsVue)) {
  app.component(key, component)
}

app.use(ElementPlus)
app.use(router)
app.mount('#app')
