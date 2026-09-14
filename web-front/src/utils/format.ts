// 展示层格式化工具（design/10 §3.2 utils/format.ts）
// 支持 ISO8601 字符串（time.Time 序列化）和 unix 秒数两种输入

type TimeInput = string | number | undefined | null

function toDate(ts: TimeInput): Date | null {
  if (!ts) return null
  if (typeof ts === 'string') {
    const d = new Date(ts)
    return isNaN(d.getTime()) ? null : d
  }
  if (typeof ts === 'number' && ts > 0) {
    return new Date(ts * 1000)
  }
  return null
}

// ISO8601 字符串或 unix 秒 → 'yyyy-MM-dd HH:mm:ss'
export function formatTime(ts: TimeInput): string {
  const d = toDate(ts)
  if (!d) return '-'
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`
}

// ISO8601 字符串或 unix 秒 → 'yyyy-MM-dd'（仅日期，供列表时间戳展示）
export function formatDate(ts: TimeInput): string {
  const d = toDate(ts)
  if (!d) return ''
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}`
}

// ISO8601 字符串或 unix 秒 → 'YYYY-MM-DD'（本地时区，供日期选择器展示）
export function unixToDateInput(ts: TimeInput): string {
  const d = toDate(ts)
  if (!d) return ''
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}`
}

// 'YYYY-MM-DD' → ISO8601 字符串（本地时区当天 00:00:00，供 DB time.Time 字段）
export function dateInputToISO(date: string): string {
  if (!date) return ''
  return new Date(`${date}T00:00:00`).toISOString()
}

// 'YYYY-MM-DD' → unix 秒（保留，供后端仍接收 int64 的场景）
export function dateInputToUnix(date: string): number {
  if (!date) return 0
  return Math.floor(new Date(`${date}T00:00:00`).getTime() / 1000)
}
