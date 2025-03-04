import request from '../utils/request'

export default {
  list() {
    return request({
      url: '/api/compose/list',
      method: 'get'
    })
  },
  
  deploy(data) {
    return request({
      url: '/api/compose/project',
      method: 'post',
      data
    })
  },
  
  start(name) {
    return request({
      url: `/api/compose/${name}/start`,
      method: 'post'
    })
  },

  stop(name) {
    return request({
      url: `/api/compose/${name}/stop`,
      method: 'post'
    })
  },
  remove(name) {
    // 修改为适应当前后端路由格式
    return request({
      url: `/api/compose/remove/${name}`,
      method: 'delete'
    })
  },
  getStatus(name) {
    return request({
      url: `/api/compose/${name}/status`,
      method: 'get'
    })
  }
}