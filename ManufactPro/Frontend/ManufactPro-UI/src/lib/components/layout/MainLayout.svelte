<script lang="ts">
  import { cn } from '$lib/utils/cn'
  import Sidebar from './Sidebar.svelte'
  import TopBar from './TopBar.svelte'

  interface Props {
    currentPath?: string
    class?: string
    children: import('svelte').Snippet
  }

  let { currentPath = '/', class: className = '', children }: Props = $props()
</script>

<div class="flex min-h-screen bg-background">
  <Sidebar {currentPath} />

  <div class="flex-1 flex flex-col md:ml-64 min-h-screen">
    <TopBar />
    <main class={cn('flex-1 p-4 md:p-8', className)}>
      {@render children()}
    </main>
  </div>
</div>
