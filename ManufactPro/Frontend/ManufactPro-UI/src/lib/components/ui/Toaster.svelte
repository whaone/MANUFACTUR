<script lang="ts">
  import { toast } from '$lib/stores/toast'
  import { CheckCircle2, AlertTriangle, XCircle, Info, X } from '@lucide/svelte'

  const ICON = {
    success: CheckCircle2,
    warning: AlertTriangle,
    error: XCircle,
    info: Info,
  }

  const STYLE = {
    success: 'border-status-success/40 bg-status-success/10 text-status-success',
    warning: 'border-status-warning/40 bg-status-warning/10 text-status-warning',
    error: 'border-status-critical/40 bg-status-critical/10 text-status-critical',
    info: 'border-status-info/40 bg-status-info/10 text-status-info',
  }
</script>

<div class="fixed top-4 right-4 z-[100] flex flex-col gap-3 w-[min(92vw,380px)]">
  {#each $toast as t (t.id)}
    {@const Icon = ICON[t.kind]}
    <div
      class="flex items-start gap-3 p-4 rounded-xl border shadow-soft bg-surface-container-lowest animate-[zoom-in_0.15s_ease-out] {STYLE[t.kind]}"
      role="alert"
    >
      <Icon class="w-5 h-5 shrink-0 mt-0.5" />
      <div class="flex-1 min-w-0">
        <p class="text-sm font-semibold text-on-surface">{t.title}</p>
        {#if t.lines.length}
          <ul class="mt-1 space-y-0.5">
            {#each t.lines as line}
              <li class="text-xs text-on-surface-variant break-words">{line}</li>
            {/each}
          </ul>
        {/if}
      </div>
      <button
        onclick={() => toast.dismiss(t.id)}
        class="shrink-0 text-on-surface-variant hover:text-on-surface transition-colors cursor-pointer"
        aria-label="Dismiss"
      >
        <X class="w-4 h-4" />
      </button>
    </div>
  {/each}
</div>
