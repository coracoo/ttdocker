<template>
  <el-container class="layout-container">
    <el-aside width="200px">
      <el-menu
        default-active="overview"
        class="side-menu"
        @select="handleSelect"
      >
        <el-menu-item index="overview">
          <el-icon><Monitor /></el-icon>
          <span>概览</span>
        </el-menu-item>
        <el-sub-menu index="containers">
          <template #title>
            <el-icon><Box /></el-icon>
            <span>容器管理</span>
          </template>
          <el-menu-item index="containers-list">容器列表</el-menu-item>
          <el-menu-item index="app-store">应用商店</el-menu-item>
        </el-sub-menu>
        <el-menu-item index="images">
          <el-icon><Picture /></el-icon>
          <span>镜像管理</span>
        </el-menu-item>
        <el-menu-item index="volumes">
          <el-icon><Files /></el-icon>
          <span>数据卷</span>
        </el-menu-item>
        <el-menu-item index="networks">
          <el-icon><Connection /></el-icon>
          <span>网络管理</span>
        </el-menu-item>
      </el-menu>
    </el-aside>
    
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
import { Monitor, Box, Picture, Files, Connection } from '@element-plus/icons-vue'

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