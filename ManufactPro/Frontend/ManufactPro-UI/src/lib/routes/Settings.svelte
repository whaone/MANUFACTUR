<script lang="ts">
  import Card from '$lib/components/ui/Card.svelte'
  import { Building2, FileText, DollarSign, Ruler, QrCode, Bell, RefreshCw, Cog, Plus, X } from '@lucide/svelte'

  const DEFAULT_UNITS = ['kg', 'gram', 'meter', 'cm', 'pcs', 'lusin', 'liter', 'ml', 'dus']

  const DEFAULTS = {
    company_name: '',
    address: '',
    phone: '',
    po_prefix: 'PO',
    prod_prefix: 'PROD',
    grn_prefix: 'GRN',
    doc_padding: 4,
    costing_method: 'FIFO',
    currency: 'IDR',
    language: 'Bahasa Indonesia',
    ppn_enabled: false,
    ppn_rate: 11,
    units: DEFAULT_UNITS,
    barcode_format: 'QR',
    label_size: 'A4',
    low_stock_threshold: 10,
    sync_interval: 30,
    max_queue: 500,
  }

  function loadSettings() {
    try {
      const s = localStorage.getItem('manufactpro_settings')
      return s ? { ...DEFAULTS, ...JSON.parse(s) } : { ...DEFAULTS }
    } catch { return { ...DEFAULTS } }
  }

  let settings = $state(loadSettings())
  let newUnit = $state('')

  $effect(() => {
    localStorage.setItem('manufactpro_settings', JSON.stringify(settings))
  })

  function addUnit() {
    const u = newUnit.trim()
    if (!u || settings.units.includes(u)) return
    settings.units = [...settings.units, u]
    newUnit = ''
  }

  function removeUnit(u: string) {
    settings.units = settings.units.filter((x: string) => x !== u)
  }

  const selectClass = 'px-3 py-1.5 rounded-lg text-xs bg-surface-container-low border border-outline-variant cursor-pointer focus:outline-none focus:ring-2 focus:ring-primary/30'
  const rowClass = 'flex items-center justify-between p-3 rounded-lg bg-surface-container-low/30'
  const inputClass = 'w-32 px-2 py-1 rounded-lg text-xs bg-surface-container-low border border-outline-variant focus:outline-none focus:ring-2 focus:ring-primary/30'
</script>

<div class="space-y-6">
  <div>
    <h1 class="text-[32px] font-bold text-on-surface tracking-tight">Settings</h1>
    <p class="text-on-surface-variant mt-1">Konfigurasi workspace ManufactPro</p>
  </div>

  <div class="grid grid-cols-1 md:grid-cols-2 gap-6">

    <!-- 1. Profil Workspace -->
    <Card variant="glass">
      <div class="flex items-center gap-2 mb-4">
        <Building2 class="w-5 h-5 text-primary" />
        <h2 class="text-sm font-semibold text-on-surface uppercase tracking-wider">Profil Workspace</h2>
      </div>
      <div class="space-y-3">
        <div class="space-y-1">
          <label for="s-company" class="text-xs text-on-surface-variant">Nama Perusahaan</label>
          <input id="s-company" bind:value={settings.company_name} placeholder="PT Manufaktur Jaya" class="{inputClass} w-full" />
        </div>
        <div class="space-y-1">
          <label for="s-address" class="text-xs text-on-surface-variant">Alamat</label>
          <input id="s-address" bind:value={settings.address} placeholder="Jl. Industri No. 1" class="{inputClass} w-full" />
        </div>
        <div class="space-y-1">
          <label for="s-phone" class="text-xs text-on-surface-variant">No. Telepon</label>
          <input id="s-phone" bind:value={settings.phone} placeholder="021-xxxx" class="{inputClass} w-full" />
        </div>
      </div>
    </Card>

    <!-- 2. Keuangan -->
    <Card variant="glass">
      <div class="flex items-center gap-2 mb-4">
        <DollarSign class="w-5 h-5 text-primary" />
        <h2 class="text-sm font-semibold text-on-surface uppercase tracking-wider">Keuangan</h2>
      </div>
      <div class="space-y-3">
        <div class={rowClass}>
          <div><p class="text-sm font-medium text-on-surface">Metode Costing</p></div>
          <select bind:value={settings.costing_method} class={selectClass}>
            <option>FIFO</option>
            <option>Weighted Average</option>
          </select>
        </div>
        <div class={rowClass}>
          <div><p class="text-sm font-medium text-on-surface">Mata Uang</p></div>
          <select bind:value={settings.currency} class={selectClass}>
            <option>IDR</option>
            <option>USD</option>
          </select>
        </div>
        <div class={rowClass}>
          <div><p class="text-sm font-medium text-on-surface">Bahasa</p></div>
          <select bind:value={settings.language} class={selectClass}>
            <option>Bahasa Indonesia</option>
            <option>English</option>
          </select>
        </div>
        <div class={rowClass}>
          <div>
            <p class="text-sm font-medium text-on-surface">PPN</p>
            <p class="text-xs text-on-surface-variant">{settings.ppn_enabled ? `${settings.ppn_rate}%` : 'Nonaktif'}</p>
          </div>
          <div class="flex items-center gap-2">
            {#if settings.ppn_enabled}
              <input type="number" bind:value={settings.ppn_rate} min="0" max="100" class="{inputClass} w-16" />
              <span class="text-xs text-on-surface-variant">%</span>
            {/if}
            <input type="checkbox" bind:checked={settings.ppn_enabled} class="w-4 h-4 accent-primary cursor-pointer" />
          </div>
        </div>
      </div>
    </Card>

    <!-- 3. Dokumen & Nomor Urut -->
    <Card variant="glass">
      <div class="flex items-center gap-2 mb-4">
        <FileText class="w-5 h-5 text-primary" />
        <h2 class="text-sm font-semibold text-on-surface uppercase tracking-wider">Dokumen & Nomor Urut</h2>
      </div>
      <div class="space-y-3">
        {#each [['Prefix PO', 'po_prefix', 'PO-0001'], ['Prefix Produksi', 'prod_prefix', 'PROD-0001'], ['Prefix GRN', 'grn_prefix', 'GRN-0001']] as [label, key, example]}
          <div class={rowClass}>
            <div>
              <p class="text-sm font-medium text-on-surface">{label}</p>
              <p class="text-xs text-on-surface-variant">Contoh: {settings[key]}-{'0'.repeat(settings.doc_padding - 1)}1</p>
            </div>
            <input bind:value={settings[key]} class={inputClass} placeholder={example.split('-')[0]} />
          </div>
        {/each}
        <div class={rowClass}>
          <div><p class="text-sm font-medium text-on-surface">Padding Angka</p><p class="text-xs text-on-surface-variant">Digit nomor urut</p></div>
          <select bind:value={settings.doc_padding} class={selectClass}>
            {#each [3,4,5,6] as n}<option value={n}>{n} digit</option>{/each}
          </select>
        </div>
      </div>
    </Card>

    <!-- 4. Satuan -->
    <Card variant="glass">
      <div class="flex items-center gap-2 mb-4">
        <Ruler class="w-5 h-5 text-primary" />
        <h2 class="text-sm font-semibold text-on-surface uppercase tracking-wider">Satuan (Unit)</h2>
      </div>
      <div class="flex flex-wrap gap-2 mb-3">
        {#each settings.units as u}
          <span class="inline-flex items-center gap-1 px-2.5 py-1 rounded-full text-xs bg-primary/10 text-primary font-medium">
            {u}
            <button onclick={() => removeUnit(u)} class="hover:text-status-critical transition-colors cursor-pointer"><X class="w-3 h-3" /></button>
          </span>
        {/each}
      </div>
      <div class="flex gap-2">
        <input bind:value={newUnit} placeholder="Tambah satuan..." onkeydown={(e) => e.key === 'Enter' && addUnit()} class="{inputClass} flex-1 w-auto" />
        <button onclick={addUnit} class="px-3 py-1 rounded-lg text-xs bg-primary text-on-primary hover:opacity-90 cursor-pointer flex items-center gap-1">
          <Plus class="w-3 h-3" /> Tambah
        </button>
      </div>
    </Card>

    <!-- 5. Label & Barcode -->
    <Card variant="glass">
      <div class="flex items-center gap-2 mb-4">
        <QrCode class="w-5 h-5 text-primary" />
        <h2 class="text-sm font-semibold text-on-surface uppercase tracking-wider">Label & Barcode</h2>
      </div>
      <div class="space-y-3">
        <div class={rowClass}>
          <div><p class="text-sm font-medium text-on-surface">Format Barcode</p></div>
          <select bind:value={settings.barcode_format} class={selectClass}>
            <option>QR</option>
            <option>Code128</option>
            <option>EAN-13</option>
          </select>
        </div>
        <div class={rowClass}>
          <div><p class="text-sm font-medium text-on-surface">Ukuran Kertas Label</p></div>
          <select bind:value={settings.label_size} class={selectClass}>
            <option>A4</option>
            <option>A5</option>
            <option>Letter</option>
            <option>Label 100x50</option>
          </select>
        </div>
      </div>
    </Card>

    <!-- 6. Notifikasi Stok -->
    <Card variant="glass">
      <div class="flex items-center gap-2 mb-4">
        <Bell class="w-5 h-5 text-primary" />
        <h2 class="text-sm font-semibold text-on-surface uppercase tracking-wider">Notifikasi Stok</h2>
      </div>
      <div class="space-y-3">
        <div class={rowClass}>
          <div>
            <p class="text-sm font-medium text-on-surface">Threshold Stok Minimum</p>
            <p class="text-xs text-on-surface-variant">Alert jika stok ≤ nilai ini</p>
          </div>
          <div class="flex items-center gap-2">
            <input type="number" bind:value={settings.low_stock_threshold} min="0" class="{inputClass} w-20" />
            <span class="text-xs text-on-surface-variant">unit</span>
          </div>
        </div>
      </div>
    </Card>

    <!-- 7. Sync & Offline -->
    <Card variant="glass">
      <div class="flex items-center gap-2 mb-4">
        <RefreshCw class="w-5 h-5 text-primary" />
        <h2 class="text-sm font-semibold text-on-surface uppercase tracking-wider">Sync & Offline</h2>
      </div>
      <div class="space-y-3">
        <div class={rowClass}>
          <div>
            <p class="text-sm font-medium text-on-surface">Interval Sync</p>
            <p class="text-xs text-on-surface-variant">Auto-sync setiap X detik</p>
          </div>
          <div class="flex items-center gap-2">
            <input type="number" bind:value={settings.sync_interval} min="5" max="3600" class="{inputClass} w-20" />
            <span class="text-xs text-on-surface-variant">detik</span>
          </div>
        </div>
        <div class={rowClass}>
          <div>
            <p class="text-sm font-medium text-on-surface">Maks Offline Queue</p>
            <p class="text-xs text-on-surface-variant">Batas operasi disimpan offline</p>
          </div>
          <div class="flex items-center gap-2">
            <input type="number" bind:value={settings.max_queue} min="10" max="10000" class="{inputClass} w-20" />
            <span class="text-xs text-on-surface-variant">ops</span>
          </div>
        </div>
      </div>
    </Card>

  </div>
</div>
