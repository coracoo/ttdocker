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
  },
  
  getYaml(name) {
    return request({
      url: `/api/compose/${name}/yaml`,
      method: 'get'
    })
  },
  
  saveYaml(name, content) {
    return request({
      url: `/api/compose/${name}/yaml`,
      method: 'post',
      data: { content }
    })
  }
}