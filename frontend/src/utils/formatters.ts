/**
 * 格式化缓存 token 数量（1K/1M 缩写）
 */
export function formatCacheTokens(tokens: number): string {
  if (tokens >= 1000000) return `${(tokens / 1000000).toFixed(1)}M`
  if (tokens >= 1000) return `${(tokens / 1000).toFixed(1)}K`
  return tokens.toLocaleString()
}

/**
 * 自适应精度格式化倍率：保留至多 4 位小数并去掉末尾多余的 0，
 * 但至少保留 2 位小数（0.035 -> "0.035"，0.3 -> "0.30"，1 -> "1.00"）
 */
export function formatMultiplier(val: number): string {
  if (val < 0.0001) return val.toPrecision(2)
  return val.toFixed(4).replace(/(\.\d{2}\d*?)0+$/, '$1')
}

/** 计算分组倍率与用户专属分组倍率的乘积；未配置用户倍率按 1。 */
export function combinedGroupUserRate(
  rateMultiplier: number | null | undefined,
  userRateMultiplier: number | null | undefined
): number | null {
  if (typeof rateMultiplier !== 'number' || !Number.isFinite(rateMultiplier) || rateMultiplier < 0) return null
  const userRate = userRateMultiplier == null ? 1 : userRateMultiplier
  if (typeof userRate !== 'number' || !Number.isFinite(userRate) || userRate < 0) return null
  return rateMultiplier * userRate
}
