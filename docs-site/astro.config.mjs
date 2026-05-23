// @ts-check
import { defineConfig } from 'astro/config';
import starlight from '@astrojs/starlight';

export default defineConfig({
	site: 'https://loomi.local',
	integrations: [
		starlight({
			title: 'Loomi Docs',
			description: 'Loomi 的产品路线、架构设计、Spec Kit 规格和开发日志。',
			social: [],
			sidebar: [
				{
					label: '开始',
					items: [
						{ label: '文档首页', slug: 'index' },
						{ label: '以后如何开发', slug: 'workflow/how-to-develop' },
						{ label: '开发同步规则', slug: 'workflow/documentation-sync' },
					],
				},
				{
					label: '路线图',
					items: [{ autogenerate: { directory: 'roadmap' } }],
				},
				{
					label: '架构',
					items: [{ autogenerate: { directory: 'architecture' } }],
				},
				{
					label: 'Spec Kit',
					items: [{ autogenerate: { directory: 'spec-kit' } }],
				},
				{
					label: '开发日志',
					items: [{ autogenerate: { directory: 'devlog' } }],
				},
				{
					label: 'API 与运行手册',
					items: [
						{ autogenerate: { directory: 'api' } },
						{ autogenerate: { directory: 'runbooks' } },
					],
				},
				{
					label: '技术决策',
					items: [{ autogenerate: { directory: 'adr' } }],
				},
			],
		}),
	],
});
