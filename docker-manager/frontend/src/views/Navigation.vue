<template>
  <div class="container">
    <div class="header">
      <div class="title">导航栏</div>
      <div class="actions">
        <el-button type="primary" @click="handleAdd">
          <el-icon><Plus /></el-icon>添加应用
        </el-button>
        <el-button @click="handleRefresh">
          <el-icon><Refresh /></el-icon>刷新
        </el-button>
      </div>
    </div>

    <div class="app-grid">
      <el-card 
        v-for="app in apps" 
        :key="app.id" 
        class="app-card"
        :body-style="{ padding: '0px' }"
      >
        <div class="app-content">
          <div class="app-icon">
            <img :src="app.icon" :alt="app.name">
          </div>
          <div class="app-info">
            <h3>{{ app.name }}</h3>
            <p>{{ app.description }}</p>
            <div class="app-url">{{ app.url }}</div>
          </div>
        </div>
        <div class="app-actions">
          <el-button-group>
            <el-button size="small" type="primary" @click="handleEdit(app)">
              编辑
            </el-button>
            <el-button size="small" type="danger" @click="handleDelete(app)">
              删除
            </el-button>
          </el-button-group>
        </div>
      </el-card>
    </div>

    <!-- 添加/编辑对话框 -->
    <el-dialog
      :title="dialogTitle"
      v-model="dialogVisible"
      width="500px"
    >
      <el-form :model="form" label-width="80px">
        <el-form-item label="名称" required>
          <el-input v-model="form.name" placeholder="请输入应用名称" />
        </el-form-item>
        <el-form-item label="图标">
          <el-upload
            class="icon-upload"
            action="/api/upload"
            :show-file-list="false"
            :on-success="handleIconSuccess"
          >
            <img v-if="form.icon" :src="form.icon" class="preview-icon">
            <el-icon v-else class="upload-icon"><Plus /></el-icon>
          </el-upload>
        </el-form-item>
        <el-form-item label="URL" required>
          <el-input v-model="form.url" placeholder="请输入应用访问地址" />
        </el-form-item>
        <el-form-item label="描述">
          <el-input
            v-model="form.description"
            type="textarea"
            placeholder="请输入应用描述"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="dialogVisible = false">取消</el-button>
          <el-button type="primary" @click="handleSave">确定</el-button>
        </span>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Refresh } from '@element-plus/icons-vue'

const apps = ref([
  {
    id: 1,
    name: 'Transmission',
    icon: '/icons/transmission.png',
    url: 'http://nas.local:9091',
    description: '下载工具'
  },
  {
    id: 2,
    name: 'Jellyfin',
    icon: '/icons/jellyfin.png',
    url: 'http://nas.local:8096',
    description: '媒体服务器'
  }
])

const dialogVisible = ref(false)
const dialogTitle = ref('添加应用')
const form = ref({
  name: '',
  icon: '',
  url: '',
  description: ''
})

const handleAdd = () => {
  dialogTitle.value = '添加应用'
  form.value = {
    name: '',
    icon: '',
    url: '',
    description: ''
  }
  dialogVisible.value = true
}

const handleEdit = (app) => {
  dialogTitle.value = '编辑应用'
  form.value = { ...app }
  dialogVisible.value = true
}

const handleDelete = (app) => {
  ElMessageBox.confirm(
    `确定要删除应用 "${app.name}" 吗？`,
    '警告',
    {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    }
  ).then(() => {
    // TODO: 调用后端 API 删除应用
    ElMessage.success('删除成功')
  })
}

const handleRefresh = () => {
  // TODO: 调用后端 API 刷新应用列表
  ElMessage.success('刷新成功')
}

const handleIconSuccess = (response) => {
  form.value.icon = response.url
}

const handleSave = () => {
  // TODO: 调用后端 API 保存应用
  dialogVisible.value = false
  ElMessage.success('保存成功')
}

onMounted(() => {
  // TODO: 获取应用列表
})
</script>

<style scoped>
.container {
  padding: 20px;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.title {
  font-size: 20px;
  font-weight: bold;
}

.actions {
  display: flex;
  gap: 10px;
}

.app-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
  gap: 20px;
}

.app-card {
  border-radius: 8px;
  transition: all 0.3s;
}

.app-card:hover {
  transform: translateY(-5px);
  box-shadow: 0 2px 12px 0 rgba(0,0,0,0.1);
}

.app-content {
  padding: 20px;
  display: flex;
  gap: 15px;
}

.app-icon {
  width: 48px;
  height: 48px;
  flex-shrink: 0;
}

.app-icon img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  border-radius: 8px;
}

.app-info {
  flex: 1;
}

.app-info h3 {
  margin: 0 0 8px 0;
  font-size: 16px;
}

.app-info p {
  margin: 0 0 8px 0;
  color: #666;
  font-size: 14px;
}

.app-url {
  color: #409EFF;
  font-size: 12px;
}

.app-actions {
  padding: 10px 20px;
  border-top: 1px solid #EBEEF5;
  display: flex;
  justify-content: flex-end;
}

.icon-upload {
  width: 100px;
  height: 100px;
  border: 1px dashed #d9d9d9;
  border-radius: 6px;
  cursor: pointer;
  position: relative;
  overflow: hidden;
  display: flex;
  justify-content: center;
  align-items: center;
}

.preview-icon {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.upload-icon {
  font-size: 28px;
  color: #8c939d;
}
</style>