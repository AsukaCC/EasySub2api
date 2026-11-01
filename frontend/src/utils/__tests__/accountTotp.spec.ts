import { describe, expect, it } from 'vitest'
import { Secret } from 'otpauth'
import { parseAccountTotp } from '../accountTotp'

describe('account TOTP', () => {
  it.each([
    ['SHA1', '12345678901234567890', '94287082'],
    ['SHA256', '12345678901234567890123456789012', '46119246'],
    ['SHA512', '1234567890123456789012345678901234567890123456789012345678901234', '90693936']
  ])('matches RFC 6238 %s vectors', (algorithm, secret, expected) => {
    const encoded = Secret.fromUTF8(secret).base32
    const generator = parseAccountTotp(`otpauth://totp/RFC?secret=${encoded}&algorithm=${algorithm}&digits=8`)
    expect(generator.generate({ timestamp: 59_000 })).toBe(expected)
  })

  it('accepts grouped Base32 and the default six-digit period', () => {
    const generator = parseAccountTotp('jbsw y3dp-ehpk 3pxp')
    expect(generator.secret.base32).toBe('JBSWY3DPEHPK3PXP')
    expect(generator.digits).toBe(6)
    expect(generator.period).toBe(30)
  })

  it.each(['A', 'MZ', 'invalid!', 'otpauth://hotp/Test?secret=MY',
    'otpauth://totp/Test?secret=MY&algorithm=MD5', 'otpauth://totp/Test?secret=MY&period=0',
    'otpauth://totp/Test?secret=MY&period=301', 'otpauth://totp/Test?secret=MY&digits=7',
    'otpauth://totp/Test?secret=MY&secret=MY'])('rejects invalid configuration %s', value => {
    expect(() => parseAccountTotp(value)).toThrow()
  })
})
