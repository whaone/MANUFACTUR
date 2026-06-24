<script lang="ts">
  import { cn } from '$lib/utils/cn'

  interface Props {
    label?: string
    value?: string
    options: Array<{ value: string; label: string }>
    placeholder?: string
    disabled?: boolean
    required?: boolean
    error?: string
    class?: string
    onchange?: (value: string) => void
  }

  let {
    label = '',
    value = $bindable(''),
    options = [],
    placeholder = 'Select...',
    disabled = false,
    required = false,
    error = '',
    class: className = '',
    onchange,
  }: Props = $props()

  const id = `select-${Math.random().toString(36).slice(2, 9)}`
</script>

<div class={cn('flex flex-col gap-1', className)}>
  {#if label}
    <label for={id} class="text-xs font-semibold text-on-surface-variant tracking-wide uppercase">
      {label}
      {#if required}
        <span class="text-error">*</span>
      {/if}
    </label>
  {/if}
  <select
    {id}
    {disabled}
    {required}
    bind:value
    onchange={() => onchange?.(value)}
    class={cn(
      'w-full px-3 py-2 rounded-lg text-sm bg-surface-container-low border transition-all duration-150 appearance-none cursor-pointer',
      'focus:outline-none focus:ring-2 focus:ring-primary/30 focus:border-primary',
      'disabled:opacity-50 disabled:cursor-not-allowed',
      error ? 'border-status-critical' : 'border-outline-variant',
    )}
  >
    <option value="" disabled>{placeholder}</option>
    {#each options as opt}
      <option value={opt.value}>{opt.label}</option>
    {/each}
  </select>
  {#if error}
    <p class="text-xs text-status-critical">{error}</p>
  {/if}
</div>
