export interface NavigationIndicatorGeometry {
  x: number
  y: number
  width: number
  height: number
}

type NavigationIndicatorKey = 'top-navigation' | 'admin-sidebar'

const indicatorGeometry = new Map<NavigationIndicatorKey, NavigationIndicatorGeometry>()

export function getNavigationIndicatorGeometry(
  key: NavigationIndicatorKey
): NavigationIndicatorGeometry | null {
  const geometry = indicatorGeometry.get(key)
  return geometry ? { ...geometry } : null
}

export function setNavigationIndicatorGeometry(
  key: NavigationIndicatorKey,
  geometry: NavigationIndicatorGeometry
): void {
  indicatorGeometry.set(key, { ...geometry })
}
