import { viteLogo } from './src/core/config'
import Banner from 'vite-plugin-banner'
import * as path from 'path'
import { loadEnv } from 'vite'
import vuePlugin from '@vitejs/plugin-vue'
import vueDevTools from 'vite-plugin-vue-devtools'
import VueFilePathPlugin from 'vite-vue-path-map'
import { svgBuilder } from 'vite-auto-import-svg'
import vueRootValidator from 'vite-check-multiple-dom'
import { AddSecret } from './vitePlugin/secret'
import UnoCSS from '@unocss/vite'

// @see https://cn.vitejs.dev/config/
export default ({ mode }) => {
  AddSecret('')
  const env = loadEnv(mode, process.cwd())
  viteLogo()

  const timestamp = Date.parse(new Date())

  const alias = {
    '@': path.resolve(import.meta.dirname, './src'),
    vue$: 'vue/dist/vue.runtime.esm-bundler.js'
  }

  // base 支持环境变量覆盖：VITE_BASE=/admin/ 时构建产物资源带 /admin/ 前缀（子路径部署）
  // 默认 '/'（根路径部署，与 GVA 原版一致）
  const base = env.VITE_BASE || '/'
  const root = './'
  const outDir = 'admin'

  const config = {
    base: base, // 编译后js导入的资源路径
    root: root, // index.html文件所在位置
    publicDir: 'public', // 静态资源文件夹
    resolve: {
      alias
    },
    css: {
      preprocessorOptions: {
        scss: {
          api: 'modern-compiler' // or "modern"
        }
      }
    },
    server: {
      // 如果使用docker-compose开发模式，设置为false
      open: true,
      port: Number(env.VITE_CLI_PORT),
      proxy: {
        // hrc 业务接口（/api/v1/**）原样转发，不能去掉 /api 前缀
        // （后端 hrc 路由就注册在 /api/v1 下，见 server/initialize/router_biz.go）
        // 注意：必须声明在 VITE_BASE_API 之前，Vite 按声明顺序匹配代理
        [`${env.VITE_BASE_API}/v1`]: {
          // 需要代理的路径   例如 '/api/v1'
          target: `${env.VITE_BASE_PATH}:${env.VITE_SERVER_PORT}/`, // 代理到 目标路径
          changeOrigin: true
        },
        // 上传文件静态资源（uploads/file/xxx，企业 Logo/证照预览用）
        '/uploads': {
          target: `${env.VITE_BASE_PATH}:${env.VITE_SERVER_PORT}/`, // 代理到 目标路径
          changeOrigin: true
        },
        // 把key的路径代理到target位置
        // detail: https://cli.vuejs.org/config/#devserver-proxy
        [env.VITE_BASE_API]: {
          // 需要代理的路径   例如 '/api'
          target: `${env.VITE_BASE_PATH}:${env.VITE_SERVER_PORT}/`, // 代理到 目标路径
          changeOrigin: true,
          rewrite: (path) =>
            path.replace(new RegExp('^' + env.VITE_BASE_API), '')
        },
        '/plugin': {
          // 需要代理的路径   例如 '/api'
          target: `https://plugin.gin-vue-admin.com/api/`, // 代理到 目标路径
          changeOrigin: true,
          rewrite: (path) =>
            path.replace(new RegExp('^/plugin'), '')
        }
      }
    },
    build: {
      manifest: false, // 是否产出manifest.json
      sourcemap: false, // 是否产出sourcemap.json
      outDir: outDir, // 产出目录
      target: 'es2015',
      rolldownOptions: {
        output: {
          entryFileNames: 'assets/087AC4D233B64EB0[name].[hash].js',
          chunkFileNames: 'assets/087AC4D233B64EB0[name].[hash].js',
          assetFileNames: 'assets/087AC4D233B64EB0[name].[hash].[ext]'
        }
      }
    },
    plugins: [
      env.VITE_POSITION === 'open' &&
      vueDevTools({ launchEditor: env.VITE_EDITOR }),
      vuePlugin(),
      svgBuilder(['./src/plugin/', './src/assets/icons/'], base, outDir, 'assets', mode),
      [Banner(`\n Build based on gin-vue-admin \n Time : ${timestamp}`)],
      VueFilePathPlugin('./src/pathInfo.json'),
      UnoCSS(),
      vueRootValidator()
    ]
  }
  return config
}
