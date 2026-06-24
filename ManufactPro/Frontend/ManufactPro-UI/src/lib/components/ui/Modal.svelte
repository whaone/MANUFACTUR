<script lang="ts">
  import { cn } from '$lib/utils/cn'
  import { X } from '@lucide/svelte'

  interface Props {
    open: boolean
    title?: string
    size?: 'sm' | 'md' | 'lg'
    class?: string
    children: import('svelte').Snippet
    footer?: import('svelte').Snippet
    onclose?: () => void
  }

  let {
    open = $bindable(false),
    title = '',
    size = 'md',
    class: className = '',
    children,
    footer,
    onclose,
  }: Props = $props()

  const sizes = {
    sm: 'max-w-md',
    md: 'max-w-2xl',
    lg: 'max-w-4xl',
  }

  function handleBackdropClick(e: MouseEvent) {
    if (e.target === e.currentTarget) {
      open = false
      onclose?.()
    }
  }

  function handleKeydown(e: KeyboardEvent) {
    if (e.key === 'Escape') {
      open = false
      onclose?.()
    }
  }
</script>

{#if open}
  <div
    class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-on-background/40 backdrop-blur-sm"
    onclick={handleBackdropClick}
    onkeydown={handleKeydown}
    role="dialog"
    aria-modal="true"
    tabindex="-1"
  >
    <div class={cn('w-full bg-surface-container-lowest rounded-xl shadow-soft border border-outline-variant/50 overflow-hidden', sizes[size], className)}>
      {#if title}
        <div class="flex items-center justify-between px-4 md:px-6 py-4 border-b border-outline-variant">
          <h2 class="text-lg font-semibold text-on-surface">{title}</h2>
          <button
            onclick={() => { open = false; onclose?.() }}
            class="p-1 rounded-lg hover:bg-surface-container-high transition-colors cursor-pointer"
          >
            <X class="w-5 h-5 text-on-surface-variant" />
          </button>
        </div>
      {/if}
      <div class="px-4 md:px-6 py-4 max-h-[70vh] overflow-y-auto">
        {@render children()}
      </div>
      {#if footer}
        <div class="px-4 md:px-6 py-4 border-t border-outline-variant bg-surface-container-low">
          {@render footer()}
        </div>
      {/if}
    </div>
  </div>
{/if}

