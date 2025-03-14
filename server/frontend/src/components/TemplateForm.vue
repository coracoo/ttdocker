<template>
  <el-form :model="form" label-width="120px">
    <el-form-item label="项目名称" required>
      <el-input v-model="form.name" />
    </el-form-item>

    <el-form-item label="项目分类" required>
      <el-select v-model="form.category">
        <el-option label="Web服务" value="web" />
        <el-option label="数据库" value="database" />
        <el-option label="开发工具" value="development" />
        <el-option label="其他" value="other" />
      </el-select>
    </el-form-item>

    <el-form-item label="项目描述">
      <el-input v-model="form.description" type="textarea" />
    </el-form-item>

    <el-form-item label="版本">
      <el-input v-model="form.version" />
    </el-form-item>

    <el-form-item label="项目主页">
      <el-input v-model="form.website" />
    </el-form-item>

    <el-form-item label="项目Logo">
      <el-upload
        class="logo-uploader"
        action="/api/upload"
        :show-file-list="false"
        :on-success="handleLogoSuccess"
      >
        <img v-if="form.logo" :src="form.logo" class="logo">
        <el-icon v-else><Plus /></el-icon>
      </el-upload>
    </el-form-item>

    <el-form-item label="项目截图">
      <el-upload
        action="/api/upload"
        list-type="picture-card"
        :on-success="handleScreenshotSuccess"
      >
        <el-icon><Plus /></el-icon>
      </el-upload>
    </el-form-item>

    <el-form-item label="Compose文件" required>
      <el-upload
        class="compose-uploader"
        action="/api/upload"
        :show-file-list="false"
        :on-success="handleComposeSuccess"
        accept=".yml,.yaml"
      >
        <el-button type="primary">上传Compose文件</el-button>
      </el-upload>
      <div v-if="form.compose" class="compose-preview">
        <el-input type="textarea" v-model="form.compose" rows="10" readonly />
      </div>
    </el-form-item>

    <el-form-item label="使用教程">
      <el-tabs v-model="activeTab">
        <el-tab-pane label="编辑" name="edit">
          <el-input
            type="textarea"
            v-model="form.tutorial"
            rows="10"
            placeholder="支持Markdown格式"
          />
        </el-tab-pane>
        <el-tab-pane label="预览" name="preview">
          <div v-html="markdownHtml" class="markdown-body"></div>
        </el-tab-pane>
      </el-tabs>
    </el-form-item>

    <el-form-item>
      <el-button type="primary" @click="handleSubmit">保存</el-button>
      <el-button @click="handleReset">重置</el-button>
    </el-form-item>
  </el-form>
</template>

<script setup>
import { ref, computed } from 'vue'
import { Plus } from '@element-plus/icons-vue'
import { marked } from 'marked'
import { ElMessage } from 'element-plus'

const form = ref({
  name: '',
  category: '',
  description: '',
  version: '',
  website: '',
  logo: '',
  screenshots: [],
  compose: '',
  tutorial: ''
})

const activeTab = ref('edit')

const markdownHtml = computed(() => {
  return marked(form.value.tutorial || '')
})

const handleLogoSuccess = (response) => {
  form.value.logo = response.url
}

const handleScreenshotSuccess = (response) => {
  form.value.screenshots.push(response.url)
}

const handleComposeSuccess = async (response) => {
  try {
    const res = await fetch(response.url)
    form.value.compose = await res.text()
    ElMessage.success('Compose文件上传成功')
  } catch (error) {
    ElMessage.error('读取Compose文件失败')
  }
}

const handleSubmit = () => {
  // 这里将添加保存模板的逻辑
  console.log('表单数据:', form.value)
}

const handleReset = () => {
  form.value = {
    name: '',
    category: '',
    description: '',
    version: '',
    website: '',
    logo: '',
    screenshots: [],
    compose: '',
    tutorial: ''
  }
}
</script>

<style scoped>
.logo-uploader {
  width: 178px;
  height: 178px;
  border: 1px dashed #d9d9d9;
  border-radius: 6px;
  cursor: pointer;
  overflow: hidden;
}

.logo {
  width: 178px;
  height: 178px;
  object-fit: contain;
}

.compose-preview {
  margin-top: 10px;
}

.markdown-body {
  padding: 20px;
  border: 1px solid #dcdfe6;
  border-radius: 4px;
}
</style>