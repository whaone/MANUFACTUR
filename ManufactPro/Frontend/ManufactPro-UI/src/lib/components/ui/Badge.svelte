<script lang="ts">
  import { cn } from '$lib/utils/cn'
  import type { BadgeVariant } from '$lib/types'

  interface Props {
    variant?: BadgeVariant
    size?: 'sm' | 'md'
    class?: string
    children: import('svelte').Snippet
  }

  let {
    variant = 'neutral',
    size = 'sm',
    class: className = '',
    children,
  }: Props = $props()

  const variants = {
    success: 'bg-status-success/15 text-status-success border-status-success/20',
    warning: 'bg-status-warning/15 text-status-warning border-status-warning/20',
    critical: 'bg-status-critical/15 text-status-critical border-status-critical/20',
    info: 'bg-status-info/15 text-status-info border-status-info/20',
    neutral: 'bg-surface-container-high text-on-surface-variant border-outline-variant',
    primary: 'bg-primary/15 text-primary border-primary/20',
  }

  const sizes = {
    sm: 'px-2 py-0.5 text-xs',
    md: 'px-3 py-1 text-sm',
  }
</script>

<span class={cn('inline-flex items-center font-semibold rounded-full border tracking-wide', variants[variant], sizes[size], className)}>
  {@render children()}
</span>
