import { describe, it, expect, vi, beforeEach } from 'vitest'
import { flushPromises } from '@vue/test-utils'
import { mount } from '@vue/test-utils'
import { defineComponent } from 'vue'
import { useCrudList } from '../useCrudList'

// Mock useConfirm — auto-confirm deletes by default
const mockConfirm = vi.fn().mockResolvedValue(true)
vi.mock('../useConfirm', () => ({
  useConfirm: () => ({ confirm: mockConfirm }),
}))

type Item = { id: string; name: string }
type Form = { name: string }

function mockApi() {
  return {
    list: vi.fn<() => Promise<Item[]>>().mockResolvedValue([
      { id: '1', name: 'Alpha' },
      { id: '2', name: 'Bravo' },
    ]),
    create: vi.fn<(f: Form) => Promise<Item>>().mockImplementation(
      (f) => Promise.resolve({ id: '3', name: f.name }),
    ),
    update: vi.fn<(id: string, f: Form) => Promise<Item>>().mockImplementation(
      (id, f) => Promise.resolve({ id, name: f.name }),
    ),
    remove: vi.fn<(id: string) => Promise<void>>().mockResolvedValue(undefined),
  }
}

const emptyForm = (): Form => ({ name: '' })
const toForm = (item: Item): Form => ({ name: item.name })

// Helper: mount a wrapper component so onMounted fires and we get reactive refs
function setup(api = mockApi()) {
  let result!: ReturnType<typeof useCrudList<Item, Form>>
  mount(defineComponent({
    setup() {
      result = useCrudList<Item, Form>(api, { emptyForm, toForm, confirmMessage: 'Delete?' })
      return {}
    },
    template: '<div />',
  }))
  return { result, api }
}

describe('useCrudList', () => {
  beforeEach(() => {
    mockConfirm.mockResolvedValue(true)
  })

  it('loads items on mount', async () => {
    const { result } = setup()
    await flushPromises()
    expect(result.items.value).toHaveLength(2)
    expect(result.items.value[0].name).toBe('Alpha')
  })

  it('sets error when load fails', async () => {
    const api = mockApi()
    api.list.mockRejectedValue(new Error('network error'))
    const { result } = setup(api)
    await flushPromises()
    expect(result.error.value).toBe('network error')
    expect(result.items.value).toHaveLength(0)
  })

  it('openCreate resets form and shows it', async () => {
    const { result } = setup()
    await flushPromises()
    result.openCreate()
    expect(result.showForm.value).toBe(true)
    expect(result.editingId.value).toBeNull()
    expect(result.form.value).toEqual({ name: '' })
  })

  it('openEdit populates form with item data', async () => {
    const { result } = setup()
    await flushPromises()
    result.openEdit(result.items.value[0])
    expect(result.showForm.value).toBe(true)
    expect(result.editingId.value).toBe('1')
    expect(result.form.value).toEqual({ name: 'Alpha' })
  })

  it('closeForm resets state', async () => {
    const { result } = setup()
    await flushPromises()
    result.openCreate()
    result.form.value.name = 'test'
    result.closeForm()
    expect(result.showForm.value).toBe(false)
    expect(result.editingId.value).toBeNull()
    expect(result.form.value).toEqual({ name: '' })
    expect(result.error.value).toBe('')
  })

  it('submit creates a new item', async () => {
    const { result, api } = setup()
    await flushPromises()
    result.openCreate()
    result.form.value.name = 'Charlie'
    await result.submit()
    expect(api.create).toHaveBeenCalledWith({ name: 'Charlie' })
    expect(result.items.value[0].name).toBe('Charlie')
    expect(result.items.value).toHaveLength(3)
    expect(result.showForm.value).toBe(false)
  })

  it('submit updates an existing item', async () => {
    const { result, api } = setup()
    await flushPromises()
    result.openEdit(result.items.value[0])
    result.form.value.name = 'Alpha Updated'
    await result.submit()
    expect(api.update).toHaveBeenCalledWith('1', { name: 'Alpha Updated' })
    expect(result.items.value[0].name).toBe('Alpha Updated')
    expect(result.items.value).toHaveLength(2)
    expect(result.showForm.value).toBe(false)
  })

  it('submit sets error on failure', async () => {
    const { result, api } = setup()
    await flushPromises()
    api.create.mockRejectedValue(new Error('validation failed'))
    result.openCreate()
    result.form.value.name = 'Bad'
    await result.submit()
    expect(result.error.value).toBe('validation failed')
    expect(result.showForm.value).toBe(true)
  })

  it('remove deletes an item after confirmation', async () => {
    const { result, api } = setup()
    await flushPromises()
    await result.remove('1')
    expect(mockConfirm).toHaveBeenCalledWith('Delete?')
    expect(api.remove).toHaveBeenCalledWith('1')
    expect(result.items.value).toHaveLength(1)
    expect(result.items.value[0].id).toBe('2')
  })

  it('remove does nothing when confirmation is denied', async () => {
    mockConfirm.mockResolvedValue(false)
    const { result, api } = setup()
    await flushPromises()
    await result.remove('1')
    expect(api.remove).not.toHaveBeenCalled()
    expect(result.items.value).toHaveLength(2)
  })

  it('remove sets error on failure', async () => {
    const { result, api } = setup()
    await flushPromises()
    api.remove.mockRejectedValue(new Error('forbidden'))
    await result.remove('1')
    expect(result.error.value).toBe('forbidden')
    expect(result.items.value).toHaveLength(2)
  })
})
