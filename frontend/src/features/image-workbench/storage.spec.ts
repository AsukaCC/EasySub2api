import { describe, expect, it } from 'vitest'
import { reactive } from 'vue'

import { toPersistedTask, type ImageTaskItem } from './storage'

describe('image workbench task persistence', () => {
  it('converts Vue reactive task records into cloneable plain data', () => {
    const task = reactive<ImageTaskItem>({
      taskId: 'imgtask_test',
      keyId: 'key_test',
      platform: 'openai',
      prompt: 'a lighthouse',
      model: 'gpt-image-1',
      params: { size: '1024x1024', n: 1 },
      status: 'processing',
      createdAt: 1,
      error: { type: 'api_error', message: 'temporary failure' },
    })

    const persisted = toPersistedTask(task)

    expect(persisted).toEqual(task)
    expect(Object.getPrototypeOf(persisted)).toBe(Object.prototype)
    expect(Object.getPrototypeOf(persisted.params)).toBe(Object.prototype)
    expect(structuredClone(persisted)).toEqual(persisted)
  })
})
