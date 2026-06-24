<script lang="ts">
  import { cn } from '$lib/utils/cn'
  import { link } from 'svelte-routing'
  import { auth } from '$lib/stores/auth'
  import { isOffline } from '$lib/stores/offline'
  import { toggleSidebar } from '$lib/stores/sidebar'
  import { Menu, Bell, Cloud, CloudOff, User, LogOut, Settings, AlertTriangle, CheckCheck } from '@lucide/svelte'

  let isOfflineState = $derived($isOffline)
  let showUserMenu = $state(false)
  let showNotifMenu = $state(false)

  // Mock low stock notifications
  let notifications = $state([
    { id: 1, title: 'Low Stock Alert', message: 'Rib Polos Hitam is running low (8/30 meter)', time: '10m ago', read: false, type: 'critical' },
    { id: 2, title: 'Low Stock Alert', message: 'Cat Sablon Putih is critically low (2/10 kg)', time: '1h ago', read: false, type: 'critical' },
    { id: 3, title: 'Production Completed', message: 'PO-2024-0042 has finished production', time: '2h ago', read: true, type: 'success' }
  ])

  let unreadCount = $derived(notifications.filter(n => !n.read).length)

  function toggleUserMenu() {
    showUserMenu = !showUserMenu
    if (showUserMenu) showNotifMenu = false
  }

  function toggleNotifMenu() {
    showNotifMenu = !showNotifMenu
    if (showNotifMenu) showUserMenu = false
  }

  function markAllRead() {
    notifications = notifications.map(n => ({ ...n, read: true }))
  }

  function handleClickOutside(e: MouseEvent) {
    const target = e.target as HTMLElement
    if (!target.closest('[data-user-menu]')) showUserMenu = false
    if (!target.closest('[data-notif-menu]')) showNotifMenu = false
  }
</script>

<svelte:window onclick={handleClickOutside} />

<header class="sticky top-0 z-30 bg-surface-container-lowest/80 backdrop-blur-md border-b border-outline-variant/30">
  <div class="flex items-center justify-between h-16 px-4 md:px-8">
    <div class="flex items-center gap-3">
      <button onclick={toggleSidebar} class="md:hidden p-2 rounded-lg hover:bg-surface-container-high transition-colors cursor-pointer">
        <Menu class="w-5 h-5 text-on-surface" />
      </button>
      <div class="md:hidden flex items-center gap-2">
        <div class="w-7 h-7 rounded-lg primary-gradient flex items-center justify-center">
          <span class="text-xs font-bold text-on-primary">M</span>
        </div>
        <span class="font-bold text-on-surface tracking-tight">ManufactPro</span>
      </div>
    </div>

    <div class="flex items-center gap-2">
      <div class={cn(
        'flex items-center gap-1.5 px-3 py-1.5 rounded-full text-xs font-medium',
        isOfflineState
          ? 'bg-status-warning/10 text-status-warning'
          : 'bg-status-success/10 text-status-success',
      )}>
        {#if isOfflineState}
          <CloudOff class="w-3.5 h-3.5" />
          <span>Offline</span>
        {:else}
          <Cloud class="w-3.5 h-3.5" />
          <span>Online</span>
        {/if}
      </div>

      <div class="relative" data-notif-menu>
        <button
          class="p-2 rounded-lg hover:bg-surface-container-high transition-colors relative cursor-pointer"
          onclick={toggleNotifMenu}
        >
          <Bell class="w-5 h-5 text-on-surface-variant" />
          {#if unreadCount > 0}
            <span class="absolute top-1 right-1 w-2.5 h-2.5 border-2 border-surface-container-lowest bg-status-critical rounded-full"></span>
          {/if}
        </button>

        {#if showNotifMenu}
          <div class="absolute right-0 top-full mt-2 w-80 sm:w-96 bg-surface-container-lowest rounded-xl shadow-soft border border-outline-variant/50 overflow-hidden animate-[zoom-in_0.15s_ease-out]">
            <div class="flex items-center justify-between px-4 py-3 border-b border-outline-variant/30 bg-surface-container-low/50">
              <p class="text-sm font-semibold text-on-surface">Notifications</p>
              {#if unreadCount > 0}
                <button onclick={markAllRead} class="text-xs font-medium text-primary hover:text-primary-container flex items-center gap-1 cursor-pointer">
                  <CheckCheck class="w-3.5 h-3.5" /> Mark all read
                </button>
              {/if}
            </div>
            
            <div class="max-h-80 overflow-y-auto divide-y divide-outline-variant/20">
              {#if notifications.length === 0}
                <div class="p-8 text-center text-outline">
                  <Bell class="w-8 h-8 mx-auto mb-2 opacity-50" />
                  <p class="text-sm">No notifications</p>
                </div>
              {:else}
                {#each notifications as notif}
                  <div class={cn('flex gap-3 p-4 hover:bg-surface-container-low/50 transition-colors cursor-pointer', !notif.read ? 'bg-primary/5' : '')}>
                    <div class={cn('mt-0.5 w-8 h-8 rounded-full flex items-center justify-center shrink-0', notif.type === 'critical' ? 'bg-status-critical/10 text-status-critical' : 'bg-status-success/10 text-status-success')}>
                      <AlertTriangle class="w-4 h-4" />
                    </div>
                    <div>
                      <p class={cn('text-sm font-medium text-on-surface', !notif.read ? 'font-semibold' : '')}>{notif.title}</p>
                      <p class="text-xs text-on-surface-variant mt-0.5 line-clamp-2">{notif.message}</p>
                      <p class="text-xs text-outline mt-1.5">{notif.time}</p>
                    </div>
                  </div>
                {/each}
              {/if}
            </div>
            
            <div class="p-2 border-t border-outline-variant/30 bg-surface-container-lowest text-center">
              <a href="/" class="text-xs font-medium text-primary hover:underline p-2 block w-full" onclick={(e) => e.preventDefault()}>View all notifications</a>
            </div>
          </div>
        {/if}
      </div>

      <div class="relative" data-user-menu>
        <button
          onclick={toggleUserMenu}
          class="p-2 rounded-lg hover:bg-surface-container-high transition-colors cursor-pointer"
        >
          <User class="w-5 h-5 text-on-surface-variant" />
        </button>

        {#if showUserMenu}
          <div class="absolute right-0 top-full mt-2 w-48 bg-surface-container-lowest rounded-xl shadow-soft border border-outline-variant/50 overflow-hidden animate-[zoom-in_0.15s_ease-out]">
            <div class="px-4 py-3 border-b border-outline-variant/30">
              <p class="text-sm font-medium text-on-surface">Admin</p>
              <p class="text-xs text-outline">admin@manufactpro.id</p>
            </div>
            <div class="py-1">
              <a href="/settings" use:link onclick={() => { showUserMenu = false }} class="flex items-center gap-3 px-4 py-2.5 text-sm text-on-surface-variant hover:bg-surface-container-high hover:text-on-surface transition-colors">
                <Settings class="w-4 h-4" />
                Settings
              </a>
              <button onclick={() => auth.logout()} class="w-full flex items-center gap-3 px-4 py-2.5 text-sm text-status-critical hover:bg-status-critical/5 transition-colors cursor-pointer">
                <LogOut class="w-4 h-4" />
                Logout
              </button>
            </div>
          </div>
        {/if}
      </div>
    </div>
  </div>
</header>
