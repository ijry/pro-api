<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { NCard, NTabs, NTabPane, NForm, NFormItem, NInput, NSwitch, NButton, NInputNumber, NSpin, useMessage } from 'naive-ui'
import { settingApi, type SettingGroup, type SettingItem } from '@/api/setting'

const message = useMessage()
const groups = ref<SettingGroup[]>([])
const loading = ref(false)
const saving = ref<Record<string, boolean>>({})

// Local editable values per key
const editValues = ref<Record<string, unknown>>({})

async function load() {
  loading.value = true
  try {
    const res = await settingApi.all()
    groups.value = res.groups
    for (const g of res.groups) {
      for (const item of g.items) {
        editValues.value[item.key] = item.value
      }
    }
  } catch (_) { /* handled */ } finally { loading.value = false }
}

onMounted(load)

async function saveItem(item: SettingItem) {
  saving.value[item.key] = true
  try {
    await settingApi.patch(item.key, editValues.value[item.key])
    message.success(`${item.key} 已保存`)
  } catch (_) { /* handled */ } finally { saving.value[item.key] = false }
}

function isBoolean(v: unknown): v is boolean { return typeof v === 'boolean' }
function isNumber(v: unknown): v is number { return typeof v === 'number' }
</script>

<template>
  <div>
    <h2 class="text-2xl font-semibold mb-4">系统设置</h2>
    <NSpin :show="loading">
      <NCard size="small" :bordered="false">
        <NTabs type="line" animated>
          <NTabPane
            v-for="group in groups"
            :key="group.name"
            :name="group.name"
            :tab="group.name"
          >
            <NForm label-placement="left" label-width="220px" class="mt-4">
              <div v-for="item in group.items" :key="item.key" class="mb-4">
                <NFormItem :label="item.key">
                  <div class="flex items-center gap-2 w-full">
                    <template v-if="isBoolean(item.value)">
                      <NSwitch v-model:value="(editValues[item.key] as boolean)" />
                    </template>
                    <template v-else-if="isNumber(item.value)">
                      <NInputNumber v-model:value="(editValues[item.key] as number)" style="width: 220px" />
                    </template>
                    <template v-else>
                      <NInput
                        v-model:value="(editValues[item.key] as string)"
                        :type="item.is_sensitive ? 'password' : 'text'"
                        :show-password-on="item.is_sensitive ? 'click' : undefined"
                        style="flex: 1; max-width: 400px"
                        :placeholder="item.description"
                      />
                    </template>
                    <NButton size="small" :loading="saving[item.key]" @click="saveItem(item)">保存</NButton>
                  </div>
                  <span v-if="item.description" class="text-xs text-gray-400 ml-1">{{ item.description }}</span>
                </NFormItem>
              </div>
            </NForm>
          </NTabPane>
        </NTabs>
      </NCard>
    </NSpin>
  </div>
</template>
