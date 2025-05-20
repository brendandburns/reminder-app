import { defineConfig } from 'vite';

export default defineConfig({
  root: '.',
  build: {
    outDir: 'dist',
  },
  server: {
    port: 5173,
    proxy: {
      '/families': 'http://localhost:8080',
      '/reminders': 'http://localhost:8080',
    },
  },
});
