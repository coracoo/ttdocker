import axios from 'axios'

const api = {
  // 获取容器列表
  listContainers() {
    return axios.get('/api/containers')
  },

  // 启动容器
  startContainer(id) {
    return axios.post(`/api/containers/${id}/start`)
  },

  // 停止容器
  stopContainer(id) {
    return axios.post(`/api/containers/${id}/stop`)
  },

  // 删除容器
  removeContainer(id) {
    return axios.delete(`/api/containers/${id}`)
  },

  // 获取容器日志
  getContainerLogs(id) {
    return axios.get(`/api/containers/${id}/logs`)
  }
}

export default api