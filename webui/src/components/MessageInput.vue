<!-- Notes:
     - Simple input + send button; emits "send" with trimmed text.
     - Keeps internal state minimal; styling is scoped to avoid bleed. -->

<template>
  <form @submit.prevent="submit" class="flex gap-2">
    <input v-model="text" class="flex-1 input" placeholder="Type a message..." />
    <button class="btn">Send</button>
  </form>
</template>

<script setup>
import { ref } from "vue";
const emit = defineEmits(["send"]);
const text = ref("");

function submit() {
  const t = text.value.trim();
  if (!t) return;
  emit("send", t);
  text.value = "";
}
</script>

<style scoped>
.input { width:100%; border:1px solid #ddd; padding:.5rem .75rem; border-radius:.375rem; }
.btn { background:#111827; color:#fff; padding:.5rem .75rem; border-radius:.375rem; }
</style>
