import { defineStore } from 'pinia'
import { ref } from 'vue'
import { getMessageUnreadCount } from '@/api/message'
import type { MessageScope } from '@/api/message'

export const useMessageStore = defineStore('message', () => {
  const unread = ref(0)

  const refreshUnread = async (scope: MessageScope) => {
    const { data } = await getMessageUnreadCount(scope)
    unread.value = data.unread
  }

  return { unread, refreshUnread }
})
