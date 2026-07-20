import { defineConfig } from '@vben/vite-config';

export default defineConfig(async () => {
  return {
    application: {},
    vite: {
      server: {
        port: 3001,
        strictPort: true,
        proxy: {
          '/api': {
            changeOrigin: true,
            target: 'http://localhost:8080',
            ws: true,
          },
        },
      },
    },
  };
});
