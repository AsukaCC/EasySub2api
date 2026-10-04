export function normalizePlanType(value?: string | null): string {
  return (value || '').trim().toLowerCase().replace(/[\s_-]+/g, '')
}

export function openAIPlanTypeKey(value?: string | null): string {
  const key = normalizePlanType(value)
  return key === 'chatgptpro' ? 'pro' : key
}

export const openAIPlanTypes = [
  'free', 'go', 'plus', 'prolite', 'pro', 'promax', 'team',
  'self_serve_business_usage_based', 'self_serve_business_prolite', 'business',
  'enterprise', 'ent26', 'enterprise_cbp_usage_based', 'enterprise_cbp_automation',
  'edu', 'edu_plus', 'edu_pro', 'unknown'
] as const

export function openAIPlanTypeLabel(value?: string | null): string {
  switch (openAIPlanTypeKey(value)) {
    case 'free': return 'Free'
    case 'go': return 'Go'
    case 'plus': return 'Plus'
    case 'prolite': return 'Pro 100'
    case 'pro': return 'Pro 200'
    case 'promax': return 'Pro 500'
    case 'team':
    case 'selfservebusinessusagebased': return 'Business'
    case 'business': return 'Enterprise'
    case 'selfservebusinessprolite': return 'Business Premium'
    case 'enterprisecbpautomation': return 'Enterprise (Automation)'
    case 'enterprise':
    case 'ent26':
    case 'enterprisecbpusagebased': return 'Enterprise'
    case 'edu': return 'Edu'
    case 'eduplus': return 'Edu Plus'
    case 'edupro': return 'Edu Pro'
    case 'unknown': return 'Unknown'
    default: return ''
  }
}
