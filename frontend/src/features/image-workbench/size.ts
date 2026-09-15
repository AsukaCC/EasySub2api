const SIZE_PATTERN = /^\s*(\d+)\s*[xX×]\s*(\d+)\s*$/
const RATIO_PATTERN = /^\s*(\d+(?:\.\d+)?)\s*[:xX×]\s*(\d+(?:\.\d+)?)\s*$/
const SIZE_MULTIPLE = 16
const MAX_EDGE = 3840
const MAX_ASPECT_RATIO = 3
const MIN_PIXELS = 655_360
const MAX_PIXELS = 8_294_400
const MAX_1K_PIXELS = 1_572_864
const MAX_RATIO_ERROR = 0.01

export type SizeTier = '1K' | '2K' | '4K'
export type SizeMode = 'auto' | 'ratio' | 'resolution'
type PresetRatio = '1:1' | '3:2' | '2:3' | '16:9' | '9:16' | '4:3' | '3:4' | '21:9'

export const SIZE_TIERS: SizeTier[] = ['1K', '2K', '4K']
export const RATIO_PRESETS: Array<{ label: string; value: PresetRatio }> = [
  { label: '1:1', value: '1:1' },
  { label: '3:2', value: '3:2' },
  { label: '2:3', value: '2:3' },
  { label: '16:9', value: '16:9' },
  { label: '9:16', value: '9:16' },
  { label: '4:3', value: '4:3' },
  { label: '3:4', value: '3:4' },
  { label: '21:9', value: '21:9' },
]

const PRESET_RATIO_SET = new Set<string>(RATIO_PRESETS.map((item) => item.value))

const TIER_PIXEL_BUDGET: Record<SizeTier, number> = {
  '1K': MAX_1K_PIXELS,
  '2K': 4_194_304,
  '4K': MAX_PIXELS,
}

// Longest-edge limits must match backend ClassifyImageBillingTier.
const TIER_MAX_EDGE: Record<SizeTier, number> = {
  '1K': 1024,
  '2K': 2048,
  '4K': MAX_EDGE,
}

const COMMON_SIZE_PRESETS: Record<SizeTier, Partial<Record<PresetRatio, string>>> = {
  '1K': {
    '1:1': '1024x1024',
    '3:2': '1024x688',
    '2:3': '688x1024',
    '4:3': '1024x768',
    '3:4': '768x1024',
  },
  '2K': {
    '1:1': '2048x2048',
    '3:2': '2048x1360',
    '2:3': '1360x2048',
    '16:9': '2048x1152',
    '9:16': '1152x2048',
    '4:3': '2048x1536',
    '3:4': '1536x2048',
    '21:9': '2048x880',
  },
  '4K': {
    '1:1': '2880x2880',
    '3:2': '3456x2304',
    '2:3': '2304x3456',
    '16:9': '3840x2160',
    '9:16': '2160x3840',
    '4:3': '3200x2400',
    '3:4': '2400x3200',
    '21:9': '3840x1600',
  },
}

function roundToMultiple(value: number, multiple: number) {
  return Math.max(multiple, Math.round(value / multiple) * multiple)
}

function floorToMultiple(value: number, multiple: number) {
  return Math.max(multiple, Math.floor(value / multiple) * multiple)
}

function ceilToMultiple(value: number, multiple: number) {
  return Math.max(multiple, Math.ceil(value / multiple) * multiple)
}

function gcd(a: number, b: number): number {
  return b === 0 ? a : gcd(b, a % b)
}

function normalizeDimensions(width: number, height: number) {
  let normalizedWidth = roundToMultiple(width, SIZE_MULTIPLE)
  let normalizedHeight = roundToMultiple(height, SIZE_MULTIPLE)

  const scaleToFit = (scale: number) => {
    normalizedWidth = floorToMultiple(normalizedWidth * scale, SIZE_MULTIPLE)
    normalizedHeight = floorToMultiple(normalizedHeight * scale, SIZE_MULTIPLE)
  }

  const scaleToFill = (scale: number) => {
    normalizedWidth = ceilToMultiple(normalizedWidth * scale, SIZE_MULTIPLE)
    normalizedHeight = ceilToMultiple(normalizedHeight * scale, SIZE_MULTIPLE)
  }

  for (let i = 0; i < 4; i++) {
    const maxEdge = Math.max(normalizedWidth, normalizedHeight)
    if (maxEdge > MAX_EDGE) scaleToFit(MAX_EDGE / maxEdge)

    if (normalizedWidth / normalizedHeight > MAX_ASPECT_RATIO) {
      normalizedWidth = floorToMultiple(normalizedHeight * MAX_ASPECT_RATIO, SIZE_MULTIPLE)
    } else if (normalizedHeight / normalizedWidth > MAX_ASPECT_RATIO) {
      normalizedHeight = floorToMultiple(normalizedWidth * MAX_ASPECT_RATIO, SIZE_MULTIPLE)
    }

    const pixels = normalizedWidth * normalizedHeight
    if (pixels > MAX_PIXELS) scaleToFit(Math.sqrt(MAX_PIXELS / pixels))
    else if (pixels < MIN_PIXELS) scaleToFill(Math.sqrt(MIN_PIXELS / pixels))
  }

  return { width: normalizedWidth, height: normalizedHeight }
}

function getPresetRatioKey(ratioWidth: number, ratioHeight: number): PresetRatio | null {
  if (!Number.isInteger(ratioWidth) || !Number.isInteger(ratioHeight)) return null
  const divisor = gcd(ratioWidth, ratioHeight)
  const key = `${ratioWidth / divisor}:${ratioHeight / divisor}`
  return PRESET_RATIO_SET.has(key) ? key as PresetRatio : null
}

function searchRatioSize(
  ratioWidth: number,
  ratioHeight: number,
  maxEdge: number,
  pixelBudget: number,
  preferLargest: boolean,
) {
  const targetRatio = ratioWidth / ratioHeight
  let bestWidth = 0
  let bestHeight = 0
  let bestPixels = preferLargest ? 0 : Number.POSITIVE_INFINITY

  for (let w = SIZE_MULTIPLE; w <= maxEdge; w += SIZE_MULTIPLE) {
    const idealH = w / targetRatio
    const candidates = [
      Math.floor(idealH / SIZE_MULTIPLE) * SIZE_MULTIPLE,
      Math.ceil(idealH / SIZE_MULTIPLE) * SIZE_MULTIPLE,
    ]

    for (const h of candidates) {
      if (h < SIZE_MULTIPLE || h > maxEdge) continue
      const pixels = w * h
      if (pixels > pixelBudget || pixels < MIN_PIXELS) continue
      if (Math.max(w / h, h / w) > MAX_ASPECT_RATIO) continue
      const ratioError = Math.abs(w / h - targetRatio) / targetRatio
      if (ratioError > MAX_RATIO_ERROR) continue
      const better = preferLargest ? pixels > bestPixels : pixels < bestPixels
      if (better) {
        bestPixels = pixels
        bestWidth = w
        bestHeight = h
      }
    }
  }

  if (!Number.isFinite(bestPixels) || bestPixels === 0) return null
  return `${bestWidth}x${bestHeight}`
}

function calculateImageSizeForTier(tier: SizeTier, ratio: string) {
  const parsed = parseRatio(ratio)
  if (!parsed) return null

  const presetRatioKey = getPresetRatioKey(parsed.width, parsed.height)
  const preset = presetRatioKey ? COMMON_SIZE_PRESETS[tier][presetRatioKey] : undefined
  if (preset) return preset

  return searchRatioSize(
    parsed.width,
    parsed.height,
    TIER_MAX_EDGE[tier],
    TIER_PIXEL_BUDGET[tier],
    true,
  )
}

export function parseSize(size: string) {
  const match = size.match(SIZE_PATTERN)
  if (!match) return null
  return { width: Number(match[1]), height: Number(match[2]) }
}

export function parseRatio(ratio: string) {
  const match = ratio.match(RATIO_PATTERN)
  if (!match) return null
  const width = Number(match[1])
  const height = Number(match[2])
  if (!Number.isFinite(width) || !Number.isFinite(height) || width <= 0 || height <= 0) return null
  return { width, height }
}

export function normalizeImageSize(size: string) {
  const trimmed = size.trim()
  const match = trimmed.match(SIZE_PATTERN)
  if (!match) return trimmed
  const { width, height } = normalizeDimensions(Number(match[1]), Number(match[2]))
  return `${width}x${height}`
}

export function imageBillingTier(size: string): SizeTier {
  const trimmed = size.trim()
  if (!trimmed || trimmed.toLowerCase() === 'auto') return '2K'
  const named = trimmed.toUpperCase()
  if (named === '1K' || named === '2K' || named === '4K') return named
  const parsed = parseSize(trimmed)
  if (!parsed) return '2K'
  const maxEdge = Math.max(parsed.width, parsed.height)
  if (maxEdge <= TIER_MAX_EDGE['1K']) return '1K'
  if (maxEdge <= TIER_MAX_EDGE['2K']) return '2K'
  return '4K'
}

export function calculateImageSize(tier: SizeTier, ratio: string) {
  const exact = calculateImageSizeForTier(tier, ratio)
  if (exact) return exact

  const parsed = parseRatio(ratio)
  if (!parsed) return null
  return searchRatioSize(parsed.width, parsed.height, MAX_EDGE, MAX_PIXELS, false)
}

export function findPresetForSize(size: string) {
  const normalized = normalizeImageSize(size)
  for (const tier of SIZE_TIERS) {
    for (const ratio of RATIO_PRESETS) {
      if (calculateImageSizeForTier(tier, ratio.value) === normalized) {
        return { tier, ratio: ratio.value }
      }
    }
  }
  return null
}

export function ratioOrientation(value: string) {
  const [width, height] = value.split(':').map(Number)
  if (!width || !height) return 'custom'
  if (width === height) return 'square'
  return width > height ? 'landscape' : 'portrait'
}
