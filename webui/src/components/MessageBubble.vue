<template>
  <div class="msg" :class="{ mine }">
    <div class="bubble">{{ text }}</div>
    <div class="meta">{{ meta }}</div>
  </div>
</template>

<script setup>
defineProps({
  mine: { type: Boolean, default: false },
  text: { type: String, default: '' },
  meta: { type: String, default: '' },
})
</script>
