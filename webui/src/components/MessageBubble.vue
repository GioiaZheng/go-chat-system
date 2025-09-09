<template>
  <div :class="['flex', mine ? 'justify-end' : 'justify-start']">
    <div :class="bubbleCls">
      <div class="text-sm whitespace-pre-wrap">{{ text }}</div>
      <div class="text-[11px] opacity-70 mt-1">{{ meta }}</div>
    </div>
  </div>
</template>

<script setup>
// English: minimal bubble
const props = defineProps({
  mine: { type: Boolean, default: false },
  text: { type: String, default: "" },
  meta: { type: String, default: "" },
});

const bubbleCls = computed(() =>
  props.mine
    ? "max-w-[80%] bg-blue-600 text-white p-2 rounded-lg"
    : "max-w-[80%] bg-gray-200 text-gray-900 p-2 rounded-lg"
);
</script>
