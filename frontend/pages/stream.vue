<template>
  <div class="mx-auto max-w-3xl space-y-6">
    <div>
      <h2 class="text-2xl font-bold">Stream</h2>
      <p class="text-sm text-gray-400 mt-1">RTMP / SRT live streaming</p>
    </div>

    <!-- Source -->
    <FileSelector v-model="form.source" label="Source File" placeholder="Select video file or enter stream URL..." />

    <div>
      <label class="text-sm text-gray-400">Or enter URL (RTSP/HTTP)</label>
      <input
        v-model="form.sourceURL"
        type="text"
        placeholder="rtsp://... or http://..."
        class="mt-1 w-full rounded-lg border border-gray-700 bg-gray-900 px-3 py-2 text-sm text-gray-100 outline-none focus:border-primary-500"
        @focus="form.source = ''"
      />
    </div>

    <!-- Protocol -->
    <div>
      <label class="text-sm text-gray-400">Protocol</label>
      <div class="mt-1 flex gap-3">
        <button
          class="rounded-lg px-4 py-2 text-sm font-medium transition-colors"
          :class="form.protocol === 'rtmp' ? 'bg-primary-600 text-white' : 'bg-gray-800 text-gray-400 hover:bg-gray-700'"
          @click="form.protocol = 'rtmp'"
        >
          RTMP
        </button>
        <button
          class="rounded-lg px-4 py-2 text-sm font-medium transition-colors"
          :class="form.protocol === 'srt' ? 'bg-primary-600 text-white' : 'bg-gray-800 text-gray-400 hover:bg-gray-700'"
          @click="form.protocol = 'srt'"
        >
          SRT
        </button>
      </div>
    </div>

    <!-- Server URL -->
    <div>
      <label class="text-sm text-gray-400">
        {{ form.protocol === 'rtmp' ? 'RTMP URL' : 'SRT URL' }}
      </label>
      <input
        v-model="form.url"
        type="text"
        :placeholder="form.protocol === 'rtmp' ? 'rtmp://a.rtmp.youtube.com/live2/YOUR-STREAM-KEY' : 'srt://host:port?mode=caller'"
        class="mt-1 w-full rounded-lg border border-gray-700 bg-gray-900 px-3 py-2 text-sm text-gray-100 outline-none focus:border-primary-500"
      />
    </div>

    <!-- Presets -->
    <div v-if="form.protocol === 'rtmp'" class="grid grid-cols-3 gap-2">
      <button
        v-for="preset in rtmpPresets" :key="preset.name"
        class="rounded-lg border border-gray-700 bg-gray-900 px-3 py-2 text-xs text-gray-300 hover:border-primary-500 transition-colors"
        @click="form.url = preset.url"
      >
        {{ preset.name }}
      </button>
    </div>

    <!-- Encoding Settings -->
    <div class="grid grid-cols-2 gap-4">
      <div>
        <label class="text-sm text-gray-400">Video Codec</label>
        <select v-model="form.videoCodec" class="mt-1 w-full rounded-lg border border-gray-700 bg-gray-900 px-3 py-2 text-sm text-gray-100">
          <option value="libx264">H.264</option>
          <option value="libx265">H.265</option>
        </select>
      </div>
      <div>
        <label class="text-sm text-gray-400">Preset</label>
        <select v-model="form.preset" class="mt-1 w-full rounded-lg border border-gray-700 bg-gray-900 px-3 py-2 text-sm text-gray-100">
          <option value="ultrafast">Ultrafast</option>
          <option value="veryfast">Veryfast</option>
          <option value="fast">Fast</option>
          <option value="medium">Medium</option>
          <option value="slow">Slow</option>
        </select>
      </div>
    </div>

    <div class="grid grid-cols-2 gap-4">
      <div>
        <label class="text-sm text-gray-400">Bitrate</label>
        <input
          v-model="form.bitrate"
          type="text"
          placeholder="3000k"
          class="mt-1 w-full rounded-lg border border-gray-700 bg-gray-900 px-3 py-2 text-sm text-gray-100 outline-none focus:border-primary-500"
        />
      </div>
      <div v-if="form.protocol === 'srt'">
        <label class="text-sm text-gray-400">Latency (ms)</label>
        <input
          v-model.number="form.latency"
          type="number"
          placeholder="200"
          class="mt-1 w-full rounded-lg border border-gray-700 bg-gray-900 px-3 py-2 text-sm text-gray-100 outline-none focus:border-primary-500"
        />
      </div>
    </div>

    <!-- Live toggle -->
    <div class="flex items-center gap-2">
      <input
        id="isLive"
        v-model="form.isLive"
        type="checkbox"
        class="rounded border-gray-700 bg-gray-900"
      />
      <label for="isLive" class="text-sm text-gray-400">Live source (no -re flag)</label>
    </div>

    <!-- Active Stream -->
    <div v-if="activeTask" class="space-y-2">
      <ProgressBar
        :value="0"
        :status="activeTask.status"
        :label="`Stream: ${activeTask.status}`"
        v-bind="currentProgressDetails"
      />
      <LogViewer :lines="currentLogs" />
      <button
        v-if="activeTask.status === 'running'"
        class="rounded-lg bg-red-600 px-4 py-2 text-sm font-medium hover:bg-red-700 transition-colors"
        @click="cancelActiveTask"
      >
        Stop Stream
      </button>
    </div>

    <!-- Start -->
    <button
      v-if="!activeTask || activeTask.status !== 'running'"
      :disabled="!canSubmit"
      class="w-full rounded-lg bg-red-600 py-3 text-sm font-semibold hover:bg-red-700 disabled:opacity-40 disabled:cursor-not-allowed transition-colors"
      @click="handleSubmit"
    >
      🔴 Start Streaming
    </button>
  </div>
</template>

<script setup lang="ts">
const {
  startStream,
  cancelTask,
  getTaskLogs,
  getTaskProgress,
} = useFFmpeg()

const rtmpPresets = [
  { name: 'YouTube', url: 'rtmp://a.rtmp.youtube.com/live2/' },
  { name: 'Twitch', url: 'rtmp://live.twitch.tv/app/' },
  { name: 'Custom', url: '' },
]

const form = reactive({
  source: '',
  sourceURL: '',
  protocol: 'rtmp' as 'rtmp' | 'srt',
  url: '',
  videoCodec: 'libx264',
  preset: 'veryfast',
  bitrate: '3000k',
  latency: 200,
  isLive: false,
})

const activeTask = ref<any>(null)

const effectiveSource = computed(() => form.sourceURL || form.source)
const canSubmit = computed(() => effectiveSource.value && form.url)

const currentLogs = computed(() => {
  if (!activeTask.value) return []
  return getTaskLogs(activeTask.value.id)
})

const currentProgressDetails = computed(() => {
  if (!activeTask.value) return {}
  const p = getTaskProgress(activeTask.value.id)
  if (!p) return {}
  return { fps: p.fps, bitrate: p.bitrate, speed: p.speed, time: p.time }
})

onMounted(() => {
  // Listeners are set up in app.vue
})

async function handleSubmit() {
  if (!canSubmit.value) return

  activeTask.value = await startStream({
    source: effectiveSource.value,
    protocol: form.protocol,
    url: form.url,
    videoCodec: form.videoCodec,
    preset: form.preset,
    bitrate: form.bitrate,
    latency: form.latency,
    isLive: form.isLive,
  })
}

async function cancelActiveTask() {
  if (activeTask.value) {
    await cancelTask(activeTask.value.id)
  }
}
</script>
