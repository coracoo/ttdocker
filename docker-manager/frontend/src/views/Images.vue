<template>
  <div class="container">
    <!-- 顶部操作栏 -->
    <div class="operation-bar">
      <el-button @click="fetchImages">
        <el-icon><Refresh /></el-icon>
      </el-button>
      <el-button type="primary" @click="pullImage">拉取镜像</el-button>
      <el-button @click="importImage">导入镜像</el-button>
      <el-button @click="buildImage">构建镜像</el-button>
      <el-button @click="clearBuildCache">清理构建缓存</el-button>
      <el-button @click="clearImages">清理镜像</el-button>
      <el-button @click="showProxyDialog">配置加速器/代理</el-button>
    </div>

    <!-- 镜像列表 -->
    <!-- 修改 el-table 的属性 -->
    <el-table 
      :data="images" 
      style="width: 100%" 
      v-loading="loading"
      @sort-change="handleSortChange"
      :default-sort="{ prop: 'RepoTags', order: 'ascending' }">
      <!-- 修改每个可排序列 -->
      <el-table-column 
        label="IMAGE ID" 
        width="120" 
        prop="Id" 
        sortable="custom"
        :sort-orders="['ascending', 'descending']"
        sort-by="click">
        <template #default="scope">
          {{ scope.row.Id.substring(7, 19) }}
        </template>
      </el-table-column>
      <el-table-column 
        label="状态" 
        width="100" 
        prop="isInUse" 
        sortable="custom">
        <template #default="scope">
          <el-tag :type="scope.row.isInUse ? 'success' : ''" effect="plain">
            {{ scope.row.isInUse ? '使用中' : '未使用' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column 
        label="镜像名称" 
        prop="RepoTags" 
        sortable="custom">
        <template #default="scope">
          {{ getImageName(scope.row.RepoTags?.[0]) }}
        </template>
      </el-table-column>
      <!-- 修改镜像标签列，移除排序功能 -->
      <el-table-column 
        label="镜像标签" 
        prop="RepoTags">
        <template #default="scope">
          {{ getImageTag(scope.row.RepoTags?.[0]) }}
        </template>
      </el-table-column>
      <el-table-column 
        label="大小" 
        prop="Size" 
        sortable="custom">
        <template #default="scope">
          {{ formatSize(scope.row.Size) }}
        </template>
      </el-table-column>
      <el-table-column 
        label="创建时间" 
        prop="Created" 
        sortable="custom">
        <template #default="scope">
          {{ formatTime(scope.row.Created) }}
        </template>
      </el-table-column>
      <!-- 操作列保持不变 -->
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
    <!-- 添加配置对话框 -->
    <el-dialog
      v-model="proxyDialogVisible"
      title="Docker 配置"
      width="600px"
    >
      <el-form :model="proxyForm" label-width="120px">
        <el-form-item label="HTTP 代理">
          <el-input v-model="proxyForm.proxies.http" placeholder="http://proxy:port" />
        </el-form-item>
        <el-form-item label="HTTPS 代理">
          <el-input v-model="proxyForm.proxies.https" placeholder="https://proxy:port" />
        </el-form-item>
        <el-form-item label="无需代理">
          <el-input v-model="proxyForm.proxies.no" placeholder="localhost,127.0.0.1" />
        </el-form-item>
        <el-form-item label="镜像加速器">
          <el-select
            v-model="proxyForm.mirrors"
            multiple
            filterable
            allow-create
            default-first-option
            placeholder="请选择或输入镜像加速器地址"
          >
            <el-option
              v-for="mirror in defaultMirrors"
              :key="mirror.value"
              :label="mirror.label"
              :value="mirror.value"
            />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="proxyDialogVisible = false">取消</el-button>
          <el-button type="primary" @click="updateProxy">确定</el-button>
        </span>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted, h } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Refresh } from '@element-plus/icons-vue'  // 添加图标导入
import api from '../api'
import { formatTime } from '../utils/format'
import DockerSettings from '../components/DockerSettings.vue'

const loading = ref(false)
const images = ref([])
const currentPage = ref(1)
const pageSize = ref(10)
const total = ref(0)

// 添加处理镜像名称和标签的函数
const getImageName = (repoTag) => {
  if (!repoTag) return '<none>'
  const parts = repoTag.split(':')
  return parts[0]
}

const getImageTag = (repoTag) => {
  if (!repoTag) return '<none>'
  const parts = repoTag.split(':')
  return parts[1] || 'latest'
}

const settingsVisible = ref(false)
const showProxyDialog = () => {
  settingsVisible.value = true
}

// 修改获取镜像列表的函数，添加更详细的错误处理
const fetchImages = async () => {
  loading.value = true
  try {
    const imagesData = await api.images.list()
    const containersData = await api.containers.list({ all: true })
    
    // 获取使用中的镜像信息
    const usedImages = new Set(containersData.map(container => {
      const imageName = container.Image
      // 如果镜像名称中没有标签，添加 :latest
      return imageName.includes(':') ? imageName : `${imageName}:latest`
    }))
    
    // 处理镜像数据，将每个标签作为单独的行
    const processedImages = []
    imagesData.forEach(image => {
      if (!image.RepoTags || image.RepoTags.length === 0) {
        processedImages.push({
          ...image,
          RepoTags: ['<none>:<none>'],
          isInUse: false
        })
      } else {
        // 为每个标签创建一个记录
        image.RepoTags.forEach(tag => {
          processedImages.push({
            ...image,
            RepoTags: [tag],
            // 检查该标签是否被使用
            isInUse: usedImages.has(tag)
          })
        })
      }
    })
    
    images.value = processedImages
    total.value = processedImages.length
    
    // 添加默认排序
    handleSortChange({ prop: 'RepoTags', order: 'ascending' })
  } catch (error) {
    console.error('获取镜像列表错误:', error)
    ElMessage.error('获取镜像列表失败')
    images.value = []
    total.value = 0
  } finally {
    loading.value = false
  }
}

// 拉取镜像
// 修改拉取镜像的对话框
const pullImage = async () => {
  try {
    const { value: formData } = await ElMessageBox.prompt('', '拉取镜像', {
      title: '拉取镜像',
      message: '请输入镜像名称',
      inputPlaceholder: '例如：nginx:latest',
      confirmButtonText: '确定',
      cancelButtonText: '取消',
    })

    if (formData) {
      loading.value = true
      await api.images.pull({ name: formData })
      ElMessage.success('镜像拉取成功')
      fetchImages()
    }
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('拉取失败：' + error.message)
    }
  } finally {
    loading.value = false
  }
}

// 删除镜像
const deleteImage = async (image) => {
  try {
    await ElMessageBox.confirm('确定要删除该镜像吗？', '警告', {
      type: 'warning'
    })
    await api.images.remove(image.Id)
    ElMessage.success('镜像已删除')
    fetchImages()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除失败')
    }
  }
}

// 分页处理
const handleSizeChange = (val) => {
  pageSize.value = val
  fetchImages()
}

const handleCurrentChange = (val) => {
  currentPage.value = val
  fetchImages()
}

// 格式化文件大小
const formatSize = (size) => {
  if (!size) return '0 MB'
  const mb = size / (1024 * 1024)
  return `${mb.toFixed(2)} MB`
}

// 添加排序相关变量
const sortBy = ref('')
const sortOrder = ref('ascending')

// 添加排序处理函数
const handleSortChange = ({ prop, order }) => {
  if (!prop || !order) {
    images.value = [...images.value]
    return
  }

  // 移除空值检查，保证排序状态
  images.value.sort((a, b) => {
    let aValue, bValue

    switch (prop) {
      case 'Id':
        aValue = a.Id
        bValue = b.Id
        break
      case 'isInUse':
        aValue = a.isInUse ? 1 : 0
        bValue = b.isInUse ? 1 : 0
        break
      case 'RepoTags':
        // 只比较镜像名称部分
        aValue = getImageName(a.RepoTags?.[0] || '')
        bValue = getImageName(b.RepoTags?.[0] || '')
        break
      case 'Size':
        aValue = a.Size
        bValue = b.Size
        break
      case 'Created':
        aValue = a.Created
        bValue = b.Created
        break
      default:
        aValue = a[prop]
        bValue = b[prop]
    }

    return order === 'ascending' ? 
      (aValue > bValue ? 1 : -1) : 
      (aValue < bValue ? 1 : -1)
  })
}

onMounted(() => {
  fetchImages()
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

.el-button .el-icon {
  margin-right: 0;
}
</style>