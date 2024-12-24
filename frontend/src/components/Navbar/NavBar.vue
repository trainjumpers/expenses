<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import {
  PhHouse,
  PhCodeBlock,
  PhLightbulb,
  PhEnvelope,
  PhQuotes,
  PhPaintBrush
} from '@phosphor-icons/vue'
import type { NavItem } from '@/types/navbar'

// Props and emits
const props = defineProps<{
  name: string
}>()

// Composables
const router = useRouter()
const theme = ref(localStorage.getItem('theme') || 'light')
const isDrawerOpen = ref(false)

// Navigation items
const navItems = ref<NavItem[]>([
  { name: 'Home', section: 'intro', icon: PhHouse },
  { name: 'Skills', section: 'skills', icon: PhPaintBrush },
  { name: 'Projects', section: 'projects', icon: PhCodeBlock },
  { name: 'Experience', section: 'experience', icon: PhLightbulb },
  { name: 'Contact', section: 'contact', icon: PhEnvelope },
  { name: 'Blogs', section: 'blogs', icon: PhQuotes }
])

// Methods
const scrollToSection = (section: string) => {
  const element = document.getElementById(section)
  element?.scrollIntoView({ behavior: 'smooth' })
  isDrawerOpen.value = false
}

const handleNavigation = (section: string) => {
  if (section === 'blogs') {
    router.push('/blogs')
    return
  }

  if (router.currentRoute.value.path !== '/') {
    router.push('/')
    // Wait for route change before scrolling
    setTimeout(() => scrollToSection(section), 100)
    return
  }

  scrollToSection(section)
}

const setTheme = (newTheme: string) => {
  theme.value = newTheme
  localStorage.setItem('theme', newTheme)
  document.documentElement.setAttribute('data-theme', newTheme)
}

// Lifecycle
onMounted(() => {
  setTheme(theme.value)
})
</script>

<template>
  <div class="navbar-container">
    <!-- Mobile drawer toggle -->
    <div class="drawer-overlay" :class="{ active: isDrawerOpen }" @click="isDrawerOpen = false" />

    <!-- Main navbar -->
    <nav class="navbar bg-base-300">
      <div class="navbar-start">
        <button class="btn btn-ghost lg:hidden" @click="isDrawerOpen = !isDrawerOpen">
          <!-- <component :is="PiTextOutdentDuotone" /> -->
        </button>
        <span class="text-xl font-bold px-2">{{ props.name.split(' ')[0] }}</span>
      </div>

      <!-- Desktop navigation -->
      <div class="navbar-center hidden lg:flex">
        <ul class="menu menu-horizontal">
          <li v-for="item in navItems" :key="item.section">
            <button class="btn btn-ghost" @click="handleNavigation(item.section)">
              <component :is="item.icon" class="w-5 h-5" />
              <span>{{ item.name }}</span>
            </button>
          </li>
        </ul>
      </div>

      <!-- Theme selector -->
      <div class="navbar-end">
        <div class="dropdown dropdown-end">
          <button class="btn btn-ghost">
            <!-- <component :is="PiPaletteDuotone" class="w-5 h-5" /> -->
            <span>{{ theme }}</span>
          </button>
          <ul class="dropdown-content menu bg-base-200 rounded-box w-52">
            <li v-for="t in ['light', 'dark', 'system']" :key="t">
              <button @click="setTheme(t)">
                {{ t.charAt(0).toUpperCase() + t.slice(1) }}
              </button>
            </li>
          </ul>
        </div>
      </div>
    </nav>

    <!-- Mobile drawer -->
    <div class="drawer lg:hidden" :class="{ open: isDrawerOpen }">
      <ul class="menu bg-base-200 w-80 p-4">
        <li v-for="item in navItems" :key="item.section">
          <button class="btn btn-ghost" @click="handleNavigation(item.section)">
            <component :is="item.icon" class="w-5 h-5" />
            <span>{{ item.name }}</span>
          </button>
        </li>
      </ul>
    </div>
  </div>
</template>

<style scoped>
.navbar-container {
  position: sticky;
  top: 0;
  z-index: 50;
}

.drawer {
  position: fixed;
  top: 0;
  left: -320px;
  height: 100vh;
  transition: transform 0.3s ease-in-out;
}

.drawer.open {
  transform: translateX(320px);
}

.drawer-overlay {
  display: none;
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.5);
  z-index: 40;
}

.drawer-overlay.active {
  display: block;
}
</style>
