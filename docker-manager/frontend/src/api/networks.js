import request from '../utils/request'

export default {
  list: () => {
    return request({
      url: '/api/networks',
      method: 'get'
    })
  },
  create: (data) => {
    return request({
      url: '/api/networks',
      method: 'post',
      data
    })
  },
  remove: (id) => {
    return request({
      url: `/api/networks/${id}`,
      method: 'delete'
    })
  }
}