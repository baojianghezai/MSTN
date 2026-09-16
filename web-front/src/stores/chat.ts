// 在线对话连接管理：优先 WebSocket 实时，断线自动退避重连并降级为轮询未读
import { defineStore } from 'pinia'
import { ref } from 'vue'
import { getChatUnread } from '@/api/chat'
import type { ChatMessageItem, ChatScope } from '@/api/chat'

type MessageListener = (message: ChatMessageItem) => void

const RECONNECT_DELAY = 5000
const POLL_INTERVAL = 5000

export const useChatStore = defineStore('chat', () => {
  const unread = ref(0)
  const connected = ref(false)
  const scope = ref<ChatScope>('personal')

  let socket: WebSocket | null = null
  let reconnectTimer: number | null = null
  let pollTimer: number | null = null
  let currentToken = ''
  let manualClose = false
  const listeners = new Set<MessageListener>()

  const refreshUnread = async () => {
    try {
      const { data } = await getChatUnread(scope.value)
      unread.value = data.unread
    } catch {
      // 未登录/网络异常时忽略
    }
  }

  const onMessage = (listener: MessageListener) => {
    listeners.add(listener)
    return () => listeners.delete(listener)
  }

  const emit = (message: ChatMessageItem) => {
    listeners.forEach((listener) => listener(message))
  }

  const stopPolling = () => {
    if (pollTimer !== null) {
      window.clearInterval(pollTimer)
      pollTimer = null
    }
  }

  const startPolling = () => {
    if (pollTimer !== null) return
    pollTimer = window.setInterval(refreshUnread, POLL_INTERVAL)
  }

  const scheduleReconnect = () => {
    if (manualClose || reconnectTimer !== null) return
    reconnectTimer = window.setTimeout(() => {
      reconnectTimer = null
      openSocket()
    }, RECONNECT_DELAY)
  }

  const openSocket = () => {
    if (!currentToken) return
    const protocol = window.location.protocol === 'https:' ? 'wss' : 'ws'
    const url = `${protocol}://${window.location.host}/api/v1/ws/chat?token=${encodeURIComponent(currentToken)}`
    try {
      socket = new WebSocket(url)
    } catch {
      startPolling()
      scheduleReconnect()
      return
    }
    socket.onopen = () => {
      connected.value = true
      stopPolling()
    }
    socket.onmessage = (event) => {
      try {
        const payload = JSON.parse(event.data)
        if (payload?.type === 'message' && payload.data) {
          unread.value += 1
          emit(payload.data as ChatMessageItem)
        }
      } catch {
        // 忽略非 JSON 帧（如心跳）
      }
    }
    socket.onclose = () => {
      connected.value = false
      socket = null
      if (manualClose) return
      startPolling()
      scheduleReconnect()
    }
    socket.onerror = () => {
      try {
        socket?.close()
      } catch {
        // ignore
      }
    }
  }

  /** 登录后连接：scope 决定未读接口，token 用于 WS 鉴权 */
  const connect = (nextScope: ChatScope, token: string) => {
    scope.value = nextScope
    currentToken = token || ''
    manualClose = false
    if (reconnectTimer !== null) {
      window.clearTimeout(reconnectTimer)
      reconnectTimer = null
    }
    if (socket) {
      socket.onclose = null
      socket.close()
      socket = null
    }
    connected.value = false
    if (!currentToken) {
      startPolling()
      return
    }
    openSocket()
    refreshUnread()
  }

  /** 断开（登出时调用） */
  const disconnect = () => {
    manualClose = true
    if (reconnectTimer !== null) {
      window.clearTimeout(reconnectTimer)
      reconnectTimer = null
    }
    stopPolling()
    if (socket) {
      socket.onclose = null
      socket.close()
      socket = null
    }
    connected.value = false
    unread.value = 0
  }

  return { unread, connected, scope, refreshUnread, onMessage, connect, disconnect }
})