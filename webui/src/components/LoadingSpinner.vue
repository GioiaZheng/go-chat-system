<!-- Notes:
     - Minimal loading indicator component.
     - Kept intentionally small to avoid layout shifts.
     - Can optionally accept a `text` prop for accessibility/UX. -->

<template>
  <span class="inline-block animate-spin" aria-live="polite" aria-busy="true">⏳</span>
  <span v-if="text" class="ml-2 text-sm opacity-75">{{ text }}</span>
</template>

<script setup>
const props = defineProps({
  text: { type: String, default: "" }
});
</script>

<style scoped>
/* Fallback for environments without Tailwind: simple spin animation */
@keyframes spin {
  to { transform: rotate(360deg); }
}
.animate-spin {
  display: inline-block;
  animation: spin 1s linear infinite;
}
</style>
