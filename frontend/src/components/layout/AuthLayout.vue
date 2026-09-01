<template>
  <div class="components-layout-auth-layout__panel">
    <!-- Background -->
    <div
      class="components-layout-auth-layout__panel-2"
    ></div>

    <!-- Decorative Elements -->
    <div class="components-layout-auth-layout__panel-3">
      <!-- Gradient Orbs -->
      <div
        class="components-layout-auth-layout__panel-4"
      ></div>
      <div
        class="components-layout-auth-layout__panel-5"
      ></div>
      <div
        class="components-layout-auth-layout__panel-6"
      ></div>

      <!-- Grid Pattern -->
      <div
        class="components-layout-auth-layout__panel-7"
      ></div>
    </div>

    <!-- Content Container -->
    <div class="components-layout-auth-layout__panel-8">
      <!-- Logo/Brand -->
      <div class="components-layout-auth-layout__panel-9">
        <!-- Custom Logo or Default Logo -->
        <template v-if="settingsLoaded">
          <div
            class="components-layout-auth-layout__panel-10"
          >
            <img :src="siteLogo || '/logo.svg'" alt="Logo" class="components-layout-auth-layout__image" />
          </div>
          <h1 class="components-layout-auth-layout__heading text-gradient">
            {{ siteName }}
          </h1>
          <p class="components-layout-auth-layout__description">
            {{ siteSubtitle }}
          </p>
        </template>
      </div>

      <!-- Card Container -->
      <div class="components-layout-auth-layout__panel-11 card-glass">
        <slot />
      </div>

      <!-- Footer Links -->
      <div class="components-layout-auth-layout__panel-12">
        <slot name="footer" />
      </div>

      <!-- Copyright -->
      <div class="components-layout-auth-layout__panel-13">
        &copy; {{ currentYear }} {{ siteName }}. All rights reserved.
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { useAppStore } from '@/stores'
import { sanitizeUrl } from '@/utils/url'

const appStore = useAppStore()

const siteName = computed(() => appStore.siteName || 'EasySub2api')
const siteLogo = computed(() => sanitizeUrl(appStore.siteLogo || '', { allowRelative: true, allowDataUrl: true }))
const siteSubtitle = computed(() => appStore.cachedPublicSettings?.site_subtitle || 'Subscription to API Conversion Platform')
const settingsLoaded = computed(() => appStore.publicSettingsLoaded)

const currentYear = computed(() => new Date().getFullYear())

onMounted(() => {
  appStore.fetchPublicSettings()
})
</script>

<style scoped>
.text-gradient {
  background: linear-gradient(90deg, color-mix(in srgb, var(--theme-accent) 88%, black), var(--theme-accent));
  -webkit-background-clip: text;
  background-clip: text;
  color: transparent;
}
</style>
