// @ts-check
import {defineConfig} from 'astro/config';
import starlight from '@astrojs/starlight';
import sitemap from "@astrojs/sitemap";
import starlightThemeNova from 'starlight-theme-nova'

// https://astro.build/config
export default defineConfig({
  site: 'https://miyamo2.github.io',
  base: '/qilin',
  integrations: [
    starlight({
      plugins: [starlightThemeNova()],
      title: 'Qilin MCP Framework',
      favicon: '/favicon.ico',
      logo: {
        src: "./src/assets/logo.png",
      },
      social: [{ icon: 'github', label: 'GitHub', href: 'https://github.com/miyamo2/qilin' }],
      expressiveCode: {
        themes: ['github-dark', 'github-light'],
      },
      lastUpdated: true,
      sidebar: [
        {
          label: "👋 Introduction",
          slug: "introduction",
        },
        {
          label: 'Guide',
          items: [
            { label: '⏱ Quick Start', slug: 'guides/quickstart' },
            {
              label: '🏗️ Building MCP Servers',
              items: [
                {
                  label: '🛠️ Tools',
                  items: [
                    { label: 'Calling Tools', slug: 'guides/mcp/tools/calling' },
                  ],
                },
                {
                  label: '📚 Resources',
                  items: [
                    { label: 'Listing Resources', slug: 'guides/mcp/resources/listing' },
                    { label: 'Reading Resources', slug: 'guides/mcp/resources/reading' },
                    { label: 'Resources Subscriptions', slug: 'guides/mcp/resources/subscribe' },
                    { label: 'Resources List Changed Notification', slug: 'guides/mcp/resources/list_changed' },
                  ],
                },
                {
                  label: '🤝 Sessions',
                  items: [
                    { label: 'Session Management', slug: 'guides/mcp/sessions/manage' },
                  ],
                },
              ]
            },
            {
              label: '🔄 Transport',
              items: [
                { label: 'Stdio', slug: 'guides/transport/stdio' },
                { label: 'Streamable HTTP', slug: 'guides/transport/streamable_http' },
              ],
            }
          ],
        },
      ],
    }),
    sitemap()
  ],
});
