import { Secret, TOTP } from 'otpauth'

export function parseAccountTotp(value: string): TOTP {
  let secret = value.trim()
  let algorithm = 'SHA1'
  let digits = 6
  let period = 30
  if (secret.includes('://')) {
    const uri = new URL(secret)
    if (uri.protocol !== 'otpauth:' || uri.host !== 'totp' || uri.username || uri.password) throw new Error('Invalid TOTP URI')
    for (const key of ['secret', 'algorithm', 'digits', 'period']) {
      if (uri.searchParams.getAll(key).length > 1) throw new Error('Duplicate TOTP parameter')
    }
    secret = uri.searchParams.get('secret') || ''
    algorithm = uri.searchParams.get('algorithm') || algorithm
    const rawDigits = uri.searchParams.get('digits') || '6'
    const rawPeriod = uri.searchParams.get('period') || '30'
    if (!/^(6|8)$/.test(rawDigits) || !/^\d+$/.test(rawPeriod)) throw new Error('Invalid TOTP parameters')
    digits = Number(rawDigits)
    period = Number(rawPeriod)
  }
  if (!['SHA1', 'SHA256', 'SHA512'].includes(algorithm) || period < 1 || period > 300) throw new Error('Invalid TOTP parameters')
  secret = secret.replace(/[\s-]/g, '').replace(/=+$/, '').toUpperCase()
  if (!/^[A-Z2-7]+$/.test(secret)) throw new Error('Invalid TOTP secret')
  const decoded = Secret.fromBase32(secret)
  if (!decoded.buffer.byteLength || decoded.base32 !== secret) throw new Error('Invalid TOTP secret')
  return new TOTP({ secret: decoded, algorithm, digits, period })
}
