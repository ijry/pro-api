<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { rechargeApi, type ManualRechargeOrder } from '@/api/recharge'
import { useToast } from '@/composables/useToast'
import Skeleton from '@/components/ui/Skeleton.vue'
import Button from '@/components/ui/Button.vue'

const route = useRoute()
const router = useRouter()
const toast = useToast()
const order = ref<ManualRechargeOrder | null>(null)
const loading = ref(true)
const cancelling = ref(false)

async function load() {
  loading.value = true
  try {
    const id = route.params.orderId as string
    order.value = await rechargeApi.get(id)
  } catch { toast.error('加载订单失败'); router.push('/recharge') } finally { loading.value = false }
}

onMounted(load)

async function cancel() {
  if (!order.value) return
  cancelling.value = true
  try {
    await rechargeApi.cancel(order.value.id)
    toast.success('订单已取消')
    await load()
  } catch (e: unknown) {
    toast.error((e as { response?: { data?: { message?: string } } })?.response?.data?.message || '取消失败')
  } finally { cancelling.value = false }
}

const statusLabels: Record<string, { label: string; cls: string }> = {
  pending: { label: '待审核', cls: 'text-yellow-500' },
  approved: { label: '已通过', cls: 'text-green-500' },
  rejected: { label: '已拒绝', cls: 'text-rose-500' },
  cancelled: { label: '已取消', cls: 'text-fg-muted' },
}
</script>

<template>
  <div class="max-w-lg space-y-5">
    <button @click="router.push('/recharge')" class="flex items-center gap-1 text-sm text-fg-muted hover:text-fg transition-colors">
      <span class="i-lucide-arrow-left w-4 h-4" />
      返回充值记录
    </button>

    <div v-if="loading">
      <Skeleton class="h-48" />
    </div>

    <div v-else-if="order" class="bg-bg-elevated border border-border rounded-xl p-6 space-y-4">
      <div class="flex items-center justify-between">
        <h2 class="text-lg font-semibold text-fg">充值订单详情</h2>
        <span class="text-sm font-medium" :class="statusLabels[order.status]?.cls">
          {{ statusLabels[order.status]?.label ?? order.status }}
        </span>
      </div>

      <div class="space-y-3 text-sm">
        <div class="flex justify-between">
          <span class="text-fg-muted">订单 ID</span>
          <span class="font-mono text-fg text-xs">{{ order.id }}</span>
        </div>
        <div class="flex justify-between">
          <span class="text-fg-muted">充值金额</span>
          <span class="text-fg font-semibold">¥{{ order.amount_cny }}</span>
        </div>
        <div v-if="order.granted_usd !== undefined" class="flex justify-between">
          <span class="text-fg-muted">发放额度</span>
          <span class="text-green-500 font-semibold">${{ order.granted_usd }}</span>
        </div>
        <div class="flex justify-between">
          <span class="text-fg-muted">申请备注</span>
          <span class="text-fg text-right max-w-xs">{{ order.remark || '--' }}</span>
        </div>
        <div v-if="order.reject_reason" class="flex justify-between">
          <span class="text-fg-muted">拒绝原因</span>
          <span class="text-rose-500 text-right max-w-xs">{{ order.reject_reason }}</span>
        </div>
        <div class="flex justify-between">
          <span class="text-fg-muted">申请时间</span>
          <span class="text-fg">{{ new Date(order.created_at).toLocaleString('zh-CN') }}</span>
        </div>
        <div class="flex justify-between">
          <span class="text-fg-muted">更新时间</span>
          <span class="text-fg">{{ new Date(order.updated_at).toLocaleString('zh-CN') }}</span>
        </div>
      </div>

      <Button
        v-if="order.status === 'pending'"
        variant="ghost"
        :disabled="cancelling"
        @click="cancel"
        class="w-full border-rose-300 text-rose-500 hover:bg-rose-50 dark:hover:bg-rose-950/20"
      >
        <span v-if="cancelling" class="i-lucide-loader-circle w-4 h-4 mr-1 animate-spin" />
        取消申请
      </Button>
    </div>
  </div>
</template>
