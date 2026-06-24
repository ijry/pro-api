<script setup lang="ts">
import { ref, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import ClipboardButton from '@/components/ui/ClipboardButton.vue'

const { t } = useI18n()

const tabs = ['curl', 'python', 'node', 'go'] as const
type Tab = typeof tabs[number]
const active = ref<Tab>('curl')

const BASE = 'https://api.proapi.io/v1'
const TOKEN_PLACEHOLDER = 'pa-xxx...'

const code: Record<Tab, string> = {
  curl: `curl ${BASE}/chat/completions \\
  -H "Content-Type: application/json" \\
  -H "Authorization: Bearer ${TOKEN_PLACEHOLDER}" \\
  -d '{
    "model": "gpt-4o",
    "messages": [{"role": "user", "content": "Hello!"}]
  }'`,
  python: `from openai import OpenAI

client = OpenAI(
    base_url="${BASE}",
    api_key="${TOKEN_PLACEHOLDER}",
)

completion = client.chat.completions.create(
    model="gpt-4o",
    messages=[{"role": "user", "content": "Hello!"}],
)
print(completion.choices[0].message.content)`,
  node: `import OpenAI from "openai";

const client = new OpenAI({
  baseURL: "${BASE}",
  apiKey: "${TOKEN_PLACEHOLDER}",
});

const completion = await client.chat.completions.create({
  model: "gpt-4o",
  messages: [{ role: "user", content: "Hello!" }],
});
console.log(completion.choices[0].message.content);`,
  go: `client := openai.NewClient(
    option.WithBaseURL("${BASE}"),
    option.WithAPIKey("${TOKEN_PLACEHOLDER}"),
)
completion, _ := client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
    Model:    openai.F(openai.ChatModelGPT4o),
    Messages: openai.F([]openai.ChatCompletionMessageParamUnion{
        openai.UserMessage("Hello!"),
    }),
})
fmt.Println(completion.Choices[0].Message.Content)`,
}

const currentCode = computed(() => code[active.value])
</script>

<template>
  <div class="rounded-xl border border-border bg-bg-elevated backdrop-blur-md p-5 shadow-sm">
    <h3 class="font-semibold text-fg mb-4">{{ t('home.quickstart.title') }}</h3>

    <!-- Tab bar -->
    <div class="flex gap-1 mb-3">
      <button
        v-for="tab in tabs"
        :key="tab"
        @click="active = tab"
        :class="[
          'px-3 h-7 rounded-md text-xs font-medium transition-colors',
          active === tab ? 'bg-primary text-white' : 'text-fg-muted hover:text-fg hover:bg-bg',
        ]"
      >{{ t(`home.quickstart.tab.${tab}`, tab) }}</button>
    </div>

    <!-- Code block -->
    <div class="relative rounded-lg bg-bg border border-border overflow-hidden">
      <div class="absolute top-2 right-2">
        <ClipboardButton :text="currentCode" :success-msg="t('home.quickstart.copy')" size="sm" />
      </div>
      <pre class="text-xs text-fg-muted p-4 pr-10 overflow-x-auto leading-relaxed font-mono whitespace-pre">{{ currentCode }}</pre>
    </div>

    <div class="flex gap-2 mt-3">
      <router-link to="/apikeys"
        class="inline-flex items-center gap-1 px-3 h-7 rounded-md border border-border text-xs text-fg hover:bg-bg transition-colors">
        <span class="i-lucide-key-round w-3.5 h-3.5" />{{ t('home.quickstart.my_tokens') }}
      </router-link>
    </div>
  </div>
</template>
