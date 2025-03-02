import request from '../utils/request'

export function getRegistries() {
  return request({
    url: '/api/image-registry',
    method: 'get'
  })
}

export function updateRegistries(data) {
  return request({
    url: '/api/image-registry',
    method: 'post',
    data
  })
}