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
import { ref, onMounted } from 'vue'
import * as echarts from 'echarts'

const statistics = ref([
  { title: '运行容器', value: 8 },
  { title: '镜像总数', value: 10 },
  { title: '数据卷', value: 6 },
  { title: '网络', value: 6 }
])

const systemInfo = ref({
  '系统版本': 'Docker 24.0.7',
  'API版本': '1.42',
  '运行时间': '7天 5小时',
  'CPU使用率': '2.77%',
  '内存使用': '823.43 MB / 3576.83 MB',
  '磁盘使用': '6.65 GB / 78.37 GB'
})

// 初始化图表
onMounted(() => {
  const cpuChart = echarts.init(document.querySelector('.chart-item:first-child div'))
  const memoryChart = echarts.init(document.querySelector('.chart-item:last-child div'))
  
  // CPU使用率图表配置
  cpuChart.setOption({
    title: { text: 'CPU使用率' },
    tooltip: { trigger: 'axis' },
    xAxis: { type: 'time' },
    yAxis: { type: 'value', max: 100 },
    series: [{
      name: 'CPU',
      type: 'line',
      smooth: true,
      areaStyle: {},
      data: []  // 这里需要接入实时数据
    }]
  })
  
  // 内存使用图表配置
  memoryChart.setOption({
    title: { text: '内存使用' },
    tooltip: { trigger: 'axis' },
    xAxis: { type: 'time' },
    yAxis: { type: 'value' },
    series: [{
      name: '内存',
      type: 'line',
      smooth: true,
      areaStyle: {},
      data: []  // 这里需要接入实时数据
    }]
  })
})
</script>

<style scoped>
.overview {
  padding: 20px;
}

.stat-card {
  text-align: center;
}

.stat-number {
  font-size: 24px;
  font-weight: bold;
  color: #409EFF;
  margin-top: 10px;
}

.mt-20 {
  margin-top: 20px;
}

.monitor-charts {
  display: flex;
  gap: 20px;
}

.chart-item {
  flex: 1;
}
</style>