<template>
  <div class="container">
    <!-- 顶部操作栏 -->
    <div class="operation-bar">
      <el-button @click="fetchContainers">
		<el-icon><Refresh /></el-icon>
      </el-button>
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
          <el-button 
            link 
            type="primary" 
            @click="goToContainerDetail(scope.row)"
          >
            {{ scope.row.Names?.[0]?.replace(/^\//, '') || '-' }}
          </el-button>
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
import { ref, onMounted, computed, nextTick } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { ArrowDown } from '@element-plus/icons-vue'
import dayjs from 'dayjs'
import { formatTime } from '../utils/format'
import api from '../api'
import ContainerTerminal from '../components/ContainerTerminal.vue'
import ContainerLogs from '../components/ContainerLogs.vue'
import { useRouter } from 'vue-router'

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

// 批量操作函数
const batchAction = async (action) => {
  if (selectedContainers.value.length === 0) {
    ElMessage.warning('请选择容器')
    return
  }
  
  try {
    const actionMap = {
      'start': '启动',
      'stop': '停止',
      'restart': '重启',
      'kill': '强制停止',
      'pause': '暂停',
      'unpause': '恢复',
      'remove': '删除'
    }
    
    await ElMessageBox.confirm(`确定要${actionMap[action]}选中的 ${selectedContainers.value.length} 个容器吗？`, '确认', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    
    await Promise.all(
      selectedContainers.value.map(container => 
        api.containers[action](container.Id)
      )
    )
    
    ElMessage.success(`已${actionMap[action]}${selectedContainers.value.length}个容器`)
    fetchContainers()
  } catch (error) {
    if (error !== 'cancel') {
      console.error('批量操作失败:', error)
      ElMessage.error(`操作失败: ${error.message || '未知错误'}`)
    }
  }
}

// 清理容器函数
const clearContainers = async () => {
  try {
    await ElMessageBox.confirm('确定要清理所有已停止的容器吗？', '警告', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    
    await api.containers.prune()
    ElMessage.success('已清理所有已停止的容器')
    fetchContainers()
  } catch (error) {
    if (error !== 'cancel') {
      console.error('清理容器失败:', error)
      ElMessage.error(`清理失败: ${error.message || '未知错误'}`)
    }
  }
}

// 添加处理单个容器操作的函数
const handleAction = async (container, action) => {
  try {
    const actionMap = {
      'start': '启动',
      'stop': '停止',
      'restart': '重启',
      'pause': '暂停',
      'unpause': '恢复'
    }
    
    await api.containers[action](container.Id)
    ElMessage.success(`容器已${actionMap[action]}`)
    fetchContainers()
  } catch (error) {
    console.error(`容器操作失败:`, error)
    ElMessage.error(`操作失败: ${error.message || '未知错误'}`)
  }
}

// 添加处理单个容器删除的函数
const handleDelete = async (container) => {
  try {
    const containerName = container.Names?.[0]?.replace(/^\//, '') || container.Id.substring(0, 12)
    
    await ElMessageBox.confirm(`确定要删除容器 "${containerName}" 吗？`, '警告', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    
    await api.containers.remove(container.Id)
    ElMessage.success('容器已删除')
    fetchContainers()
  } catch (error) {
    if (error !== 'cancel') {
      console.error('删除容器失败:', error)
      ElMessage.error(`删除失败: ${error.message || '未知错误'}`)
    }
  }
}

// 创建容器函数
const createContainer = () => {
  ElMessage.info('创建容器功能正在开发中')
  // 这里可以添加创建容器的对话框逻辑
}

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

// 获取容器列表
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

// 格式化端口映射
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
// 添加容器详情页面跳转方法
const goToContainerDetail = (container) => {
  const containerName = container.Names?.[0]?.replace(/^\//, '') || ''
  if (containerName) {
    router.push(`/containers/${containerName}`)
  }
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