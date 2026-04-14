import type { Ref } from 'vue'

export interface Task {
  id: string
  type: string
  command: string
  status: string
  progress: number
  input: string
  output: string
  createdAt: string
  error?: string
  warning?: string
}

export interface ConvertPayload {
  input: string
  output: string
  videoCodec?: string
  audioCodec?: string
  resolution?: string
  fps?: number
  crf?: number
  bitrate?: string
  audioBitrate?: string
  subtitleFile?: string
  format?: string
  extraArgs?: string
}

export interface StreamPayload {
  source: string
  protocol: string
  url: string
  videoCodec?: string
  audioCodec?: string
  bitrate?: string
  preset?: string
  latency?: number
  isLive?: boolean
}

export interface ProgressEvent {
  id: string
  progress?: number
  fps?: number
  bitrate?: string
  time?: string
  frame?: number
  speed?: string
}

export interface LogEvent {
  id: string
  line: string
}

// Wails runtime type declarations
declare global {
  interface Window {
    go?: {
      backend?: {
        App?: {
          StartTask: (payload: string) => Promise<Task>
          GetTaskStatus: (id: string) => Promise<Task>
          CancelTask: (id: string) => Promise<void>
          ListTasks: () => Promise<Task[]>
          GetFFmpegVersion: () => Promise<string>
          SelectFile: () => Promise<string>
          SelectDirectory: () => Promise<string>
        }
      }
    }
    runtime?: {
      EventsOn: (event: string, callback: (...args: any[]) => void) => void
      EventsOff: (event: string) => void
    }
  }
}

function getWailsBindings() {
  const app = window.go?.backend?.App
  if (!app) {
    return null
  }
  return app
}

// Shared state (module-level, singleton across all component instances)
const tasks: Ref<Task[]> = ref<Task[]>([])
const logs = ref<Map<string, string[]>>(new Map())
const progress = ref<Map<string, ProgressEvent>>(new Map())
const ffmpegVersion = ref('')
const listenersReady = ref(false)

export function useFFmpeg() {
  const isWailsReady = computed(() => !!getWailsBindings())

  function setupListeners() {
    if (listenersReady.value) return
    const rt = window.runtime
    if (!rt) return

    rt.EventsOn('task:progress', (data: ProgressEvent) => {
      progress.value.set(data.id, data)
    })

    rt.EventsOn('task:log', (data: LogEvent) => {
      const existing = logs.value.get(data.id) || []
      existing.push(data.line)
      logs.value.set(data.id, existing)
    })

    rt.EventsOn('task:done', (data: { id: string; status: string; error?: string }) => {
      refreshTasks()
    })

    listenersReady.value = true
  }

  async function startConvert(payload: ConvertPayload): Promise<Task | null> {
    const app = getWailsBindings()
    if (!app) return null
    const task = await app.StartTask(JSON.stringify(payload))
    await refreshTasks()
    return task
  }

  async function startStream(payload: StreamPayload): Promise<Task | null> {
    const app = getWailsBindings()
    if (!app) return null
    const task = await app.StartTask(JSON.stringify(payload))
    await refreshTasks()
    return task
  }

  async function cancelTask(id: string): Promise<void> {
    const app = getWailsBindings()
    if (!app) return
    await app.CancelTask(id)
    await refreshTasks()
  }

  async function refreshTasks(): Promise<void> {
    const app = getWailsBindings()
    if (!app) return
    tasks.value = await app.ListTasks()
  }

  async function selectFile(): Promise<string> {
    const app = getWailsBindings()
    if (!app) return ''
    return await app.SelectFile()
  }

  async function selectDirectory(): Promise<string> {
    const app = getWailsBindings()
    if (!app) return ''
    return await app.SelectDirectory()
  }

  async function fetchVersion(): Promise<void> {
    const app = getWailsBindings()
    if (!app) return
    ffmpegVersion.value = await app.GetFFmpegVersion()
  }

  function getTaskLogs(id: string): string[] {
    return logs.value.get(id) || []
  }

  function getTaskProgress(id: string): ProgressEvent | undefined {
    return progress.value.get(id)
  }

  return {
    tasks: readonly(tasks),
    ffmpegVersion: readonly(ffmpegVersion),
    isWailsReady,
    setupListeners,
    startConvert,
    startStream,
    cancelTask,
    refreshTasks,
    selectFile,
    selectDirectory,
    fetchVersion,
    getTaskLogs,
    getTaskProgress,
  }
}
