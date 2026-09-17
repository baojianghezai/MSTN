import { defineConfig, loadEnv } from 'vite'
import vue from '@vitejs/plugin-vue'
import UnoCSS from '@unocss/vite'
import path from 'node:path'

// 前台开发服务器
// 端口 8081（避开后台 8080）；/api/v1 原样转发到后端（后端 hrc 路由就注册在 /api/v1 下）
// 教训（后台踩过）：代理绝不能再去掉 /api 前缀，否则接口 404
export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd())
  const target = env.VITE_API_TARGET || 'http://127.0.0.1:8888'

  return {
    // 必须注册 vue 插件，否则 .vue 文件不会被编译（模板会被当普通 JS 解析报错）
    plugins: [vue(), UnoCSS()],
    resolve: {
      alias: {
        '@': path.resolve(import.meta.dirname, 'src')
      }
    },
     build: {
      outDir: 'www',
    },
    server: {
      port: Number(env.VITE_PORT || 8081),
      watch: {
        // AI 编辑工具的原子写临时目录（.xxx.vue.<pid>.<uuid>.tmpdir）会让 chokidar 在 Windows 上
        // watch 到被锁文件抛 EBUSY 并整进程崩溃（vite 8 实测两次），直接忽略该模式
        ignored: ['**/.*.tmpdir/**']
      },
      proxy: {
        // hrc 业务接口：原样转发，保留 /api/v1 前缀
        '/api/v1': {
          target,
          changeOrigin: true,
          // 在线对话 WebSocket（/api/v1/ws/chat）需要 ws 升级转发
          ws: true
        },
        // 上传文件静态资源（如 uploads/file/xxx）
        '/uploads': {
          target,
          changeOrigin: true
        }
      }
    }
  }
})
