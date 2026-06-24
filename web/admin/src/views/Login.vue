<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { NCard, NForm, NFormItem, NInput, NButton, NAlert, NDivider } from 'naive-ui'
import { useI18n } from 'vue-i18n'
import { useRouter, useRoute } from 'vue-router'
import { useUserStore } from '@/stores/user'

const { t } = useI18n()
const router = useRouter()
const route = useRoute()
const userStore = useUserStore()

const form = ref({ identity: '', password: '' })
const loading = ref(false)
const errorMsg = ref('')

// If already logged in, redirect
onMounted(async () => {
  if (userStore.user) {
    router.replace((route.query.redirect as string) || '/')
  }
})

async function submit() {
  if (!form.value.identity || !form.value.password) {
    errorMsg.value = '请填写用户名和密码'
    return
  }
  loading.value = true
  errorMsg.value = ''
  try {
    await userStore.login(form.value.identity, form.value.password)
    router.replace((route.query.redirect as string) || '/')
  } catch (err: unknown) {
    const e = err as { response?: { data?: { code?: number; message?: string }; status?: number } }
    if (e?.response?.status === 403) {
      errorMsg.value = t('login.errors.not_admin')
    } else if (e?.response?.data?.message) {
      errorMsg.value = e.response.data.message
    } else {
      errorMsg.value = t('login.errors.wrong_password')
    }
  } finally {
    loading.value = false
  }
}

function handleEnter(e: KeyboardEvent) {
  if (e.key === 'Enter') submit()
}
</script>

<template>
  <div class="min-h-screen flex flex-col items-center justify-center p-4">
    <NCard class="w-full max-w-md" :title="t('login.title')">
      <template #header-extra>
        <span class="text-sm opacity-60">{{ t('login.subtitle') }}</span>
      </template>

      <NAlert v-if="errorMsg" type="error" class="mb-4" :title="errorMsg" />

      <NForm label-placement="top" @keydown="handleEnter">
        <NFormItem :label="t('login.identity')">
          <NInput
            v-model:value="form.identity"
            :placeholder="t('login.identity_placeholder')"
            :disabled="loading"
          />
        </NFormItem>
        <NFormItem :label="t('login.password')">
          <NInput
            v-model:value="form.password"
            type="password"
            show-password-on="click"
            :disabled="loading"
          />
        </NFormItem>
        <NButton type="primary" block :loading="loading" @click="submit">
          {{ t('login.submit') }}
        </NButton>
      </NForm>

      <NDivider />
      <NButton
        ghost block
        tag="a"
        href="/api/auth/oauth/github/start?redirect=/admin/"
      >
        GitHub {{ t('login.oauth_login') }}
      </NButton>
    </NCard>

    <p class="mt-4 text-sm opacity-40">pro-api admin &copy; {{ new Date().getFullYear() }}</p>
  </div>
</template>
