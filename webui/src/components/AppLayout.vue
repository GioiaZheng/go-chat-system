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
  display: grid;
  grid-template-columns: 230px 1fr;
  min-height: 100vh;
  height: 100%;
  width: 100%;
  background: #e5e7eb;
}

.app-layout__sidebar {
  position: static;
  min-height: 100%;
}

.app-layout__main {
  min-width: 0;
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  padding: 0;
}

@media (max-width: 992px) {
  .app-layout {
    grid-template-columns: 1fr;
    grid-template-rows: auto 1fr;
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
