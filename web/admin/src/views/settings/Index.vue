<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { NCard, NTabs, NTabPane, NForm, NFormItem, NInput, NSwitch, NButton, NInputNumber, NSpin, useMessage } from 'naive-ui'
import { settingApi, type SettingGroup, type SettingItem } from '@/api/setting'

const message = useMessage()
const { t } = useI18n()
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
        editValues.value[item.key] = editValueFor(item)
      }
    }
  } catch (_) { /* handled */ } finally { loading.value = false }
}

onMounted(load)

async function saveItem(item: SettingItem) {
  saving.value[item.key] = true
  try {
    const value = parseEditValue(item)
    if (value === invalidJSON) return
    await settingApi.patch(item.key, value)
    message.success(t('settings.toast.saved', { key: item.key }))
  } catch (_) { /* handled */ } finally { saving.value[item.key] = false }
}

function isBoolean(v: unknown): v is boolean { return typeof v === 'boolean' }
function isNumber(v: unknown): v is number { return typeof v === 'number' }
function isObject(v: unknown): v is Record<string, unknown> { return v !== null && typeof v === 'object' && !Array.isArray(v) }

const invalidJSON = Symbol('invalidJSON')

function editValueFor(item: SettingItem) {
  const value = item.value
  if (isObject(value)) return JSON.stringify(value, null, 2)
  return value
}

function parseEditValue(item: SettingItem) {
  const original = item.value
  const value = editValues.value[item.key]
  if (!isObject(original)) return value
  if (typeof value !== 'string') return value
  try {
    return JSON.parse(value)
  } catch {
    message.error(t('settings.toast.invalid_json'))
    return invalidJSON
  }
}

function groupLabel(name: string) {
  return t(`settings.groups.${name}`, name)
}

function itemLabel(key: string) {
  return t(`settings.items.${key}.label`, key)
}

function itemDescription(item: SettingItem) {
  return t(`settings.items.${item.key}.description`, item.description || item.key)
}
</script>

<template>
  <div>
    <h2 class="text-2xl font-semibold mb-4">{{ t('settings.title') }}</h2>
    <NSpin :show="loading">
      <NCard size="small" :bordered="false">
        <NTabs type="line" animated>
          <NTabPane
            v-for="group in groups"
            :key="group.name"
            :name="group.name"
            :tab="groupLabel(group.name)"
          >
            <NForm label-placement="left" label-width="220px" class="mt-4">
              <div v-for="item in group.items" :key="item.key" class="mb-4">
                <NFormItem :label="itemLabel(item.key)">
                  <div class="flex items-center gap-2 w-full">
                    <template v-if="isBoolean(item.value)">
                      <NSwitch v-model:value="(editValues[item.key] as boolean)" />
                    </template>
                    <template v-else-if="isNumber(item.value)">
                      <NInputNumber v-model:value="(editValues[item.key] as number)" style="width: 220px" />
                    </template>
                    <template v-else-if="isObject(item.value)">
                      <NInput
                        v-model:value="(editValues[item.key] as string)"
                        type="textarea"
                        :autosize="{ minRows: 3, maxRows: 8 }"
                        style="flex: 1; max-width: 520px"
                        :placeholder="itemDescription(item)"
                      />
                    </template>
                    <template v-else>
                      <NInput
                        v-model:value="(editValues[item.key] as string)"
                        :type="item.is_sensitive ? 'password' : 'text'"
                        :show-password-on="item.is_sensitive ? 'click' : undefined"
                        style="flex: 1; max-width: 400px"
                        :placeholder="itemDescription(item)"
                      />
                    </template>
                    <NButton size="small" :loading="saving[item.key]" @click="saveItem(item)">{{ t('settings.actions.save') }}</NButton>
                  </div>
                  <span v-if="item.description" class="text-xs text-gray-400 ml-1">{{ itemDescription(item) }}</span>
                </NFormItem>
              </div>
            </NForm>
          </NTabPane>
        </NTabs>
      </NCard>
    </NSpin>
  </div>
</template>
