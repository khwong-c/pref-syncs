import { resolve } from 'node:path';
import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';
import dts from 'vite-plugin-dts';

export default defineConfig({
  plugins: [
    react(),
    dts({
      tsconfigPath: './tsconfig.json',
      include: ['src'],
    }),
    {
      name: 'example-root-redirect',
      configureServer(server) {
        server.middlewares.use((req, res, next) => {
          if (req.url === '/' || req.url === '/index.html') {
            req.url = '/example/minimal/index.html';
          }
          next();
        });
      },
    },
  ],
  build: {
    lib: {
      entry: resolve(import.meta.dirname, 'src/index.ts'),
      name: 'PrefSyncReact',
      fileName: (format) => `index.${format === 'es' ? 'js' : 'cjs'}`,
      formats: ['es', 'cjs'],
    },
    rollupOptions: {
      external: [
        'react',
        'react-dom',
        'react/jsx-runtime',
        'pref-sync',
        /^node:.*/,
      ],
      output: {
        globals: {
          react: 'React',
          'react-dom': 'ReactDOM',
          'pref-sync': 'PrefSync',
        },
      },
    },
    sourcemap: true,
    emptyOutDir: true,
  },
  server: {
    open: false,
    proxy: {
      '/app': 'http://127.0.0.1:7086',
      '/user': 'http://127.0.0.1:7086',
      '/pref': 'http://127.0.0.1:7086',
      '/idp': 'http://127.0.0.1:7086',
    },
  },
});
