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

    <div class="app-layout__content">
      <!-- Primary Navigation -->
      <AppTopNav />

      <!-- Main Content -->
      <main class="app-layout__main">
        <AppSidebar v-if="isAdmin" admin-menu-only />
        <div class="app-layout__page">
          <slot />
        </div>
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
    radial-gradient(at 40% 20%, color-mix(in srgb, var(--theme-accent) 6%, transparent) 0, transparent 50%),
    radial-gradient(at 80% 0, color-mix(in srgb, var(--theme-accent) 4%, transparent) 0, transparent 50%),
    radial-gradient(at 0 50%, rgba(255, 255, 255, 0.4) 0, transparent 50%);
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
  background-color: color-mix(in srgb, var(--theme-accent) 8%, transparent);
}

:global(.dark) .app-layout__glow--top {
  background-color: color-mix(in srgb, var(--theme-accent) 10%, transparent);
}

.app-layout__glow--bottom {
  bottom: -10rem;
  left: -8rem;
  background-color: color-mix(in srgb, var(--theme-accent) 5%, transparent);
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
  --app-page-padding-top: 1.25rem;
  --app-page-padding-bottom: 2.5rem;

  display: flex;
  align-items: stretch;
  gap: 1rem;
  flex: 1 1 auto;
  width: 100%;
  min-height: 0;
  padding: 0 clamp(1rem, 2vw, 2rem);
  overflow: hidden;
}

.app-layout__main > :global(#admin-sidebar) {
  align-self: stretch;
  min-height: 0;
  margin-block: 1rem;
}

.app-layout__page {
  flex: 1 1 auto;
  min-width: 0;
  min-height: 0;
  padding-block: var(--app-page-padding-top) var(--app-page-padding-bottom);
  overflow-y: auto;
  overscroll-behavior-y: contain;
  scroll-padding-block: var(--app-page-padding-top) var(--app-page-padding-bottom);
  scrollbar-width: none;
}

.app-layout__page::-webkit-scrollbar {
  display: none;
}

@media (max-width: 1023px) {
  .app-layout__main > :global(#admin-sidebar) {
    margin-block: 0;
  }
}

</style>
