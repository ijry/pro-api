import { reactive, computed } from 'vue'

type Rules<T> = { [K in keyof T]?: (v: T[K]) => string }

export function useForm<T extends Record<string, unknown>>(options: {
  initial: T
  rules?: Rules<T>
}) {
  const values = reactive({ ...options.initial }) as T
  const errors = reactive({} as Record<keyof T, string>)
  const touched = reactive({} as Record<keyof T, boolean>)

  function validate(): boolean {
    let ok = true
    if (!options.rules) return true
    for (const key in options.rules) {
      const rule = options.rules[key]
      if (rule) {
        const msg = rule(values[key])
        ;(errors as Record<string, string>)[key] = msg
        if (msg) ok = false
      }
    }
    return ok
  }

  function touch(key: keyof T) {
    (touched as Record<string, boolean>)[key as string] = true
    if (options.rules?.[key]) {
      ;(errors as Record<string, string>)[key as string] = options.rules[key]!(values[key]) ?? ''
    }
  }

  function reset() {
    Object.assign(values, options.initial)
    for (const key in errors) delete (errors as Record<string, string>)[key]
    for (const key in touched) delete (touched as Record<string, boolean>)[key]
  }

  function setValues(partial: Partial<T>) {
    Object.assign(values, partial)
  }

  const isValid = computed(() => {
    if (!options.rules) return true
    for (const key in options.rules) {
      if (options.rules[key]?.(values[key])) return false
    }
    return true
  })

  return { values, errors, touched, validate, touch, reset, setValues, isValid }
}
