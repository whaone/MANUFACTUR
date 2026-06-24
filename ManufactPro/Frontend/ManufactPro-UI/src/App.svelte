<script lang="ts">
  import { Router, Route } from 'svelte-routing'
  import { auth, isAuthenticated as authStore } from '$lib/stores/auth'
  import MainLayout from '$lib/components/layout/MainLayout.svelte'
  import Dashboard from '$lib/routes/Dashboard.svelte'
  import Materials from '$lib/routes/Materials.svelte'
  import Products from '$lib/routes/Products.svelte'
  import Bom from '$lib/routes/Bom.svelte'
  import Production from '$lib/routes/Production.svelte'
  import Procurement from '$lib/routes/Procurement.svelte'
  import Inventory from '$lib/routes/Inventory.svelte'
  import Reports from '$lib/routes/Reports.svelte'
  import Settings from '$lib/routes/Settings.svelte'
  import Planning from '$lib/routes/Planning.svelte'
  import NotFound from '$lib/routes/NotFound.svelte'
  import Login from '$lib/routes/Login.svelte'
  import Users from '$lib/routes/Users.svelte'
  import Warehouses from '$lib/routes/Warehouses.svelte'
  import Suppliers from '$lib/routes/Suppliers.svelte'
  import OAuthCallback from '$lib/routes/OAuthCallback.svelte'
  import { onMount } from 'svelte'
  import { initSyncListener, flushSyncQueue } from '$lib/sync/flush'
  import Toaster from '$lib/components/ui/Toaster.svelte'

  let isAuthenticated = $state(false)
  authStore.subscribe((v) => { isAuthenticated = v })

  let currentPath = $state(window.location.pathname)

  onMount(() => {
    initSyncListener()
    if (navigator.onLine) flushSyncQueue().catch((e) => console.error('sync flush failed:', e))

    // svelte-routing navigates via history.pushState, which does NOT emit popstate.
    // Patch pushState to broadcast a 'locationchange' event so the sidebar highlight tracks navigation.
    const origPush = history.pushState.bind(history)
    history.pushState = function (...args: Parameters<typeof history.pushState>) {
      origPush(...args)
      window.dispatchEvent(new Event('locationchange'))
    }
    const sync = () => { currentPath = window.location.pathname }
    window.addEventListener('locationchange', sync)
    window.addEventListener('popstate', sync)
    return () => {
      window.removeEventListener('locationchange', sync)
      window.removeEventListener('popstate', sync)
      history.pushState = origPush
    }
  })
  let basepath = '/'
</script>

<svelte:window on:popstate={() => { currentPath = window.location.pathname }} />

<Toaster />

{#if currentPath === '/auth/callback'}
  <OAuthCallback />
{:else if isAuthenticated}
  <Router {basepath}>
    <MainLayout {currentPath}>
      <Route path="/">
        <Dashboard />
      </Route>
      <Route path="/materials">
        <Materials />
      </Route>
      <Route path="/products">
        <Products />
      </Route>
      <Route path="/bom">
        <Bom />
      </Route>
      <Route path="/production">
        <Production />
      </Route>
      <Route path="/procurement">
        <Procurement />
      </Route>
      <Route path="/inventory">
        <Inventory />
      </Route>
      <Route path="/planning">
        <Planning />
      </Route>
      <Route path="/reports">
        <Reports />
      </Route>
      <Route path="/settings">
        <Settings />
      </Route>
      <Route path="/users">
        <Users />
      </Route>
      <Route path="/warehouses">
        <Warehouses />
      </Route>
      <Route path="/suppliers">
        <Suppliers />
      </Route>
      <Route path="*">
        <NotFound />
      </Route>
    </MainLayout>
  </Router>
{:else}
  <Login />
{/if}

