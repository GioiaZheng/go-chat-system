<template>
  <div class="app-layout">
    <Sidebar class="app-layout__sidebar" />
    <main class="app-layout__main">
      <slot />
    </main>
  </div>
</template>

<script setup>
import Sidebar from '@/components/Sidebar.vue'
</script>

<style scoped>
.app-layout {
  display: flex;
  min-height: 100vh;
  height: 100vh;
  width: 100%;
  background: #f7fafc;
}

.app-layout__sidebar {
  flex: 0 0 230px;
  min-height: 100%;
}

.app-layout__main {
  flex: 1 1 auto;
  min-width: 0;
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  padding: 16px;
  overflow: hidden;
}

@media (max-width: 992px) {
  .app-layout {
    flex-direction: column;
  }

  .app-layout__sidebar {
    width: 100%;
  }

  .app-layout__main {
    min-height: auto;
    padding: 12px;
  }
}
</style>
