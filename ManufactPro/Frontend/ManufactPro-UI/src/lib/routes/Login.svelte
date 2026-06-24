<script lang="ts">
  import { auth } from '$lib/stores/auth'
  import { Eye, EyeOff, Loader2 } from '@lucide/svelte'

  type FieldErrors = { email: string; password: string }

  let email = $state('')
  let password = $state('')
  let remember = $state(false)
  let showPassword = $state(false)
  let loading = $state(false)
  let generalError = $state('')
  let touched = $state<FieldErrors>({ email: '', password: '' })

  let isValid = $derived(email.trim().length > 0 && password.length > 0 && !touched.email && !touched.password)

  const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/

  function validateEmail(v: string): string {
    if (!v.trim()) return ''
    if (!emailRegex.test(v.trim())) return 'Enter a valid email address'
    return ''
  }

  function validatePassword(v: string): string {
    if (!v) return ''
    if (v.length < 6) return 'Must be at least 6 characters'
    return ''
  }

  function onEmailInput(v: string) {
    email = v
    touched = { ...touched, email: email.trim() ? validateEmail(v) : '' }
    generalError = ''
  }

  function onPasswordInput(v: string) {
    password = v
    touched = { ...touched, password: password ? validatePassword(v) : '' }
    generalError = ''
  }

  async function handleSubmit(e: Event) {
    e.preventDefault()
    const eErr = validateEmail(email)
    const pErr = validatePassword(password)
    touched = { email: eErr, password: pErr }
    if (eErr || pErr) return

    loading = true
    generalError = ''

    if (remember) {
      try { localStorage.setItem('manufactpro_remember', email) } catch {}
    }

    try {
      await auth.login(email.trim(), password)
    } catch (err) {
      generalError = err instanceof Error ? err.message : 'Invalid email or password. Please try again.'
    } finally {
      loading = false
    }
  }

  const API_URL = import.meta.env.VITE_API_URL ?? 'http://localhost:8080'

  function handleSocialLogin(provider: 'google' | 'github') {
    window.location.href = `${API_URL}/api/auth/oauth/${provider}`
  }
</script>

<svelte:head>
  <title>Sign In — ManufactPro</title>
</svelte:head>

<!-- Background with subtle gradient -->
<div class="flex min-h-dvh items-center justify-center bg-gradient-to-br from-primary/5 via-background to-surface-container-low p-4">
  <div class="w-full max-w-[440px]">
    <!-- Logo & heading -->
    <div class="mb-8 flex flex-col items-center">
      <div class="mb-4 flex h-16 w-16 items-center justify-center rounded-2xl primary-gradient shadow-lg shadow-primary/20">
        <span class="text-3xl font-bold text-on-primary">M</span>
      </div>
      <h1 class="text-2xl font-bold tracking-tight text-on-surface">Sign in to your account</h1>
      <p class="mt-1 text-sm text-on-surface-variant">Enter your credentials to continue</p>
    </div>

    <!-- Card -->
    <form
      onsubmit={handleSubmit}
      novalidate
      class="space-y-5 rounded-2xl border border-outline-variant/40 bg-surface-container-lowest/80 p-6 sm:p-8 shadow-xl shadow-primary/5 backdrop-blur-sm"
    >
      <!-- General error banner -->
      {#if generalError}
        <div role="alert" class="flex items-center gap-2 rounded-lg border border-status-critical/20 bg-status-critical/10 px-4 py-3 text-sm text-status-critical">
          <svg class="h-4 w-4 shrink-0" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><line x1="12" y1="8" x2="12" y2="12"/><line x1="12" y1="16" x2="12.01" y2="16"/></svg>
          <span>{generalError}</span>
        </div>
      {/if}

      <!-- Email -->
      <div class="space-y-1.5">
        <label for="login-email" class="text-sm font-medium text-on-surface">Email address</label>
        <div class="relative">
          <svg class="absolute left-3.5 top-1/2 h-4 w-4 -translate-y-1/2 text-outline" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="2" y="4" width="20" height="16" rx="2"/><path d="m22 7-8.97 5.7a1.94 1.94 0 0 1-2.06 0L2 7"/></svg>
          <input
            id="login-email"
            type="email"
            autocomplete="email"
            inputmode="email"
            placeholder="you@company.com"
            aria-describedby={touched.email ? 'email-error' : undefined}
            aria-invalid={!!touched.email}
            disabled={loading}
            bind:value={email}
            oninput={(e) => onEmailInput(e.currentTarget.value)}
            class="w-full rounded-xl border bg-surface-container-lowest py-3 pl-11 pr-4 text-sm placeholder:text-outline transition-all duration-200 focus:outline-none focus:ring-2 focus:ring-primary/40 focus:border-transparent disabled:cursor-not-allowed disabled:opacity-50 {touched.email ? 'border-status-critical' : 'border-outline-variant'}"
          />
          {#if touched.email}
            <span class="pointer-events-none absolute right-3.5 top-1/2 -translate-y-1/2 text-status-critical">
              <svg class="h-4 w-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><line x1="15" y1="9" x2="9" y2="15"/><line x1="9" y1="9" x2="15" y2="15"/></svg>
            </span>
          {/if}
        </div>
        {#if touched.email}
          <p id="email-error" class="text-xs text-status-critical" role="alert">{touched.email}</p>
        {/if}
      </div>

      <!-- Password -->
      <div class="space-y-1.5">
        <label for="login-password" class="text-sm font-medium text-on-surface">Password</label>
        <div class="relative">
          <svg class="absolute left-3.5 top-1/2 h-4 w-4 -translate-y-1/2 text-outline" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="3" y="11" width="18" height="11" rx="2" ry="2"/><path d="M7 11V7a5 5 0 0 1 10 0v4"/></svg>
          <input
            id="login-password"
            type={showPassword ? 'text' : 'password'}
            autocomplete="current-password"
            placeholder="Enter password"
            aria-describedby={touched.password ? 'password-error' : undefined}
            aria-invalid={!!touched.password}
            disabled={loading}
            bind:value={password}
            oninput={(e) => onPasswordInput(e.currentTarget.value)}
            class="w-full rounded-xl border bg-surface-container-lowest py-3 pl-11 pr-11 text-sm placeholder:text-outline transition-all duration-200 focus:outline-none focus:ring-2 focus:ring-primary/40 focus:border-transparent disabled:cursor-not-allowed disabled:opacity-50 {touched.password ? 'border-status-critical' : 'border-outline-variant'}"
          />
          <button
            type="button"
            onclick={() => { showPassword = !showPassword }}
            aria-label={showPassword ? 'Hide password' : 'Show password'}
            disabled={loading}
            tabindex="-1"
            class="absolute right-2.5 top-1/2 -translate-y-1/2 rounded-lg p-1.5 text-outline hover:text-on-surface hover:bg-surface-container-high transition-colors cursor-pointer disabled:opacity-50"
          >
            {#if showPassword}
              <EyeOff class="h-4 w-4" />
            {:else}
              <Eye class="h-4 w-4" />
            {/if}
          </button>
        </div>
        {#if touched.password}
          <p id="password-error" class="text-xs text-status-critical" role="alert">{touched.password}</p>
        {/if}
      </div>

      <!-- Remember me + Forgot password -->
      <div class="flex items-center justify-between">
        <label class="flex cursor-pointer items-center gap-2 text-sm text-on-surface-variant">
          <input
            type="checkbox"
            bind:checked={remember}
            disabled={loading}
            class="h-4 w-4 rounded border-outline-variant text-primary focus:ring-primary/40 disabled:opacity-50"
          />
          Remember me
        </label>
        <a href="/" class="text-sm font-medium text-primary hover:text-primary-container transition-colors focus:outline-none focus:underline" onclick={(e) => e.preventDefault()}>
          Forgot password?
        </a>
      </div>

      <!-- Submit -->
      <button
        type="submit"
        disabled={!isValid || loading}
        class="flex w-full items-center justify-center gap-2 rounded-xl py-3 text-sm font-semibold primary-gradient text-on-primary shadow-md shadow-primary/20 transition-all hover:shadow-lg hover:shadow-primary/30 hover:opacity-95 active:scale-[0.98] focus:outline-none focus:ring-2 focus:ring-primary/50 focus:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50 disabled:shadow-none"
      >
        {#if loading}
          <Loader2 class="h-4 w-4 animate-spin" />
          Signing in...
        {:else}
          <svg class="h-4 w-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M15 3h4a2 2 0 0 1 2 2v14a2 2 0 0 1-2 2h-4"/><polyline points="10 17 15 12 10 7"/><line x1="15" y1="12" x2="3" y2="12"/></svg>
          Sign in
        {/if}
      </button>

      <!-- Social divider -->
      <div class="relative">
        <div class="absolute inset-0 flex items-center"><span class="w-full border-t border-outline-variant/40"></span></div>
        <div class="relative flex justify-center"><span class="bg-surface-container-lowest/80 px-4 text-xs text-on-surface-variant">or continue with</span></div>
      </div>

      <!-- Social buttons (full-width stacked) -->
      <div class="flex flex-col gap-3">
        <button
          type="button"
          onclick={() => handleSocialLogin('google')}
          disabled={loading}
          class="flex w-full items-center justify-center gap-3 rounded-xl border border-outline-variant/60 bg-surface-container-lowest py-3 text-sm font-medium text-on-surface transition-all hover:bg-surface-container-low hover:border-outline-variant hover:shadow-sm cursor-pointer disabled:opacity-50 focus:outline-none focus:ring-2 focus:ring-primary/30"
        >
          <span class="flex h-5 w-5 items-center justify-center rounded-full overflow-hidden">
            <svg class="h-5 w-5" viewBox="0 0 24 24"><path fill="#4285F4" d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92a5.06 5.06 0 0 1-2.2 3.32v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.1z"/><path fill="#34A853" d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z"/><path fill="#FBBC05" d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l2.85-2.22.81-.62z"/><path fill="#EA4335" d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z"/></svg>
          </span>
          Continue with Google
        </button>
        <button
          type="button"
          disabled={loading}
          class="flex w-full items-center justify-center gap-3 rounded-xl border border-outline-variant/60 bg-surface-container-lowest py-3 text-sm font-medium text-on-surface transition-all hover:bg-surface-container-low hover:border-outline-variant hover:shadow-sm cursor-pointer disabled:opacity-50 focus:outline-none focus:ring-2 focus:ring-primary/30"
        >
          <span class="flex h-5 w-5 items-center justify-center">
            <svg class="h-5 w-5" viewBox="0 0 24 24" fill="currentColor"><path d="M12 0C5.37 0 0 5.37 0 12c0 5.31 3.435 9.795 8.205 11.385.6.105.825-.255.825-.57 0-.285-.015-1.23-.015-2.235-3.015.555-3.795-.735-4.035-1.41-.135-.345-.72-1.41-1.23-1.695-.42-.225-1.02-.78-.015-.795.945-.015 1.62.87 1.845 1.23 1.08 1.815 2.805 1.305 3.495.99.105-.78.42-1.305.765-1.605-2.67-.3-5.46-1.335-5.46-5.925 0-1.305.465-2.385 1.23-3.225-.12-.3-.54-1.53.12-3.18 0 0 1.005-.315 3.3 1.23.96-.27 1.98-.405 3-.405s2.04.135 3 .405c2.295-1.56 3.3-1.23 3.3-1.23.66 1.65.24 2.88.12 3.18.765.84 1.23 1.905 1.23 3.225 0 4.605-2.805 5.625-5.475 5.925.435.375.81 1.095.81 2.22 0 1.605-.015 2.895-.015 3.3 0 .315.225.69.825.57A12.02 12.02 0 0 0 24 12c0-6.63-5.37-12-12-12z"/></svg>
          </span>
          Continue with GitHub
        </button>
      </div>

      <!-- Sign up link -->
      <p class="text-center text-sm text-on-surface-variant pt-1">
        Don't have an account?{' '}
        <a href="/" class="font-semibold text-primary hover:text-primary-container transition-colors focus:outline-none focus:underline" onclick={(e) => e.preventDefault()}>
          Sign up
        </a>
      </p>
    </form>
  </div>
</div>
