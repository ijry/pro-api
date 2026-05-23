<script setup lang="ts">
interface Column {
  key: string
  label: string
  align?: 'left' | 'right' | 'center'
  class?: string
}
defineProps<{ columns: Column[]; rows: Record<string, unknown>[] }>()
</script>

<template>
  <div class="overflow-x-auto rounded-lg border border-border">
    <table class="w-full text-sm">
      <thead>
        <tr class="border-b border-border bg-bg-elevated/50">
          <th
            v-for="col in columns"
            :key="col.key"
            :class="[
              'px-4 py-3 font-medium text-fg-muted whitespace-nowrap',
              col.align === 'right' ? 'text-right' : col.align === 'center' ? 'text-center' : 'text-left',
              col.class,
            ]"
          >{{ col.label }}</th>
        </tr>
      </thead>
      <tbody>
        <tr
          v-for="(row, i) in rows"
          :key="i"
          class="border-b border-border/50 last:border-0 hover:bg-bg-elevated/40 transition-colors"
        >
          <td
            v-for="col in columns"
            :key="col.key"
            :class="[
              'px-4 py-3 text-fg',
              col.align === 'right' ? 'text-right' : col.align === 'center' ? 'text-center' : 'text-left',
              col.class,
            ]"
          >
            <slot :name="col.key" :row="row" :value="row[col.key]">
              {{ row[col.key] }}
            </slot>
          </td>
        </tr>
        <tr v-if="!rows.length">
          <td :colspan="columns.length" class="px-4 py-8 text-center text-fg-muted">
            <slot name="empty">暂无数据</slot>
          </td>
        </tr>
      </tbody>
    </table>
  </div>
</template>
