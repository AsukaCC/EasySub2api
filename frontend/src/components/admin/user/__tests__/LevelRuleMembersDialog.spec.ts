import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import LevelRuleMembersDialog from '../LevelRuleMembersDialog.vue'
import type { UserLevelRule } from '@/api/admin/users'

const api = vi.hoisted(() => ({ getLevelRuleMembers: vi.fn(), list: vi.fn(), batchAssignLevelRules: vi.fn() }))
vi.mock('@/api/admin', () => ({ adminAPI: { users: api } }))
vi.mock('vue-i18n', () => ({ useI18n: () => ({ t: (key: string) => key }) }))
const rule = { id: 'r1', name: 'VIP', enabled: true, is_default: false } as UserLevelRule
const rules = [rule, { id: 'r2', name: 'Default', enabled: true, is_default: true } as UserLevelRule]
const make = () => mount(LevelRuleMembersDialog, {
  props: { rule, rules },
  global: { stubs: { BaseDialog: { template: '<div><slot /></div>' }, Pagination: true, Icon: true } }
})
describe('LevelRuleMembersDialog', () => {
  beforeEach(() => {
    vi.resetAllMocks()
    api.getLevelRuleMembers.mockResolvedValue({ items: [{ id: 'u1', email: 'one@example.test', username: 'One' }], total: 1 })
    api.list.mockResolvedValue({ items: [{ id: 'u2', email: 'two@example.test' }], total: 1 })
    api.batchAssignLevelRules.mockResolvedValue({ affected: 1 })
    vi.spyOn(window, 'confirm').mockReturnValue(true)
  })
  it('searches members and atomically reassigns a selected user', async () => {
    const wrapper = make()
    await flushPromises()
    await wrapper.get('form input').setValue('one')
    await wrapper.get('form').trigger('submit')
    await flushPromises()
    expect(api.getLevelRuleMembers).toHaveBeenLastCalledWith('r1', 1, 20, 'one')
    await wrapper.get('select').setValue('r2')
    await wrapper.get('tbody button').trigger('click')
    await flushPromises()
    expect(api.batchAssignLevelRules).toHaveBeenCalledWith({ user_ids: ['u1'], rule_ids: ['r2'], operation: 'replace' })
    expect(wrapper.emitted('changed')).toHaveLength(1)
  })
  it('adds users by replacement only after confirmation', async () => {
    const wrapper = make()
    await flushPromises()
    await wrapper.findAll('[role=tab]')[1].trigger('click')
    await flushPromises()
    vi.mocked(window.confirm).mockReturnValueOnce(false)
    await wrapper.get('tbody button').trigger('click')
    expect(api.batchAssignLevelRules).not.toHaveBeenCalled()
    await wrapper.get('tbody button').trigger('click')
    await flushPromises()
    expect(api.batchAssignLevelRules).toHaveBeenCalledWith({ user_ids: ['u2'], rule_ids: ['r1'], operation: 'replace' })
  })
  it('resets selected members to the server default', async () => {
    const wrapper = make()
    await flushPromises()
    await wrapper.get('tbody input[type=checkbox]').setValue(true)
    const reset = wrapper.findAll('button').find(b => b.text() === 'admin.users.levels.resetDefault')!
    await reset.trigger('click')
    await flushPromises()
    expect(api.batchAssignLevelRules).toHaveBeenCalledWith({ user_ids: ['u1'], rule_ids: [], operation: 'replace' })
  })
})
