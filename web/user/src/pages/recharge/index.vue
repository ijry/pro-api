<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { rechargeApi, type ManualRechargeOrder } from '@/api/recharge'
import { useWalletStore } from '@/stores/wallet'
import { useToast } from '@/composables/useToast'
import WalletCard from '@/components/biz/WalletCard.vue'
import Dialog from '@/components/ui/Dialog.vue'
import Badge from '@/components/ui/Badge.vue'
import Pagination from '@/components/ui/Pagination.vue'
import Skeleton from '@/components/ui/Skeleton.vue'

const { t } = useI18n()
const toast = useToast()
const wallet = useWalletStore()

const PRESETS = [100, 200, 500, 1000]
const amount = ref<number | null>(null)
const customAmount = ref<number | null>(null)
const remark = ref('')
const submitting = ref(false)

const history = ref<ManualRechargeOrder[]>([])
const historyTotal = ref(0)
const historyPage = ref(1)
const historyLoading = ref(true)

const cancelTarget = ref<ManualRechargeOrder | null>(null)
const cancelConfirmOpen = ref(false)

function selectPreset(v: number) {
  amount.value = v
  customAmount.value = null
}

function onCustomInput() {
  amount.value = customAmount.value
}

async function loadHistory() {
  historyLoading.value = true
  try {
    const r = await rechargeApi.list(historyPage.value)
    history.value = r.items
    historyTotal.value = r.total
  } finally {
    historyLoading.value = false
  }
}

onMounted(loadHistory)

async function submitRecharge() {
  if (!amount.value || amount.value < 1) {
    toast.warn('请输入充值金额')
    return
  }
  submitting.value = true
  try {
    await rechargeApi.create(amount.value, remark.value)
    toast.success('充值申请已提交，请等待管理员审核')
    amount.value = null
    customAmount.value = null
    remark.value = ''
    await Promise.all([loadHistory(), wallet.refresh()])
  } catch (e: unknown) {
    const msg = (e as { response?: { data?: { message?: string } } })?.response?.data?.message || '提交失败'
    toast.error(msg)
  } finally {
    submitting.value = false
  }
}

function showCancelConfirm(order: ManualRechargeOrder) {
  cancelTarget.value = order
  cancelConfirmOpen.value = true
}

async function doCancel() {
  if (!cancelTarget.value) return
  try {
    await rechargeApi.cancel(cancelTarget.value.id)
    toast.success('申请已取消')
    cancelConfirmOpen.value = false
    await loadHistory()
  } catch {
    toast.error('取消失败')
  }
}

function statusVariant(status: string) {
  const m: Record<string, string> = { pending: 'warn', approved: 'success', rejected: 'error', cancelled: 'muted' }
  return (m[status] || 'muted') as 'warn' | 'success' | 'error' | 'muted'
}
function statusLabel(status: string) {
  const m: Record<string, string> = { pending: '待审核', approved: '已通过', rejected: '已拒绝', cancelled: '已取消' }
  return m[status] || status
}
</script>

<template>
  <div class="space-y-6 max-w-2xl">
    <h1 class="text-2xl font-bold text-fg">{{ t('nav.recharge') }}</h1>

    <WalletCard variant="recharge-hero" />

    <!-- Manual recharge card -->
    <div class="rounded-xl border border-border bg-bg-elevated p-6 space-y-4">
      <h3 class="font-semibold text-fg">① 手动充值</h3>
      <p class="text-sm text-fg-muted">转账到银行账户后，提交申请等待管理员审核。</p>

      <!-- Presets -->
      <div class="flex flex-wrap gap-2">
        <button
          v-for="preset in PRESETS"
          :key="preset"
          @click="selectPreset(preset)"
          :class="[
            'px-4 h-9 rounded-md border text-sm font-medium transition-colors',
            amount === preset ? 'border-primary bg-primary/10 text-primary' : 'border-border text-fg hover:bg-bg',
          ]"
        >¥ {{ preset }}</button>
        <input
          v-model.number="customAmount"
          @input="onCustomInput"
          type="number"
          placeholder="自定义"
          class="px-3 h-9 w-28 rounded-md border border-border bg-bg text-sm text-fg placeholder:text-fg-muted outline-none focus:border-primary"
        />
      </div>

      <!-- Remark -->
      <div>
        <label class="block text-sm font-medium text-fg mb-1">备注 <span class="text-fg-muted font-normal">（转账凭证/单号，≤ 512 字）</span></label>
        <textarea
          v-model="remark"
          maxlength="512"
          rows="3"
          class="w-full rounded-md border border-border bg-bg px-3 py-2 text-sm text-fg placeholder:text-fg-muted outline-none focus:border-primary resize-none"
          placeholder="请填写转账凭证或单号"
        />
      </div>

      <button @click="submitRecharge" :disabled="submitting || !amount"
        class="w-full h-10 rounded-md bg-primary text-white text-sm font-medium hover:bg-primary-hover disabled:opacity-50 transition-colors">
        {{ submitting ? '提交中...' : '提交申请' }}
      </button>
    </div>

    <!-- Redeem link -->
    <div class="rounded-xl border border-border bg-bg-elevated p-6">
      <h3 class="font-semibold text-fg mb-2">② 兑换码</h3>
      <router-link to="/redeem" class="text-sm text-primary hover:underline">已有兑换码？去兑换 →</router-link>
    </div>

    <!-- Coming soon -->
    <div class="rounded-xl border border-border/50 bg-bg-elevated/50 p-6 opacity-60">
      <h3 class="font-semibold text-fg mb-1">Stripe / 支付宝 / 微信支付</h3>
      <p class="text-sm text-fg-muted">Coming in M2</p>
    </div>

    <!-- Recharge history -->
    <div class="rounded-xl border border-border bg-bg-elevated p-6">
      <h3 class="font-semibold text-fg mb-4">我的充值历史</h3>
      <div v-if="historyLoading" class="space-y-2">
        <Skeleton v-for="i in 3" :key="i" class="h-12" />
      </div>
      <div v-else-if="!history.length" class="text-sm text-fg-muted text-center py-6">暂无充值记录</div>
      <div v-else class="space-y-2">
        <div
          v-for="order in history"
          :key="order.id"
          class="flex items-center gap-3 py-2 px-3 rounded-lg hover:bg-bg transition-colors"
        >
          <div class="flex-1">
            <div class="flex items-center gap-2">
              <span class="text-sm text-fg font-medium">#{{ order.id.slice(-6) }}</span>
              <span class="text-sm text-fg">¥ {{ order.amount_cny }}</span>
              <Badge :variant="statusVariant(order.status)" size="sm">{{ statusLabel(order.status) }}</Badge>
            </div>
            <p v-if="order.reject_reason" class="text-xs text-rose-400 mt-0.5">{{ order.reject_reason }}</p>
            <p v-if="order.granted_usd" class="text-xs text-emerald-400 mt-0.5">+${{ order.granted_usd }}</p>
          </div>
          <span class="text-xs text-fg-muted">{{ new Date(order.created_at).toLocaleDateString() }}</span>
          <button
            v-if="order.status === 'pending'"
            @click="showCancelConfirm(order)"
            class="text-xs text-rose-400 hover:text-rose-300 px-2 py-1 rounded hover:bg-rose-500/10 transition-colors"
          >取消</button>
        </div>
      </div>
      <Pagination v-if="historyTotal > 20" v-model="historyPage" :total="historyTotal" :size="20" @update:model-value="loadHistory" />
    </div>
  </div>

  <Dialog v-model:open="cancelConfirmOpen" title="确认取消充值申请？" size="sm">
    <div class="space-y-4">
      <p class="text-sm text-fg-muted">取消后申请将作废，请重新提交。</p>
      <div class="flex gap-3">
        <button @click="cancelConfirmOpen = false"
          class="flex-1 h-9 rounded-md border border-border text-sm hover:bg-bg-elevated transition-colors">取消</button>
        <button @click="doCancel"
          class="flex-1 h-9 rounded-md bg-rose-500 text-white text-sm font-medium hover:bg-rose-600 transition-colors">确认取消</button>
      </div>
    </div>
  </Dialog>
</template>
