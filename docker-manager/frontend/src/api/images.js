// ... 现有代码 ...

export default {
  // ... 现有方法 ...
  
  // 获取 Docker 代理配置
  getProxy: () => {
    return request.get('/api/images/proxy')
  },
  
  // 更新 Docker 代理配置
  updateProxy: (data) => {
    return request.post('/api/images/proxy', data)
  }
}