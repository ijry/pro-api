<script setup lang="ts">
import { NDrawer, NDrawerContent, NForm, NSpace, NButton } from 'naive-ui'
import type { FormRules } from 'naive-ui'
import { ref } from 'vue'

const props = defineProps<{
  show: boolean
  mode?: 'create' | 'edit'
  title: string
  loading?: boolean
  rules?: FormRules
}>()

const emit = defineEmits<{
  'update:show': [val: boolean]
  submit: []
  cancel: []
}>()

const formRef = ref<InstanceType<typeof NForm> | null>(null)

function close() {
  emit('update:show', false)
  emit('cancel')
}

async function handleSubmit() {
  try {
    await formRef.value?.validate()
    emit('submit')
  } catch (_) { /* validation failed */ }
}
</script>

<template>
  <NDrawer :show="props.show" width="480" placement="right" @update:show="emit('update:show', $event)">
    <NDrawerContent :title="props.title" closable>
      <NForm ref="formRef" label-placement="top" :rules="props.rules ?? {}">
        <slot />
      </NForm>
      <template #footer>
        <NSpace justify="end">
          <NButton @click="close">取消</NButton>
          <NButton type="primary" :loading="props.loading" @click="handleSubmit">
            {{ props.mode === 'edit' ? '保存' : '创建' }}
          </NButton>
        </NSpace>
      </template>
    </NDrawerContent>
  </NDrawer>
</template>
