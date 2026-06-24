import { enqueue } from '$lib/sync/flush'
import type { SyncQueueItem } from '$lib/types'

/** Entities the backend /api/sync endpoint can replay. */
export type SyncEntity = 'material' | 'supplier' | 'product'

/**
 * fetch() rejects with a TypeError on network failure (offline, DNS, refused
 * connection). HTTP error statuses do NOT reject — apiFetch converts them to
 * Errors. So only a TypeError (or navigator.onLine === false) means "no
 * network"; genuine server errors must still surface to the caller.
 */
function noNetwork(e?: unknown): boolean {
  if (typeof navigator !== 'undefined' && !navigator.onLine) return true
  return e instanceof TypeError
}

/**
 * Run a write online; when the network is unavailable, queue the operation for
 * later replay (flushSyncQueue runs on reconnect) and resolve optimistically so
 * the UI stays responsive instead of losing the edit.
 *
 * Reads are server-authoritative: an offline create's temporary client id is
 * superseded by the real server id on the next list() after the queue syncs.
 */
export async function syncedWrite<T>(
  entity: SyncEntity,
  operation: SyncQueueItem['operation'],
  payload: Record<string, unknown>,
  run: () => Promise<T>,
  optimistic: () => T,
): Promise<T> {
  if (!noNetwork()) {
    try {
      return await run()
    } catch (e) {
      if (!noNetwork(e)) throw e
      // network died mid-flight → fall through and queue
    }
  }
  await enqueue(operation, entity, payload)
  return optimistic()
}
