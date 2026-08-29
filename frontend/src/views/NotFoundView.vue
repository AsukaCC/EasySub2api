<template>
  <div
    class="views-not-found-view__panel"
  >
    <!-- Background Decoration -->
    <div class="views-not-found-view__panel-2">
      <div
        class="views-not-found-view__panel-3"
      ></div>
      <div
        class="views-not-found-view__panel-4"
      ></div>
    </div>

    <div class="views-not-found-view__panel-5">
      <!-- 404 Display -->
      <div class="views-not-found-view__panel-6">
        <div class="views-not-found-view__panel-7">
          <span class="views-not-found-view__text"
            >404</span
          >
          <div class="views-not-found-view__panel-8">
            <div
              class="views-not-found-view__panel-9"
            >
              <svg
                class="views-not-found-view__icon"
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
                stroke-width="1.5"
              >
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  d="M12 9v3.75m-9.303 3.376c-.866 1.5.217 3.374 1.948 3.374h14.71c1.73 0 2.813-1.874 1.948-3.374L13.949 3.378c-.866-1.5-3.032-1.5-3.898 0L2.697 16.126zM12 15.75h.007v.008H12v-.008z"
                />
              </svg>
            </div>
          </div>
        </div>
      </div>

      <!-- Text Content -->
      <div class="views-not-found-view__panel-6">
        <h1 class="views-not-found-view__heading">
          {{ t('errors.pageNotFound') }}
        </h1>
        <p class="views-not-found-view__description">
          The page you are looking for doesn't exist or has been moved.
        </p>
      </div>

      <!-- Action Buttons -->
      <div class="views-not-found-view__panel-10">
        <button @click="goBack" class="btn btn-secondary">
          <Icon name="arrowLeft" size="md" class="views-not-found-view__icon-2" />
          Go Back
        </button>
        <router-link to="/dashboard" class="btn btn-primary">
          <Icon name="home" size="md" class="views-not-found-view__icon-2" />
          Go to Dashboard
        </router-link>
      </div>

      <!-- Help Link -->
      <p class="views-not-found-view__description-2">
        Need help?
        <a
          href="#"
          class="views-not-found-view__link"
        >
          Contact support
        </a>
      </p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router'
import Icon from '@/components/icons/Icon.vue'

const { t } = useI18n()
const router = useRouter()

function goBack(): void {
  router.back()
}
</script>
