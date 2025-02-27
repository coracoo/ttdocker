<template>
  <el-dialog
    v-model="dialogVisible"
    title="Docker 设置"
    width="800px"
  >
    <el-tabs v-model="activeTab">
      <el-tab-pane label="镜像仓库" name="registry">
        <el-form :model="registryForm" label-width="100px">
          <el-form-item>
            <div class="registry-header">
              <el-button type="primary" @click="addRegistry">新建</el-button>
            </div>
          </el-form-item>
          
          <el-table :data="registryForm.registries" style="width: 100%">
            <el-table-column prop="name" label="别名" width="180" />
            <el-table-column prop="url" label="仓库链接" />
            <el-table-column label="操作" width="150">
              <template #default="scope">
                <el-button link type="primary" @click="editRegistry(scope.row)">编辑</el-button>
                <el-button link type="danger" @click="removeRegistry(scope.$index)">移除</el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-form>
      </el-tab-pane>

      <el-tab-pane label="代理设置" name="proxy">
        <el-form :model="proxyForm" label-width="120px">
          <el-form-item label="启用代理">
            <el-switch v-model="proxyForm.enabled" />
          </el-form-item>
          
          <template v-if="proxyForm.enabled">
            <el-form-item label="HTTP 代理">
              <el-input v-model="proxyForm.http" placeholder="http://proxy:port" />
            </el-form-item>
            <el-form-item label="HTTPS 代理">
              <el-input v-model="proxyForm.https" placeholder="https://proxy:port" />
            </el-form-item>
            <el-form-item label="无需代理">
              <el-input v-model="proxyForm.no" placeholder="localhost,127.0.0.1" />
            </el-form-item>
          </template>
        </el-form>
      </el-tab-pane>

      <el-tab-pane label="镜像加速" name="mirrors">
        <el-form :model="mirrorForm" label-width="120px">
          <el-form-item label="镜像加速器">
            <el-select
              v-model="mirrorForm.mirrors"
              multiple
              filterable
              allow-create
              default-first-option
              style="width: 100%"
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
      </el-tab-pane>
    </el-tabs>

    <template #footer>
      <span class="dialog-footer">
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="saveSettings">保存</el-button>
      </span>
    </template>
  </el-dialog>
</template>

<script setup>
import { ref, defineEmits, defineProps } from 'vue'
import { ElMessage } from 'element-plus'
import api from '../api'

const props = defineProps({
  modelValue: Boolean
})

const emit = defineEmits(['update:modelValue'])

const dialogVisible = computed({
  get: () => props.modelValue,
  set: (value) => emit('update:modelValue', value)
})

const activeTab = ref('registry')

const registryForm = ref({
  registries: []
})

const proxyForm = ref({
  enabled: false,
  http: '',
  https: '',
  no: ''
})

const mirrorForm = ref({
  mirrors: []
})

const defaultMirrors = [
  { label: '阿里云', value: 'https://mirror.aliyuncs.com' },
  { label: '腾讯云', value: 'https://mirror.ccs.tencentyun.com' },
  { label: '网易', value: 'https://hub-mirror.c.163.com' },
  { label: '中科大', value: 'https://docker.mirrors.ustc.edu.cn' }
]

// 加载配置
const loadSettings = async () => {
  try {
    const config = await api.images.getProxy()
    proxyForm.value = {
      enabled: !!(config['http-proxy'] || config['https-proxy']),
      http: config['http-proxy'] || '',
      https: config['https-proxy'] || '',
      no: config['no-proxy'] || ''
    }
    mirrorForm.value.mirrors = config['registry-mirrors'] || []
    registryForm.value.registries = config.registries || []
  } catch (error) {
    ElMessage.error('加载配置失败')
  }
}

// 保存配置
const saveSettings = async () => {
  try {
    const config = {
      'http-proxy': proxyForm.value.enabled ? proxyForm.value.http : '',
      'https-proxy': proxyForm.value.enabled ? proxyForm.value.https : '',
      'no-proxy': proxyForm.value.enabled ? proxyForm.value.no : '',
      'registry-mirrors': mirrorForm.value.mirrors,
      registries: registryForm.value.registries
    }
    
    await api.images.updateProxy(config)
    ElMessage.success('配置已更新，请重启 Docker 服务以生效')
    dialogVisible.value = false
  } catch (error) {
    ElMessage.error('保存配置失败')
  }
}

// 仓库管理
const addRegistry = () => {
  registryForm.value.registries.push({
    name: '',
    url: ''
  })
}

const editRegistry = (registry) => {
  // 实现编辑逻辑
}

const removeRegistry = (index) => {
  registryForm.value.registries.splice(index, 1)
}

// 监听对话框打开
watch(() => dialogVisible.value, (val) => {
  if (val) {
    loadSettings()
  }
})
</script>

<style scoped>
.registry-header {
  margin-bottom: 16px;
}
</style>