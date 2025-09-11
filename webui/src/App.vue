<!-- Reviewer Notes (2025-09-11):
     - Keep the root component minimal: just render <router-view/>.
     - Added a global body style reset for consistent typography and spacing.
     - This file deliberately avoids extra logic to focus routing and styles. -->

<template>
  <router-view></router-view>
</template>

<script>
export default {
  name: 'App'
}
</script>

<style>
/* Global style reset */
body {
  margin: 0;
  padding: 0;
  font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
}
</style>
