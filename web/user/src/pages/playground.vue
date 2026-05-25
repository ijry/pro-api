<script setup lang="ts">
import { ref } from 'vue'
import { post } from '@/api/http'

interface Message {
  role: 'user' | 'assistant' | 'system'
  content: string
}

interface ChatResponse {
  choices?: Array<{ message?: { content?: string } }>
}

const model = ref('gpt-4o-mini')
const input = ref('')
const messages = ref<Message[]>([])
const loading = ref(false)
const error = ref('')

async function send() {
  const text = input.value.trim()
  if (!text || loading.value) return
  input.value = ''
  error.value = ''
  messages.value.push({ role: 'user', content: text })
  loading.value = true
  try {
    const res = await post<ChatResponse>('/api/user/playground/chat', {
      model: model.value,
      messages: messages.value,
    })
    const choice = res?.choices?.[0]
    const reply = choice?.message?.content ?? JSON.stringify(res)
    messages.value.push({ role: 'assistant', content: reply })
  } catch (e: unknown) {
    const err = e as { response?: { data?: { message?: string } }; message?: string }
    error.value = err?.response?.data?.message ?? err?.message ?? '请求失败'
    // Remove the user message that failed so they can retry
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="space-y-5">
    <div>
      <h1 class="text-2xl font-bold text-fg">Playground</h1>
      <p class="text-sm text-fg-muted mt-1">在线测试模型对话效果</p>
    </div>

    <!-- Model selector -->
    <div class="flex gap-3 items-center">
      <label class="text-sm font-medium text-fg w-12 shrink-0">模型</label>
      <input
        v-model="model"
        class="border border-border rounded-md px-3 py-1.5 text-sm flex-1 bg-bg text-fg focus:outline-none focus:ring-2 focus:ring-primary/40 max-w-xs"
        placeholder="gpt-4o-mini"
      />
    </div>

    <!-- Chat history -->
    <div class="border border-border rounded-lg p-4 space-y-3 bg-bg-elevated min-h-64 max-h-[480px] overflow-y-auto">
      <div v-if="!messages.length" class="flex items-center justify-center h-48 text-fg-muted text-sm">
        <span>发送消息开始对话</span>
      </div>
      <div
        v-for="(msg, i) in messages"
        :key="i"
        :class="msg.role === 'user' ? 'flex justify-end' : 'flex justify-start'"
      >
        <span
          :class="[
            'inline-block rounded-lg px-3 py-2 text-sm max-w-[70%] break-words whitespace-pre-wrap',
            msg.role === 'user'
              ? 'bg-primary text-white'
              : 'bg-bg border border-border text-fg'
          ]"
        >{{ msg.content }}</span>
      </div>
      <div v-if="loading" class="flex justify-start">
        <span class="inline-block rounded-lg px-3 py-2 text-sm bg-bg border border-border text-fg-muted">
          <span class="i-lucide-loader-2 w-4 h-4 animate-spin inline-block mr-1" />思考中…
        </span>
      </div>
    </div>

    <!-- Input area -->
    <div class="flex gap-2">
      <textarea
        v-model="input"
        rows="2"
        class="border border-border rounded-md px-3 py-2 text-sm flex-1 resize-none bg-bg text-fg focus:outline-none focus:ring-2 focus:ring-primary/40 placeholder:text-fg-muted"
        placeholder="输入消息，按 Enter 发送…"
        @keydown.enter.exact.prevent="send"
      />
      <button
        class="px-4 py-2 bg-primary text-white rounded-md text-sm hover:bg-primary-hover transition-colors disabled:opacity-50 disabled:cursor-not-allowed shrink-0"
        :disabled="loading || !input.trim()"
        @click="send"
      >发送</button>
    </div>
    <p v-if="error" class="text-rose-500 text-sm">{{ error }}</p>
  </div>
</template>
