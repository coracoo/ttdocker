<template>
  <div class="template-list">
    <div class="toolbar">
      <el-button type="primary" @click="handleCreate">
        新建模板
      </el-button>
    </div>

    <el-table :data="templates" style="width: 100%">
      <el-table-column prop="name" label="名称" />
      <el-table-column prop="category" label="分类" />
      <el-table-column prop="description" label="描述" />
      <el-table-column prop="version" label="版本" />
      <el-table-column label="操作" width="200">
        <template #default="{ row }">
          <el-button-group>
            <el-button size="small" @click="handleEdit(row)">编辑</el-button>
            <el-button size="small" type="danger" @click="handleDelete(row)">删除</el-button>
          </el-button-group>
        </template>
      </el-table-column>
    </el-table>

    <el-dialog
      :title="dialogTitle"
      v-model="dialogVisible"
      width="80%"
    >
      <template-form
        ref="formRef"
        :template="currentTemplate"
        @submit="handleSubmit"
      />
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import TemplateForm from '../components/TemplateForm.vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { templateApi } from '../api/template'

const templates = ref([])
const dialogVisible = ref(false)
const dialogTitle = ref('新建模板')
const currentTemplate = ref(null)
const formRef = ref(null)

// 获取模板列表
const fetchTemplates = async () => {
  try {
    const response = await templateApi.list()
    templates.value = response.data
  } catch (error) {
    ElMessage.error('获取模板列表失败')
  }
}

const handleCreate = () => {
  currentTemplate.value = null
  dialogTitle.value = '新建模板'
  dialogVisible.value = true
}

const handleEdit = (row) => {
  currentTemplate.value = { ...row }
  dialogTitle.value = '编辑模板'
  dialogVisible.value = true
}

const handleDelete = async (row) => {
  try {
    await ElMessageBox.confirm('确定要删除该模板吗？')
    await templateApi.delete(row.id)
    ElMessage.success('删除成功')
    await fetchTemplates()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除失败')
    }
  }
}

const handleSubmit = async (formData) => {
  try {
    if (currentTemplate.value) {
      await templateApi.update(currentTemplate.value.id, formData)
    } else {
      await templateApi.create(formData)
    }
    dialogVisible.value = false
    ElMessage.success('保存成功')
    await fetchTemplates()
  } catch (error) {
    ElMessage.error('保存失败')
  }
}

onMounted(() => {
  fetchTemplates()
})
</script>

<style scoped>
.template-list {
  padding: 20px;
}

.toolbar {
  margin-bottom: 20px;
}
</style>