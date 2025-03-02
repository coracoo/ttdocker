//import volumes from './volumes'
//import networks from './networks'
import imagesApi from './images'
import volumesApi from './volumes'
import networksApi from './networks'
import axios from 'axios'
import { ElMessage } from 'element-plus'

// 创建 axios 实例
const instance = axios.create({
  baseURL: 'http://192.168.0.110:8080',  // 修改为实际的后端地址
  timeout: 300000,  // 修改为5分钟
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

// 定义 API 对象
const api = {
  containers: {
    list: (params) => instance.get('/api/containers', { params }),
    start: (id) => instance.post(`/api/containers/${id}/start`),
    stop: (id) => instance.post(`/api/containers/${id}/stop`),
    restart: (id) => instance.post(`/api/containers/${id}/restart`),
    pause: (id) => instance.post(`/api/containers/${id}/pause`),
    unpause: (id) => instance.post(`/api/containers/${id}/unpause`),
    remove: (id) => instance.delete(`/api/containers/${id}`),
    logs: (id) => instance.get(`/api/containers/${id}/logs`, {
      responseType: 'text',
      timeout: 0
    })
  },
  // 修正 images 对象，移除嵌套的 images 对象
  images: {
    list: () => instance.get('/api/images'),
    pull: (data) => instance.post('/api/images/pull', data),
    remove: (id) => instance.delete(`/api/images/${id}`),
    tag: (data) => instance.post('/api/images/tag', data),
    export: (id) => instance.get(`/api/images/export/${id}`, {
      responseType: 'blob'
    }),
    import: (formData) => instance.post('/api/images/import', formData, {
      headers: {
        'Content-Type': 'multipart/form-data'
      }
    }),
    getProxy: () => instance.get('/api/images/proxy'),
    updateProxy: (data) => instance.post('/api/images/proxy', data)
  },
  volumes: volumesApi,
  networks: networksApi
}

export default api
