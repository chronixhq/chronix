import {defineConfig} from 'vitest/config'
import react from '@vitejs/plugin-react'
import svgr from 'vite-plugin-svgr'

export default defineConfig({
    build: {
        rollupOptions: {
            output: {
                manualChunks(id) {
                    if (!id.includes('node_modules')) return;
                    if (id.includes('@mui/x-data-grid')) return 'mui-data-grid';
                    if (id.includes('@mui/x-date-pickers')) return 'mui-date-pickers';
                    if (id.includes('@mui/icons-material')) return 'mui-icons';
                    if (
                        id.includes('@mui/material') ||
                        id.includes('@mui/system') ||
                        id.includes('@emotion') ||
                        id.includes('@popperjs')
                    ) {
                        return 'mui-core';
                    }
                    if (
                        id.includes('react-markdown') ||
                        id.includes('remark-gfm') ||
                        id.includes('rehype-slug')
                    ) {
                        return 'markdown';
                    }
                    if (id.includes('/node_modules/react/') || id.includes('/node_modules/react-dom/') || id.includes('/node_modules/scheduler/')) return 'react-core';
                    if (id.includes('/node_modules/react-router/')) return 'react-router';
                    if (id.includes('/node_modules/@remix-run/router/')) return 'react-router';
                    if (id.includes('/node_modules/@dsherwin/mui-kit/')) return 'dsherwin-kit';
                    if (id.includes('/node_modules/@dsherwin/react-api-interface/')) return 'dsherwin-runtime';
                    if (id.includes('/node_modules/@dsherwin/react-sse/')) return 'dsherwin-runtime';
                },
            },
        },
    },
    test: {
        environment: 'jsdom',
        globals: true,
        setupFiles: './src/test/setupTests.ts',
        css: true,
        restoreMocks: true,
        server: {
            deps: {
                inline: ['@dsherwin/mui-kit', '@fontsource/roboto'],
            },
        },
    },
    plugins: [
        react(),
        svgr({
            svgrOptions: {
            },
        }),
    ],
    server: {
        host: '0.0.0.0',
    },
    resolve: {dedupe: ['react', 'react-dom']},
})
