<template>
  <div class="mx-auto max-w-3xl space-y-6">
    <div>
      <h2 class="text-2xl font-bold">Convert</h2>
      <p class="text-sm text-gray-400 mt-1">Video / audio conversion and processing</p>
    </div>

    <!-- File Selection -->
    <div class="space-y-3">
      <FileSelector v-model="form.input" label="Input File" placeholder="Select input file..." />
      <FileSelector v-model="form.output" label="Output File" type="directory" placeholder="Select output directory..." />
    </div>

    <!-- Output Filename -->
    <div>
      <label class="text-sm text-gray-400">Output Filename</label>
      <input
        v-model="form.outputName"
        type="text"
        placeholder="output.mp4"
        class="mt-1 w-full rounded-lg border border-gray-700 bg-gray-900 px-3 py-2 text-sm text-gray-100 outline-none focus:border-primary-500"
      />
    </div>

    <!-- Video Settings -->
    <div class="grid grid-cols-2 gap-4">
      <div>
        <label class="text-sm text-gray-400">Video Codec</label>
        <select v-model="form.videoCodec" class="mt-1 w-full rounded-lg border border-gray-700 bg-gray-900 px-3 py-2 text-sm text-gray-100">
          <option value="">Auto</option>
          <option value="libx264">H.264 (libx264)</option>
          <option value="libx265">H.265 (libx265)</option>
          <option value="libvpx-vp9">VP9</option>
          <option value="libaom-av1">AV1</option>
          <option value="copy">Copy (no re-encode)</option>
        </select>
      </div>
      <div>
        <label class="text-sm text-gray-400">Audio Codec</label>
        <select v-model="form.audioCodec" class="mt-1 w-full rounded-lg border border-gray-700 bg-gray-900 px-3 py-2 text-sm text-gray-100">
          <option value="">Auto</option>
          <option value="aac">AAC</option>
          <option value="libmp3lame">MP3</option>
          <option value="libopus">Opus</option>
          <option value="copy">Copy</option>
        </select>
      </div>
    </div>

    <div class="grid grid-cols-3 gap-4">
      <div>
        <label class="text-sm text-gray-400">Resolution</label>
        <select v-model="form.resolution" class="mt-1 w-full rounded-lg border border-gray-700 bg-gray-900 px-3 py-2 text-sm text-gray-100">
          <option value="">Original</option>
          <option value="1920:1080">1080p</option>
          <option value="1280:720">720p</option>
          <option value="854:480">480p</option>
          <option value="640:360">360p</option>
        </select>
      </div>
      <div>
        <label class="text-sm text-gray-400">FPS</label>
        <input
          v-model.number="form.fps"
          type="number"
          min="1"
          max="120"
          placeholder="Original"
          class="mt-1 w-full rounded-lg border border-gray-700 bg-gray-900 px-3 py-2 text-sm text-gray-100 outline-none focus:border-primary-500"
        />
      </div>
      <div>
        <label class="text-sm text-gray-400">CRF</label>
        <input
          v-model.number="form.crf"
          type="number"
          min="0"
          max="51"
          placeholder="23"
          class="mt-1 w-full rounded-lg border border-gray-700 bg-gray-900 px-3 py-2 text-sm text-gray-100 outline-none focus:border-primary-500"
        />
      </div>
    </div>

    <div class="grid grid-cols-2 gap-4">
      <div>
        <label class="text-sm text-gray-400">Video Bitrate</label>
        <input
          v-model="form.bitrate"
          type="text"
          placeholder="e.g. 3000k"
          class="mt-1 w-full rounded-lg border border-gray-700 bg-gray-900 px-3 py-2 text-sm text-gray-100 outline-none focus:border-primary-500"
        />
      </div>
      <div>
        <label class="text-sm text-gray-400">Audio Bitrate</label>
        <input
          v-model="form.audioBitrate"
          type="text"
          placeholder="e.g. 128k"
          class="mt-1 w-full rounded-lg border border-gray-700 bg-gray-900 px-3 py-2 text-sm text-gray-100 outline-none focus:border-primary-500"
        />
      </div>
    </div>

    <!-- Subtitle -->
    <div>
      <FileSelector v-model="form.subtitleFile" label="Subtitle File (optional)" placeholder="Select subtitle file (.srt, .ass)..." />
    </div>

    <!-- Extra Args -->
    <div>
      <label class="text-sm text-gray-400">Extra FFmpeg Arguments</label>
      <input
        v-model="form.extraArgs"
        type="text"
        placeholder="e.g. -vf eq=brightness=0.1"
        class="mt-1 w-full rounded-lg border border-gray-700 bg-gray-900 px-3 py-2 text-sm text-gray-100 outline-none focus:border-primary-500"
      />
    </div>

    <!-- Active Task Progress -->
    <div v-if="activeTask" class="space-y-2">
      <ProgressBar
        :value="activeTask.progress"
        :status="activeTask.status"
        v-bind="currentProgressDetails"
      />
      <LogViewer :lines="currentLogs" />
      <button
        v-if="activeTask.status === 'running'"
        class="rounded-lg bg-red-600 px-4 py-2 text-sm font-medium hover:bg-red-700 transition-colors"
        @click="cancelActiveTask"
      >
        Cancel
      </button>
    </div>

    <!-- Submit -->
    <button
      v-if="!activeTask || activeTask.status !== 'running'"
      :disabled="!canSubmit"
      class="w-full rounded-lg bg-primary-600 py-3 text-sm font-semibold hover:bg-primary-700 disabled:opacity-40 disabled:cursor-not-allowed transition-colors"
      @click="handleSubmit"
    >
      Start Conversion
    </button>
  </div>
</template>

<script setup lang="ts">
const {
  startConvert,
  cancelTask,
  selectDirectory,
  getTaskLogs,
  getTaskProgress,
  setupListeners,
} = useFFmpeg()

const form = reactive({
  input: '',
  output: '',
  outputName: 'output.mp4',
  videoCodec: '',
  audioCodec: '',
  resolution: '',
  fps: 0,
  crf: 0,
  bitrate: '',
  audioBitrate: '',
  subtitleFile: '',
  extraArgs: '',
})

const activeTask = ref<any>(null)

const canSubmit = computed(() => form.input && form.output && form.outputName)

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
  setupListeners()
})

async function handleSubmit() {
  if (!canSubmit.value) return

  const outputPath = form.output.endsWith('/')
    ? `${form.output}${form.outputName}`
    : `${form.output}/${form.outputName}`

  activeTask.value = await startConvert({
    input: form.input,
    output: outputPath,
    videoCodec: form.videoCodec || undefined,
    audioCodec: form.audioCodec || undefined,
    resolution: form.resolution || undefined,
    fps: form.fps || undefined,
    crf: form.crf || undefined,
    bitrate: form.bitrate || undefined,
    audioBitrate: form.audioBitrate || undefined,
    subtitleFile: form.subtitleFile || undefined,
    extraArgs: form.extraArgs || undefined,
  })
}

async function cancelActiveTask() {
  if (activeTask.value) {
    await cancelTask(activeTask.value.id)
  }
}
</script>
