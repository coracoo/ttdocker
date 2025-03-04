<template>
  <div class="container">
    <div class="header">
      <div class="title">
        <el-button link @click="goBack">
          <el-icon><Back /></el-icon>
        </el-button>
        {{ containerName }}
      </div>
      <div class="status">
        <el-tag :type="containerStatus === '运行中' ? 'success' : 'info'">
          {{ containerStatus }}
        </el-tag>
      </div>
      <div class="actions">
        <el-button-group>
          <el-button 
            type="primary" 
            :disabled="containerStatus === '运行中'"
            @click="handleStart"
          >
            启动
          </el-button>
          <el-button 
            type="warning" 
            :disabled="containerStatus !== '运行中'"
            @click="handleStop"
          >
            停止
          </el-button>
          <el-button 
            type="primary"
            @click="handleRestart"
          >
            重启
          </el-button>
        </el-button-group>
      </div>
    </div>

    <div class="content">
      <el-tabs v-model="activeTab">
        <el-tab-pane label="基本信息" name="info">
          <div class="info-section">
            <div class="resource-usage">
              <el-row :gutter="20">
                <el-col :span="6">
                  <div class="metric-card">
                    <div class="metric-title">CPU</div>
                    <div class="metric-value">{{ cpuUsage }}%</div>
                    <div class="metric-chart">
                      <el-progress 
                        :percentage="cpuUsage" 
                        :color="getProgressColor(cpuUsage)"
                      />
                    </div>
                  </div>
                </el-col>
                <el-col :span="6">
                  <div class="metric-card">
                    <div class="metric-title">内存</div>
                    <div class="metric-value">{{ memoryUsage }}MB</div>
                    <div class="metric-chart">
                      <el-progress 
                        :percentage="(memoryUsage / memoryLimit) * 100" 
                        :color="getProgressColor((memoryUsage / memoryLimit) * 100)"
                      />
                    </div>
                  </div>
                </el-col>
                <el-col :span="6">
                  <div class="metric-card">
                    <div class="metric-title">网络(上传)</div>
                    <div class="metric-value">{{ networkUp }}</div>
                  </div>
                </el-col>
                <el-col :span="6">
                  <div class="metric-card">
                    <div class="metric-title">网络(下载)</div>
                    <div class="metric-value">{{ networkDown }}</div>
                  </div>
                </el-col>
              </el-row>
            </div>

            <div class="detail-info">
              <el-descriptions :column="2" border>
                <el-descriptions-item label="容器名称">{{ containerName }}</el-descriptions-item>
                <el-descriptions-item label="镜像">{{ imageInfo }}</el-descriptions-item>
                <el-descriptions-item label="创建时间">{{ createTime }}</el-descriptions-item>
                <el-descriptions-item label="运行时长">{{ uptime }}</el-descriptions-item>
                <el-descriptions-item label="端口映射">{{ ports }}</el-descriptions-item>
                <el-descriptions-item label="存储卷">{{ volumes }}</el-descriptions-item>
                <el-descriptions-item label="网络">{{ networks }}</el-descriptions-item>
                <el-descriptions-item label="重启策略">{{ restartPolicy }}</el-descriptions-item>
              </el-descriptions>
            </div>
          </div>
        </el-tab-pane>

        <el-tab-pane label="日志" name="logs">
          <div class="logs-container">
            <div class="logs-header">
              <div class="logs-options">
                <el-switch
                  v-model="autoScroll"
                  active-text="自动滚动"
                />
                <el-input
                  v-model="logFilter"
                  placeholder="过滤日志"
                  style="width: 200px"
                >
                  <template #prefix>
                    <el-icon><Search /></el-icon>
                  </template>
                </el-input>
              </div>
              <el-button @click="handleClearLogs">清空日志</el-button>
            </div>
            <div class="logs-content" ref="logsRef">
              <pre v-for="(log, index) in filteredLogs" 
                   :key="index" 
                   :class="getLogClass(log)">{{ log.content }}</pre>
            </div>
          </div>
        </el-tab-pane>
      </el-tabs>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted, computed, watch, nextTick } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Back, Search } from '@element-plus/icons-vue'

const route = useRoute()
const router = useRouter()
const containerName = ref(route.params.name || '')
const containerStatus = ref('运行中')
const activeTab = ref('info')

// 基本信息数据
const cpuUsage = ref(1.5)
const memoryUsage = ref(256)
const memoryLimit = ref(1024)
const networkUp = ref('1.2 MB/s')
const networkDown = ref('500 KB/s')
const imageInfo = ref('nginx:latest')
const createTime = ref('2024-03-02 10:00:00')
const uptime = ref('13 days')
const ports = ref('80:80, 443:443')
const volumes = ref('/data:/data')
const networks = ref('bridge')
const restartPolicy = ref('always')

// 日志相关
const autoScroll = ref(true)
const logFilter = ref('')
const logs = ref([])
const logsRef = ref(null)
let logWebSocket = null

// 计算属性和方法
const filteredLogs = computed(() => {
  if (!logFilter.value) return logs.value
  return logs.value.filter(log => 
    log.content.toLowerCase().includes(logFilter.value.toLowerCase())
  )
})

const getProgressColor = (percentage) => {
  if (percentage < 60) return '#67C23A'
  if (percentage < 80) return '#E6A23C'
  return '#F56C6C'
}

const getLogClass = (log) => ({
  'error': log.level === 'error',
  'warning': log.level === 'warning',
  'info': log.level === 'info'
})

// 操作方法
const goBack = () => {
  router.push('/projects')
}

const handleStart = () => {
  containerStatus.value = '运行中'
  ElMessage.success('容器已启动')
}

const handleStop = () => {
  containerStatus.value = '已停止'
  ElMessage.success('容器已停止')
}

const handleRestart = () => {
  ElMessage.success('容器重启中')
}

const handleClearLogs = () => {
  logs.value = []
}

// 自动滚动
const scrollToBottom = () => {
  if (logsRef.value && autoScroll.value) {
    nextTick(() => {
      logsRef.value.scrollTop = logsRef.value.scrollHeight
    })
  }
}

watch(logs, scrollToBottom)
watch(filteredLogs, scrollToBottom)

onMounted(() => {
  // TODO: 获取容器信息
  // TODO: 建立WebSocket连接获取日志
})

onUnmounted(() => {
  if (logWebSocket) {
    logWebSocket.close()
  }
})
</script>

<style scoped>
.container {
  padding: 20px;
}

.header {
  display: flex;
  align-items: center;
  margin-bottom: 20px;
  gap: 20px;
}

.title {
  display: flex;
  align-items: center;
  font-size: 20px;
  font-weight: bold;
  gap: 10px;
}

.content {
  background: #fff;
  border-radius: 4px;
  padding: 20px;
}

.metric-card {
  background: #f5f7fa;
  padding: 15px;
  border-radius: 4px;
  margin-bottom: 20px;
}

.metric-title {
  font-size: 14px;
  color: #606266;
  margin-bottom: 10px;
}

.metric-value {
  font-size: 24px;
  font-weight: bold;
  margin-bottom: 10px;
}

.detail-info {
  margin-top: 20px;
}

.logs-container {
  height: 600px;
  display: flex;
  flex-direction: column;
}

.logs-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 10px;
  background: #f5f7fa;
  border: 1px solid #dcdfe6;
  border-bottom: none;
}

.logs-options {
  display: flex;
  align-items: center;
  gap: 20px;
}

.logs-content {
  flex: 1;
  overflow-y: auto;
  background: #1e1e1e;
  color: #fff;
  padding: 10px;
  font-family: monospace;
  border: 1px solid #dcdfe6;
}

.logs-content pre {
  margin: 0;
  white-space: pre-wrap;
  word-wrap: break-word;
}

.logs-content .error {
  color: #ff4949;
}

.logs-content .warning {
  color: #e6a23c;
}

.logs-content .info {
  color: #67c23a;
}
</style>