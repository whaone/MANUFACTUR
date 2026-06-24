<script lang="ts">
  import { cn } from '$lib/utils/cn'

  interface Props {
    label?: string
    placeholder?: string
    value?: string | number
    type?: string
    error?: string
    hint?: string
    disabled?: boolean
    required?: boolean
    class?: string
    oninput?: (e: Event) => void
    onchange?: (e: Event) => void
  }

  let {
    label = '',
    placeholder = '',
    value = $bindable(''),
    type = 'text',
    error = '',
    hint = '',
    disabled = false,
    required = false,
    class: className = '',
    oninput,
    onchange,
  }: Props = $props()

  const id = `input-${Math.random().toString(36).slice(2, 9)}`

  // Svelte only coerces bind:value to number when type="number" is a static attribute.
  // type is dynamic here, so coerce manually to keep numeric fields as numbers (not strings).
  function handleInput(e: Event) {
    const el = e.target as HTMLInputElement
    if (type === 'number') {
      value = el.value === '' ? '' : Number(el.value)
    } else {
      value = el.value
    }
    oninput?.(e)
  }
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
  <input
    {id}
    {type}
    {placeholder}
    {disabled}
    {required}
    value={value ?? ''}
    oninput={handleInput}
    {onchange}
    class={cn(
      'w-full px-3 py-2 rounded-lg text-sm bg-surface-container-low border transition-all duration-150',
      'placeholder:text-outline focus:outline-none focus:ring-2 focus:ring-primary/30 focus:border-primary',
      'disabled:opacity-50 disabled:cursor-not-allowed',
      error ? 'border-status-critical' : 'border-outline-variant',
    )}
  />
  {#if error}
    <p class="text-xs text-status-critical">{error}</p>
  {/if}
  {#if hint && !error}
    <p class="text-xs text-outline">{hint}</p>
  {/if}
</div>
