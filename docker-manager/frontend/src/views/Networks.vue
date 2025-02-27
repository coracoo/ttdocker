<template>
  <div class="container">
    <!-- 顶部操作栏 -->
    <div class="operation-bar">
      <el-button @click="fetchNetworks">
        <el-icon><Refresh /></el-icon>
      </el-button>
      <el-button type="primary" @click="createNetwork">创建网络</el-button>
      <el-button @click="clearNetworks">清理网络</el-button>
      <el-button @click="deleteSelected">删除</el-button>
    </div>

    <!-- 网络列表 -->
    <el-table 
      :data="sortedNetworks" 
      style="width: 100%" 
      v-loading="loading">
      <el-table-column type="selection" width="55" />
      <el-table-column 
        prop="Name" 
        label="名称" />
      <el-table-column 
        prop="Driver" 
        label="模式">
        <template #default="scope">
          <el-tag size="small" type="info">{{ scope.row.Driver }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="子网">
        <template #default="scope">
          {{ scope.row.IPAM?.Config?.[0]?.Subnet || '-' }}
        </template>
      </el-table-column>
      <el-table-column label="网关">
        <template #default="scope">
          {{ scope.row.IPAM?.Config?.[0]?.Gateway || '-' }}
        </template>
      </el-table-column>
      <el-table-column 
        prop="Created" 
        label="创建时间">
        <template #default="scope">
          {{ formatTime(scope.row.Created) }}
        </template>
      </el-table-column>
      <el-table-column label="操作" width="100">
        <template #default="scope">
          <el-button 
            size="small" 
            type="danger" 
            @click="deleteNetwork(scope.row)">删除</el-button>
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
        :total="total"
        @size-change="handleSizeChange"
        @current-change="handleCurrentChange"
      />
    </div>

    <!-- 创建网络对话框 -->
    <el-dialog v-model="dialogVisible" title="创建网络" width="500px">
      <el-form :model="networkForm" label-width="100px">
        <el-form-item label="网络名称">
          <el-input v-model="networkForm.name" placeholder="请输入网络名称" />
        </el-form-item>
        <el-form-item label="网络模式">
          <el-select v-model="networkForm.driver" placeholder="请选择网络模式">
            <el-option label="bridge" value="bridge" />
            <el-option label="host" value="host" />
            <el-option label="none" value="none" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="dialogVisible = false">取消</el-button>
          <el-button type="primary" @click="submitNetwork">确定</el-button>
        </span>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Refresh } from '@element-plus/icons-vue'
import api from '../api'
import { formatTime } from '../utils/format'
const loading = ref(false)
const networks = ref([])
const currentPage = ref(1)
const pageSize = ref(10)
const total = ref(0)
const dialogVisible = ref(false)
const networkForm = ref({
  name: '',
  driver: 'bridge'
})

// 获取网络列表
const fetchNetworks = async () => {
  loading.value = true
  try {
    const data = await api.networks.list()
    networks.value = Array.isArray(data) ? data : []
    total.value = networks.value.length
  } catch (error) {
    ElMessage.error('获取网络列表失败')
    networks.value = []
    total.value = 0
  } finally {
    loading.value = false
  }
}

// 创建网络
const submitNetwork = async () => {
  try {
    await api.networks.create(networkForm.value)
    ElMessage.success('网络创建成功')
    dialogVisible.value = false
    fetchNetworks()
  } catch (error) {
    ElMessage.error('创建网络失败')
  }
}

// 删除网络
const deleteNetwork = async (network) => {
  try {
    await ElMessageBox.confirm('确定要删除该网络吗？', '警告', {
      type: 'warning'
    })
    await api.networks.remove(network.Id)
    ElMessage.success('网络已删除')
    fetchNetworks()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除失败')
    }
  }
}

// 分页处理
const handleSizeChange = (val) => {
  pageSize.value = val
  fetchNetworks()
}

const handleCurrentChange = (val) => {
  currentPage.value = val
  fetchNetworks()
}

// 添加计算属性处理特殊排序
const sortedNetworks = computed(() => {
  const specialOrder = ['none', 'bridge', 'host']
  return [...networks.value].sort((a, b) => {
    const aIndex = specialOrder.indexOf(a.Name)
    const bIndex = specialOrder.indexOf(b.Name)
    
    if (aIndex !== -1 && bIndex !== -1) return aIndex - bIndex
    if (aIndex !== -1) return -1
    if (bIndex !== -1) return 1
    return a.Name.localeCompare(b.Name)
  })
})

// 删除 handleSortChange 函数

onMounted(() => {
  fetchNetworks()
})
</script>

<style scoped>
.container {
  padding: 20px;
}

.operation-bar {
  margin-bottom: 20px;
  display: flex;
  gap: 10px;
}

.pagination {
  margin-top: 20px;
  display: flex;
  justify-content: flex-end;
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
}

.el-button .el-icon {
  margin-right: 0;
}
</style>