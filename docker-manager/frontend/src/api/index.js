import axios from 'axios'
import { ElMessage } from 'element-plus'

// 创建 axios 实例
const instance = axios.create({
  baseURL: 'http://192.168.0.110:8080',  // 修改为实际的后端地址
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json'
  }
})

// 响应拦截器
instance.interceptors.response.use(
  response => response.data,
  error => {
    console.error('API Error:', error.response?.data?.error || error.message)
    ElMessage.error(error.response?.data?.error || '请求失败')
    return Promise.reject(error)
  }
)
const api = {
  containers: {
    list: () => instance.get('/api/containers'),
    start: (id) => instance.post(`/api/containers/${id}/start`),
    stop: (id) => instance.post(`/api/containers/${id}/stop`),
    restart: (id) => instance.post(`/api/containers/${id}/restart`),
    pause: (id) => instance.post(`/api/containers/${id}/pause`),
    unpause: (id) => instance.post(`/api/containers/${id}/unpause`),
    remove: (id) => instance.delete(`/api/containers/${id}`),  // 添加逗号
    logs: (id) => instance.get(`/api/containers/${id}/logs`, {
      responseType: 'text',
      timeout: 0 // 禁用超时
    })
  },  // 添加逗号
  images: {
    list: () => instance.get('/api/images'),
    pull: (data) => instance.post('/api/images/pull', data),
    remove: (id) => instance.delete(`/api/images/${id}`)
  },
  networks: {
    list: () => instance.get('/api/networks'),
    create: (data) => instance.post('/api/networks', data),
    remove: (id) => instance.delete(`/api/networks/${id}`)
  },
  volumes: {
    list: () => instance.get('/api/volumes'),
    create: (data) => instance.post('/api/volumes', data),
    remove: (name) => instance.delete(`/api/volumes/${name}`)
  }
}

export default api