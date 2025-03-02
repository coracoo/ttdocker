<template>
  <div class="container">
    <!-- 修改顶部操作栏 -->
    <div class="operation-bar">
      <el-button @click="fetchVolumes">
        <el-icon><Refresh /></el-icon>
      </el-button>
      <el-button type="primary" @click="createVolume">创建存储卷</el-button>
      <el-button type="warning" @click="pruneVolumes">清除无用卷</el-button>
    </div>

    <!-- 存储卷列表 -->
    <el-table 
      :data="sortedVolumes" 
      style="width: 100%" 
      v-loading="loading">
      <el-table-column type="selection" width="55" />
      <el-table-column prop="Name" label="名称" />
      <el-table-column prop="Mountpoint" label="存储卷目录" />
      <el-table-column prop="Driver" label="模式" />
      <!-- 添加使用状态列 -->
      <el-table-column label="使用状态" width="100">
        <template #default="scope">
          <el-tag :type="scope.row.InUse ? 'success' : 'info'" effect="plain">
            {{ scope.row.InUse ? '使用中' : '未使用' }}
          </el-tag>
        </template>
      </el-table-column>
      <!-- 添加使用容器列 -->
      <el-table-column label="使用容器">
        <template #default="scope">
          <template v-if="scope.row.Containers && Object.keys(scope.row.Containers).length">
            <div v-for="(container, id) in scope.row.Containers" :key="id">
              {{ container.Name.substring(1) }}
            </div>
          </template>
          <span v-else>-</span>
        </template>
      </el-table-column>
      <el-table-column label="创建时间">
        <template #default="scope">
          {{ formatTime(scope.row.CreatedAt) }}
        </template>
      </el-table-column>
      <el-table-column label="操作" width="100">
        <template #default="scope">
          <el-button 
            size="small" 
            type="danger" 
            :disabled="scope.row.InUse"
            @click="deleteVolume(scope.row)">删除</el-button>
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

    <!-- 创建存储卷对话框 -->
    <el-dialog v-model="dialogVisible" title="创建存储卷" width="500px">
      <el-form :model="volumeForm" label-width="100px">
        <el-form-item label="存储卷名称">
          <el-input v-model="volumeForm.name" placeholder="请输入存储卷名称" />
        </el-form-item>
        <el-form-item label="驱动类型">
          <el-select v-model="volumeForm.driver" placeholder="请选择驱动类型">
            <el-option label="local" value="local" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="dialogVisible = false">取消</el-button>
          <el-button type="primary" @click="submitVolume">确定</el-button>
        </span>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import api from '../api'  // Update this line
import { formatTime } from '../utils/format'
import { Refresh } from '@element-plus/icons-vue'

const loading = ref(false)
const volumes = ref([])
const currentPage = ref(1)
const pageSize = ref(10)
const total = ref(0)
const dialogVisible = ref(false)
const volumeForm = ref({
  name: '',
  driver: 'local'
})

// 获取存储卷列表
const fetchVolumes = async () => {
  loading.value = true
  try {
    const response = await api.volumes.list()
    volumes.value = Array.isArray(response.Volumes) ? response.Volumes.map(volume => ({
      ...volume,
      InUse: volume.Containers && Object.keys(volume.Containers).length > 0
    })) : []
    total.value = volumes.value.length
  } catch (error) {
    ElMessage.error('获取存储卷列表失败')
    volumes.value = []
    total.value = 0
  } finally {
    loading.value = false
  }
}

// 创建存储卷
const submitVolume = async () => {
  try {
    await api.volumes.create(volumeForm.value)
    ElMessage.success('存储卷创建成功')
    dialogVisible.value = false
    fetchVolumes()
  } catch (error) {
    ElMessage.error('创建存储卷失败')
  }
}

// 删除存储卷
const deleteVolume = async (volume) => {
  try {
    await ElMessageBox.confirm('确定要删除该存储卷吗？', '警告', {
      type: 'warning'
    })
    await api.volumes.remove(volume.Name)
    ElMessage.success('存储卷已删除')
    fetchVolumes()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除失败')
    }
  }
}

// 添加清除无用卷的方法
const pruneVolumes = async () => {
  try {
    await ElMessageBox.confirm(
      '此操作将清除所有未被使用的存储卷，是否继续？',
      '警告',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )
    
    await api.volumes.prune()
    ElMessage.success('无用存储卷已清除')
    fetchVolumes()  // 刷新列表
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('清除失败：' + (error.response?.data?.error || '未知错误'))
    }
  }
}

// 分页处理
const handleSizeChange = (val) => {
  pageSize.value = val
  fetchVolumes()
}

const handleCurrentChange = (val) => {
  currentPage.value = val
  fetchVolumes()
}

// 添加计算属性处理排序
const sortedVolumes = computed(() => {
  return [...volumes.value].sort((a, b) => {
    return a.Name.localeCompare(b.Name)
  })
})

// 删除 handleSortChange 函数，因为我们使用计算属性来处理排序

onMounted(() => {
  fetchVolumes()
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
</style>