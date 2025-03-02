import request from '../utils/request'

export default {
  list: () => {
    return request({
      url: '/api/volumes',
      method: 'get'
    })
  },
  create: (data) => {
    return request({
      url: '/api/volumes',
      method: 'post',
      data
    })
  },
  remove: (name) => {
    return request({
      url: `/api/volumes/${name}`,
      method: 'delete'
    })
  },
  // 添加清理无用卷的方法
  prune: () => {
    return request({
      url: '/api/volumes/prune',
      method: 'post'
    })
  }
}