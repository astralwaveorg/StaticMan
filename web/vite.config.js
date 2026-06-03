import { defineConfig } from 'vite';
import vue from '@vitejs/plugin-vue';
export default defineConfig({
    plugins: [vue()],
    server: {
        host: '0.0.0.0',
        proxy: {
            '/api': 'http://localhost:8080',
            '/d': 'http://localhost:8080',
            '/raw': 'http://localhost:8080',
        },
    },
    build: {
        outDir: '../internal/web/dist',
        emptyOutDir: true,
    },
});
