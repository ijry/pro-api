import { useMessage, type MessageReactive } from 'naive-ui'
import { useI18n } from 'vue-i18n'
import { accountApi, type Account } from '@/api/account'

export function useAccountActions(reload: () => void | Promise<void>) {
  const message = useMessage()
  const { t } = useI18n()

  async function doTest(row: Account) {
    let msg: MessageReactive | null = null
    try {
      msg = message.loading(t('accounts.actions.test') + '…', { duration: 0 })
      await accountApi.test(row.id)
      msg?.destroy()
      message.success(t('accounts.toast.test_ok'))
      reload()
    } catch (_) {
      msg?.destroy()
      message.error(t('accounts.toast.test_fail'))
    }
  }

  async function doRefresh(row: Account) {
    try {
      await accountApi.refresh(row.id)
      message.success(t('accounts.toast.refreshed'))
      reload()
    } catch (_) { /* handled */ }
  }

  async function doClearCooldown(row: Account) {
    try {
      await accountApi.clearCooldown(row.id)
      message.success(t('accounts.toast.cd_cleared'))
      reload()
    } catch (_) { /* handled */ }
  }

  async function doEnable(row: Account) {
    try {
      await accountApi.enable(row.id)
      message.success(t('accounts.toast.enabled'))
      reload()
    } catch (_) { /* handled */ }
  }

  async function doDisable(row: Account) {
    try {
      await accountApi.disable(row.id)
      message.success(t('accounts.toast.disabled'))
      reload()
    } catch (_) { /* handled */ }
  }

  async function doDelete(row: Account) {
    try {
      await accountApi.delete(row.id)
      message.success(t('accounts.toast.deleted'))
      reload()
    } catch (_) { /* handled */ }
  }

  return { doTest, doRefresh, doClearCooldown, doEnable, doDisable, doDelete }
}
