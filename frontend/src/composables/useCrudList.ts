import { type Ref, ref, onMounted } from 'vue'
import { useConfirm } from './useConfirm'

interface CrudApi<T, F> {
  list: () => Promise<T[]>
  create: (form: F) => Promise<T>
  update: (id: string, form: F) => Promise<T>
  remove: (id: string) => Promise<void>
}

export function useCrudList<T extends { id: string }, F>(api: CrudApi<T, F>, opts: {
  emptyForm: () => F
  toForm: (item: T) => F
  confirmMessage?: string
}) {
  const { confirm } = useConfirm()
  const items = ref([]) as Ref<T[]>
  const showForm = ref(false)
  const editingId = ref<string | null>(null)
  const error = ref('')
  const form = ref(opts.emptyForm()) as Ref<F>

  async function load() {
    try {
      items.value = await api.list() ?? []
    } catch (e: any) {
      error.value = e.message
    }
  }

  function openCreate() {
    editingId.value = null
    form.value = opts.emptyForm()
    showForm.value = true
  }

  function openEdit(item: T) {
    editingId.value = item.id
    form.value = opts.toForm(item)
    showForm.value = true
  }

  function closeForm() {
    showForm.value = false
    editingId.value = null
    form.value = opts.emptyForm()
    error.value = ''
  }

  async function submit() {
    try {
      if (editingId.value) {
        const updated = await api.update(editingId.value, form.value)
        const idx = items.value.findIndex(i => i.id === editingId.value)
        if (idx !== -1) items.value[idx] = updated
      } else {
        const created = await api.create(form.value)
        items.value.unshift(created)
      }
      closeForm()
    } catch (e: any) {
      error.value = e.message
    }
  }

  async function remove(id: string) {
    if (!await confirm(opts.confirmMessage ?? 'Delete this item?')) return
    try {
      await api.remove(id)
      items.value = items.value.filter(i => i.id !== id)
    } catch (e: any) {
      error.value = e.message
    }
  }

  onMounted(load)

  return { items, showForm, editingId, error, form, openCreate, openEdit, closeForm, submit, remove }
}
