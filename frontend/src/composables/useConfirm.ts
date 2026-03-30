import { ref } from 'vue'

const visible = ref(false)
const message = ref('')
let resolve: ((value: boolean) => void) | null = null

export function useConfirm() {
  function confirm(msg: string): Promise<boolean> {
    message.value = msg
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

  return { visible, message, confirm, respond }
}
