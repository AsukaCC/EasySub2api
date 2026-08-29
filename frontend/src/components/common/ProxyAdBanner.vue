<template>
  <a
    class="proxy-ad-banner"
    href="https://sub2api.io/proxyip"
    target="_blank"
    rel="noopener noreferrer"
  >
    <span class="proxy-ad-banner__label">{{ t('admin.proxies.ad.inline') }}</span>
    <Icon name="externalLink" size="xs" />
  </a>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'

const { t } = useI18n()
</script>

<style scoped>
.proxy-ad-banner {
  display: inline-flex;
  max-width: 100%;
  flex-shrink: 0;
  align-items: center;
  gap: 0.25rem;
  overflow: hidden;
  color: var(--color-text-brand);
  font-size: var(--type-caption-size);
  font-weight: var(--font-weight-normal);
  line-height: var(--type-caption-line-height);
  text-overflow: ellipsis;
  white-space: nowrap;
  transition: color 150ms ease, backdrop-filter 150ms ease;
}

.proxy-ad-banner__label {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.proxy-ad-banner:hover {
  text-decoration: underline;
  backdrop-filter: blur(var(--glass-layer-inset-blur-hover)) saturate(var(--glass-saturate-hover));
}

.proxy-ad-banner:focus-visible {
  border-radius: var(--radius-sm);
  outline: 2px solid var(--color-primary);
  outline-offset: 2px;
}
</style>
