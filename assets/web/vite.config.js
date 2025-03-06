import fs from 'fs'
import path from 'path'
import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react-swc';
import eslint from 'vite-plugin-eslint';
import { viteStaticCopy } from 'vite-plugin-static-copy';
import { ViteMinifyPlugin } from 'vite-plugin-minify'

export default defineConfig(({ mode }) => {
  return {
    plugins: [
      eslint({
        failOnError: false,
      }),
      react(),
      ViteMinifyPlugin({}),
      viteStaticCopy({
        targets: [
          {
            src: 'src/images/*',
            dest: 'images',
          },
        ],
      }),
      {
        name: 'vite-postbuild',
        closeBundle: () => {
          fs.rename(
            path.join(__dirname, '../static/index.html'),
            path.join(__dirname, '../templates/index.html.tmpl'),
            (err) => {
              if (err) throw err;
              console.log(`'index.html' successfully moved to 'templates' folder`);
            });
        },
      },
    ],
    // See: https://github.com/twbs/bootstrap/issues/40962.
    css: {
      preprocessorOptions: {
        scss: {
          silenceDeprecations: [
            'mixed-decls', 'color-functions', 'global-builtin', 'import',
          ],
        },
      }
    },
    build: {
      outDir: path.resolve(__dirname, '../static'),
      chunkSizeWarningLimit: 2048,
      rollupOptions: {
        input: './index.html',
        output: {
          entryFileNames: 'app.js',
          manualChunks: false,
          inlineDynamicImports: true,
          assetFileNames: (assetInfo) => {
            const name = assetInfo.names?.[0] ?? '';
            if (/\.(woff(2)?|eot|ttf|otf|svg)$/.test(name)) {
              return 'fonts/[name][extname]';
            }
            if (name.endsWith('.css')) {
              return 'styles.css';
            }
            throw new Error(`Unexpected asset: ${name}`);
          },
        },
      },
      watch:
        mode === 'development'
          ? {
              chokidar: {
                usePolling: true,
                interval: 1000,
              },
              exclude: 'node_modules/**',
            }
          : undefined,
    },
  };
});
