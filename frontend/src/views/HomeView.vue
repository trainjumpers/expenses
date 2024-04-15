<script setup lang="ts">
import { fetchSomething } from "@/api/main";
import { useCounterStore } from "@/stores/counter";
import { Button } from "ant-design-vue";
import { computed, ref } from "vue";

const counterStore = useCounterStore();

const loading = ref(false);
const fetchedData = ref(null);

// When fetching from store, use computed.
// Otherwise state is not reactive.
const count = computed(() => counterStore.count);

const handleLoadingToggle = async () => {
  loading.value = true;
  const data = await fetchSomething();
  loading.value = false;
  fetchedData.value = data;
  counterStore.increment();
};
</script>

<template>
  <p>This is the home page</p>
  <Button style="margin: 10px" type="dashed" :loading="loading"
    >Primary Button</Button
  >
  <Button style="margin: 10px" @click="handleLoadingToggle" type="primary"
    >Toggle Loading</Button
  >
  <p v-if="fetchedData">{{ fetchedData }}</p>
  <p>Counter: {{ count }}</p>
</template>
