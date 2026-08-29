import { readdirSync, readFileSync } from 'node:fs'
import { dirname, relative, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import { describe, expect, it } from 'vitest'

const srcRoot = resolve(dirname(fileURLToPath(import.meta.url)), '../..')

interface SourceFile {
  path: string
  source: string
}

function collectSourceFiles(directory: string): SourceFile[] {
  return readdirSync(directory, { withFileTypes: true }).flatMap((entry) => {
    const absolutePath = resolve(directory, entry.name)
    if (entry.isDirectory()) {
      if (entry.name === '__tests__') return []
      return collectSourceFiles(absolutePath)
    }
    if (!entry.name.endsWith('.vue') && !entry.name.endsWith('.ts')) return []
    if (entry.name.endsWith('.d.ts')) return []
    return [{
      path: relative(srcRoot, absolutePath).replaceAll('\\', '/'),
      source: readFileSync(absolutePath, 'utf8'),
    }]
  })
}

const sourceFiles = collectSourceFiles(srcRoot)

// Every hard-coded dollar display must identify a genuine USD field or an upstream-account UI.
// CNY payment views deliberately use formatCNY and therefore do not need a dollar exception.
const hardcodedDollarAllowlist: Record<string, RegExp[]> = {
  'components/account/AccountStatsModal.vue': [
    /stats\.summary\.(?:total_cost|total_standard_cost|avg_daily_cost|(?:today|highest_cost_day|highest_request_day)\?\.cost)/,
  ],
  'components/admin/account/AccountStatsModal.vue': [
    /stats\.summary\.(?:total_cost|total_standard_cost|avg_daily_cost|(?:today|highest_cost_day|highest_request_day)\?\.cost)/,
  ],
  'components/admin/usage/UsageTable.vue': [
    /accountBilled\(/,
    /tooltipData[^}]*\b(?:input_cost|image_input_cost|output_cost|image_output_cost|total_cost|cache_creation_cost|cache_read_cost)\b/,
    /imageUnitPrice\(tooltipData\)/,
  ],
  'components/admin/usage/UsageStatsCards.vue': [
    /totalAccountCost/,
    /stats\?\.total_cost/,
  ],
  'components/account/AccountUsageCell.vue': [
    /grokPrepaidMoneyLine\.prepaid/,
    /formatKeyCost/,
  ],
  'components/account/QuotaBadge.vue': [
    /fmt\((?:used|limit)\)/,
  ],
  'components/account/UsageProgressBar.vue': [
    /formatAccountCost/,
  ],
  'components/charts/UserBreakdownSubTable.vue': [
    /user\.(?:account_cost|cost)/,
  ],
  'components/charts/ModelDistributionChart.vue': [
    /model\.(?:account_cost|cost)/,
  ],
  'components/charts/GroupDistributionChart.vue': [
    /group\.(?:account_cost|cost)/,
  ],
  'components/charts/EndpointDistributionChart.vue': [
    /item\.cost/,
  ],
  'views/admin/DashboardView.vue': [
    /stats\.(?:today_account_cost|today_cost|total_account_cost|total_cost)/,
  ],
  'views/admin/GroupsView.vue': [
    /price\.toFixed/,
  ],
  'utils/usagePricing.ts': [
    /formatted/,
  ],
  'utils/pricing.ts': [
    /\bs\b/,
  ],
  'components/charts/TokenUsageTrend.vue': [
    /formatCost\(data\.cost\)/,
  ],
  'views/HomeView.vue': [
    /\$/,
  ],
  'components/admin/channel/PricingEntryCard.vue': [
    /\$/,
  ],
  'components/admin/channel/IntervalRow.vue': [
    /\$/,
  ],
  'components/account/CreateAccountModal.vue': [
    /\$/,
  ],
  'components/account/EditAccountModal.vue': [
    /\$/,
  ],
  'components/account/QuotaDimensionRow.vue': [
    /\$/,
  ],
  'components/account/QuotaNotifyToggle.vue': [
    /\$/,
  ],
}

const pointFieldPattern = /\b(?:actual_cost|user_cost|total_actual_cost|today_actual_cost|balance|available_balance|recharge_balance|bonus_balance|affiliate_balance|quota|quota_used|rebate_amount|commission_amount|daily_(?:usage|limit)_points|weekly_(?:usage|limit)_points|monthly_(?:usage|limit)_points)\b/

function lineNumber(source: string, index: number): number {
  return source.slice(0, index).split(/\r?\n/).length
}

function allowedDollarDisplay(path: string, snippet: string): boolean {
  return (hardcodedDollarAllowlist[path] ?? []).some((pattern) => pattern.test(snippet))
}

describe('platform-point display boundaries', () => {
  it('allows hard-coded dollar signs only for explicit USD and upstream-account displays', () => {
    const violations: string[] = []
    const displayPatterns = [
      /\$\{\{[\s\S]*?\}\}/g,
      /\$\$\{[^}\n]*\}/g,
      />\s*\$\s*</g,
    ]

    for (const file of sourceFiles) {
      for (const displayPattern of displayPatterns) {
        for (const match of file.source.matchAll(displayPattern)) {
          const snippet = match[0]
          if (!allowedDollarDisplay(file.path, snippet)) {
            violations.push(`${file.path}:${lineNumber(file.source, match.index ?? 0)} ${snippet.replace(/\s+/g, ' ')}`)
          }
        }
      }
    }

    expect(violations, violations.join('\n')).toEqual([])
  })

  it('does not pass point-domain fields to money formatters', () => {
    const violations: string[] = []
    const moneyFormatterCall = /\b(?:formatCurrency|formatUSD|formatCNY|formatPaymentAmount)\(([^)\n]*)\)/g

    for (const file of sourceFiles) {
      for (const match of file.source.matchAll(moneyFormatterCall)) {
        if (pointFieldPattern.test(match[1])) {
          violations.push(`${file.path}:${lineNumber(file.source, match.index ?? 0)} ${match[0]}`)
        }
      }
    }

    expect(violations, violations.join('\n')).toEqual([])
  })

  it('keeps account history USD and point series on explicit fields', () => {
    for (const path of [
      'components/account/AccountStatsModal.vue',
      'components/admin/account/AccountStatsModal.vue',
    ]) {
      const source = sourceFiles.find((file) => file.path === path)?.source ?? ''
      expect(source).toMatch(/data: stats\.value\.history\.map\(\(h\) => h\.account_cost\),\s+unit: 'usd'/)
      expect(source).toMatch(/data: stats\.value\.history\.map\(\(h\) => h\.actual_cost\),\s+unit: 'points'/)
      expect(source).not.toMatch(/\$[^\n]*h\.(?:actual_cost|user_cost)/)
    }
  })

  it('describes multiplier output as platform points rather than dollars', () => {
    const zh = sourceFiles.find((file) => file.path === 'i18n/locales/zh/misc.ts')?.source ?? ''
    const en = sourceFiles.find((file) => file.path === 'i18n/locales/en/misc.ts')?.source ?? ''

    expect(zh).toContain('标准成本 $1，扣除 1.5 平台积分')
    expect(zh).not.toMatch(/扣除\s*\$\d/)
    expect(en).toContain('Standard cost $1, charge 1.5 platform points')
    expect(en).not.toMatch(/charged?\s*\$\d/i)
  })

  it('treats distribution actual_cost as points without an account-cost fallback', () => {
    for (const path of [
      'components/charts/EndpointDistributionChart.vue',
      'components/charts/ModelDistributionChart.vue',
      'components/charts/GroupDistributionChart.vue',
      'components/charts/UserBreakdownSubTable.vue',
    ]) {
      const source = sourceFiles.find((file) => file.path === path)?.source ?? ''
      expect(source).toMatch(/formatPoints\([^)]*actual_cost\)/)
      expect(source).not.toMatch(/account_cost\s*\?\?\s*[^\n]*actual_cost/)
      expect(source).not.toMatch(/\$\{\{[^}]*actual_cost/)
    }
  })

  it('never falls back from point-denominated actual_cost to USD cost', () => {
    const keyUsage = sourceFiles.find((file) => file.path === 'views/KeyUsageView.vue')?.source ?? ''

    expect(keyUsage).not.toMatch(/actual_cost\s*!=\s*null\s*\?\s*[^:]+\s*:\s*[^\n]*\.cost/)
    expect(keyUsage).not.toMatch(/actual_cost\s*\?\?\s*[^\n]*\.cost/)
  })

  it('uses point-named fields for new admin quota writes', () => {
    const groups = sourceFiles.find((file) => file.path === 'views/admin/GroupsView.vue')?.source ?? ''
    const userQuotas = sourceFiles.find((file) => file.path === 'components/admin/user/UserPlatformQuotaModal.vue')?.source ?? ''

    for (const window of ['daily', 'weekly', 'monthly']) {
      expect(groups).toContain(`${window}_limit_points: normalizeOptionalLimit(`)
      expect(groups).not.toContain(`${window}_limit_usd: normalizeOptionalLimit(`)
      expect(userQuotas).toContain(`${window}_limit_points: normalizeLimit(`)
      expect(userQuotas).not.toContain(`${window}_limit_usd: normalizeLimit(`)
    }
  })
})
