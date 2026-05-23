<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { NCard, NForm, NFormItem, NInput, NSelect, NButton, NSpace, NSwitch, NDatePicker, NSpin, useMessage } from 'naive-ui'
import { noticeApi, type NoticeInput } from '@/api/notice'

const route = useRoute()
const router = useRouter()
const message = useMessage()

const isNew = computed(() => route.meta.mode === 'new' || route.path === '/notices/new')
const noticeId = computed(() => isNew.value ? 0 : Number(route.params.id))

const loading = ref(false)
const saving = ref(false)

const form = ref<NoticeInput & { publish_at_ts: number|null }>({
  title: '', content: '', level: 'info', target: 'all',
  publish_at: null, expires_at: null, pinned: false,
  publish_at_ts: null,
})

async function load() {
  if (isNew.value) return
  loading.value = true
  try {
    const n = await noticeApi.detail(noticeId.value)
    form.value = {
      title: n.title, content: n.content, level: n.level,
      target: n.target, publish_at: n.publish_at, expires_at: n.expires_at,
      pinned: n.pinned, publish_at_ts: n.publish_at ? new Date(n.publish_at).getTime() : null,
    }
  } catch (_) { /* handled */ } finally { loading.value = false }
}

onMounted(load)

async function save() {
  saving.value = true
  try {
    const payload: NoticeInput = {
      title: form.value.title, content: form.value.content,
      level: form.value.level, target: form.value.target,
      publish_at: form.value.publish_at_ts ? new Date(form.value.publish_at_ts).toISOString() : null,
      expires_at: form.value.expires_at, pinned: form.value.pinned,
    }
    if (isNew.value) {
      await noticeApi.create(payload)
      message.success('公告已创建')
      router.push('/notices')
    } else {
      await noticeApi.patch(noticeId.value, payload)
      message.success('公告已保存')
    }
  } catch (_) { /* handled */ } finally { saving.value = false }
}

const levelOptions = [
  { label: '信息', value: 'info' }, { label: '警告', value: 'warning' },
  { label: '危险', value: 'danger' }, { label: '成功', value: 'success' },
]
const targetOptions = [
  { label: '所有人', value: 'all' }, { label: '普通用户', value: 'user' }, { label: '管理员', value: 'admin' },
]
</script>

<template>
  <div>
    <div class="flex items-center gap-2 mb-4">
      <NButton text @click="router.push('/notices')">← 返回公告列表</NButton>
    </div>
    <NCard :title="isNew ? '新建公告' : '编辑公告'" size="small">
      <NSpin :show="loading">
        <NForm label-placement="top" style="max-width: 720px">
          <NFormItem label="标题">
            <NInput v-model:value="form.title" placeholder="公告标题" />
          </NFormItem>
          <NFormItem label="内容（支持 Markdown）">
            <NInput v-model:value="form.content" type="textarea" :rows="10" placeholder="正文内容..." />
          </NFormItem>
          <div class="grid grid-cols-2 gap-4">
            <NFormItem label="级别">
              <NSelect v-model:value="form.level" :options="levelOptions" />
            </NFormItem>
            <NFormItem label="目标受众">
              <NSelect v-model:value="form.target" :options="targetOptions" />
            </NFormItem>
          </div>
          <NFormItem label="发布时间">
            <NDatePicker v-model:value="form.publish_at_ts" type="datetime" clearable style="width:100%" />
          </NFormItem>
          <NFormItem label="置顶">
            <NSwitch v-model:value="form.pinned" />
          </NFormItem>
          <NSpace>
            <NButton type="primary" :loading="saving" @click="save">保存</NButton>
            <NButton @click="router.push('/notices')">取消</NButton>
          </NSpace>
        </NForm>
      </NSpin>
    </NCard>
  </div>
</template>
