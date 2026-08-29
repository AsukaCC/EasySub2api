const LEGACY_PREFIX = 'sub2api:'
const CANONICAL_PREFIX = 'easysub2api:'

function migrateStorage(storage: Storage): void {
  try {
    const legacyKeys: string[] = []
    for (let index = 0; index < storage.length; index += 1) {
      const key = storage.key(index)
      if (key?.startsWith(LEGACY_PREFIX)) legacyKeys.push(key)
    }

    for (const legacyKey of legacyKeys) {
      const canonicalKey = `${CANONICAL_PREFIX}${legacyKey.slice(LEGACY_PREFIX.length)}`
      if (storage.getItem(canonicalKey) === null) {
        const value = storage.getItem(legacyKey)
        if (value !== null) storage.setItem(canonicalKey, value)
      }
      storage.removeItem(legacyKey)
    }
  } catch {
    // Storage may be unavailable in privacy mode. Migration is best-effort.
  }
}

function migrateExactKey(storage: Storage, legacyKey: string, canonicalKey: string): void {
  try {
    if (storage.getItem(canonicalKey) === null) {
      const value = storage.getItem(legacyKey)
      if (value !== null) storage.setItem(canonicalKey, value)
    }
    storage.removeItem(legacyKey)
  } catch {
    // Best-effort compatibility migration.
  }
}

export function migrateLegacyBrowserStorage(): void {
  migrateStorage(globalThis.localStorage)
  migrateStorage(globalThis.sessionStorage)
  migrateExactKey(
    globalThis.localStorage,
    'sub2api_login_agreement_consent',
    'easysub2api_login_agreement_consent'
  )
}
