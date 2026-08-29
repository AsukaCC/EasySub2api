<template>
  <div class="components-user-profile-profile-info-card__panel">
    <section
      data-testid="profile-overview-hero"
      class="components-user-profile-profile-info-card__section card"
    >
      <div class="components-user-profile-profile-info-card__panel-2">
        <div class="components-user-profile-profile-info-card__panel-3">
          <div
            class="components-user-profile-profile-info-card__panel-4"
          >
            <img
              v-if="avatarUrl"
              :src="avatarUrl"
              :alt="displayName"
              class="components-user-profile-profile-info-card__image"
            >
            <span v-else>{{ avatarInitial }}</span>
          </div>

          <div class="components-user-profile-profile-info-card__panel-5">
            <div class="components-user-profile-profile-info-card__panel-6">
              <div class="components-user-profile-profile-info-card__panel-7">
                <h2 class="components-user-profile-profile-info-card__heading">
                  {{ displayName }}
                </h2>
                <span :class="['badge', user?.role === 'admin' ? 'badge-primary' : 'badge-gray']">
                  {{ user?.role === 'admin' ? t('profile.administrator') : t('profile.user') }}
                </span>
                <span
                  :class="['badge', user?.status === 'active' ? 'badge-success' : 'badge-danger']"
                >
                  {{
                    user?.status === 'active'
                      ? t('common.active')
                      : t('common.disabled')
                  }}
                </span>
              </div>

              <div class="components-user-profile-profile-info-card__panel-8">
                <p class="components-user-profile-profile-info-card__description">
                  {{ primaryEmailDisplay }}
                </p>
                <div
                  v-if="sourceHints.length"
                  class="components-user-profile-profile-info-card__panel-9"
                >
                  <span
                    v-for="hint in sourceHints"
                    :key="hint.key"
                    class="components-user-profile-profile-info-card__text"
                  >
                    <Icon name="link" size="sm" />
                    {{ hint.text }}
                  </span>
                </div>
              </div>
            </div>

            <div class="components-user-profile-profile-info-card__panel-10">
              <div
                data-testid="profile-overview-metric-balance"
                class="components-user-profile-profile-info-card__panel-11"
              >
                <p class="components-user-profile-profile-info-card__description-2">
                  {{ t('profile.accountBalance') }}
                </p>
                <p class="components-user-profile-profile-info-card__description-3">
                  {{ formatPoints(user?.balance || 0) }}
                </p>
              </div>
			  <div class="components-user-profile-profile-info-card__panel-11">
				<p class="components-user-profile-profile-info-card__description-2">{{ t('common.rechargeBalance') }}</p>
				<p class="components-user-profile-profile-info-card__description-3">{{ formatPoints(user?.recharge_balance ?? user?.balance ?? 0) }}</p>
			  </div>
			  <div class="components-user-profile-profile-info-card__panel-11">
				<p class="components-user-profile-profile-info-card__description-2">{{ t('common.bonusBalance') }}</p>
				<p class="components-user-profile-profile-info-card__description-3">{{ formatPoints(user?.bonus_balance ?? 0) }}</p>
			  </div>
              <div
                data-testid="profile-overview-metric-member-since"
                class="components-user-profile-profile-info-card__panel-11"
              >
                <p class="components-user-profile-profile-info-card__description-2">
                  {{ t('profile.memberSince') }}
                </p>
                <p class="components-user-profile-profile-info-card__description-3">
                  {{ memberSinceLabel }}
                </p>
              </div>
            </div>
          </div>
        </div>
      </div>
    </section>

    <div class="components-user-profile-profile-info-card__panel">
      <div data-testid="profile-main-column" class="components-user-profile-profile-info-card__panel">
        <section
          data-testid="profile-basics-panel"
          class="components-user-profile-profile-info-card__section-2 card"
        >
          <div class="components-user-profile-profile-info-card__panel-12">
            <div>
              <h3 class="components-user-profile-profile-info-card__heading-2">
                {{ t('profile.basicsTitle') }}
              </h3>
              <p class="components-user-profile-profile-info-card__description-4">
                {{ t('profile.basicsDescription') }}
              </p>
            </div>
          </div>

          <div class="components-user-profile-profile-info-card__panel-13">
            <div class="components-user-profile-profile-info-card__panel-14">
              <ProfileAvatarCard
                :user="user"
                embedded
              />
            </div>

            <div class="components-user-profile-profile-info-card__panel-14">
              <ProfileEditForm
                :initial-username="user?.username || ''"
                embedded
              />
            </div>
          </div>
        </section>

        <section
          data-testid="profile-auth-bindings-panel"
          class="components-user-profile-profile-info-card__section-2 card"
        >
          <ProfileIdentityBindingsSection
            :user="user"
            :dingtalk-enabled="dingtalkEnabled"
            :oidc-enabled="oidcEnabled"
            :oidc-provider-name="oidcProviderName"
            :wechat-enabled="wechatEnabled"
            :wechat-open-enabled="wechatOpenEnabled"
            :wechat-mp-enabled="wechatMpEnabled"
            embedded
            compact
          />
        </section>
      </div>

      <div data-testid="profile-side-column" class="components-user-profile-profile-info-card__panel">
        <section
          v-if="sourceHints.length"
          class="components-user-profile-profile-info-card__section-2 card"
        >
          <h3 class="components-user-profile-profile-info-card__heading-2">
            {{ t('profile.linkedProfileSources') }}
          </h3>
          <p class="components-user-profile-profile-info-card__description-4">
            {{ t('profile.linkedProfileSourcesDescription') }}
          </p>

          <div class="components-user-profile-profile-info-card__panel-15">
            <div
              v-for="hint in sourceHints"
              :key="hint.key"
              class="components-user-profile-profile-info-card__panel-16"
            >
              <Icon name="link" size="sm" class="components-user-profile-profile-info-card__icon" />
              <span>{{ hint.text }}</span>
            </div>
          </div>
        </section>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import ProfileAvatarCard from '@/components/user/profile/ProfileAvatarCard.vue'
import ProfileEditForm from '@/components/user/profile/ProfileEditForm.vue'
import ProfileIdentityBindingsSection from '@/components/user/profile/ProfileIdentityBindingsSection.vue'
import type { User, UserAuthBindingStatus, UserAuthProvider, UserProfileSourceContext } from '@/types'
import { formatPoints } from '@/utils/format'

const props = withDefaults(defineProps<{
  user: User | null
  dingtalkEnabled?: boolean
  oidcEnabled?: boolean
  oidcProviderName?: string
  wechatEnabled?: boolean
  wechatOpenEnabled?: boolean
  wechatMpEnabled?: boolean
}>(), {
  dingtalkEnabled: false,
  oidcEnabled: false,
  oidcProviderName: 'OIDC',
  wechatEnabled: false,
  wechatOpenEnabled: undefined,
  wechatMpEnabled: undefined,
})

const { t } = useI18n()

function normalizeBindingStatus(binding: boolean | UserAuthBindingStatus | undefined): boolean | null {
  if (typeof binding === 'boolean') {
    return binding
  }
  if (!binding) {
    return null
  }
  if (typeof binding.bound === 'boolean') {
    return binding.bound
  }
  return Boolean(binding.provider_subject || binding.issuer || binding.provider_key)
}

function isEmailBound(user: User | null | undefined): boolean {
  if (typeof user?.email_bound === 'boolean') {
    return user.email_bound
  }

  const nested = user?.auth_bindings?.email ?? user?.identity_bindings?.email
  const normalized = normalizeBindingStatus(nested)
  return normalized ?? false
}

const avatarUrl = computed(() => props.user?.avatar_url?.trim() || '')
const displayName = computed(() => props.user?.username?.trim() || props.user?.email?.trim() || t('profile.user'))
const primaryEmailDisplay = computed(() => {
  const email = props.user?.email?.trim() || ''
  if (!email) {
    return ''
  }
  if (email.endsWith('.invalid') && !isEmailBound(props.user)) {
    return ''
  }
  return email
})
const avatarInitial = computed(() => displayName.value.charAt(0).toUpperCase() || 'U')
const memberSinceLabel = computed(() => {
  const raw = props.user?.created_at?.trim()
  if (!raw) {
    return '-'
  }

  const date = new Date(raw)
  if (Number.isNaN(date.getTime())) {
    return '-'
  }

  return new Intl.DateTimeFormat(undefined, {
    year: 'numeric',
    month: 'short',
  }).format(date)
})

const providerLabels = computed<Record<UserAuthProvider, string>>(() => ({
  email: t('profile.authBindings.providers.email'),
  dingtalk: t('profile.authBindings.providers.dingtalk'),
  oidc: t('profile.authBindings.providers.oidc', { providerName: props.oidcProviderName }),
  wechat: t('profile.authBindings.providers.wechat'),
  github: 'GitHub',
  google: 'Google'
}))

function normalizeProvider(value: string): UserAuthProvider | null {
  const normalized = value.trim().toLowerCase()
  if (
    normalized === 'email' ||
    normalized === 'wechat' ||
    normalized === 'github' ||
    normalized === 'google'
  ) {
    return normalized
  }
  if (normalized === 'oidc' || normalized.startsWith('oidc:') || normalized.startsWith('oidc/')) {
    return 'oidc'
  }
  return null
}

function readObjectString(source: Record<string, unknown>, ...keys: string[]): string {
  for (const key of keys) {
    const value = source[key]
    if (typeof value === 'string' && value.trim()) {
      return value.trim()
    }
  }
  return ''
}

function resolveThirdPartySource(
  rawSource: string | UserProfileSourceContext | null | undefined
): { provider: UserAuthProvider; label: string } | null {
  if (!rawSource) {
    return null
  }

  if (typeof rawSource === 'string') {
    const provider = normalizeProvider(rawSource)
    if (!provider || provider === 'email') {
      return null
    }
    return {
      provider,
      label: providerLabels.value[provider]
    }
  }

  const sourceRecord = rawSource as Record<string, unknown>
  const provider = normalizeProvider(
    readObjectString(sourceRecord, 'provider', 'source', 'provider_type', 'auth_provider')
  )
  if (!provider || provider === 'email') {
    return null
  }

  const explicitLabel = readObjectString(
    sourceRecord,
    'provider_label',
    'label',
    'provider_name',
    'providerName'
  )

  return {
    provider,
    label: explicitLabel || providerLabels.value[provider]
  }
}

const sourceHints = computed(() => {
  const currentUser = props.user
  if (!currentUser) {
    return []
  }

  const hints: Array<{ key: string; text: string }> = []
  const avatarSource = resolveThirdPartySource(
    currentUser.profile_sources?.avatar ?? currentUser.avatar_source
  )
  const usernameSource = resolveThirdPartySource(
    currentUser.profile_sources?.username ??
      currentUser.profile_sources?.display_name ??
      currentUser.profile_sources?.nickname ??
      currentUser.display_name_source ??
      currentUser.username_source ??
      currentUser.nickname_source
  )

  if (avatarSource) {
    hints.push({
      key: 'avatar',
      text: t('profile.authBindings.source.avatar', { providerName: avatarSource.label })
    })
  }

  if (usernameSource) {
    hints.push({
      key: 'username',
      text: t('profile.authBindings.source.username', { providerName: usernameSource.label })
    })
  }

  return hints
})
</script>
