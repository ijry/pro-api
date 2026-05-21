import { defineConfig, type DefaultTheme } from 'vitepress'

const zhSidebar: DefaultTheme.Sidebar = [
  {
    text: '快速上手',
    collapsed: false,
    items: [
      { text: '项目介绍', link: '/zh/guide/introduction' },
      { text: '5 分钟跑起来', link: '/zh/guide/quickstart' },
      { text: '安装', link: '/zh/guide/installation' },
      { text: '配置', link: '/zh/guide/configuration' },
      { text: '升级', link: '/zh/guide/upgrade' },
    ],
  },
  {
    text: '架构',
    collapsed: true,
    items: [
      { text: '架构总览', link: '/zh/architecture/overview' },
      { text: '适配器层', link: '/zh/architecture/adapter-layer' },
      { text: '计费机制', link: '/zh/architecture/billing' },
      { text: '限流策略', link: '/zh/architecture/ratelimit' },
      { text: '渠道调度与熔断', link: '/zh/architecture/channel-scheduling' },
    ],
  },
  {
    text: '功能模块',
    collapsed: true,
    items: [
      { text: '用户体系', link: '/zh/modules/user-system' },
      { text: 'API 令牌', link: '/zh/modules/token-system' },
      { text: '渠道管理', link: '/zh/modules/channel-management' },
      { text: '定价与倍率', link: '/zh/modules/pricing' },
      { text: '充值方式', link: '/zh/modules/payment' },
    ],
  },
  {
    text: 'API',
    collapsed: true,
    items: [
      { text: 'API 概览', link: '/zh/api/overview' },
      { text: 'OpenAI 兼容协议', link: '/zh/api/openai-compat' },
      { text: '管理 API 索引', link: '/zh/api/admin-api' },
    ],
  },
  {
    text: '账号对接',
    collapsed: true,
    items: [
      { text: '总览', link: '/zh/integration/overview' },
      { text: 'GitHub OAuth', link: '/zh/integration/oauth-github' },
    ],
  },
  {
    text: '二次开发',
    collapsed: true,
    items: [
      { text: '总览', link: '/zh/development/overview' },
      { text: '新增上游适配器', link: '/zh/development/adapter-guide' },
      { text: '新增支付方式', link: '/zh/development/payment-guide' },
      { text: '贡献指南', link: '/zh/development/contributing' },
    ],
  },
  {
    text: '部署',
    collapsed: true,
    items: [
      { text: 'Docker', link: '/zh/deployment/docker' },
      { text: 'Docker Compose', link: '/zh/deployment/docker-compose' },
      { text: '反向代理', link: '/zh/deployment/reverse-proxy' },
    ],
  },
  {
    text: '计费说明',
    collapsed: true,
    items: [
      { text: '计费如何工作', link: '/zh/billing-guide/how-billing-works' },
      { text: '模型价格表', link: '/zh/billing-guide/model-pricing' },
    ],
  },
  { text: 'FAQ', link: '/zh/faq' },
  { text: '更新日志', link: '/zh/changelog' },
]

// 英文 sidebar:与 zhSidebar 同构,只替换前缀 /zh/ → /en/
const enSidebar: DefaultTheme.Sidebar = [
  {
    text: 'Getting Started',
    collapsed: false,
    items: [
      { text: 'Introduction', link: '/en/guide/introduction' },
      { text: 'Quickstart', link: '/en/guide/quickstart' },
      { text: 'Installation', link: '/en/guide/installation' },
      { text: 'Configuration', link: '/en/guide/configuration' },
      { text: 'Upgrade', link: '/en/guide/upgrade' },
    ],
  },
  {
    text: 'Architecture',
    collapsed: true,
    items: [
      { text: 'Overview', link: '/en/architecture/overview' },
      { text: 'Adapter Layer', link: '/en/architecture/adapter-layer' },
      { text: 'Billing', link: '/en/architecture/billing' },
      { text: 'Rate Limiting', link: '/en/architecture/ratelimit' },
      { text: 'Channel Scheduling', link: '/en/architecture/channel-scheduling' },
    ],
  },
  {
    text: 'Modules',
    collapsed: true,
    items: [
      { text: 'User System', link: '/en/modules/user-system' },
      { text: 'API Tokens', link: '/en/modules/token-system' },
      { text: 'Channel Management', link: '/en/modules/channel-management' },
      { text: 'Pricing & Ratio', link: '/en/modules/pricing' },
      { text: 'Payment', link: '/en/modules/payment' },
    ],
  },
  {
    text: 'API',
    collapsed: true,
    items: [
      { text: 'Overview', link: '/en/api/overview' },
      { text: 'OpenAI Compatibility', link: '/en/api/openai-compat' },
      { text: 'Admin API Index', link: '/en/api/admin-api' },
    ],
  },
  {
    text: 'Identity Integration',
    collapsed: true,
    items: [
      { text: 'Overview', link: '/en/integration/overview' },
      { text: 'GitHub OAuth', link: '/en/integration/oauth-github' },
    ],
  },
  {
    text: 'Development',
    collapsed: true,
    items: [
      { text: 'Overview', link: '/en/development/overview' },
      { text: 'Adding an Adapter', link: '/en/development/adapter-guide' },
      { text: 'Adding a Payment Method', link: '/en/development/payment-guide' },
      { text: 'Contributing', link: '/en/development/contributing' },
    ],
  },
  {
    text: 'Deployment',
    collapsed: true,
    items: [
      { text: 'Docker', link: '/en/deployment/docker' },
      { text: 'Docker Compose', link: '/en/deployment/docker-compose' },
      { text: 'Reverse Proxy', link: '/en/deployment/reverse-proxy' },
    ],
  },
  {
    text: 'Billing Guide',
    collapsed: true,
    items: [
      { text: 'How Billing Works', link: '/en/billing-guide/how-billing-works' },
      { text: 'Model Pricing', link: '/en/billing-guide/model-pricing' },
    ],
  },
  { text: 'FAQ', link: '/en/faq' },
  { text: 'Changelog', link: '/en/changelog' },
]

export default defineConfig({
  title: 'proapi',
  description: '一站式大模型 API 中转网关 / All-in-one LLM API Gateway',
  cleanUrls: true,
  lastUpdated: true,
  themeConfig: {
    logo: '/logo.svg',
    socialLinks: [{ icon: 'github', link: 'https://github.com/ijry/pro-api' }],
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
          { text: '更新日志', link: '/zh/changelog' },
        ],
        sidebar: {
          '/zh/': zhSidebar,
        },
        outline: { level: [2, 3], label: '本页目录' },
        docFooter: { prev: '上一页', next: '下一页' },
        lastUpdatedText: '最后更新',
        editLink: {
          pattern: 'https://github.com/ijry/pro-api/edit/main/docs-site/:path',
          text: '在 GitHub 编辑此页',
        },
      },
    },
    en: {
      label: 'English',
      lang: 'en-US',
      link: '/en/',
      themeConfig: {
        nav: [
          { text: 'Guide', link: '/en/guide/introduction' },
          { text: 'Architecture', link: '/en/architecture/overview' },
          { text: 'API', link: '/en/api/overview' },
          { text: 'Deployment', link: '/en/deployment/docker' },
          { text: 'Development', link: '/en/development/overview' },
          { text: 'Changelog', link: '/en/changelog' },
        ],
        sidebar: {
          '/en/': enSidebar,
        },
        outline: { level: [2, 3], label: 'On this page' },
        docFooter: { prev: 'Previous', next: 'Next' },
        lastUpdatedText: 'Last updated',
        editLink: {
          pattern: 'https://github.com/ijry/pro-api/edit/main/docs-site/:path',
          text: 'Edit this page on GitHub',
        },
      },
    },
  },
})
