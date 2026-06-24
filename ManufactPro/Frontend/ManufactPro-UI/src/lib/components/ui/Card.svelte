<script lang="ts">
  import { cn } from '$lib/utils/cn'

  interface Props {
    variant?: 'default' | 'glass' | 'gradient'
    padding?: 'none' | 'sm' | 'md' | 'lg'
    class?: string
    children: import('svelte').Snippet
  }

  let {
    variant = 'default',
    padding = 'md',
    class: className = '',
    children,
  }: Props = $props()

  const variants = {
    default: 'bg-surface-container-lowest border border-outline-variant shadow-card',
    glass: 'glass-card',
    gradient: 'primary-gradient text-on-primary',
  }

  const paddings = {
    none: '',
    sm: 'p-4',
    md: 'p-6',
    lg: 'p-8',
  }
</script>

<div class={cn('rounded-xl', variants[variant], paddings[padding], className)}>
  {@render children()}
</div>
