import { createI18n } from 'vue-i18n'

type LocaleCode = 'en' | 'zh'
type LocaleNamespace = 'core' | 'app' | 'admin' | 'monitor'
type LocaleMessages = Record<string, any>

const LOCALE_KEY = 'easysub2api_locale'
const LEGACY_LOCALE_KEY = 'sub2api_locale'
const DEFAULT_LOCALE: LocaleCode = 'en'

const namespaceLoaders: Record<
  LocaleCode,
  Record<LocaleNamespace, () => Promise<{ default: LocaleMessages }>>
> = {
  en: {
    core: () => import('./locales/en/core'),
    app: () => import('./locales/en/app'),
    admin: () => import('./locales/en/admin'),
    monitor: () => import('./locales/en/channelMonitorV2'),
  },
  zh: {
    core: () => import('./locales/zh/core'),
    app: () => import('./locales/zh/app'),
    admin: () => import('./locales/zh/admin'),
    monitor: () => import('./locales/zh/channelMonitorV2'),
  },
}

const CORE_ONLY_PATHS = new Set([
  '/',
  '/home',
  '/login',
  '/register',
  '/forgot-password',
  '/reset-password',
  '/email-verify',
  '/setup',
])

const CORE_ONLY_PREFIXES = ['/oauth', '/wechat', '/dingtalk', '/oidc']

function isLocaleCode(value: string): value is LocaleCode {
  return value === 'en' || value === 'zh'
}

function getDefaultLocale(): LocaleCode {
  const saved = localStorage.getItem(LOCALE_KEY) ?? localStorage.getItem(LEGACY_LOCALE_KEY)
  if (saved && isLocaleCode(saved)) {
    localStorage.setItem(LOCALE_KEY, saved)
    localStorage.removeItem(LEGACY_LOCALE_KEY)
    return saved
  }

  const browserLang = navigator.language.toLowerCase()
  if (browserLang.startsWith('zh')) {
    return 'zh'
  }

  return DEFAULT_LOCALE
}

export const i18n = createI18n({
  legacy: false,
  locale: getDefaultLocale(),
  fallbackLocale: DEFAULT_LOCALE,
  messages: {},
  // 禁用 HTML 消息警告 - 引导步骤使用富文本内容（driver.js 支持 HTML）
  // 这些内容是内部定义的，不存在 XSS 风险
  warnHtmlMessage: false
})

const loadedNamespaces = new Map<LocaleCode, Set<LocaleNamespace>>()

function namespacesForPath(path: string): LocaleNamespace[] {
  const namespaces = new Set<LocaleNamespace>(['core'])
  const isAdmin = path.startsWith('/admin')
  const isMonitor = path.includes('monitor') || path.includes('channel-status')
  const isCoreOnly =
    CORE_ONLY_PATHS.has(path) ||
    CORE_ONLY_PREFIXES.some((prefix) => path === prefix || path.startsWith(`${prefix}/`))

  if (isAdmin) {
    namespaces.add('app')
    namespaces.add('admin')
  }
  if (isMonitor) {
    namespaces.add('app')
    namespaces.add('monitor')
  }
  if (!isCoreOnly && !isAdmin && !isMonitor) {
    namespaces.add('app')
  }

  return [...namespaces]
}

async function loadNamespace(locale: LocaleCode, namespace: LocaleNamespace): Promise<void> {
  const loaded = loadedNamespaces.get(locale) ?? new Set<LocaleNamespace>()
  if (loaded.has(namespace)) {
    return
  }

  const module = await namespaceLoaders[locale][namespace]()
  const messages = namespace === 'admin' ? { admin: module.default } : module.default
  i18n.global.mergeLocaleMessage(locale, messages)
  loaded.add(namespace)
  loadedNamespaces.set(locale, loaded)
}

export async function ensureRouteMessages(path: string, locale: LocaleCode = getLocale()): Promise<void> {
  await Promise.all(namespacesForPath(path).map((namespace) => loadNamespace(locale, namespace)))
}

export async function loadLocaleMessages(locale: LocaleCode): Promise<void> {
  await loadNamespace(locale, 'core')
}

export async function initI18n(): Promise<void> {
  const current = getLocale()
  await loadNamespace(current, 'core')
  document.documentElement.setAttribute('lang', current)
}

export async function setLocale(locale: string): Promise<void> {
  if (!isLocaleCode(locale)) {
    return
  }

  const previous = loadedNamespaces.get(getLocale())
  const needed = new Set<LocaleNamespace>(['core'])
  previous?.forEach((namespace) => needed.add(namespace))

  const { default: router } = await import('@/router')
  namespacesForPath(router.currentRoute.value.path).forEach((namespace) => needed.add(namespace))

  await Promise.all([...needed].map((namespace) => loadNamespace(locale, namespace)))
  i18n.global.locale.value = locale
  localStorage.setItem(LOCALE_KEY, locale)
  document.documentElement.setAttribute('lang', locale)

  // 同步更新浏览器页签标题，使其跟随语言切换
  const { resolveRouteDocumentTitle } = await import('@/router/title')
  const { useAppStore } = await import('@/stores/app')
  const { useAuthStore } = await import('@/stores/auth')
  const { useAdminSettingsStore } = await import('@/stores/adminSettings')
  const route = router.currentRoute.value
  const appStore = useAppStore()
  const authStore = useAuthStore()
  const adminSettingsStore = useAdminSettingsStore()
  const customMenuItems = [
    ...(appStore.cachedPublicSettings?.custom_menu_items ?? []),
    ...(authStore.isAdmin ? adminSettingsStore.customMenuItems : []),
  ]
  document.title = resolveRouteDocumentTitle(route, appStore.siteName, customMenuItems)
}

export function getLocale(): LocaleCode {
  const current = i18n.global.locale.value
  return isLocaleCode(current) ? current : DEFAULT_LOCALE
}

export const availableLocales = [
  { code: 'en', name: 'English', flag: '🇺🇸' },
  { code: 'zh', name: '中文', flag: '🇨🇳' }
] as const

export default i18n
