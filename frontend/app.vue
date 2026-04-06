<template>
  <div class="min-h-screen bg-gray-950 text-gray-100">
    <nav class="flex items-center gap-6 border-b border-gray-800 px-6 py-3">
      <h1 class="text-lg font-bold text-primary-400">FFmpeg-Wails</h1>
      <NuxtLink to="/convert" class="nav-link" active-class="text-primary-400">
        Convert
      </NuxtLink>
      <NuxtLink to="/stream" class="nav-link" active-class="text-primary-400">
        Stream
      </NuxtLink>
      <NuxtLink to="/tasks" class="nav-link" active-class="text-primary-400">
        Tasks
      </NuxtLink>
    </nav>
    <main class="p-6">
      <NuxtPage />
    </main>
  </div>
</template>

<script setup lang="ts">
const { setupListeners, fetchVersion } = useFFmpeg()

onMounted(() => {
  setupListeners()
  fetchVersion()
})
</script>

<style>
.nav-link {
  @apply text-sm text-gray-400 hover:text-gray-100 transition-colors;
}
</style>
