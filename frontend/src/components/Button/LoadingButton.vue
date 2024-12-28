<template>
  <button :class="className" :disabled="loading" @click="handleClick">
    <span v-if="loading" class="loading loading-spinner"></span>
    <span v-else><slot></slot></span>
  </button>
</template>

<script setup lang="ts">
defineProps<{
  loading: boolean;
  className: string;
  handleClick: () => void;
}>();
</script>
