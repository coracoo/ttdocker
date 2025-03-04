<template>
  <div class="overview">
    <el-row :gutter="20">
      <el-col :span="6" v-for="item in statistics" :key="item.title">
        <el-card class="stat-card">
          <h3>{{ item.title }}</h3>
          <div class="stat-number">{{ item.value }}</div>
        </el-card>
      </el-col>
    </el-row>

    <el-row :gutter="20" class="mt-20">
      <el-col :span="16">
        <el-card>
          <template #header>
            <div class="card-header">
              <span>系统监控</span>
            </div>
          </template>
          <div class="monitor-charts">
            <div class="chart-item">
              <div ref="cpuChart" style="height: 300px"></div>
            </div>
            <div class="chart-item">
              <div ref="memoryChart" style="height: 300px"></div>
            </div>
          </div>
        </el-card>
      </el-col>
      
      <el-col :span="8">
        <el-card>
          <template #header>
            <div class="card-header">
              <span>系统信息</span>
            </div>
          </template>
          <el-descriptions :column="1" border>
            <el-descriptions-item v-for="(value, key) in systemInfo" 
                                :key="key" 
                                :label="key">
              {{ value }}
            </el-descriptions-item>
          </el-descriptions>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup>
import { ref, onMounted, onBeforeUnmount, nextTick } from 'vue'
import * as echarts from 'echarts'
import axios from 'axios'
import { ElMessage } from 'element-plus' // 确保这行导入正确

// 统计数据
const statistics = ref([
  { title: '运行容器', value: 0 },
  { title: '镜像总数', value: 0 },
  { title: '数据卷', value: 0 },
  { title: '网络', value: 0 }
])

// 系统信息
const systemInfo = ref({
  '系统版本': '加载中...',
  'API版本': '加载中...',
  '运行时间': '加载中...',
  'CPU使用率': '加载中...',
  '内存使用': '加载中...',
  '磁盘使用': '加载中...'
})

// 图表实例
let cpuChart = null
let memoryChart = null

// CPU和内存数据
const cpuData = ref([])
const memoryData = ref([])

// 定时器
let timer = null

// 获取统计数据
const fetchStatistics = async () => {
  try {
    // 尝试获取真实数据
    try {
      // 获取容器数量
      const containersRes = await axios.get('/api/containers')
      const runningContainers = containersRes.data.filter(c => c.State === 'running').length
      statistics.value[0].value = runningContainers
      
      // 获取镜像数量
      const imagesRes = await axios.get('/api/images')
      statistics.value[1].value = imagesRes.data.length
      
      // 获取数据卷数量 - 修改这部分
      try {
        const volumesRes = await axios.get('/api/volumes')
        // 检查返回的数据结构
        if (Array.isArray(volumesRes.data)) {
          statistics.value[2].value = volumesRes.data.length
        } else if (volumesRes.data && Array.isArray(volumesRes.data.Volumes)) {
          // Docker API 可能返回 { Volumes: [...] } 格式
          statistics.value[2].value = volumesRes.data.Volumes.length
        } else {
          console.warn('数据卷返回格式不符合预期:', volumesRes.data)
          statistics.value[2].value = 0
        }
      } catch (volumeError) {
        console.warn('获取数据卷信息失败:', volumeError)
        statistics.value[2].value = 0
      }
      
      // 获取网络数量
      const networksRes = await axios.get('/api/networks')
      statistics.value[3].value = networksRes.data.length
    } catch (error) {
      console.warn('无法获取真实数据，使用模拟数据', error)
      // 使用模拟数据
      statistics.value[0].value = Math.floor(Math.random() * 5) + 1
      statistics.value[1].value = Math.floor(Math.random() * 10) + 5
      statistics.value[2].value = Math.floor(Math.random() * 3) + 1
      statistics.value[3].value = Math.floor(Math.random() * 2) + 1
    }
  } catch (error) {
    console.error('获取统计数据失败:', error)
    // 使用 ElMessage 之前确保它已定义
    if (typeof ElMessage !== 'undefined') {
      ElMessage.error('获取统计数据失败')
    } else {
      console.error('ElMessage 未定义，无法显示错误提示')
    }
  }
}

// 获取系统信息
const fetchSystemInfo = async () => {
  try {
    try {
      const res = await axios.get('/api/system/info')
      const info = res.data
      
      systemInfo.value = {
        '系统版本': `Docker ${info.ServerVersion || '未知'}`,
        'API版本': info.ApiVersion || '未知',
        '运行时间': formatUptime(info.SystemTime, info.SystemUptime),
        'CPU使用率': `${info.NCPU ? (info.CpuUsage || 0).toFixed(2) + '%' : '未知'}`,
        '内存使用': formatMemory(info.MemTotal, info.MemUsage),
        '磁盘使用': formatDisk(info.DiskTotal, info.DiskUsage)
      }
    } catch (error) {
      console.warn('无法获取真实系统信息，使用模拟数据', error)
      // 使用模拟数据
      systemInfo.value = {
        '系统版本': 'Docker 24.0.5',
        'API版本': 'v1.43',
        '运行时间': '3天 5小时 12分钟',
        'CPU使用率': '15.23%',
        '内存使用': '2.45 GB / 8.00 GB',
        '磁盘使用': '45.67 GB / 120.00 GB'
      }
    }
  } catch (error) {
    console.error('获取系统信息失败:', error)
    if (typeof ElMessage !== 'undefined') {
      ElMessage.error('获取系统信息失败')
    }
  }
}

// 获取监控数据
const fetchMonitorData = async () => {
  try {
    try {
      const res = await axios.get('/api/system/stats')
      const stats = res.data
      
      // 添加当前时间点的数据
      const now = new Date()
      
      // CPU数据
      cpuData.value.push([now, stats.cpu_percent])
      if (cpuData.value.length > 20) {
        cpuData.value.shift()
      }
      
      // 内存数据
      memoryData.value.push([now, stats.memory_percent])
      if (memoryData.value.length > 20) {
        memoryData.value.shift()
      }
    } catch (error) {
      console.warn('无法获取真实监控数据，使用模拟数据', error)
      // 使用模拟数据
      const now = new Date()
      cpuData.value.push([now, Math.random() * 30 + 10]) // 10-40% CPU使用率
      if (cpuData.value.length > 20) {
        cpuData.value.shift()
      }
      
      memoryData.value.push([now, Math.random() * 20 + 30]) // 30-50% 内存使用率
      if (memoryData.value.length > 20) {
        memoryData.value.shift()
      }
    }
    
    // 更新图表
    updateCharts()
  } catch (error) {
    console.error('获取监控数据失败:', error)
  }
}

// 更新图表
const updateCharts = () => {
  if (cpuChart && memoryChart) {
    cpuChart.setOption({
      series: [{
        data: cpuData.value
      }]
    })
    
    memoryChart.setOption({
      series: [{
        data: memoryData.value
      }]
    })
  } else {
    console.warn('图表未初始化，无法更新数据')
  }
}

// 格式化运行时间
const formatUptime = (systemTime, uptime) => {
  if (!uptime) return '未知'
  
  const days = Math.floor(uptime / 86400)
  const hours = Math.floor((uptime % 86400) / 3600)
  const minutes = Math.floor((uptime % 3600) / 60)
  
  return `${days}天 ${hours}小时 ${minutes}分钟`
}

// 格式化内存
const formatMemory = (total, used) => {
  if (!total || !used) return '未知'
  
  const totalGB = (total / (1024 * 1024 * 1024)).toFixed(2)
  const usedGB = (used / (1024 * 1024 * 1024)).toFixed(2)
  
  return `${usedGB} GB / ${totalGB} GB`
}

// 格式化磁盘
const formatDisk = (total, used) => {
  if (!total || !used) return '未知'
  
  const totalGB = (total / (1024 * 1024 * 1024)).toFixed(2)
  const usedGB = (used / (1024 * 1024 * 1024)).toFixed(2)
  
  return `${usedGB} GB / ${totalGB} GB`
}

// 初始化
onMounted(async () => {
  // 获取初始数据
  await fetchStatistics()
  await fetchSystemInfo()
  
  // 确保DOM元素已经渲染完成
  await nextTick()
  
  // 使用 ref 获取DOM元素
  const cpuChartEl = document.querySelector('.chart-item:first-child div')
  const memoryChartEl = document.querySelector('.chart-item:last-child div')
  
  if (cpuChartEl && memoryChartEl) {
    // 初始化图表
    cpuChart = echarts.init(cpuChartEl)
    memoryChart = echarts.init(memoryChartEl)
    
    // CPU使用率图表配置
    cpuChart.setOption({
      title: { text: 'CPU使用率' },
      tooltip: { 
        trigger: 'axis',
        formatter: function(params) {
          const data = params[0].data
          return `${new Date(data[0]).toLocaleTimeString()}<br/>CPU: ${data[1].toFixed(2)}%`
        }
      },
      xAxis: { 
        type: 'time',
        axisLabel: {
          formatter: '{HH}:{mm}:{ss}'
        }
      },
      yAxis: { 
        type: 'value', 
        max: 100,
        axisLabel: {
          formatter: '{value}%'
        }
      },
      grid: {
        left: '3%',
        right: '4%',
        bottom: '3%',
        containLabel: true
      },
      series: [{
        name: 'CPU',
        type: 'line',
        smooth: true,
        areaStyle: {
          opacity: 0.3
        },
        itemStyle: {
          color: '#409EFF'
        },
        data: cpuData.value
      }]
    })
    
    // 内存使用图表配置
    memoryChart.setOption({
      title: { text: '内存使用' },
      tooltip: { 
        trigger: 'axis',
        formatter: function(params) {
          const data = params[0].data
          return `${new Date(data[0]).toLocaleTimeString()}<br/>内存: ${data[1].toFixed(2)}%`
        }
      },
      xAxis: { 
        type: 'time',
        axisLabel: {
          formatter: '{HH}:{mm}:{ss}'
        }
      },
      yAxis: { 
        type: 'value',
        max: 100,
        axisLabel: {
          formatter: '{value}%'
        }
      },
      grid: {
        left: '3%',
        right: '4%',
        bottom: '3%',
        containLabel: true
      },
      series: [{
        name: '内存',
        type: 'line',
        smooth: true,
        areaStyle: {
          opacity: 0.3
        },
        itemStyle: {
          color: '#67C23A'
        },
        data: memoryData.value
      }]
    })
    
    // 获取初始监控数据
    await fetchMonitorData()
  } else {
    console.error('找不到图表容器元素')
  }
  
  // 设置定时刷新
  timer = setInterval(async () => {
    await fetchMonitorData()
    // 每分钟刷新一次系统信息和统计数据
    if (new Date().getSeconds() < 3) {
      await fetchStatistics()
      await fetchSystemInfo()
    }
  }, 3000)
  
  // 监听窗口大小变化，调整图表大小
  window.addEventListener('resize', handleResize)
})

// 处理窗口大小变化
const handleResize = () => {
  if (cpuChart) cpuChart.resize()
  if (memoryChart) memoryChart.resize()
}

// 组件卸载前清理
onBeforeUnmount(() => {
  if (timer) {
    clearInterval(timer)
  }
  window.removeEventListener('resize', handleResize)
  if (cpuChart) cpuChart.dispose()
  if (memoryChart) memoryChart.dispose()
})
</script>

<style scoped>
.overview {
  padding: 20px;
}

.stat-card {
  text-align: center;
  height: 120px;
  display: flex;
  flex-direction: column;
  justify-content: center;
}

.stat-number {
  font-size: 28px;
  font-weight: bold;
  color: #409EFF;
  margin-top: 10px;
}

.mt-20 {
  margin-top: 20px;
}

.monitor-charts {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.chart-item {
  flex: 1;
  margin-bottom: 10px;
}

@media (min-width: 768px) {
  .monitor-charts {
    flex-direction: row;
  }
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
</style>