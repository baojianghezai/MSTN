/**
 * 网站配置文件
 */

const greenText = (text) => `\x1b[32m${text}\x1b[0m`

export const config = {
  appName: 'MSTN',
  showViteLogo: false,
  keepAliveTabs: false,
  logs: []
}

export const viteLogo = () => {
  if (config.showViteLogo) {
    console.log(
      greenText(
        `> 欢迎使用MSTN系统`
      )
    )
    console.log('\n')
  }
}

export default config
