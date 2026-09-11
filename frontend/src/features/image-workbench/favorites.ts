export const ALL_FAVORITES_COLLECTION_ID = '__all_favorites__'
export const DEFAULT_FAVORITE_COLLECTION_ID = '__default_favorites__'
const STORAGE_KEY = 'image-workbench-collections'

export interface FavoriteCollection {
  id: string
  name: string
  createdAt: number
  updatedAt: number
}

export interface FavoriteState {
  collections: FavoriteCollection[]
  defaultFavoriteCollectionId: string | null
}

function normalizeName(value: string) {
  return value.trim().replace(/\s+/g, ' ').slice(0, 60)
}

export function createDefaultFavoriteCollection(now = Date.now()): FavoriteCollection {
  return {
    id: DEFAULT_FAVORITE_COLLECTION_ID,
    name: '默认',
    createdAt: now,
    updatedAt: now,
  }
}

export function normalizeFavoriteCollectionIds(value: unknown) {
  if (!Array.isArray(value)) return []
  return Array.from(new Set(value.map(String).filter((id) => id && id !== ALL_FAVORITES_COLLECTION_ID)))
}

export function normalizeFavoriteCollections(value: unknown, now = Date.now()): FavoriteCollection[] {
  const collections = Array.isArray(value) ? value : []
  const normalized: FavoriteCollection[] = []
  const ids = new Set<string>()
  for (const item of collections) {
    if (!item || typeof item !== 'object') continue
    const collection = item as Partial<FavoriteCollection>
    if (typeof collection.id !== 'string' || !collection.id.trim()) continue
    if (collection.id === ALL_FAVORITES_COLLECTION_ID || ids.has(collection.id)) continue
    const name = normalizeName(typeof collection.name === 'string' ? collection.name : '')
    if (!name) continue
    ids.add(collection.id)
    normalized.push({
      id: collection.id,
      name,
      createdAt: typeof collection.createdAt === 'number' ? collection.createdAt : now,
      updatedAt: typeof collection.updatedAt === 'number' ? collection.updatedAt : now,
    })
  }
  return normalized
}

export function ensureDefaultFavoriteCollection(collections: FavoriteCollection[], now = Date.now()) {
  if (collections.length > 0) return collections
  return [createDefaultFavoriteCollection(now)]
}

export function resolveDefaultFavoriteCollectionId(collections: FavoriteCollection[], preferredId: unknown) {
  if (preferredId === null) return collections[0]?.id ?? null
  if (typeof preferredId === 'string' && collections.some((collection) => collection.id === preferredId)) return preferredId
  if (collections.some((collection) => collection.id === DEFAULT_FAVORITE_COLLECTION_ID)) return DEFAULT_FAVORITE_COLLECTION_ID
  return collections[0]?.id ?? null
}

export function itemCollectionIds(item: { favorite?: boolean; favoriteCollectionIds?: string[] }, defaultFavoriteCollectionId: string | null) {
  const ids = normalizeFavoriteCollectionIds(item.favoriteCollectionIds)
  if (ids.length > 0) return ids
  return item.favorite && defaultFavoriteCollectionId ? [defaultFavoriteCollectionId] : []
}

export function isItemFavorite(item: { favorite?: boolean; favoriteCollectionIds?: string[] }, defaultFavoriteCollectionId: string | null) {
  return itemCollectionIds(item, defaultFavoriteCollectionId).length > 0 || Boolean(item.favorite)
}

export function loadFavoriteState(): FavoriteState {
  try {
    const saved = JSON.parse(localStorage.getItem(STORAGE_KEY) || '{}') as Partial<FavoriteState>
    const collections = ensureDefaultFavoriteCollection(normalizeFavoriteCollections(saved.collections))
    return {
      collections,
      defaultFavoriteCollectionId: resolveDefaultFavoriteCollectionId(collections, saved.defaultFavoriteCollectionId),
    }
  } catch {
    const collections = [createDefaultFavoriteCollection()]
    return { collections, defaultFavoriteCollectionId: collections[0].id }
  }
}

export function saveFavoriteState(state: FavoriteState) {
  localStorage.setItem(STORAGE_KEY, JSON.stringify({
    collections: state.collections,
    defaultFavoriteCollectionId: state.defaultFavoriteCollectionId,
  }))
}

export function createFavoriteCollection(name: string, collections: FavoriteCollection[]) {
  const normalized = normalizeName(name)
  if (!normalized) return null
  if (collections.some((collection) => collection.name === normalized)) return null
  const now = Date.now()
  return {
    id: crypto.randomUUID(),
    name: normalized,
    createdAt: now,
    updatedAt: now,
  } satisfies FavoriteCollection
}

export function mergeImportedFavoriteState(current: FavoriteState, imported: unknown, importedDefaultId?: unknown): FavoriteState {
  const incoming = normalizeFavoriteCollections(imported)
  const collections = ensureDefaultFavoriteCollection(normalizeFavoriteCollections([...current.collections, ...incoming]))
  return {
    collections,
    defaultFavoriteCollectionId: resolveDefaultFavoriteCollectionId(collections, importedDefaultId ?? current.defaultFavoriteCollectionId),
  }
}
