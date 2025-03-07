import fs from 'fs'
import path from 'path'
import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react-swc';
import eslint from 'vite-plugin-eslint';
import { viteStaticCopy } from 'vite-plugin-static-copy';
import { ViteMinifyPlugin } from 'vite-plugin-minify'

export default defineConfig(({ mode }) => {
  const isDevelopment = mode === 'development';

  return {
    plugins: [
      eslint({
        failOnError: false,
      }),
      react(),
      // HTML minification is not enabled in development mode.
      !isDevelopment && ViteMinifyPlugin({}),
      // Skip the build pipeline for assets not directly referenced in the
      // JavaScript implementation (e.g., 'src/fonts/*' is not included here
      // because they are indirectly referenced via the SCSS files, same for
      // 'src/images/*' which are referenced in the HTML entry point, etc.).
      viteStaticCopy({
        targets: [],
      }),
      // Simple workaround to move 'index.html' to the 'templates' folder after
      // the build is completed. The file is a Go template that is hydrated with
      // data by varnishmon before being served to clients.
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
    // Temporary workaround to avoid a ton of warning messages. See:
    // https://github.com/twbs/bootstrap/issues/40962.
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
      // All build output except 'index.html' is placed in the 'static' folder.
      outDir: path.resolve(__dirname, '../static'),
      // Emptying the 'outDir' on build is generally a good idea, but during
      // development, it can be annoying due to errors related to '.fuse_hidden*'
      // files being blocked by the varnishmon agent.
      emptyOutDir: !isDevelopment,
      // We don't really care about the size of the assets in this context.
      chunkSizeWarningLimit: 1024 * 16,
      // A couple of adjustments trying to make the build faster.
      reportCompressedSize: !isDevelopment,
      minify: !isDevelopment,
      // Fine tune to ensure the build output is as expected.
      rollupOptions: {
        input: './index.html',
        output: {
          entryFileNames: 'app.js',
          manualChunks: false,
          inlineDynamicImports: true,
          assetFileNames: (assetInfo) => {
            const originalFileName = assetInfo.originalFileNames?.[0] ?? '';
            const name = assetInfo.names?.[0] ?? '';
            if (originalFileName.startsWith('src/images/') &&
                /\.(png|svg|jpg|jpeg|gif|ico)$/i.test(name)) {
              return 'images/[name][extname]';
            }
            if ((originalFileName.startsWith('src/fonts/') || originalFileName.includes('/@fortawesome/')) &&
                /\.(woff(2)?|eot|ttf|otf|svg)$/i.test(name)) {
              return 'fonts/[name][extname]';
            }
            if (name.endsWith('.css')) {
              return 'styles.css';
            }
            throw new Error(`Unexpected asset: ${name} (${originalFileName})`);
          },
        },
      },
      // Vite is not used to serve the application. Therefore, during development,
      // the Rollup watcher needs to be used.
      watch:
        isDevelopment
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
