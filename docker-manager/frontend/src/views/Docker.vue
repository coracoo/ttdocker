<template>
  <div class="container">
    <!-- 顶部操作栏 -->
    <div class="operation-bar">
      <el-button type="primary" @click="createContainer">创建容器</el-button>
      <el-button @click="clearContainers">清理容器</el-button>
      <el-button @click="batchStart">启动</el-button>
      <el-button @click="batchStop">停止</el-button>
      <el-button @click="batchRestart">重启</el-button>
      <el-button @click="batchForceStop">强制停止</el-button>
      <el-button @click="batchPause">暂停</el-button>
      <el-button @click="batchResume">恢复</el-button>
      <el-button @click="batchDelete">删除</el-button>
    </div>

    <!-- 状态筛选 -->
    <el-select v-model="statusFilter" placeholder="状态" clearable class="status-filter">
      <el-option label="所有" value="" />
      <el-option label="运行中" value="running" />
      <el-option label="已停止" value="stopped" />
      <el-option label="已暂停" value="paused" />
    </el-select>

    <!-- 容器列表 -->
    <el-table 
      :data="filteredContainers" 
      style="width: 100%" 
      v-loading="loading"
      @selection-change="handleSelectionChange">
      <el-table-column type="selection" width="55" />
      <el-table-column prop="Names" label="名称" sortable>
        <template #default="scope">
          {{ scope.row.Names?.[0]?.replace(/^\//, '') || '-' }}
        </template>
      </el-table-column>
      <el-table-column prop="Image" label="镜像" />
      <el-table-column prop="State" label="状态">
        <template #default="scope">
          <el-tag :type="getStatusType(scope.row.State)">
            {{ stateMap[scope.row.State.toLowerCase()] || scope.row.State }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="资源使用率">
        <template #default="scope">
          <div>CPU: {{ scope.row.CPUPerc || '0.00%' }}</div>
          <div>内存: {{ scope.row.MemPerc || '0.00%' }}</div>
        </template>
      </el-table-column>
      <!-- IP 地址列 -->
      <!-- 修改 IP 地址列的展示 -->
      <el-table-column label="IP 地址">
        <template #default="scope">
          <div>{{ getContainerIP(scope.row) }}</div>
          <div v-if="scope.row.HostConfig?.NetworkMode === 'host'" class="text-gray">
            (host 网络模式)
          </div>
        </template>
      </el-table-column>
      <!-- 修改端口映射列的展示 -->
      <el-table-column label="端口映射">
        <template #default="scope">
          <template v-if="scope.row.Ports && scope.row.Ports.length > 0">
            <div v-for="(port, index) in scope.row.Ports" :key="index">
              {{ formatPortWithIP(port) }}
            </div>
          </template>
          <template v-else>-</template>
        </template>
      </el-table-column>
      <el-table-column label="运行时长">
        <template #default="scope">
          {{ scope.row.RunningTime || '-' }}
        </template>
      </el-table-column>
      <el-table-column label="创建时间">
        <template #default="scope">
          {{ formatTime(scope.row.Created) }}
        </template>
      </el-table-column>
      <!-- 修改操作列的终端按钮，将其放在正确的位置 -->
      <el-table-column label="操作" width="250">
        <template #default="scope">
          <!-- 修改这里的函数名 -->
          <el-button size="small" @click="openTerminal(scope.row)">终端</el-button>
          <el-button size="small" @click="openLogs(scope.row)">日志</el-button>
          <el-dropdown>
            <el-button size="small">
              更多<el-icon class="el-icon--right"><arrow-down /></el-icon>
            </el-button>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item @click="handleAction(scope.row, 'start')">启动</el-dropdown-item>
                <el-dropdown-item @click="handleAction(scope.row, 'stop')">停止</el-dropdown-item>
                <el-dropdown-item @click="handleAction(scope.row, 'restart')">重启</el-dropdown-item>
                <el-dropdown-item @click="handleAction(scope.row, 'pause')">暂停</el-dropdown-item>
                <el-dropdown-item @click="handleAction(scope.row, 'unpause')">恢复</el-dropdown-item>
                <el-dropdown-item divided @click="handleDelete(scope.row)">删除</el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </template>
      </el-table-column>
    </el-table>

    <!-- 分页 -->
    <div class="pagination">
      <el-pagination
        v-model:current-page="currentPage"
        v-model:page-size="pageSize"
        :page-sizes="[10, 20, 30, 50]"
        layout="total, sizes, prev, pager, next, jumper"
        :total="filteredContainers.length"
        @size-change="handleSizeChange"
        @current-change="handleCurrentChange"
      />
    </div>
	
	<!-- 添加组件使用 -->
    <ContainerTerminal
      v-model="terminalDialogVisible"
      :container="currentContainer"
    />
    
    <ContainerLogs
      v-model="logDialogVisible"
      :container="currentContainer"
    />
	
  </div>
</template>

<!-- 在 script setup 中添加相关变量和方法 -->
<script setup>
import { ref, onMounted, computed, nextTick } from 'vue'  // 添加 nextTick
import { ElMessage, ElMessageBox } from 'element-plus'
import { ArrowDown } from '@element-plus/icons-vue'
import dayjs from 'dayjs'
import { formatTime } from '../utils/format'
// 修改导入语句
import api from '../api'
import ContainerTerminal from '../components/ContainerTerminal.vue'
import ContainerLogs from '../components/ContainerLogs.vue'

// 变量定义
const loading = ref(false)
const containers = ref([])
const selectedContainers = ref([])
const currentPage = ref(1)
const pageSize = ref(10)
const total = ref(0)
const statusFilter = ref('')
const currentContainer = ref(null)
const terminalDialogVisible = ref(false)
const logDialogVisible = ref(false)
const logs = ref('')
const batchStart = () => batchAction('start')
const batchStop = () => batchAction('stop')
const batchRestart = () => batchAction('restart')
const batchForceStop = () => batchAction('kill')
const batchPause = () => batchAction('pause')
const batchResume = () => batchAction('unpause')
const batchDelete = () => batchAction('remove')


// 添加打开终端和日志的方法
const openTerminal = (container) => {
  currentContainer.value = container
  nextTick(() => {
    terminalDialogVisible.value = true;
  });
};

const openLogs = (container) => {
  currentContainer.value = container
  logDialogVisible.value = true
}

// 格式化端口映射 (移到前面，因为模板中使用了这个函数)
const formatPorts = (ports) => {
  if (!Array.isArray(ports)) return '-'
  return ports.map(port => {
    if (port.PublicPort) {
      return `${port.PublicPort}:${port.PrivatePort}/${port.Type}`
    }
    return `${port.PrivatePort}/${port.Type}`
  }).join(', ')
}
// 添加格式化端口函数
const formatPortWithIP = (port) => {
  if (port.PublicPort) {
    const ip = port.IP || '0.0.0.0'
    return `${ip}:${port.PublicPort}:${port.PrivatePort}/${port.Type}`
  }
  return `${port.PrivatePort}/${port.Type}`
}

// 获取容器列表 (移到前面，因为其他函数都依赖它)
const fetchContainers = async () => {
  loading.value = true
  try {
    const data = await api.containers.list()
    containers.value = Array.isArray(data) ? data : []
    total.value = containers.value.length
  } catch (error) {
    console.error('Error fetching containers:', error)
    ElMessage.error('获取容器列表失败')
    containers.value = []
    total.value = 0
  } finally {
    loading.value = false
  }
}

// 添加状态映射
const stateMap = {
  'running': '运行中',
  'exited': '已停止',
  'created': '已创建',
  'paused': '已暂停',
  'restarting': '重启中',
  'removing': '删除中',
  'dead': '已死亡'
}

// 状态标签类型获取函数
const getStatusType = (status) => {
  const types = {
    'running': 'success',
    'exited': 'danger',
    'created': 'info',
    'paused': 'warning',
    'restarting': 'warning',
    'removing': 'danger',
    'dead': 'danger'
  }
  return types[status.toLowerCase()] || 'info'
}

// 添加计算属性用于过滤容器列表
const filteredContainers = computed(() => {
  if (!statusFilter.value) {
    return containers.value
  }
  return containers.value.filter(container => {
    const state = container.State.toLowerCase()
    switch (statusFilter.value) {
      case 'running':
        return state === 'running'
      case 'stopped':
        return state === 'exited'
      case 'paused':
        return state === 'paused'
      default:
        return true
    }
  })
})

// 表格选择变化
const handleSelectionChange = (selection) => {
  selectedContainers.value = selection
}

// 分页处理
const handleSizeChange = (val) => {
  pageSize.value = val
  fetchContainers()
}

const handleCurrentChange = (val) => {
  currentPage.value = val
  fetchContainers()
}

onMounted(() => {
  fetchContainers()
})
// 添加获取容器 IP 的函数
const getContainerIP = (container) => {
  // 如果是 host 网络模式，返回 host
  if (container.NetworkSettings?.Networks?.host) {
    return 'host'
  }
  
  // 获取容器 IP
  const ip = container.NetworkSettings?.Networks?.bridge?.IPAddress || '-'
  return ip
}

</script>

<style scoped>
.container {
  padding: 20px;
}

.operation-bar {
  margin-bottom: 20px;
  display: flex;
  gap: 10px;
  flex-wrap: wrap;
}

.status-filter {
  margin-bottom: 20px;
  width: 200px;
}

.pagination {
  margin-top: 20px;
  display: flex;
  justify-content: flex-end;
}

.text-gray {
  color: #909399;
  font-size: 12px;
}
</style>