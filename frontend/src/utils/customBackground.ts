/**
 * 自定义页面背景图(仅存储在当前浏览器 localStorage,不上传服务器)。
 * 图片经 canvas 压缩为 JPEG 后以 data URL 保存,通过 CSS 变量 + html 类名生效,
 * 玻璃组件的 backdrop-filter 会直接对背景图产生模糊折射效果。
 */

const STORAGE_KEY = 'custom_background_image'
const HTML_CLASS = 'has-custom-bg'
const MAX_DIMENSION = 2560
const JPEG_QUALITY = 0.82
// localStorage 常见上限 5MB,预留其他键的空间
const MAX_DATA_URL_LENGTH = 4 * 1024 * 1024

function applyToDocument(dataUrl: string | null): void {
  const root = document.documentElement
  if (dataUrl) {
    root.style.setProperty('--custom-bg-image', `url("${dataUrl}")`)
    root.classList.add(HTML_CLASS)
  } else {
    root.style.removeProperty('--custom-bg-image')
    root.classList.remove(HTML_CLASS)
  }
}

/** 启动时恢复已保存的背景图(在 app mount 前调用,避免闪烁)。 */
export function initCustomBackground(): void {
  try {
    const stored = localStorage.getItem(STORAGE_KEY)
    if (stored?.startsWith('data:image/')) {
      applyToDocument(stored)
    }
  } catch {
    // localStorage 不可用时静默跳过
  }
}

export function hasCustomBackground(): boolean {
  return document.documentElement.classList.contains(HTML_CLASS)
}

async function compressImage(file: File, maxDimension: number, quality: number): Promise<string> {
  const objectUrl = URL.createObjectURL(file)
  try {
    const image = await new Promise<HTMLImageElement>((resolve, reject) => {
      const img = new Image()
      img.onload = () => resolve(img)
      img.onerror = () => reject(new Error('image load failed'))
      img.src = objectUrl
    })

    const scale = Math.min(1, maxDimension / Math.max(image.width, image.height))
    const width = Math.max(1, Math.round(image.width * scale))
    const height = Math.max(1, Math.round(image.height * scale))

    const canvas = document.createElement('canvas')
    canvas.width = width
    canvas.height = height
    const ctx = canvas.getContext('2d')
    if (!ctx) throw new Error('canvas unavailable')
    ctx.drawImage(image, 0, 0, width, height)
    return canvas.toDataURL('image/jpeg', quality)
  } finally {
    URL.revokeObjectURL(objectUrl)
  }
}

/**
 * 压缩并保存背景图。压缩后仍超出 localStorage 限制时会自动降档重试,
 * 最终仍失败则抛出错误,由调用方提示用户。
 */
export async function setCustomBackground(file: File): Promise<void> {
  const attempts: Array<{ maxDimension: number; quality: number }> = [
    { maxDimension: MAX_DIMENSION, quality: JPEG_QUALITY },
    { maxDimension: 1920, quality: 0.75 },
    { maxDimension: 1280, quality: 0.7 }
  ]

  let lastError: unknown = null
  for (const { maxDimension, quality } of attempts) {
    try {
      const dataUrl = await compressImage(file, maxDimension, quality)
      if (dataUrl.length > MAX_DATA_URL_LENGTH) {
        lastError = new Error('image too large')
        continue
      }
      localStorage.setItem(STORAGE_KEY, dataUrl)
      applyToDocument(dataUrl)
      return
    } catch (error) {
      lastError = error
    }
  }
  throw lastError instanceof Error ? lastError : new Error('failed to save background')
}

export function clearCustomBackground(): void {
  try {
    localStorage.removeItem(STORAGE_KEY)
  } catch {
    // ignore
  }
  applyToDocument(null)
}
