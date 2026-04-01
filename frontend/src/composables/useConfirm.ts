import { ref } from 'vue'

const visible = ref(false)
const message = ref('')
const confirmLabel = ref('Confirm')
const confirmVariant = ref<'danger' | 'primary'>('danger')
let resolve: ((value: boolean) => void) | null = null

export function useConfirm() {
  function confirm(msg: string, opts?: { label?: string; variant?: 'danger' | 'primary' }): Promise<boolean> {
    message.value = msg
    confirmLabel.value = opts?.label ?? 'Delete'
    confirmVariant.value = opts?.variant ?? 'danger'
    visible.value = true
    return new Promise<boolean>((r) => {
      resolve = r
    })
  }

  function respond(value: boolean) {
    visible.value = false
    resolve?.(value)
    resolve = null
  }

  return { visible, message, confirmLabel, confirmVariant, confirm, respond }
}
