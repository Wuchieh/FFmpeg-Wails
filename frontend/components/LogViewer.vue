<template>
  <div class="rounded-lg border border-gray-800 bg-gray-900">
    <div class="flex items-center justify-between border-b border-gray-800 px-3 py-2">
      <span class="text-xs font-medium text-gray-400">Logs</span>
      <span class="text-xs text-gray-600">{{ lines.length }} lines</span>
    </div>
    <div
      ref="logContainer"
      class="max-h-64 overflow-y-auto p-3 font-mono text-xs leading-relaxed text-gray-500"
    >
      <div v-if="lines.length === 0" class="text-gray-700 italic">No output yet...</div>
      <div v-for="(line, i) in lines" :key="i" class="hover:text-gray-300 transition-colors">
        {{ line }}
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
const props = withDefaults(defineProps<{
  lines: string[]
}>(), {
  lines: () => [],
})

const logContainer = ref<HTMLElement | null>(null)

watch(() => props.lines.length, () => {
  nextTick(() => {
    if (logContainer.value) {
      logContainer.value.scrollTop = logContainer.value.scrollHeight
    }
  })
})
</script>
