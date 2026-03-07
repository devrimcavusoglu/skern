import { defineConfig } from 'vitepress'

export default defineConfig({
  title: 'Skern',
  description: 'System-wide skill registry for AI agents',

  base: '/',
  cleanUrls: true,
  lastUpdated: true,

  head: [
    ['link', { rel: 'icon', href: '/logo.png' }],
    ['meta', { property: 'og:type', content: 'website' }],
    ['meta', { property: 'og:title', content: 'Skern' }],
    ['meta', { property: 'og:description', content: 'System-wide skill registry for AI agents' }],
    ['meta', { property: 'og:url', content: 'https://skern.dev' }],
    ['meta', { property: 'og:image', content: 'https://skern.dev/logo.png' }],
    ['meta', { name: 'twitter:card', content: 'summary' }],
    ['meta', { name: 'twitter:title', content: 'Skern' }],
    ['meta', { name: 'twitter:description', content: 'System-wide skill registry for AI agents' }],
    ['meta', { name: 'twitter:image', content: 'https://skern.dev/logo.png' }],
  ],

  themeConfig: {
    logo: '/logo.png',

    nav: [
      { text: 'Guide', link: '/guide/' },
      { text: 'Reference', link: '/reference/' },
      { text: 'Concepts', link: '/concepts/' },
      { text: 'Platforms', link: '/platforms/' },
    ],

    sidebar: {
      '/guide/': [
        {
          text: 'Guide',
          items: [
            { text: 'Getting Started', link: '/guide/' },
            { text: 'Installation', link: '/guide/installation' },
            { text: 'Quick Start', link: '/guide/quick-start' },
            { text: 'Agent Setup', link: '/guide/agent-setup' },
          ],
        },
      ],
      '/reference/': [
        {
          text: 'Reference',
          items: [
            { text: 'Overview', link: '/reference/' },
            { text: 'Commands', link: '/reference/commands' },
            { text: 'Validation', link: '/reference/validation' },
            { text: 'Overlap Detection', link: '/reference/overlap-detection' },
          ],
        },
      ],
      '/concepts/': [
        {
          text: 'Concepts',
          items: [
            { text: 'Architecture', link: '/concepts/' },
            { text: 'Skill Format', link: '/concepts/skill-format' },
            { text: 'Registry', link: '/concepts/registry' },
            { text: 'Platform Adapters', link: '/concepts/platform-adapters' },
          ],
        },
      ],
      '/platforms/': [
        {
          text: 'Platforms',
          items: [
            { text: 'Overview', link: '/platforms/' },
            { text: 'Claude Code', link: '/platforms/claude-code' },
            { text: 'Codex CLI', link: '/platforms/codex-cli' },
            { text: 'OpenCode', link: '/platforms/opencode' },
          ],
        },
      ],
      '/contributing/': [
        {
          text: 'Contributing',
          items: [
            { text: 'Contributing Guide', link: '/contributing/' },
            { text: 'Development', link: '/contributing/development' },
            { text: 'Manual Testing', link: '/contributing/manual-testing' },
          ],
        },
      ],
    },

    socialLinks: [
      { icon: 'github', link: 'https://github.com/devrimcavusoglu/skern' },
    ],

    editLink: {
      pattern: 'https://github.com/devrimcavusoglu/skern/edit/main/docs/:path',
      text: 'Edit this page on GitHub',
    },

    search: {
      provider: 'local',
    },

    footer: {
      message: 'Released under the Apache 2.0 License.',
      copyright: 'Copyright © 2026-present Devrim Cavusoglu',
    },
  },
})
