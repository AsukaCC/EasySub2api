export interface ImageHistoryItem {
  id: string
  createdAt: number
  prompt: string
  model: string
  platform: string
  src: string
  revisedPrompt?: string
  favorite: boolean
  favoriteCollectionIds?: string[]
  params: Record<string, unknown>
  keyName?: string
  width?: number
  height?: number
}

export interface ImageTaskItem {
  taskId: string
  keyId: string
  keyName?: string
  platform: string
  prompt: string
  model: string
  params: Record<string, unknown>
  status: string
  createdAt: number
  completedAt?: number
  error?: Record<string, unknown>
}

const DB_NAME = 'easysub2api-image-workbench'
const STORE_NAME = 'history'
const TASK_STORE_NAME = 'tasks'
const DB_VERSION = 2

function openDatabase(): Promise<IDBDatabase> {
  return new Promise((resolve, reject) => {
    const request = indexedDB.open(DB_NAME, DB_VERSION)
    request.onupgradeneeded = () => {
      const db = request.result
      if (!db.objectStoreNames.contains(STORE_NAME)) {
        const store = db.createObjectStore(STORE_NAME, { keyPath: 'id' })
        store.createIndex('createdAt', 'createdAt')
      }
      if (!db.objectStoreNames.contains(TASK_STORE_NAME)) {
        const store = db.createObjectStore(TASK_STORE_NAME, { keyPath: 'taskId' })
        store.createIndex('createdAt', 'createdAt')
      }
    }
    request.onsuccess = () => resolve(request.result)
    request.onerror = () => reject(request.error || new Error('IndexedDB unavailable'))
  })
}

export async function listTasks(): Promise<ImageTaskItem[]> {
  const db = await openDatabase()
  return new Promise((resolve, reject) => {
    const request = db.transaction(TASK_STORE_NAME).objectStore(TASK_STORE_NAME).getAll()
    request.onsuccess = () => resolve((request.result as ImageTaskItem[]).sort((a, b) => b.createdAt - a.createdAt))
    request.onerror = () => reject(request.error)
  })
}

export async function putTask(item: ImageTaskItem): Promise<void> {
  const db = await openDatabase()
  return new Promise((resolve, reject) => {
    const request = db.transaction(TASK_STORE_NAME, 'readwrite').objectStore(TASK_STORE_NAME).put(item)
    request.onsuccess = () => resolve(); request.onerror = () => reject(request.error)
  })
}

export async function deleteTask(taskId: string): Promise<void> {
  const db = await openDatabase()
  return new Promise((resolve, reject) => {
    const request = db.transaction(TASK_STORE_NAME, 'readwrite').objectStore(TASK_STORE_NAME).delete(taskId)
    request.onsuccess = () => resolve(); request.onerror = () => reject(request.error)
  })
}

export async function listHistory(): Promise<ImageHistoryItem[]> {
  const db = await openDatabase()
  return new Promise((resolve, reject) => {
    const request = db.transaction(STORE_NAME).objectStore(STORE_NAME).getAll()
    request.onsuccess = () => resolve((request.result as ImageHistoryItem[]).sort((a, b) => b.createdAt - a.createdAt))
    request.onerror = () => reject(request.error)
  })
}

export async function putHistory(item: ImageHistoryItem): Promise<void> {
  const db = await openDatabase()
  return new Promise((resolve, reject) => {
    const request = db.transaction(STORE_NAME, 'readwrite').objectStore(STORE_NAME).put(item)
    request.onsuccess = () => resolve()
    request.onerror = () => reject(request.error)
  })
}

export async function deleteHistory(id: string): Promise<void> {
  const db = await openDatabase()
  return new Promise((resolve, reject) => {
    const request = db.transaction(STORE_NAME, 'readwrite').objectStore(STORE_NAME).delete(id)
    request.onsuccess = () => resolve()
    request.onerror = () => reject(request.error)
  })
}

export async function clearHistory(): Promise<void> {
  const db = await openDatabase()
  return new Promise((resolve, reject) => {
    const request = db.transaction(STORE_NAME, 'readwrite').objectStore(STORE_NAME).clear()
    request.onsuccess = () => resolve()
    request.onerror = () => reject(request.error)
  })
}

export async function exportHistory(extra?: Record<string, unknown>): Promise<string> {
  return JSON.stringify({
    version: 2,
    exportedAt: new Date().toISOString(),
    items: await listHistory(),
    ...extra,
  }, null, 2)
}

export async function importHistory(raw: string): Promise<{ count: number; collections?: unknown; defaultFavoriteCollectionId?: unknown }> {
  const payload = JSON.parse(raw) as {
    version?: number
    items?: ImageHistoryItem[]
    collections?: unknown
    defaultFavoriteCollectionId?: unknown
  }
  if ((payload.version !== 1 && payload.version !== 2) || !Array.isArray(payload.items)) throw new Error('Invalid history export')
  const valid = payload.items.filter((item) => item && typeof item.id === 'string' && typeof item.src === 'string' && typeof item.prompt === 'string')
  for (const item of valid) {
    await putHistory({
      ...item,
      favorite: item.favorite === true || (Array.isArray(item.favoriteCollectionIds) && item.favoriteCollectionIds.length > 0),
      favoriteCollectionIds: Array.isArray(item.favoriteCollectionIds) ? item.favoriteCollectionIds : [],
    })
  }
  return {
    count: valid.length,
    collections: payload.collections,
    defaultFavoriteCollectionId: payload.defaultFavoriteCollectionId,
  }
}
