import { defineConfig } from 'vitepress'

export default defineConfig({
  base: '/AgnesCode2Api/',
  title: 'AgnesCodeProxy — 协议翻译层',
  description: '将 AgnesCode API 协议翻译为 Anthropic/OpenAI 兼容格式，让 Claude Code 等工具可直接调用',
  lang: 'zh-CN',
  lastUpdated: true,
  ignoreDeadLinks: true,
  markdown: {
    lineNumbers: true,
  },
  themeConfig: {
    search: {
      provider: 'local',
    },
    nav: [
      { text: '文档首页', link: '/' },
      { text: 'GitHub', link: 'https://github.com/vibe-coding-labs/AgnesCode2Api' },
    ],
    sidebar: [
      { text: 'AgnesCodeProxy', items: [
        { text: '首页', link: '/' },
      ]}
    ],
    socialLinks: [
      { icon: 'github', link: 'https://github.com/vibe-coding-labs/AgnesCode2Api' },
    ],
    footer: {
      message: '基于 VitePress 构建',
    },
  },
})