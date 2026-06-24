<script lang="ts">
  import { cn } from '$lib/utils/cn'
  import Card from '$lib/components/ui/Card.svelte'
  import {
    BarChart3,
    TrendingUp,
    DollarSign,
    TrendingDown,
    Download,
    Calendar,
    Search,
  } from '@lucide/svelte'
  import { Bar, Line, Doughnut } from 'svelte-chartjs'
  import {
    Chart,
    CategoryScale,
    LinearScale,
    BarElement,
    LineElement,
    PointElement,
    ArcElement,
    Title,
    Tooltip,
    Legend,
    Filler,
  } from 'chart.js'
  import { api } from '$lib/api'
  import { onMount } from 'svelte'
  import { formatDate, colorForBar, barWidth } from '$lib/utils/reports'

  Chart.register(
    CategoryScale, LinearScale, BarElement, LineElement,
    PointElement, ArcElement, Title, Tooltip, Legend, Filler
  )

  interface MarginItem {
    variant_sku: string
    product_name: string
    hpp_per_unit: number
    sell_price: number
    margin: number
    margin_pct: number
    qty_produced: number
    total_cost: number
  }

  interface DashboardData {
    completed_orders: number
    total_qty_produced: number
    received_pos: number
    stock_value: number
    material_cost_month: number
  }

  interface TrendItem {
    month: string
    qty_produced: number
    order_count: number
  }

  let marginData = $state<MarginItem[]>([])
  let dashboard = $state<DashboardData>({
    completed_orders: 0, total_qty_produced: 0, received_pos: 0, stock_value: 0, material_cost_month: 0,
  })
  let trendItems = $state<TrendItem[]>([])
  let loading = $state(true)

  onMount(async () => {
    try {
      const [dash, margin, trend] = await Promise.all([
        api.reports.dashboard(),
        api.reports.hppMargin(),
        api.reports.productionTrend(),
      ])
      dashboard = dash
      marginData = margin
      trendItems = trend
    } catch (e) {
      console.error('Failed to load reports:', e)
    } finally {
      loading = false
    }
  })

  const CHART_COLORS = [
    'rgba(99,102,241,0.8)', 'rgba(139,92,246,0.8)', 'rgba(236,72,153,0.8)',
    'rgba(245,158,11,0.8)', 'rgba(16,185,129,0.8)',
  ]

  let marginChartData = $derived({
    labels: marginData.map(d => d.variant_sku),
    datasets: [
      { label: 'HPP/unit', data: marginData.map(d => d.hpp_per_unit), backgroundColor: 'rgba(239,68,68,0.7)', borderRadius: 4 },
      { label: 'Harga Jual', data: marginData.map(d => d.sell_price), backgroundColor: 'rgba(34,197,94,0.7)', borderRadius: 4 },
    ],
  })

  let totalMarginChartData = $derived({
    labels: marginData.map(d => d.variant_sku),
    datasets: [{
      label: 'Margin/unit (Rp)',
      data: marginData.map(d => d.margin),
      backgroundColor: marginData.map((_, i) => CHART_COLORS[i % CHART_COLORS.length]),
      borderRadius: 4,
    }],
  })

  let productionTrendData = $derived({
    labels: trendItems.map(t => t.month),
    datasets: [{
      label: 'Unit Diproduksi',
      data: trendItems.map(t => t.qty_produced),
      borderColor: 'rgb(99,102,241)',
      backgroundColor: 'rgba(99,102,241,0.1)',
      fill: true, tension: 0.4, pointRadius: 4,
    }],
  })

  const chartOptions = {
    responsive: true,
    plugins: { legend: { position: 'bottom' as const } },
    scales: { y: { beginAtZero: true } },
  }

  const doughnutOptions = {
    responsive: true,
    plugins: { legend: { position: 'bottom' as const } },
  }

  // Kategori doughnut — derived from marginData grouped by product_name
  let categoryNames = $derived([...new Set(marginData.map(d => d.product_name))])
  let categoryChartData = $derived({
    labels: categoryNames,
    datasets: [{
      data: categoryNames.map(name =>
        marginData.filter(d => d.product_name === name).reduce((s, d) => s + d.qty_produced, 0)
      ),
      backgroundColor: CHART_COLORS,
      borderWidth: 0,
    }],
  })

  let search = $state('')
  let showDatePicker = $state(false)
  let startDate = $state('')
  let endDate = $state('')

  let dateLabel = $derived(
    startDate && endDate ? `${formatDate(startDate)} — ${formatDate(endDate)}`
      : startDate ? `From ${formatDate(startDate)}`
      : endDate ? `Until ${formatDate(endDate)}`
      : 'Pick a date range'
  )

  function toggleDatePicker() { showDatePicker = !showDatePicker }

  function handleClickOutside(e: MouseEvent) {
    const target = e.target as HTMLElement
    if (!target.closest('[data-date-picker]')) showDatePicker = false
  }

  function handleExport() {
    const data = filteredMargin.map(item => ({
      product: item.product_name, variant: item.variant_sku,
      hpp_per_unit: item.hpp_per_unit, sell_price: item.sell_price,
      margin: item.margin, margin_pct: item.margin_pct.toFixed(1),
      qty_produced: item.qty_produced,
    }))
    const header = Object.keys(data[0] || {}).join(',')
    const rows = data.map(r => Object.values(r).join(','))
    const csv = [header, ...rows].join('\n')
    const blob = new Blob([csv], { type: 'text/csv;charset=utf-8;' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `manufactpro_report_${startDate || 'all'}_to_${endDate || 'all'}.csv`
    a.click()
    URL.revokeObjectURL(url)
  }

  let filteredMargin = $derived(
    marginData.filter(m => !search || m.product_name.toLowerCase().includes(search.toLowerCase()) || m.variant_sku.toLowerCase().includes(search.toLowerCase()))
  )

  let totalRevenue = $derived(filteredMargin.reduce((a, i) => a + i.sell_price * i.qty_produced, 0))
  let totalHpp = $derived(filteredMargin.reduce((a, i) => a + i.hpp_per_unit * i.qty_produced, 0))
  let totalMargin = $derived(filteredMargin.reduce((a, i) => a + i.margin * i.qty_produced, 0))
  let avgMarginPct = $derived(totalRevenue > 0 ? (totalMargin / totalRevenue * 100).toFixed(1) : '0')
</script>

<svelte:window onclick={handleClickOutside} />

<div class="space-y-6">
  <div class="flex items-center justify-between">
    <div>
      <h1 class="text-[32px] font-bold text-on-surface tracking-tight">Reports & Analytics</h1>
      <p class="text-on-surface-variant mt-1">HPP, profit margins, and material consumption</p>
    </div>
    <div class="flex gap-2">
      <div class="relative" data-date-picker>
        <button
          onclick={toggleDatePicker}
          class="inline-flex items-center gap-2 px-3 py-2 rounded-lg text-sm bg-surface-container-high border border-outline-variant hover:bg-surface-container-highest transition-colors cursor-pointer"
        >
          <Calendar class="w-4 h-4 text-on-surface-variant" />
          <span class="max-w-[180px] truncate">{dateLabel}</span>
        </button>

        {#if showDatePicker}
          <div class="absolute right-0 top-full mt-2 w-72 bg-surface-container-lowest rounded-xl shadow-soft border border-outline-variant/50 overflow-hidden animate-[zoom-in_0.15s_ease-out] z-40">
            <div class="px-4 py-3 border-b border-outline-variant/30">
              <p class="text-xs font-semibold text-on-surface-variant uppercase tracking-wider">Date Range</p>
            </div>
            <div class="p-4 space-y-3">
              <div class="space-y-1.5">
                <label for="start-date" class="text-xs font-medium text-on-surface-variant">Start Date</label>
                <input
                  id="start-date"
                  type="date"
                  bind:value={startDate}
                  class="w-full px-3 py-2 rounded-lg text-sm bg-surface-container-low border border-outline-variant focus:outline-none focus:ring-2 focus:ring-primary/30 focus:border-primary cursor-pointer"
                />
              </div>
              <div class="space-y-1.5">
                <label for="end-date" class="text-xs font-medium text-on-surface-variant">End Date</label>
                <input
                  id="end-date"
                  type="date"
                  bind:value={endDate}
                  class="w-full px-3 py-2 rounded-lg text-sm bg-surface-container-low border border-outline-variant focus:outline-none focus:ring-2 focus:ring-primary/30 focus:border-primary cursor-pointer"
                />
              </div>
              {#if startDate || endDate}
                <button
                  onclick={() => { startDate = ''; endDate = '' }}
                  class="text-xs text-primary hover:underline cursor-pointer w-full text-center pt-1"
                >
                  Clear dates
                </button>
              {/if}
            </div>
          </div>
        {/if}
      </div>
      <button onclick={handleExport} class="inline-flex items-center gap-2 px-3 py-2 rounded-lg text-sm primary-gradient text-on-primary hover:opacity-90 transition-colors cursor-pointer">
        <Download class="w-4 h-4" />
        <span>Export</span>
      </button>
    </div>
  </div>

  <div class="grid grid-cols-1 sm:grid-cols-2 xl:grid-cols-4 gap-4">
    <Card variant="glass">
      <div class="flex items-start justify-between mb-2">
        <span class="text-xs font-semibold uppercase tracking-wider text-on-surface-variant">Revenue</span>
        <DollarSign class="w-4 h-4 text-primary" />
      </div>
      <p class="text-xl font-bold text-on-surface">Rp {(totalRevenue / 1e6).toFixed(1)}M</p>
      <span class="text-xs text-status-success">+12.5% vs prev month</span>
    </Card>
    <Card variant="glass">
      <div class="flex items-start justify-between mb-2">
        <span class="text-xs font-semibold uppercase tracking-wider text-on-surface-variant">Total HPP</span>
        <TrendingDown class="w-4 h-4 text-status-warning" />
      </div>
      <p class="text-xl font-bold text-on-surface">Rp {(totalHpp / 1e6).toFixed(1)}M</p>
      <span class="text-xs text-status-warning">-3.2% vs prev month</span>
    </Card>
    <Card variant="glass">
      <div class="flex items-start justify-between mb-2">
        <span class="text-xs font-semibold uppercase tracking-wider text-on-surface-variant">Total Margin</span>
        <TrendingUp class="w-4 h-4 text-status-success" />
      </div>
      <p class="text-xl font-bold text-on-surface">Rp {(totalMargin / 1e6).toFixed(1)}M</p>
      <span class="text-xs text-status-success">{avgMarginPct}% margin</span>
    </Card>
    <Card variant="gradient">
      <div class="flex items-start justify-between mb-2">
        <span class="text-xs font-semibold uppercase tracking-wider text-on-primary/80">Avg Margin</span>
        <BarChart3 class="w-4 h-4 text-on-primary" />
      </div>
      <p class="text-xl font-bold text-on-primary">{avgMarginPct}%</p>
      <span class="text-xs text-on-primary/80">per product</span>
    </Card>
  </div>

  <!-- Charts section -->
  <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
    <Card variant="glass" padding="none">
      <div class="px-5 py-4 border-b border-outline-variant/30">
        <h3 class="text-sm font-semibold text-on-surface uppercase tracking-wider">HPP vs Harga Jual per Varian</h3>
      </div>
      <div class="p-4">
        <Bar data={marginChartData} options={chartOptions} />
      </div>
    </Card>

    <Card variant="glass" padding="none">
      <div class="px-5 py-4 border-b border-outline-variant/30">
        <h3 class="text-sm font-semibold text-on-surface uppercase tracking-wider">Total Margin per Varian</h3>
      </div>
      <div class="p-4">
        <Bar data={totalMarginChartData} options={chartOptions} />
      </div>
    </Card>

    <Card variant="glass" padding="none">
      <div class="px-5 py-4 border-b border-outline-variant/30">
        <h3 class="text-sm font-semibold text-on-surface uppercase tracking-wider">Tren Produksi 6 Bulan</h3>
      </div>
      <div class="p-4">
        <Line data={productionTrendData} options={chartOptions} />
      </div>
    </Card>

    <Card variant="glass" padding="none">
      <div class="px-5 py-4 border-b border-outline-variant/30">
        <h3 class="text-sm font-semibold text-on-surface uppercase tracking-wider">Komposisi Kategori</h3>
      </div>
      <div class="p-4 flex justify-center">
        <div class="w-full max-w-xs">
          <Doughnut data={categoryChartData} options={doughnutOptions} />
        </div>
      </div>
    </Card>
  </div>

  <div class="grid grid-cols-1 xl:grid-cols-3 gap-6">
    <div class="xl:col-span-2">
      <Card variant="glass" padding="none">
        <div class="px-5 py-4 border-b border-outline-variant/30">
          <h3 class="text-sm font-semibold text-on-surface uppercase tracking-wider">Product Margin Analysis</h3>
        </div>
        <div class="p-5 space-y-4">
          <div class="px-5 py-3 border-b border-outline-variant/30">
           <div class="relative">
             <Search class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-outline" />
             <input
               type="text"
               bind:value={search}
               placeholder="Filter by product..."
               class="w-full pl-10 pr-4 py-2 rounded-lg text-sm bg-surface-container-low border border-outline-variant focus:outline-none focus:ring-2 focus:ring-primary/30 focus:border-primary"
             />
           </div>
         </div>
         {#each filteredMargin as item}
            <div class="space-y-1.5">
              <div class="flex items-end justify-between text-sm">
                <div class="flex items-center gap-2">
                  <span class="font-medium text-on-surface">{item.product_name} - {item.variant_sku}</span>
                  <span class="text-xs text-on-surface-variant">({item.qty_produced} diproduksi)</span>
                </div>
                <div class="text-right">
                  <span class="font-semibold text-on-surface">Rp {item.margin.toLocaleString('id-ID')}</span>
                  <span class="text-xs text-on-surface-variant ml-1">{item.margin_pct.toFixed(1)}%</span>
                </div>
              </div>
              <div class="h-2 rounded-full bg-surface-container-low overflow-hidden">
                <div
                  class={colorForBar(item.margin_pct)}
                  style="height: 100%; width: {barWidth(item.margin_pct)}%; border-radius: 9999px; transition: width 0.5s ease;"
                ></div>
              </div>
              <div class="flex justify-between text-xs text-outline">
                <span>HPP: Rp {item.hpp_per_unit.toLocaleString('id-ID')}</span>
                <span>Sell: Rp {item.sell_price.toLocaleString('id-ID')}</span>
              </div>
            </div>
          {/each}
        </div>
      </Card>
    </div>

    <div>
      <Card variant="glass">
        <h3 class="text-sm font-semibold text-on-surface uppercase tracking-wider mb-4">Dashboard Bulan Ini</h3>
        <div class="space-y-4">
          <div class="flex items-center justify-between">
            <span class="text-sm text-on-surface-variant">Order Selesai</span>
            <span class="font-bold text-on-surface">{dashboard.completed_orders}</span>
          </div>
          <div class="flex items-center justify-between">
            <span class="text-sm text-on-surface-variant">Total Qty Produksi</span>
            <span class="font-bold text-on-surface">{dashboard.total_qty_produced.toLocaleString('id-ID')}</span>
          </div>
          <div class="flex items-center justify-between">
            <span class="text-sm text-on-surface-variant">PO Diterima</span>
            <span class="font-bold text-on-surface">{dashboard.received_pos}</span>
          </div>
          <div class="border-t border-outline-variant/30 pt-3 flex items-center justify-between">
            <span class="text-sm text-on-surface-variant">Nilai Stok</span>
            <span class="font-bold text-primary">Rp {(dashboard.stock_value / 1e6).toFixed(1)}M</span>
          </div>
          <div class="flex items-center justify-between">
            <span class="text-sm text-on-surface-variant">Biaya Material</span>
            <span class="font-bold text-status-warning">Rp {(dashboard.material_cost_month / 1e6).toFixed(1)}M</span>
          </div>
        </div>
      </Card>
    </div>
  </div>
</div>
