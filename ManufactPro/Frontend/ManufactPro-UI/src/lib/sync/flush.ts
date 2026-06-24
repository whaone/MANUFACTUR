import { db } from '$lib/db'
import { api } from '$lib/api'
import type { SyncQueueItem } from '$lib/types'

let flushing = false

/**
 * Drain the offline sync queue to the backend.
 * Reads pending items, replays them via POST /api/sync, then reconciles
 * each queue row from the per-item result (synced rows are removed,
 * failed rows are marked with the error and retry count bumped).
 */
export async function flushSyncQueue(): Promise<{ synced: number; failed: number }> {
  if (flushing) return { synced: 0, failed: 0 }
  flushing = true
  try {
    const pending = await db.syncQueue
      .where('status')
      .anyOf('pending', 'failed')
      .sortBy('client_timestamp')

    if (pending.length === 0) return { synced: 0, failed: 0 }

    // mark in-flight
    await db.syncQueue.bulkPut(pending.map((p) => ({ ...p, status: 'syncing' as const })))

    const resp = await api.sync.push(
      pending.map((p) => ({
        id: p.id,
        operation: p.operation,
        entity: p.entity,
        payload: p.payload,
        client_timestamp: p.client_timestamp,
      })),
    )

    const byId = new Map(pending.map((p) => [p.id, p]))
    for (const r of resp.results) {
      const item = byId.get(r.id)
      if (!item) continue
      if (r.status === 'synced') {
        await db.syncQueue.delete(r.id)
      } else {
        await db.syncQueue.put({
          ...item,
          status: 'failed',
          error: r.error ?? 'sync failed',
          retries: (item.retries ?? 0) + 1,
        })
      }
    }
    return { synced: resp.synced, failed: resp.failed }
  } catch (e) {
    // network/server error — revert in-flight rows back to pending for retry
    const stuck = await db.syncQueue.where('status').equals('syncing').toArray()
    await db.syncQueue.bulkPut(stuck.map((p) => ({ ...p, status: 'pending' as const })))
    throw e
  } finally {
    flushing = false
  }
}

/** Enqueue a local mutation for later sync. */
export async function enqueue(
  operation: SyncQueueItem['operation'],
  entity: string,
  payload: unknown,
): Promise<void> {
  await db.syncQueue.put({
    id: crypto.randomUUID(),
    operation,
    entity,
    payload,
    client_timestamp: Date.now(),
    status: 'pending',
    retries: 0,
  })
}

/** Auto-flush when the browser regains connectivity. */
export function initSyncListener(): void {
  if (typeof window === 'undefined') return
  window.addEventListener('online', () => {
    flushSyncQueue().catch((e) => console.error('sync flush failed:', e))
  })
}
