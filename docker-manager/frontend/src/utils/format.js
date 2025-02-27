import dayjs from 'dayjs'

export const formatTime = (timestamp) => {
  if (!timestamp) return '-'
  
  // 处理字符串格式的时间
  if (typeof timestamp === 'string') {
    return dayjs(timestamp).format('YYYY-MM-DD HH:mm:ss')
  }
  
  // 处理 Unix 时间戳
  return dayjs(timestamp * 1000).format('YYYY-MM-DD HH:mm:ss')
}