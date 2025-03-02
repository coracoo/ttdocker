import request from '../utils/request'

// 重命名文件为 image_registry.js 并更新路径
export function getRegistries() {
  return request({
    url: '/api/image-registry',  // 更新路径
    method: 'get'
  })
}

export function updateRegistries(data) {
  return request({
    url: '/api/image-registry',  // 更新路径
    method: 'post',
    data
  })
}