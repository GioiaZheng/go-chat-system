<template>
  <div class="p-2 bg-red-50 text-red-700 border border-red-200 rounded">
    {{ text }}
  </div>
</template>

<script setup>
const props = defineProps({
  text: {
    type: String,
    default: "Error"
  }
});
</script>
