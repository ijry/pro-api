<script setup lang="ts">
import { NModal, NInput, NText } from 'naive-ui'
import { ref, computed } from 'vue'

const props = defineProps<{
  show: boolean
  title?: string
  content?: string
  type?: 'warning' | 'error' | 'info'
  confirmText?: string
  cancelText?: string
  requireInput?: string
}>()

const emit = defineEmits<{
  'update:show': [val: boolean]
  confirm: []
  cancel: []
}>()

const inputVal = ref('')
const canConfirm = computed(() => {
  if (props.requireInput) return inputVal.value === props.requireInput
  return true
})

function close() { emit('update:show', false); emit('cancel') }
function confirm() {
  if (!canConfirm.value) return
  emit('confirm')
  emit('update:show', false)
  inputVal.value = ''
}
</script>

<template>
  <NModal :show="props.show" preset="dialog" :type="props.type ?? 'warning'"
    :title="props.title ?? '确认操作'"
    :content="props.content"
    :positive-text="props.confirmText ?? '确认'"
    :negative-text="props.cancelText ?? '取消'"
    @positive-click="confirm"
    @negative-click="close"
    @update:show="emit('update:show', $event)"
  >
    <template v-if="props.requireInput" #default>
      <div>
        <p>{{ props.content }}</p>
        <p class="mt-2 text-sm text-gray-500">请输入 <NText code>{{ props.requireInput }}</NText> 确认</p>
        <NInput v-model:value="inputVal" class="mt-2" :placeholder="props.requireInput" />
      </div>
    </template>
  </NModal>
</template>
