<template>
  <div class="w-full">
    <div class="mb-1 flex items-center justify-between">
      <span class="text-xs text-gray-400">{{ label }}</span>
      <span class="text-xs font-mono text-gray-400">{{ percentage }}%</span>
    </div>
    <div class="h-2 w-full rounded-full bg-gray-800 overflow-hidden">
      <div
        class="h-full rounded-full transition-all duration-300"
        :class="barColor"
        :style="{ width: `${clamped}%` }"
      />
    </div>
    <div v-if="fps || bitrate || speed" class="mt-1 flex gap-4 text-xs text-gray-500">
      <span v-if="fps">FPS: {{ fps }}</span>
      <span v-if="bitrate">Bitrate: {{ bitrate }}</span>
      <span v-if="speed">Speed: {{ speed }}</span>
      <span v-if="time">Time: {{ time }}</span>
    </div>
  </div>
</template>

<script setup lang="ts">
const props = withDefaults(defineProps<{
  value: number
  label?: string
  fps?: number
  bitrate?: string
  speed?: string
  time?: string
  status?: string
}>(), {
  label: 'Progress',
  status: 'running',
})

const clamped = computed(() => Math.min(100, Math.max(0, props.value * 100)))
const percentage = computed(() => Math.round(clamped.value))

const barColor = computed(() => {
  if (props.status === 'completed') return 'bg-green-500'
  if (props.status === 'failed') return 'bg-red-500'
  if (props.status === 'cancelled') return 'bg-yellow-500'
  return 'bg-primary-500'
})
</script>
