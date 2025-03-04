<template>
  <el-container class="layout-container">
    <el-aside width="200px">
      <el-menu
        :router="true"
        default-active="overview"
        class="side-menu"
      >
        <el-menu-item index="/overview">
          <el-icon><Monitor /></el-icon>
          <span>概览</span>
        </el-menu-item>
        
        <el-menu-item index="/navigation">
          <el-icon><Operation /></el-icon>
          <span>导航栏</span>
        </el-menu-item>

        <el-menu-item index="/app-store">
          <el-icon><Shop /></el-icon>
          <span>应用商店</span>
        </el-menu-item>

        <el-menu-item index="/projects">
          <el-icon><Folder /></el-icon>
          <span>项目列表</span>
        </el-menu-item>

        <el-menu-item index="/containers">
          <el-icon><Box /></el-icon>
          <span>容器列表</span>
        </el-menu-item>

        <el-menu-item index="/images">
          <el-icon><Picture /></el-icon>
          <span>镜像管理</span>
        </el-menu-item>

        <el-menu-item index="/volumes">
          <el-icon><Files /></el-icon>
          <span>卷管理</span>
        </el-menu-item>

        <el-menu-item index="/networks">
          <el-icon><Connection /></el-icon>
          <span>网络管理</span>
        </el-menu-item>
      </el-menu>
    </el-aside>

    <!-- 保持其他部分不变 -->
    <el-container>
      <el-header>
        <div class="header-title">Docker Manager</div>
        <div class="header-right">
          <el-button type="primary" plain>重启服务</el-button>
          <el-button>帮助文档</el-button>
        </div>
      </el-header>
      
      <el-main>
        <router-view></router-view>
      </el-main>
    </el-container>
  </el-container>
</template>

<script setup>
import { useRouter } from 'vue-router'
import { 
  Monitor, 
  Box, 
  Picture, 
  Files, 
  Connection, 
  Operation, 
  Shop, 
  Folder 
} from '@element-plus/icons-vue'

const router = useRouter()

const handleSelect = (key) => {
  router.push(`/${key}`)
}
</script>

<style scoped>
.layout-container {
  height: 100vh;
}

.side-menu {
  height: 100%;
  border-right: none;
}

.el-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  border-bottom: 1px solid #eee;
}

.header-title {
  font-size: 18px;
  font-weight: bold;
}

.header-right {
  display: flex;
  gap: 10px;
}
</style>