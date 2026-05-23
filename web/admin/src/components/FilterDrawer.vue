<script setup lang="ts">
import { NDrawer, NDrawerContent, NForm, NSpace, NButton } from 'naive-ui'

const props = defineProps<{
  show: boolean
  title?: string
  width?: number
}>()

const emit = defineEmits<{
  'update:show': [val: boolean]
  apply: []
  reset: []
}>()

function close() { emit('update:show', false) }
function apply() { emit('apply'); close() }
function reset() { emit('reset') }
</script>

<template>
  <NDrawer :show="props.show" :width="props.width ?? 360" placement="right" @update:show="emit('update:show', $event)">
    <NDrawerContent :title="props.title ?? '筛选'" closable>
      <NForm label-placement="top">
        <slot />
      </NForm>
      <template #footer>
        <NSpace justify="end">
          <NButton @click="reset">重置</NButton>
          <NButton type="primary" @click="apply">应用筛选</NButton>
        </NSpace>
      </template>
    </NDrawerContent>
  </NDrawer>
</template>
