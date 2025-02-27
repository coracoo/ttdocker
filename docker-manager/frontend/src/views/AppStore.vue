<template>
  <div class="app-store">
    <div class="category-nav">
      <el-menu
        :default-active="activeCategory"
        mode="horizontal"
        @select="handleCategoryChange"
      >
        <el-menu-item index="all">全部</el-menu-item>
        <el-menu-item index="web">Web 服务器</el-menu-item>
        <el-menu-item index="database">数据库</el-menu-item>
        <el-menu-item index="tools">实用工具</el-menu-item>
        <el-menu-item index="storage">云存储</el-menu-item>
      </el-menu>
    </div>

    <div class="search-bar">
      <el-input
        v-model="searchQuery"
        placeholder="搜索应用..."
        :prefix-icon="Search"
        clearable
      >
        <template #append>
          <el-button :icon="Search">搜索</el-button>
        </template>
      </el-input>
    </div>

    <div class="app-grid">
      <el-card v-for="app in filteredApps" :key="app.id" class="app-card">
        <div class="app-header">
          <img :src="app.icon" :alt="app.name" class="app-icon">
          <div class="app-title">
            <h3>{{ app.name }}</h3>
            <p class="app-desc">{{ app.description }}</p>
          </div>
        </div>
        <div class="app-footer">
          <el-button type="primary" @click="handleDeploy(app)">安装</el-button>
          <el-button @click="showDetail(app)">详情</el-button>
        </div>
      </el-card>
    </div>

    <!-- 应用详情对话框 -->
    <el-dialog
      v-model="detailVisible"
      :title="currentApp?.name"
      width="50%"
    >
      <template v-if="currentApp">
        <div class="app-detail">
          <img :src="currentApp.icon" :alt="currentApp.name" class="detail-icon">
          <div class="detail-content">
            <h4>描述</h4>
            <p>{{ currentApp.description }}</p>
            <h4>版本</h4>
            <p>{{ currentApp.version }}</p>
            <h4>部署配置</h4>
            <el-form :model="deployConfig" label-width="100px">
              <el-form-item label="端口映射">
                <el-input v-model="deployConfig.port" placeholder="80:80"></el-input>
              </el-form-item>
              <el-form-item label="环境变量">
                <el-input
                  v-model="deployConfig.env"
                  type="textarea"
                  placeholder="KEY=VALUE"
                  rows="3"
                ></el-input>
              </el-form-item>
            </el-form>
          </div>
        </div>
      </template>
      <template #footer>
        <el-button @click="detailVisible = false">取消</el-button>
        <el-button type="primary" @click="confirmDeploy">确认部署</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import { Search } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'

const activeCategory = ref('all')
const searchQuery = ref('')
const detailVisible = ref(false)
const currentApp = ref(null)
const deployConfig = ref({
  port: '',
  env: ''
})

// 模拟应用数据
const apps = ref([
  {
    id: 1,
    name: 'Nginx',
    description: '高性能的 HTTP 和反向代理服务器',
    icon: 'https://www.nginx.com/wp-content/uploads/2020/05/nginx-plus-icon.svg',
    category: 'web',
    version: '1.24.0'
  },
  {
    id: 2,
    name: 'MySQL',
    description: '最流行的开源关系型数据库',
    icon: 'https://www.mysql.com/common/logos/logo-mysql-170x115.png',
    category: 'database',
    version: '8.0'
  },
  {
    id: 3,
    name: 'Redis',
    description: '开源的内存数据结构存储系统',
    icon: 'https://redis.io/images/redis-white.png',
    category: 'database',
    version: '7.2'
  },
  // 可以继续添加更多应用...
])

const filteredApps = computed(() => {
  return apps.value.filter(app => {
    const matchCategory = activeCategory.value === 'all' || app.category === activeCategory.value
    const matchSearch = app.name.toLowerCase().includes(searchQuery.value.toLowerCase()) ||
                       app.description.toLowerCase().includes(searchQuery.value.toLowerCase())
    return matchCategory && matchSearch
  })
})

const handleCategoryChange = (category) => {
  activeCategory.value = category
}

const showDetail = (app) => {
  currentApp.value = app
  detailVisible.value = true
  deployConfig.value = {
    port: '',
    env: ''
  }
}

const handleDeploy = (app) => {
  showDetail(app)
}

const confirmDeploy = () => {
  // 这里添加部署逻辑
  ElMessage.success(`${currentApp.value.name} 开始部署`)
  detailVisible.value = false
}
</script>

<style scoped>
.app-store {
  padding: 20px;
}

.category-nav {
  margin-bottom: 20px;
}

.search-bar {
  margin-bottom: 20px;
  max-width: 500px;
}

.app-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
  gap: 20px;
}

.app-card {
  height: 150px;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
}

.app-header {
  display: flex;
  align-items: center;
  gap: 15px;
}

.app-icon {
  width: 50px;
  height: 50px;
  object-fit: contain;
}

.app-title h3 {
  margin: 0;
  font-size: 16px;
}

.app-desc {
  margin: 5px 0;
  font-size: 12px;
  color: #666;
}

.app-footer {
  display: flex;
  gap: 10px;
  justify-content: flex-end;
}

.app-detail {
  display: flex;
  gap: 20px;
}

.detail-icon {
  width: 100px;
  height: 100px;
  object-fit: contain;
}

.detail-content {
  flex: 1;
}

.detail-content h4 {
  margin: 10px 0;
  color: #333;
}
</style>