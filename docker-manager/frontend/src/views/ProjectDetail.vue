<template>
  <div class="project-detail">
    <div class="header">
      <div class="title">
        <el-button link @click="goBack">
          <el-icon><Back /></el-icon>
        </el-button>
        {{ projectName }}
      </div>
      <div class="actions">
        <el-button-group>
          <el-button type="primary" :disabled="!isRunning" @click="handleStop">
            停止
          </el-button>
          <el-button type="primary" :disabled="isRunning" @click="handleStart">
            启动
          </el-button>
          <el-button type="primary" @click="handleRestart">
            重启
          </el-button>
        </el-button-group>
      </div>
    </div>

    <el-tabs v-model="activeTab" class="detail-tabs">
      <el-tab-pane label="YAML配置" name="yaml">
        <div class="yaml-editor">
          <div class="editor-header">
            <span>docker-compose.yml</span>
            <div class="editor-actions">
              <el-button type="primary" size="small" @click="handleSaveYaml">
                保存
              </el-button>
            </div>
          </div>
          <el-input
            v-model="yamlContent"
            type="textarea"
            :rows="20"
            class="yaml-textarea"
            :spellcheck="false"
          />
        </div>
      </el-tab-pane>

      <el-tab-pane label="容器" name="containers">
        <el-table :data="containerList" style="width: 100%">
          <el-table-column type="index" label="序号" width="60" />
          <el-table-column prop="name" label="名称" min-width="150" />
          <el-table-column prop="image" label="镜像" min-width="150" />
          <el-table-column prop="status" label="状态" width="100">
            <template #default="scope">
              <el-tag :type="scope.row.status === 'running' ? 'success' : 'info'">
                {{ scope.row.status }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="cpu" label="CPU" width="100" />
          <el-table-column prop="memory" label="内存" width="100" />
          <el-table-column prop="network" label="网络" width="120" />
          <el-table-column label="操作" width="150" fixed="right">
            <template #default="scope">
              <el-button-group>
                <el-button 
                  size="small" 
                  type="primary"
                  @click="handleContainerRestart(scope.row)"
                >
                  重启
                </el-button>
                <el-button 
                  size="small" 
                  type="danger"
                  @click="handleContainerStop(scope.row)"
                >
                  停止
                </el-button>
              </el-button-group>
            </template>
          </el-table-column>
        </el-table>
      </el-tab-pane>

      <el-tab-pane label="日志" name="logs">
        <div class="logs-container">
          <div class="logs-header">
            <el-switch
              v-model="autoScroll"
              active-text="自动滚动"
            />
            <el-button @click="handleClearLogs">
              清空日志
            </el-button>
          </div>
          <div class="logs-content" ref="logsRef">
            <pre v-for="(log, index) in logs" :key="index" :class="log.type">
              {{ log.content }}
            </pre>
          </div>
        </div>
      </el-tab-pane>
    </el-tabs>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted, nextTick, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Back } from '@element-plus/icons-vue'

const route = useRoute()
const router = useRouter()
const projectName = ref(route.params.name || '')
const activeTab = ref('yaml')
const isRunning = ref(true)
const autoScroll = ref(true)
const logsRef = ref(null)

// YAML配置
const yamlContent = ref('')

// 容器列表
const containerList = ref([
  {
    name: 'mtphotos',
    image: 'mtphotos/mt-photos:latest',
    status: 'running',
    cpu: '4.30%',
    memory: '267 MB',
    network: '4 MB / 120 KB'
  },
  {
    name: 'mtphotos_ai',
    image: 'mtphotos/mt-photos-ai:latest',
    status: 'running',
    cpu: '0.20%',
    memory: '2 MB',
    network: '9 KB / 13 KB'
  }
])

// 日志数据
const logs = ref([])
let logWebSocket = null

const goBack = () => {
  router.push('/projects')
}

// YAML相关方法
const handleSaveYaml = async () => {
  try {
    // TODO: 调用后端API保存YAML
    ElMessage.success('保存成功')
  } catch (error) {
    ElMessage.error('保存失败')
  }
}

// 容器相关方法
const handleContainerRestart = (container) => {
  ElMessage.success(`重启容器 ${container.name}`)
}

const handleContainerStop = (container) => {
  ElMessage.success(`停止容器 ${container.name}`)
}

// 项目操作方法
const handleStart = () => {
  ElMessage.success('启动项目')
  isRunning.value = true
}

const handleStop = () => {
  ElMessage.success('停止项目')
  isRunning.value = false
}

const handleRestart = () => {
  ElMessage.success('重启项目')
}

// 日志相关方法
const handleClearLogs = () => {
  logs.value = []
}

const setupWebSocket = () => {
  // TODO: 实现WebSocket连接
  // logWebSocket = new WebSocket(...)
}

const scrollToBottom = () => {
  if (logsRef.value && autoScroll.value) {
    nextTick(() => {
      logsRef.value.scrollTop = logsRef.value.scrollHeight
    })
  }
}

watch(logs, scrollToBottom)

onMounted(() => {
  // 获取项目信息
  // 获取YAML配置
  yamlContent.value = `version: '3'
services:
  mtphotos:
    image: mtphotos/mt-photos:latest
    container_name: mtphotos
    ports:
      - "8080:8080"
    volumes:
      - ./data:/data
    environment:
      - TZ=Asia/Shanghai`

  // 设置WebSocket连接
  setupWebSocket()
})

onUnmounted(() => {
  if (logWebSocket) {
    logWebSocket.close()
  }
})
</script>

<style scoped>
.project-detail {
  padding: 20px;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.title {
  display: flex;
  align-items: center;
  font-size: 20px;
  font-weight: bold;
  gap: 10px;
}

.detail-tabs {
  background: #fff;
  border-radius: 4px;
  padding: 20px;
}

.yaml-editor {
  border: 1px solid #dcdfe6;
  border-radius: 4px;
}

.editor-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 15px;
  background-color: #f5f7fa;
  border-bottom: 1px solid #dcdfe6;
}

.yaml-textarea {
  font-family: monospace;
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
</style>