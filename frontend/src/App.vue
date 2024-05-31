<script setup lang="ts">
import { RouterView } from "vue-router";
</script>

<template>
  <header>
    <div>This is a header. We will change this later on.</div>
  </header>
  <RouterView />
</template>

<style scoped>
header {
  display: flex;
  justify-content: center;
  line-height: 1.5;
  max-height: 100vh;
}

@media (min-width: 1024px) {
  header {
    display: flex;
    place-items: center;
  }
}
</style>
