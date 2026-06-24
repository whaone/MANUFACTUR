<script lang="ts">
  import { onMount } from 'svelte'
  import Card from '$lib/components/ui/Card.svelte'
  import Button from '$lib/components/ui/Button.svelte'
  import Modal from '$lib/components/ui/Modal.svelte'
  import Input from '$lib/components/ui/Input.svelte'
  import { UserCog, Plus, Trash2, Check, X } from '@lucide/svelte'
  import { api } from '$lib/api'
  import { toast } from '$lib/stores/toast'

  const ROLES = ['owner', 'admin', 'production', 'warehouse', 'viewer'] as const
  type Role = typeof ROLES[number]

  // Permissions matrix: what each role can do
  const PERMISSIONS: { label: string; key: string; roles: Role[] }[] = [
    { label: 'Manage Users', key: 'manage_users', roles: ['owner'] },
    { label: 'Manage Workspace', key: 'manage_workspace', roles: ['owner'] },
    { label: 'Master Data (Create/Edit)', key: 'master_data_write', roles: ['owner', 'admin'] },
    { label: 'Master Data (View)', key: 'master_data_read', roles: ['owner', 'admin', 'production', 'warehouse', 'viewer'] },
    { label: 'Production Orders (Create)', key: 'production_write', roles: ['owner', 'admin', 'production'] },
    { label: 'Production Orders (View)', key: 'production_read', roles: ['owner', 'admin', 'production', 'warehouse', 'viewer'] },
    { label: 'Execute Production (FIFO)', key: 'production_execute', roles: ['owner', 'admin', 'production'] },
    { label: 'Procurement (Create PO)', key: 'procurement_write', roles: ['owner', 'admin'] },
    { label: 'Goods Receipt', key: 'goods_receipt', roles: ['owner', 'admin', 'warehouse'] },
    { label: 'Stock Transfer', key: 'stock_transfer', roles: ['owner', 'admin', 'warehouse'] },
    { label: 'Stock Adjustment', key: 'stock_adjustment', roles: ['owner', 'admin'] },
    { label: 'View Reports', key: 'reports_read', roles: ['owner', 'admin', 'viewer'] },
    { label: 'Export Reports', key: 'reports_export', roles: ['owner', 'admin'] },
  ]

  interface User {
    id: string
    name: string
    email: string
    role: Role
    created_at: string
  }

  let users = $state<User[]>([])
  let loading = $state(true)
  let saving = $state<string | null>(null)
  let showInvite = $state(false)
  let inviteForm = $state({ name: '', email: '', role: 'viewer' as Role, password: '' })
  let inviteError = $state('')
  let inviting = $state(false)

  onMount(async () => {
    try {
      const data = await api.users.list()
      users = data
    } catch (e) {
      console.error(e)
    } finally {
      loading = false
    }
  })

  async function changeRole(userId: string, newRole: Role) {
    saving = userId
    try {
      const updated = await api.users.updateRole(userId, newRole)
      users = users.map(u => u.id === userId ? { ...u, role: updated.role } : u)
      toast.success('Role diperbarui', newRole)
    } catch (e) {
      toast.error('Gagal ubah role', e instanceof Error ? e.message : 'error')
    } finally {
      saving = null
    }
  }

  async function deleteUser(userId: string) {
    if (!confirm('Hapus user ini?')) return
    try {
      await api.users.delete(userId)
      users = users.filter(u => u.id !== userId)
      toast.success('User dihapus')
    } catch (e) {
      toast.error('Gagal menghapus user', e instanceof Error ? e.message : 'error')
    }
  }

  async function handleInvite() {
    inviteError = ''
    if (!inviteForm.name || !inviteForm.email || !inviteForm.password) {
      inviteError = 'Semua field wajib diisi'
      return
    }
    inviting = true
    try {
      const u = await api.users.invite(inviteForm)
      users = [...users, u]
      showInvite = false
      inviteForm = { name: '', email: '', role: 'viewer', password: '' }
      toast.success('User ditambahkan', u.name)
    } catch (e: any) {
      inviteError = e?.message ?? 'Gagal menambah user'
      toast.error('Gagal menambah user', inviteError)
    } finally {
      inviting = false
    }
  }

  const roleColors: Record<Role, string> = {
    owner: 'bg-purple-100 text-purple-700 border-purple-200',
    admin: 'bg-blue-100 text-blue-700 border-blue-200',
    production: 'bg-orange-100 text-orange-700 border-orange-200',
    warehouse: 'bg-green-100 text-green-700 border-green-200',
    viewer: 'bg-gray-100 text-gray-600 border-gray-200',
  }
</script>

<div class="p-6 space-y-6">
  <!-- Header -->
  <div class="flex items-center justify-between">
    <div class="flex items-center gap-3">
      <UserCog class="w-6 h-6 text-primary" />
      <div>
        <h1 class="text-xl font-semibold text-on-surface">User & Role Management</h1>
        <p class="text-sm text-on-surface-variant">{users.length} user terdaftar</p>
      </div>
    </div>
    <Button onclick={() => showInvite = true}>
      <Plus class="w-4 h-4 mr-1" /> Tambah User
    </Button>
  </div>

  <!-- Permissions Matrix -->
  <Card>
    <div class="p-4 border-b border-outline-variant/30">
      <h2 class="font-medium text-on-surface">Permission Matrix</h2>
      <p class="text-xs text-on-surface-variant mt-0.5">Hak akses per role — read-only reference</p>
    </div>
    <div class="overflow-x-auto">
      <table class="w-full text-sm">
        <thead>
          <tr class="border-b border-outline-variant/20">
            <th class="text-left p-3 font-medium text-on-surface-variant w-56">Permission</th>
            {#each ROLES as role}
              <th class="p-3 font-medium text-center">
                <span class="px-2 py-0.5 rounded-full text-xs border {roleColors[role]}">
                  {role}
                </span>
              </th>
            {/each}
          </tr>
        </thead>
        <tbody>
          {#each PERMISSIONS as perm, i}
            <tr class="border-b border-outline-variant/10 {i % 2 === 0 ? 'bg-surface-container-lowest/40' : ''}">
              <td class="p-3 text-on-surface-variant">{perm.label}</td>
              {#each ROLES as role}
                <td class="p-3 text-center">
                  {#if perm.roles.includes(role)}
                    <Check class="w-4 h-4 text-status-ok mx-auto" />
                  {:else}
                    <span class="w-4 h-4 block mx-auto text-on-surface-variant/20">—</span>
                  {/if}
                </td>
              {/each}
            </tr>
          {/each}
        </tbody>
      </table>
    </div>
  </Card>

  <!-- Users Table -->
  <Card>
    <div class="p-4 border-b border-outline-variant/30">
      <h2 class="font-medium text-on-surface">Daftar User</h2>
    </div>
    {#if loading}
      <div class="p-8 text-center text-on-surface-variant">Memuat...</div>
    {:else if users.length === 0}
      <div class="p-8 text-center text-on-surface-variant">Belum ada user</div>
    {:else}
      <div class="overflow-x-auto">
        <table class="w-full text-sm">
          <thead>
            <tr class="border-b border-outline-variant/20 text-left">
              <th class="p-3 font-medium text-on-surface-variant">Nama</th>
              <th class="p-3 font-medium text-on-surface-variant">Email</th>
              {#each ROLES as role}
                <th class="p-3 font-medium text-center">
                  <span class="px-2 py-0.5 rounded-full text-xs border {roleColors[role]}">{role}</span>
                </th>
              {/each}
              <th class="p-3"></th>
            </tr>
          </thead>
          <tbody>
            {#each users as u}
              <tr class="border-b border-outline-variant/10 hover:bg-surface-container-low/40 transition-colors">
                <td class="p-3 font-medium text-on-surface">{u.name}</td>
                <td class="p-3 text-on-surface-variant">{u.email}</td>
                {#each ROLES as role}
                  <td class="p-3 text-center">
                    <button
                      onclick={() => changeRole(u.id, role)}
                      disabled={saving === u.id}
                      class="w-6 h-6 rounded-full border-2 mx-auto flex items-center justify-center transition-all
                        {u.role === role
                          ? 'border-primary bg-primary text-on-primary'
                          : 'border-outline-variant hover:border-primary/50'}"
                      title="Set role ke {role}"
                    >
                      {#if u.role === role}
                        <Check class="w-3 h-3" />
                      {/if}
                    </button>
                  </td>
                {/each}
                <td class="p-3">
                  <button
                    onclick={() => deleteUser(u.id)}
                    class="p-1.5 rounded-lg text-on-surface-variant hover:text-status-critical hover:bg-status-critical/10 transition-colors"
                    title="Hapus user"
                  >
                    <Trash2 class="w-4 h-4" />
                  </button>
                </td>
              </tr>
            {/each}
          </tbody>
        </table>
      </div>
    {/if}
  </Card>
</div>

<!-- Invite Modal -->
<Modal open={showInvite} title="Tambah User" onclose={() => showInvite = false}>
  <div class="space-y-4 p-1">
    {#if inviteError}
      <div class="text-sm text-status-critical bg-status-critical/10 rounded-lg px-3 py-2">{inviteError}</div>
    {/if}
    <Input label="Nama" bind:value={inviteForm.name} placeholder="Nama lengkap" />
    <Input label="Email" type="email" bind:value={inviteForm.email} placeholder="email@domain.com" />
    <Input label="Password" type="password" bind:value={inviteForm.password} placeholder="Min. 6 karakter" />
    <div class="space-y-1.5">
      <label for="user-role" class="text-sm font-medium text-on-surface">Role</label>
      <select id="user-role" bind:value={inviteForm.role}
        class="w-full rounded-xl border border-outline-variant bg-surface-container-lowest px-3 py-2.5 text-sm focus:outline-none focus:ring-2 focus:ring-primary/40">
        {#each ROLES as role}
          <option value={role}>{role}</option>
        {/each}
      </select>
    </div>
    <div class="flex gap-3 pt-2">
      <Button variant="secondary" onclick={() => showInvite = false} class="flex-1">Batal</Button>
      <Button onclick={handleInvite} disabled={inviting} class="flex-1">
        {inviting ? 'Menyimpan...' : 'Tambah User'}
      </Button>
    </div>
  </div>
</Modal>
