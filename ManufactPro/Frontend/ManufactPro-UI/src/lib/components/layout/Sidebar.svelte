<script lang="ts">
  import { cn } from '$lib/utils/cn'
  import { sidebarOpen, closeSidebar } from '$lib/stores/sidebar'
  import { link } from 'svelte-routing'
  import {
    LayoutDashboard,
    Package,
    Boxes,
    ClipboardList,
    Factory,
    Truck,
    Warehouse,
    BarChart3,
    Settings,
    Calculator,
    X,
    ChevronRight,
    UserCog,
    Building2,
  } from '@lucide/svelte'

  interface Props {
    currentPath?: string
  }

  let { currentPath = '/' }: Props = $props()
  let isOpen = $derived($sidebarOpen)

  const navItems = [
    { label: 'Dashboard', icon: LayoutDashboard, href: '/' },
    { section: 'Master Data', items: [
      { label: 'Materials', icon: Package, href: '/materials' },
      { label: 'Products', icon: Boxes, href: '/products' },
      { label: 'BOM', icon: ClipboardList, href: '/bom' },
      { label: 'Supplier', icon: Truck, href: '/suppliers' },
      { label: 'Gudang', icon: Building2, href: '/warehouses' },
    ]},
    { section: 'Operations', items: [
      { label: 'Production', icon: Factory, href: '/production' },
      { label: 'Planning', icon: Calculator, href: '/planning' },
      { label: 'Procurement', icon: Warehouse, href: '/procurement' },
      { label: 'Inventory', icon: Warehouse, href: '/inventory' },
    ]},
    { section: 'Insights', items: [
      { label: 'Reports', icon: BarChart3, href: '/reports' },
    ]},
    { section: 'Admin', items: [
      { label: 'User & Roles', icon: UserCog, href: '/users' },
      { label: 'Settings', icon: Settings, href: '/settings' },
    ]},
  ]

  function isActive(href: string): boolean {
    if (href === '/') return currentPath === '/'
    return currentPath.startsWith(href)
  }
</script>

{#if isOpen}
  <div
    class="fixed inset-0 bg-on-background/30 backdrop-blur-sm z-40 md:hidden"
    onclick={closeSidebar}
    onkeydown={() => {}}
    role="button"
    tabindex="-1"
  ></div>
{/if}

<aside class={cn(
  'fixed left-0 top-0 z-50 h-screen w-64 bg-surface-container-lowest/80 backdrop-blur-md border-r border-outline-variant/50',
  'flex flex-col transition-transform duration-200 ease-out',
  'md:translate-x-0',
  isOpen ? 'translate-x-0' : '-translate-x-full',
)}>
  <div class="flex items-center justify-between px-5 py-5 border-b border-outline-variant/30">
    <div class="flex items-center gap-3">
      <div class="w-8 h-8 rounded-lg primary-gradient flex items-center justify-center">
        <Factory class="w-5 h-5 text-on-primary" />
      </div>
      <span class="text-lg font-bold text-on-surface tracking-tight">ManufactPro</span>
    </div>
    <button onclick={closeSidebar} class="md:hidden p-1 rounded hover:bg-surface-container-high transition-colors cursor-pointer">
      <X class="w-5 h-5 text-on-surface-variant" />
    </button>
  </div>

  <nav class="flex-1 overflow-y-auto py-4 px-3 space-y-1">
    {#each navItems as item}
      {#if 'section' in item}
        <div class="pt-4 pb-1 px-3">
          <span class="text-[11px] font-bold text-outline tracking-widest uppercase">{item.section}</span>
        </div>
        {#each item.items as sub}
          <a
            href={sub.href}
            use:link
            onclick={closeSidebar}
            class={cn(
              'flex items-center gap-3 px-3 py-2.5 rounded-xl text-sm font-medium transition-all duration-150 group',
              isActive(sub.href)
                ? 'bg-primary/10 text-primary'
                : 'text-on-surface-variant hover:bg-surface-container-high hover:text-on-surface',
            )}
          >
            <sub.icon class="w-5 h-5 shrink-0" />
            <span class="flex-1">{sub.label}</span>
            {#if isActive(sub.href)}
              <ChevronRight class="w-4 h-4 opacity-50" />
            {/if}
          </a>
        {/each}
      {:else}
        <a
          href={item.href}
          use:link
          onclick={closeSidebar}
          class={cn(
            'flex items-center gap-3 px-3 py-2.5 rounded-xl text-sm font-medium transition-all duration-150 group',
            isActive(item.href)
              ? 'bg-primary/10 text-primary'
              : 'text-on-surface-variant hover:bg-surface-container-high hover:text-on-surface',
          )}
        >
          <item.icon class="w-5 h-5 shrink-0" />
          <span class="flex-1">{item.label}</span>
          {#if isActive(item.href)}
            <ChevronRight class="w-4 h-4 opacity-50" />
          {/if}
        </a>
      {/if}
    {/each}
  </nav>

  <div class="p-4 border-t border-outline-variant/30">
    <div class="flex items-center gap-3 px-3 py-2">
      <div class="w-8 h-8 rounded-full bg-primary/10 flex items-center justify-center">
        <span class="text-sm font-semibold text-primary">A</span>
      </div>
      <div class="flex-1 min-w-0">
        <p class="text-sm font-medium text-on-surface truncate">Admin</p>
        <p class="text-xs text-outline truncate">admin@manufactpro.id</p>
      </div>
    </div>
  </div>
</aside>
