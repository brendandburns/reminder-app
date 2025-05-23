import { defineConfig } from 'vite';

export default defineConfig({
  root: '.',
  build: {
    outDir: 'dist',
    rollupOptions: {
      input: {
        main: 'index.html',
        events: 'events.html',
        member: 'member.html',
        menu: 'components/menu.html',
      },
    },
  },
  server: {
    port: 5173,
    proxy: {
      '/families': 'http://localhost:8080',
      '/reminders': 'http://localhost:8080',
    },
  },
});
