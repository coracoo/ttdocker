<template>
  <el-dialog
    v-model="dialogVisible"
    title="容器终端"
    width="80%"
    :close-on-click-modal="false"
    :before-close="handleClose"
    class="terminal-dialog"	
  >
    <div class="terminal-container">
      <!-- 左侧面板 -->
      <div class="left-panel">
        <!-- 常用命令区域 -->
        <div class="quick-commands">
          <h4>常用命令</h4>
          <el-form :model="terminalForm" label-width="80px">
            <el-form-item label="命令">
				<el-button type="primary" @click="connectTerminal">执行</el-button>
				<el-button @click="testConnection" type="info">测试连接</el-button>
              <el-select v-model="terminalForm.command" style="width: 100%">
                <el-option label="/bin/bash" value="/bin/bash" />
                <el-option label="/bin/sh" value="/bin/sh" />
                <el-option label="自定义" value="custom" />
              </el-select>
            </el-form-item>
            <el-form-item v-if="terminalForm.command === 'custom'" label="自定义">
              <el-input v-model="terminalForm.customCommand" placeholder="请输入命令" />
            </el-form-item>
            <el-form-item>
              <el-button type="primary" @click="connectTerminal">交互式终端</el-button>
              <el-button type="success" @click="executeSimpleCommand">执行命令</el-button>
              <el-button @click="testConnection" type="info">测试连接</el-button>
            </el-form-item>
          </el-form>
        </div>
        
        <!-- 命令历史记录 -->
        <div class="command-history">
          <h4>历史命令</h4>
          <el-scrollbar height="200px">
            <ul class="history-list">
              <li v-for="(cmd, index) in commandHistory" 
                  :key="index" 
                  @click="executeHistoryCommand(cmd)"
                  class="history-item">
                {{ cmd }}
              </li>
            </ul>
          </el-scrollbar>
        </div>
      </div>
  
      <!-- 右侧终端显示区 -->
      <div class="terminal-panel">
        <div ref="terminalContainer" class="terminal-container"></div>
      </div>
    </div>
  </el-dialog>
</template>

<script setup>
import { ref, watch, onMounted, onBeforeUnmount, defineProps, defineEmits } from 'vue'
import { Terminal } from 'xterm'
import { FitAddon } from 'xterm-addon-fit'
// import { AttachAddon } from 'xterm-addon-attach';
import 'xterm/css/xterm.css'
import { ElMessage } from 'element-plus'

const props = defineProps({
  modelValue: {
    type: Boolean,
    default: false
  },
  container: {
    type: Object,
    default: () => ({})
  }
})

const emit = defineEmits(['update:modelValue'])
const commandHistory = ref([])
const dialogVisible = ref(props.modelValue)
const terminalContainer = ref(null)
const terminal = ref(null)
const fitAddon = ref(null)
const socket = ref(null)
const terminalForm = ref({
  command: '/bin/sh',
  customCommand: ''
})

// 监听对话框可见性变化
watch(() => props.modelValue, (val) => {
  dialogVisible.value = val
  if (val && props.container) {
    // 当对话框显示且有容器信息时，初始化终端
    setTimeout(() => {
      initTerminal()
    }, 100)
  }
})

// 监听对话框内部状态变化
watch(() => dialogVisible.value, (val) => {
  emit('update:modelValue', val)
  if (!val) {
    // 关闭WebSocket连接
    closeConnection()
  }
})

// 初始化终端函数
const initTerminal = () => {
  if (!terminalContainer.value) return
  
  // 如果已经有终端实例，先销毁
  if (terminal.value) {
    // 先将插件引用置为null
    if (fitAddon.value) {
      fitAddon.value = null
    }
    
    try {
      terminal.value.dispose()
    } catch (e) {
      console.warn('销毁终端实例时出错:', e)
    }
    
    terminal.value = null
  }
  
  // 创建新的终端实例
  terminal.value = new Terminal({
    cursorBlink: true,
    theme: {
      background: '#1e1e1e',
      foreground: '#f0f0f0'
    },
    fontSize: 14,
    fontFamily: 'Consolas, "Courier New", monospace',
    scrollback: 1000
  })
  
  // 添加自适应大小插件
  fitAddon.value = new FitAddon()
  terminal.value.loadAddon(fitAddon.value)
  
  // 打开终端
  terminal.value.open(terminalContainer.value)
  fitAddon.value.fit()
  
  // 显示欢迎信息
  terminal.value.writeln('欢迎使用容器终端，请点击"执行"按钮连接到容器...')
}

// 连接终端
const connectTerminal = () => {
  if (!props.container || !props.container.Id) {
    ElMessage.error('容器信息不完整')
    return
  }
  
  // 如果已经有终端实例，先销毁
  if (terminal.value) {
    // 先检查并卸载插件，再销毁终端
    if (fitAddon.value) {
      try {
        // 不需要显式卸载，只需要将引用置为null
        fitAddon.value = null
      } catch (e) {
        console.warn('卸载终端插件时出错:', e)
      }
    }
    
    try {
      terminal.value.dispose()
    } catch (e) {
      console.warn('销毁终端实例时出错:', e)
    }
    
    terminal.value = null
  }
  
  // 创建新的终端实例
  terminal.value = new Terminal({
    cursorBlink: true,
    theme: {
      background: '#1e1e1e',
      foreground: '#f0f0f0'
    },
    fontSize: 14,
    fontFamily: 'Consolas, "Courier New", monospace',
    scrollback: 1000
  })
  
  // 添加自适应大小插件
  fitAddon.value = new FitAddon()
  terminal.value.loadAddon(fitAddon.value)
  
  // 打开终端
  terminal.value.open(terminalContainer.value)
  fitAddon.value.fit()
  
  // 连接WebSocket
  connectWebSocket()
  
  // 保存命令到历史记录
  const cmd = terminalForm.value.command === 'custom' 
    ? terminalForm.value.customCommand 
    : terminalForm.value.command
  
  if (cmd && !commandHistory.value.includes(cmd)) {
    commandHistory.value.unshift(cmd)
    // 限制历史记录数量
    if (commandHistory.value.length > 10) {
      commandHistory.value.pop()
    }
  }
}

// 执行历史命令
const executeHistoryCommand = (cmd) => {
  if (cmd === '/bin/sh' || cmd === '/bin/bash') {
    terminalForm.value.command = cmd
    terminalForm.value.customCommand = ''
  } else {
    terminalForm.value.command = 'custom'
    terminalForm.value.customCommand = cmd
  }
  connectTerminal()
}

// 连接WebSocket
const connectWebSocket = () => {
  if (!props.container || !props.container.Id) {
    ElMessage.error('容器信息不完整')
    return
  }
  
  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
  const host = window.location.host
  const containerId = props.container.Id
  const wsUrl = `${protocol}//${host}/api/containers/${containerId}/terminal`
  
  // 显示连接信息，帮助调试
  console.log('尝试连接WebSocket:', {
    protocol,
    host,
    containerId,
    wsUrl
  })
  terminal.value.writeln(`正在连接到: ${wsUrl}`)
  
  try {
    socket.value = new WebSocket(wsUrl)
    
    socket.value.onopen = () => {
      console.log('WebSocket连接成功')
      terminal.value.writeln('已连接到容器终端...')
      
      // 发送命令  发送终端大小发送到服务器端  发送到服务器端  发送到服务器端  发送到服务器端  发送到服务器端  发送到服务器端  发送到服务器端  发送到服务器端  发送到服务器端  发送到服务器端  发送到服务器端  发送到服务器端  发送到服务器端  发送到服务器端  发送到服务器端  发送到服务器端  发送到服务器端  发送到服务器端  发送到服务器端  发送到服务器端  发送到服务器端  发送到服务器端  发送到服务器端  发送到服务器端  发送到服务器端  发送到服务器端  发送到服务器端  发送到服务器端  发送到服务器端  发送到服务器端  发送到服务器端  发送到服务器端  发送到服务器端  发送到服务器端  发送到服务器端  发送到服务器端  发送到服务器端  发送到服务器端  发送到服务器端  发送到服务器端  发送到服务器端  发送到服务器端  发送到服务器端  发送到服务器端  发送到服务器端  发送到服务器端  发送到服务器端  发送到服务器端  发送到服务器端  发送到服务器端  发送到服务器端  发送到服务器端  发送到服务器端  发送到服务器端  发送到服务器端  发送到服务器端  发送到服务器端  发送到服务器端  发送到服务器端  发送到服务器端  发送到服务器端  发送到服务器端  发送到服务器端  发送到服务器端  发送到服务器端  发送到服务器端  发送到服务器端  发送到服务器端  发送到服务器端  发送到服务器端  发送到容器日志API: ${terminalHttpUrl}\x1b[0m`)
        return fetch(terminalHttpUrl)
          .then(terminalResponse => {
            if (terminalResponse.ok) {
              terminal.value.writeln(`\x1b[32m容器日志API可访问: ${terminalResponse.status}\x1b[0m`)
              terminal.value.writeln(`\x1b[32m这表明后端可能支持容器操作，但WebSocket终端可能未实现\x1b[0m`)
            } else {
              terminal.value.writeln(`\x1b[31m容器日志API返回错误: ${terminalResponse.status}\x1b[0m`)
              terminal.value.writeln(`\x1b[31m这表明后端可能不支持此容器ID或终端功能未实现\x1b[0m`)
            }
          })
          .catch(error => {
            terminal.value.writeln(`\x1b[31m容器日志API请求失败: ${error.message}\x1b[0m`)
          })
      } else {
        terminal.value.writeln(`\x1b[31mHTTP容器列表API返回错误: ${response.status}\x1b[0m`)
        terminal.value.writeln(`\x1b[31m这表明后端API可能不可用或路径错误\x1b[0m`)
      }
    })
    .catch(error => {
      terminal.value.writeln(`\x1b[31mHTTP请求失败: ${error.message}\x1b[0m`)
      terminal.value.writeln(`\x1b[31m这可能是网络问题或后端服务未运行\x1b[0m`)
    })
  
  // 检查后端路由是否正确配置
  terminal.value.writeln(`\r\n\x1b[33m检查后端路由配置:\x1b[0m`)
  terminal.value.writeln(`\x1b[33m1. 确认后端是否注册了 '/api/containers/:id/terminal' 路由\x1b[0m`)
  terminal.value.writeln(`\x1b[33m2. 确认后端是否正确处理WebSocket升级请求\x1b[0m`)
  terminal.value.writeln(`\x1b[33m3. 确认后端是否导入了 gorilla/websocket 包\x1b[0m`)
  
  // 尝试WebSocket连接
  try {
    terminal.value.writeln(`\r\n\x1b[33m尝试WebSocket连接...\x1b[0m`)
    const testSocket = new WebSocket(wsUrl)
    
    testSocket.onopen = () => {
      terminal.value.writeln(`\x1b[32mWebSocket连接成功!\x1b[0m`)
      testSocket.send(JSON.stringify({type: 'test', data: 'hello'}))
      
      // 5秒后关闭测试连接
      setTimeout(() => {
        testSocket.close()
      }, 5000)
    }
    
    testSocket.onmessage = (event) => {
      terminal.value.writeln(`\x1b[32m收到消息: ${typeof event.data === 'string' ? event.data.substring(0, 50) : '(binary data)'}\x1b[0m`)
    }
    
    testSocket.onerror = (error) => {
      terminal.value.writeln(`\x1b[31mWebSocket错误\x1b[0m`)
    }
    
    testSocket.onclose = (event) => {
      terminal.value.writeln(`\x1b[33mWebSocket连接关闭: 代码=${event.code}, 原因=${event.reason || '无'}\x1b[0m`)
      
      if (event.code === 1006) {
        terminal.value.writeln(`\x1b[31m连接异常关闭，可能是后端未处理WebSocket请求\x1b[0m`)
        terminal.value.writeln(`\x1b[33m建议检查:\x1b[0m`)
        terminal.value.writeln(`\x1b[33m1. 后端是否注册了WebSocket处理函数\x1b[0m`)
        terminal.value.writeln(`\x1b[33m2. 后端日志中是否有相关错误\x1b[0m`)
        terminal.value.writeln(`\x1b[33m3. 后端是否正确配置了CORS和WebSocket升级\x1b[0m`)
      }
    }
  } catch (error) {
    terminal.value.writeln(`\x1b[31m创建WebSocket连接失败: ${error.message}\x1b[0m`)
  }
}
</script>

<style>
/* 修改终端样式 */
.terminal-container {
  display: flex;
  min-height: 500px;
  background-color: #1e1e1e;
  padding: 5px;
  border-radius: 4px;
  gap: 10px;
}

.terminal {
  padding: 10px;
}

.left-panel {
  width: 250px;
  background-color: #f5f7fa;
  padding: 10px;
  border-radius: 4px;
}

.terminal-panel {
  flex: 1;
  background-color: #1e1e1e;
  border-radius: 4px;
  overflow: hidden;
}

.terminal-panel .terminal-container {
  height: 100%;
  width: 100%;
  padding: 0;
}

.history-item {
  cursor: pointer;
  padding: 5px 10px;
  list-style: none;
  border-bottom: 1px solid #ebeef5;
}

.history-item:hover {
  background-color: #ecf5ff;
  color: #409eff;
}

.history-list {
  padding: 0;
  margin: 0;
}

.terminal-dialog :deep(.el-dialog__body) {
  padding: 10px;
}

.quick-commands h4,
.command-history h4 {
  margin-top: 0;
  margin-bottom: 10px;
  color: #303133;
}

.command-history {
  margin-top: 20px;
}
</style>
