<template>
  <div class="app-layout">
    <!-- Background Decoration -->
    <div class="app-layout__backdrop"></div>
    <div class="app-layout__decorations" aria-hidden="true">
      <div
        class="app-layout__glow app-layout__glow--top"
      ></div>
      <div
        class="app-layout__glow app-layout__glow--bottom"
      ></div>
    </div>

    <AppSidebar v-if="isAdmin" admin-menu-only />

    <div class="app-layout__content">
      <!-- Primary Navigation -->
      <AppTopNav />

      <!-- Main Content -->
      <main class="app-layout__main">
        <slot />
      </main>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { useAuthStore } from '@/stores'
import { useOnboardingTour } from '@/composables/useOnboardingTour'
import { useOnboardingStore } from '@/stores/onboarding'
import AppTopNav from './AppTopNav.vue'
import AppSidebar from './AppSidebar.vue'

const authStore = useAuthStore()
const isAdmin = computed(() => authStore.user?.role === 'admin')

const { replayTour } = useOnboardingTour({
  storageKey: isAdmin.value ? 'admin_guide' : 'user_guide',
  autoStart: true
})

const onboardingStore = useOnboardingStore()

onMounted(() => {
  onboardingStore.setReplayCallback(replayTour)
})

defineExpose({ replayTour })
</script>

<style scoped>
.app-layout {
  min-height: 100vh;
  height: 100dvh;
  overflow: hidden;
}

.app-layout__backdrop,
.app-layout__decorations {
  position: fixed;
  inset: 0;
  pointer-events: none;
}

.app-layout__backdrop {
  background-image:
    radial-gradient(at 40% 20%, rgba(10, 132, 255, 0.12) 0, transparent 50%),
    radial-gradient(at 80% 0, rgba(94, 92, 230, 0.08) 0, transparent 50%),
    radial-gradient(at 0 50%, rgba(10, 132, 255, 0.08) 0, transparent 50%);
}

.app-layout__decorations {
  overflow: hidden;
}

.app-layout__glow {
  position: absolute;
  width: 24rem;
  height: 24rem;
  border-radius: var(--radius-full);
  filter: blur(64px);
}

.app-layout__glow--top {
  top: -8rem;
  right: -8rem;
  background-color: rgba(58, 162, 255, 0.15);
}

:global(.dark) .app-layout__glow--top {
  background-color: rgba(10, 132, 255, 0.1);
}

.app-layout__glow--bottom {
  bottom: -10rem;
  left: -8rem;
  background-color: rgba(94, 92, 230, 0.1);
}

.app-layout__content {
  position: relative;
  display: flex;
  flex-direction: column;
  min-height: 100vh;
  height: 100dvh;
  overflow: hidden;
}

.app-layout__main {
  flex: 1 1 auto;
  width: 100%;
  min-height: 0;
  padding: 1.25rem clamp(1rem, 2vw, 2rem) 2.5rem;
  overflow-y: auto;
  overscroll-behavior-y: contain;
  scrollbar-width: none;
}

.app-layout__main::-webkit-scrollbar {
  display: none;
}

</style>
