// 展示层格式化工具（design/10 §3.2 utils/format.ts）

// unix 秒 → 'yyyy-MM-dd HH:mm:ss'
export function formatTime(ts: number | undefined): string {
  if (!ts) return '-'
  const d = new Date(ts * 1000)
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`
}

// unix 秒 → 'yyyy-MM-dd'（仅日期，供列表时间戳展示）
export function formatDate(ts: number | undefined): string {
  if (!ts) return ''
  const d = new Date(ts * 1000)
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}`
}

// unix 秒 → 'YYYY-MM-DD'（本地时区，供日期选择器展示）
export function unixToDateInput(ts: number): string {
  if (!ts) return ''
  const d = new Date(ts * 1000)
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}`
}

// 'YYYY-MM-DD' → unix 秒（本地时区当天 00:00:00，避免 UTC 偏移导致差一天）
export function dateInputToUnix(date: string): number {
  if (!date) return 0
  return Math.floor(new Date(`${date}T00:00:00`).getTime() / 1000)
}
