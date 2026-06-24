import { defineConfig } from 'vitest/config'
import { svelte } from '@sveltejs/vite-plugin-svelte'
import { svelteTesting } from '@testing-library/svelte/vite'
import path from 'path'

// Dedicated test config: Svelte plugin + jsdom, without the PWA/Tailwind
// plugins from vite.config.ts (not needed for unit/component tests).
export default defineConfig({
  plugins: [svelte(), svelteTesting()],
  resolve: {
    alias: { $lib: path.resolve('./src/lib') },
  },
  test: {
    environment: 'jsdom',
    setupFiles: ['./src/test-setup.ts'],
  },
})
