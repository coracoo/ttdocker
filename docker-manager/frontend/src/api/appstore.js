import request from '../utils/request'

export default {
  // 获取所有应用
  getApps() {
    return request({
      url: '/api/appstore/apps',
      method: 'get'
    })
  },
  
  // 获取单个应用详情
  getAppDetail(id) {
    return request({
      url: `/api/appstore/apps/${id}`,
      method: 'get'
    })
  },
  
  // 部署应用
  deployApp(id) {
    return request({
      url: `/api/appstore/deploy/${id}`,
      method: 'post'
    })
  },
  
  // 检查应用状态
  checkAppStatus(id) {
    return request({
      url: `/api/appstore/status/${id}`,
      method: 'get'
    })
  }
}