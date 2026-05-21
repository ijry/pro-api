import { defineConfig } from 'vitepress'

export default defineConfig({
  title: 'proapi',
  description: '一站式大模型 API 中转网关 / All-in-one LLM API Gateway',
  cleanUrls: true,
  lastUpdated: true,
  themeConfig: {
    logo: '/logo.svg',
    socialLinks: [{ icon: 'github', link: 'https://github.com/proapi/proapi' }],
    search: { provider: 'local' },
  },
  locales: {
    root: {
      label: '简体中文',
      lang: 'zh-CN',
      link: '/zh/',
      themeConfig: {
        nav: [
          { text: '指南', link: '/zh/guide/introduction' },
          { text: '架构', link: '/zh/architecture/overview' },
          { text: 'API', link: '/zh/api/overview' },
          { text: '部署', link: '/zh/deployment/docker' },
          { text: '二次开发', link: '/zh/development/overview' },
        ],
        sidebar: {
          '/zh/guide/': [
            {
              text: '快速上手',
              items: [
                { text: '项目介绍', link: '/zh/guide/introduction' },
                { text: '5 分钟跑起来', link: '/zh/guide/quickstart' },
              ],
            },
          ],
        },
      },
    },
    en: {
      label: 'English',
      lang: 'en-US',
      link: '/en/',
      themeConfig: {
        nav: [{ text: 'Guide', link: '/en/' }],
      },
    },
  },
})
