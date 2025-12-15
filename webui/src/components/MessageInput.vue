<!-- Compact message composer; emits "send" with trimmed text and resets the field after submission. -->

<template>
  <form @submit.prevent="submit" class="flex gap-2">
    <input
      v-model="text"
      class="flex-1 input"
      type="text"
      autocomplete="off"
      placeholder="Type a message..."
    />
    <button class="btn" type="submit">Send</button>
  </form>
</template>

<script setup>
import { ref } from 'vue'

const emit = defineEmits(['send'])
const text = ref('')

function submit() {
  const value = text.value.trim()
  if (!value) return
  emit('send', value)
  text.value = ''
}
</script>

<style scoped>
.input {
  width: 100%;
  border: 1px solid #ddd;
  padding: 0.5rem 0.75rem;
  border-radius: 0.375rem;
}

.btn {
  background: #111827;
  color: #fff;
  padding: 0.5rem 0.75rem;
  border-radius: 0.375rem;
  border: none;
}
</style>
