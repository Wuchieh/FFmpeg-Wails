<template>
  <div class="mx-auto max-w-4xl space-y-6">
    <div class="flex items-center justify-between">
      <div>
        <h2 class="text-2xl font-bold">Tasks</h2>
        <p class="text-sm text-gray-400 mt-1">FFmpeg task management</p>
      </div>
      <button
        class="rounded-lg bg-gray-800 px-4 py-2 text-sm text-gray-300 hover:bg-gray-700 transition-colors"
        @click="refreshTasks"
      >
        Refresh
      </button>
    </div>

    <!-- Empty state -->
    <div v-if="tasks.length === 0" class="rounded-lg border border-gray-800 bg-gray-900 p-8 text-center">
      <p class="text-gray-500">No tasks yet. Start a conversion or stream to see tasks here.</p>
    </div>

    <!-- Task list -->
    <div v-else class="space-y-3">
      <div
        v-for="task in sortedTasks"
        :key="task.id"
        class="rounded-lg border border-gray-800 bg-gray-900 p-4 space-y-3"
      >
        <div class="flex items-center justify-between">
          <div class="flex items-center gap-3">
            <span
              class="inline-block rounded-full px-2 py-0.5 text-xs font-medium"
              :class="statusBadgeClass(task.status)"
            >
              {{ task.status }}
            </span>
            <span class="text-xs text-gray-500">{{ task.type }}</span>
          </div>
          <div class="flex items-center gap-2">
            <span class="text-xs text-gray-600">{{ formatTime(task.createdAt) }}</span>
            <button
              v-if="task.status === 'running'"
              class="rounded bg-red-600/20 px-2 py-1 text-xs text-red-400 hover:bg-red-600/30 transition-colors"
              @click="cancelTask(task.id)"
            >
              Cancel
            </button>
          </div>
        </div>

        <div class="text-xs text-gray-400 truncate">
          <span class="text-gray-500">Input:</span> {{ task.input }}
        </div>
        <div class="text-xs text-gray-400 truncate">
          <span class="text-gray-500">Output:</span> {{ task.output }}
        </div>

        <ProgressBar
          v-if="task.type === 'convert'"
          :value="task.progress"
          :status="task.status"
          v-bind="getProgressDetails(task.id)"
        />

        <div v-if="task.error" class="text-xs text-red-400 bg-red-900/20 rounded p-2">
          {{ task.error }}
        </div>

        <div v-if="task.warning" class="text-xs text-yellow-400 bg-yellow-900/20 rounded p-2">
          {{ task.warning }}
        </div>

        <!-- Collapsible logs -->
        <details v-if="getTaskLogs(task.id).length > 0" class="group">
          <summary class="cursor-pointer text-xs text-gray-500 hover:text-gray-300">
            Show logs ({{ getTaskLogs(task.id).length }} lines)
          </summary>
          <div class="mt-2">
            <LogViewer :lines="getTaskLogs(task.id)" />
          </div>
        </details>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
const { tasks, refreshTasks, cancelTask, getTaskLogs, getTaskProgress } = useFFmpeg()

onMounted(() => {
  refreshTasks()
})

const sortedTasks = computed(() =>
  [...tasks.value].sort((a, b) =>
    new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime()
  )
)

function statusBadgeClass(status: string): string {
  switch (status) {
    case 'running': return 'bg-blue-500/20 text-blue-400'
    case 'completed': return 'bg-green-500/20 text-green-400'
    case 'failed': return 'bg-red-500/20 text-red-400'
    case 'canceled': return 'bg-yellow-500/20 text-yellow-400'
    default: return 'bg-gray-500/20 text-gray-400'
  }
}

function getProgressDetails(id: string) {
  const p = getTaskProgress(id)
  if (!p) return {}
  return { fps: p.fps, bitrate: p.bitrate, speed: p.speed, time: p.time }
}

function formatTime(dateStr: string): string {
  try {
    const d = new Date(dateStr)
    return d.toLocaleString()
  } catch {
    return dateStr
  }
}
</script>
