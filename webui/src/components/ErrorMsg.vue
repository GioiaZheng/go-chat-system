<template>
  <!-- Simple red error text block -->
  <div v-if="$slots.default" class="err">
    <slot />
  </div>
</template>

<script>
/**
 * ErrorMsg
 * - Minimal error message component.
 * - Renders slot content in a styled red text block.
 */
export default { name: "ErrorMsg" };
</script>

<style scoped>
.err {
  color: crimson;
  padding: 8px 0;
}
</style>
