<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { NCard, NForm, NFormItem, NInput, NInputNumber, NButton, NDatePicker, useMessage } from 'naive-ui'
import CopyButton from '@/components/CopyButton.vue'
import { redeemApi, type BatchGenerateInput } from '@/api/redeem'

const message = useMessage()
const router = useRouter()

const form = ref<BatchGenerateInput>({ count: 10, amount_quota: 100, batch_no: '', expires_at: null })
const expiresTs = ref<number|null>(null)
const generating = ref(false)
const result = ref<{ batch_no: string; count: number; codes: string[] } | null>(null)

async function generate() {
  if (!form.value.batch_no) { message.error('请填写批次号'); return }
  generating.value = true
  try {
    const payload: BatchGenerateInput = {
      count: form.value.count,
      amount_quota: form.value.amount_quota,
      batch_no: form.value.batch_no,
      expires_at: expiresTs.value ? new Date(expiresTs.value).toISOString() : null,
    }
    result.value = await redeemApi.batchGenerate(payload)
    message.success(`已生成 ${result.value.count} 个兑换码`)
  } catch (_) { /* handled */ } finally { generating.value = false }
}

function copyAll() {
  if (!result.value) return
  navigator.clipboard.writeText(result.value.codes.join('\n')).then(() => message.success('已复制所有兑换码'))
}
</script>

<template>
  <div>
    <div class="flex items-center gap-2 mb-4">
      <NButton text @click="router.push('/payments/redeem')">← 返回兑换码列表</NButton>
    </div>
    <div class="grid grid-cols-1 lg:grid-cols-2 gap-4">
      <NCard title="批量生成兑换码" size="small">
        <NForm label-placement="top">
          <NFormItem label="批次号（唯一标识）">
            <NInput v-model:value="form.batch_no" placeholder="batch_2024_01" />
          </NFormItem>
          <NFormItem label="生成数量">
            <NInputNumber v-model:value="form.count" :min="1" :max="1000" style="width:100%" />
          </NFormItem>
          <NFormItem label="每码 Quota">
            <NInputNumber v-model:value="form.amount_quota" :min="1" style="width:100%" />
          </NFormItem>
          <NFormItem label="到期时间（可选）">
            <NDatePicker v-model:value="expiresTs" type="datetime" clearable style="width:100%" />
          </NFormItem>
          <NButton type="primary" :loading="generating" @click="generate" style="width:100%">生成</NButton>
        </NForm>
      </NCard>

      <NCard v-if="result" title="生成结果" size="small">
        <div class="flex items-center justify-between mb-2">
          <span class="text-sm text-gray-500">批次号：<code class="text-primary">{{ result.batch_no }}</code>，共 {{ result.count }} 个</span>
          <NButton size="small" @click="copyAll">复制全部</NButton>
        </div>
        <div class="overflow-auto max-h-80 border border-gray-200 dark:border-gray-700 rounded p-2">
          <div v-for="(code, i) in result.codes" :key="i" class="flex items-center justify-between py-0.5 border-b border-gray-100 dark:border-gray-800 last:border-0">
            <code class="text-xs font-mono">{{ code }}</code>
            <CopyButton :value="code" size="tiny" />
          </div>
        </div>
      </NCard>
    </div>
  </div>
</template>
