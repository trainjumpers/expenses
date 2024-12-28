<script setup lang="ts">
import Navbar from "@/components/Navbar/NavBar.vue";
</script>

<template>
  <Navbar />
  <div className="container mx-auto px-4 py-8">
    <h1 className="text-3xl font-bold mb-4">Welcome to Expense Tracker</h1>
    <p className="text-lg mb-4">
      This is a simple expense tracker application built with Vue.js and
      Tailwind CSS.
    </p>
  </div>
  <router-view />
</template>
