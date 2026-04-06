<template>
  <div class="flex flex-col gap-2">
    <label v-if="label" class="text-sm text-gray-400">{{ label }}</label>
    <div class="flex gap-2">
      <input
        :value="modelValue"
        type="text"
        readonly
        :placeholder="placeholder"
        class="flex-1 rounded-lg border border-gray-700 bg-gray-900 px-3 py-2 text-sm text-gray-100 outline-none focus:border-primary-500"
        @click="handleSelect"
      />
      <button
        class="rounded-lg bg-primary-600 px-4 py-2 text-sm font-medium hover:bg-primary-700 transition-colors"
        @click="handleSelect"
      >
        Browse
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
const props = withDefaults(defineProps<{
  modelValue: string
  label?: string
  placeholder?: string
  type?: 'file' | 'directory'
}>(), {
  label: '',
  placeholder: 'Select a file...',
  type: 'file',
})

const emit = defineEmits<{
  'update:modelValue': [value: string]
}>()

const { selectFile, selectDirectory } = useFFmpeg()

async function handleSelect() {
  const path = props.type === 'directory'
    ? await selectDirectory()
    : await selectFile()
  if (path) {
    emit('update:modelValue', path)
  }
}
</script>
