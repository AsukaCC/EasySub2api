import { flushPromises, mount } from '@vue/test-utils'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import AccountLoginFields from '../AccountLoginFields.vue'
import type { Account } from '@/types'

const { getLogin, copy } = vi.hoisted(() => ({ getLogin: vi.fn(), copy: vi.fn() }))
vi.mock('@/api/admin/accountLogin', () => ({ getAccountLoginCredentials: getLogin }))
vi.mock('@/composables/useClipboard', () => ({ useClipboard: () => ({ copyToClipboard: copy }) }))
vi.mock('vue-i18n', () => ({ useI18n: () => ({ t: (key: string) => key }) }))
const account = (id: string) => ({ id, credentials_status: { has_login_password: true, has_login_totp_secret: true } }) as Account
const mounts: ReturnType<typeof mount>[] = []
function mountFields(id = 'one') {
  const wrapper = mount(AccountLoginFields, { props: { account: account(id) }, global: { stubs: { Icon: true } } })
  mounts.push(wrapper)
  return wrapper
}

describe('account login details', () => {
  beforeEach(() => {
    vi.useFakeTimers()
    vi.setSystemTime(59_000)
    vi.spyOn(document, 'hidden', 'get').mockReturnValue(false)
    getLogin.mockReset().mockResolvedValue({ login_username: 'person@example.com', login_password: 'saved-password', login_totp_secret: 'GEZDGNBVGY3TQOJQGEZDGNBVGY3TQOJQ' })
    copy.mockReset()
  })
  afterEach(() => { mounts.splice(0).forEach(wrapper => wrapper.unmount()); vi.useRealTimers(); vi.restoreAllMocks() })

  it('reads once, masks secrets, and computes codes without backend polling', async () => {
    const wrapper = mountFields()
    await flushPromises()
    expect(wrapper.get('input[name="login_password"]').attributes('type')).toBe('password')
    expect(wrapper.get('input[name="login_totp_secret"]').attributes('type')).toBe('password')
    expect(wrapper.get('output').text()).toBe('287082')
    await vi.advanceTimersByTimeAsync(60_000)
    expect(getLogin).toHaveBeenCalledTimes(1)
    await wrapper.get('.account-login-code button').trigger('click')
    expect(copy).toHaveBeenCalledWith(wrapper.get('output').text())
    expect(wrapper.emitted('update:modelValue')).toBeUndefined()
  })

  it('sends only edited fields and permits explicitly clearing an existing secret', async () => {
    const wrapper = mountFields()
    await flushPromises()
    await wrapper.get('input[name="login_totp_secret"]').setValue('')
    expect(wrapper.emitted('update:modelValue')?.[0]).toEqual([{ login_totp_secret: '' }])
    await wrapper.setProps({ modelValue: { login_totp_secret: '' } })
    expect(wrapper.find('output').exists()).toBe(false)
    expect(vi.getTimerCount()).toBe(0)
  })

  it('discards late responses when switching accounts and clears data when closed', async () => {
    let resolveOld!: (value: unknown) => void
    getLogin.mockImplementationOnce(() => new Promise(resolve => { resolveOld = resolve }))
    const wrapper = mountFields()
    await wrapper.setProps({ account: account('two') })
    await flushPromises()
    resolveOld({ login_password: 'wrong-account-password' })
    await flushPromises()
    expect((wrapper.get('input[name="login_password"]').element as HTMLInputElement).value).toBe('saved-password')
    await wrapper.setProps({ active: false })
    expect((wrapper.get('input[name="login_password"]').element as HTMLInputElement).value).toBe('')
    expect(vi.getTimerCount()).toBe(0)
  })

  it('shows a read error and waits for an explicit retry', async () => {
    getLogin.mockRejectedValueOnce(new Error('offline'))
    const wrapper = mountFields()
    await flushPromises()
    expect(wrapper.text()).toContain('admin.accounts.loginInfo.loadFailed')
    await vi.advanceTimersByTimeAsync(60_000)
    expect(getLogin).toHaveBeenCalledTimes(1)
    await wrapper.get('[role="alert"] button').trigger('click')
    await flushPromises()
    expect(getLogin).toHaveBeenCalledTimes(2)
    expect(wrapper.get('output').text()).not.toBe('')
  })
})
