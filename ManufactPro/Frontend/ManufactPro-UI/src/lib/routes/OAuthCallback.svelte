<script lang="ts">
  import { onMount } from 'svelte'
  import { auth } from '$lib/stores/auth'
  import { Loader2 } from '@lucide/svelte'

  let error = $state('')

  onMount(() => {
    const hash = window.location.hash.slice(1)
    const params = new URLSearchParams(hash)
    const token = params.get('token')
    const refresh = params.get('refresh')

    if (!token) {
      error = 'OAuth login gagal — token tidak ditemukan.'
      return
    }

    auth.setTokens(token, refresh ?? '')
    window.location.replace('/')
  })
</script>

<div class="flex min-h-dvh items-center justify-center bg-gradient-to-br from-primary/5 via-background to-surface-container-low">
  {#if error}
    <div class="text-center space-y-3">
      <p class="text-status-critical text-sm">{error}</p>
      <a href="/login" class="text-primary text-sm underline">Kembali ke login</a>
    </div>
  {:else}
    <div class="flex flex-col items-center gap-3 text-on-surface-variant">
      <Loader2 class="w-8 h-8 animate-spin text-primary" />
      <p class="text-sm">Menyelesaikan login...</p>
    </div>
  {/if}
</div>
