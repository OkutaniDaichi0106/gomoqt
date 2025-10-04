import { defineConfig } from 'tsup';

export default defineConfig({
  entry: ['src/index.ts', 'src/internal/index.ts', 'src/io/index.ts'],
  format: ['esm'],
  dts: true,
  sourcemap: false, // ソースマップを無効化
  splitting: false, // コード分割を無効化
  clean: true,
  outDir: 'dist',
  external: [], // golikejsもバンドルに含める
  treeshake: true,
  minify: false,
  noExternal: ['golikejs'], // golikejsを明示的にバンドル
});
