export default defineNuxtConfig({
  compatibilityDate: '2025-01-01',
  future: {
    compatibilityVersion: 4,
  },
  modules: ['@unocss/nuxt'],
  css: ['@unocss/reset/tailwind.css'],
  app: {
    head: {
      title: 'FFmpeg-Wails',
      meta: [
        { charset: 'utf-8' },
        { name: 'viewport', content: 'width=device-width, initial-scale=1' },
      ],
    },
  },
})
