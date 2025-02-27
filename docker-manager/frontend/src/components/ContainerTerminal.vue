<template>
  <el-dialog
    v-model="visible"
    title="容器终端"
    width="80%"
    @close="handleClose"
    :fullscreen="false"
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
              <el-button type="primary" @click="connectTerminal">执行</el-button>
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
        <div ref="terminalContainer" class="terminal-window"></div>
      </div>
    </div>
  </el-dialog>
</template>

<script setup>
import { ref, onMounted, watch, computed, nextTick } from 'vue'
import { Terminal } from 'xterm'
import { FitAddon } from 'xterm-addon-fit'
import { AttachAddon } from 'xterm-addon-attach';  // 添加这行
import 'xterm/css/xterm.css'
import { ElMessage } from 'element-plus'

const props = defineProps({
  modelValue: Boolean,
  container: Object
})

const emit = defineEmits(['update:modelValue'])

// 修改 visible 的处理
const visible = computed({
  get: () => props.modelValue,
  set: (value) => emit('update:modelValue', value)
})

// 修改初始化终端的方法
const initTerminal = () => {
  console.log('初始化终端')
  if (!terminalContainer.value) {
    console.error('终端容器元素不存在')
    return
  }
  // 清除现有终端
  if (term.value) {
    try {
      term.value.dispose()
    } catch (error) {
      console.warn('终端清理警告:', error)
    }
  }
  
  terminalContainer.value.innerHTML = ''
  try {
    // 创建新终端
    term.value = new Terminal({
      cursorBlink: true,
      theme: {
        background: '#1e1e1e',
        foreground: '#ffffff'
      },
      fontSize: 14,
      rows: 30,
      cols: 100
    })
  fitAddon.value = new FitAddon()
  term.value.loadAddon(fitAddon.value)
  // 打开终端
  term.value.open(terminalContainer.value)
  fitAddon.value.fit()
  
  // 写入初始提示
  term.value.write('\x1b[1;32m欢迎使用终端\x1b[0m\r\n')
  
  console.log('终端初始化完成:', {
    term: !!term.value,
    container: !!terminalContainer.value
  })
} catch (error) {
  console.error('终端初始化错误:', error)
  ElMessage.error('终端初始化失败')
}
}

// 删除重复的 watch，只保留这一个
watch(() => visible.value, (newVal) => {
  if (newVal) {
    nextTick(() => {
      initTerminal()
    })
  }
}, { immediate: true })

const terminalForm = ref({
  command: '/bin/bash',
  customCommand: ''
})
const commandHistory = ref([])
const terminalContainer = ref(null)
const term = ref(null)
const fitAddon = ref(null)

// 添加 visible 属性的监听
watch(() => props.modelValue, (newVal) => {
  if (newVal) {
    // 当对话框打开时，延迟初始化终端
    setTimeout(() => {
      initTerminal()
    }, 100)
  }
})
// 修改 connectTerminal 方法
const connectTerminal = async () => {
  try {
    if (!term.value) {
      console.error('终端未初始化');
      await initTerminal(); // 确保终端已初始化
      if (!term.value) {
        ElMessage.error('终端初始化失败');
        return;
      }
    }

    // 获取当前命令
    const command = terminalForm.value.command === 'custom' 
      ? terminalForm.value.customCommand 
      : terminalForm.value.command;

    // 构建 WebSocket URL
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const host = import.meta.env.VITE_API_BASE_URL 
      ? new URL(import.meta.env.VITE_API_BASE_URL).host 
      : window.location.host;
    const wsUrl = `${protocol}//${host}/api/containers/${props.container.Id}/exec?cmd=${encodeURIComponent(command)}`;
    
    console.log('正在连接到终端服务器:', wsUrl);
    
    const ws = new WebSocket(wsUrl);
    let attachAddon = null;

    ws.onopen = () => {
      if (!term.value) return;
      console.log('终端连接成功');
      term.value.write('\r\n*** 连接成功 ***\r\n');
      
      try {
        attachAddon = new AttachAddon(ws);
        term.value.loadAddon(attachAddon);
        
        // 添加到历史记录
        if (command && !commandHistory.value.includes(command)) {
          commandHistory.value.unshift(command);
          if (commandHistory.value.length > 10) {
            commandHistory.value.pop();
          }
        }
      } catch (error) {
        console.error('AttachAddon 加载失败:', error);
        ws.close();
      }
    };

    ws.onerror = (error) => {
      console.error('终端连接错误:', error);
      if (term.value) {
        term.value.write('\r\n*** 连接错误 ***\r\n');
      }
      ElMessage.error('终端连接失败，请检查网络连接');
    };

    ws.onclose = () => {
      console.log('终端连接已关闭');
      if (term.value) {
        term.value.write('\r\n*** 连接已关闭 ***\r\n');
      }
      if (attachAddon) {
        try {
          attachAddon.dispose();
        } catch (error) {
          console.warn('AttachAddon 清理警告:', error);
        }
      }
    };

  } catch (error) {
    console.error('终端连接失败:', error);
    if (term.value) {
      term.value.write('\r\n*** 连接失败 ***\r\n');
    }
    ElMessage.error('终端连接失败: ' + error.message);
  }
};

// 添加执行历史命令的方法
const executeHistoryCommand = (cmd) => {
  terminalForm.value.command = 'custom';
  terminalForm.value.customCommand = cmd;
  connectTerminal();
};

// 在组件挂载时初始化终端
onMounted(() => {
  if (props.modelValue) {
    initTerminal()
  }
})
// 修改关闭处理
const handleClose = () => {
  if (term.value) {
    try {
      term.value.dispose()
    } catch (error) {
      console.warn('终端清理警告:', error)
    }
    term.value = null
  }
  emit('update:modelValue', false)
}
</script>

<style>
/* 修改终端样式 */
.terminal-window {
  min-height: 400px;
  background: #1e1e1e;
  padding: 10px;
  border-radius: 4px;
}

.terminal {
  padding: 10px;
}
.terminal-container {
  display: flex;
  gap: 20px;
}

.left-panel {
  width: 250px;
}

.terminal-panel {
  flex: 1;
}

.history-item {
  cursor: pointer;
  padding: 5px 10px;
  list-style: none;
}

.history-item:hover {
  background-color: #f5f7fa;
}

.history-list {
  padding: 0;
  margin: 0;
}
</style>
