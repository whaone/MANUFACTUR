<script lang="ts">
  import { link } from 'svelte-routing'
  import { Home } from '@lucide/svelte'
</script>

<div class="flex flex-col items-center justify-center py-20 space-y-4">
  <div class="p-4 rounded-full bg-surface-container-high">
    <Home class="w-8 h-8 text-outline" />
  </div>
  <h2 class="text-lg font-semibold text-on-surface">Page Not Found</h2>
  <p class="text-sm text-on-surface-variant">The page you're looking for doesn't exist.</p>
  <a href="/" use:link class="text-sm text-primary hover:underline">Go to Dashboard</a>
</div>
