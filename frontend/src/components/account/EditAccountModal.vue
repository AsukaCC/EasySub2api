<template>
  <BaseDialog
    :show="show"
    :title="t('admin.accounts.editAccount')"
    width="wide"
    @close="handleClose"
  >
    <form
      v-if="account"
      id="edit-account-form"
      @submit.prevent="handleSubmit"
      class="components-account-edit-account-modal__form"
    >
      <div>
        <label class="input-label">{{ t('common.name') }}</label>
        <input v-model="form.name" type="text" required class="input" data-tour="edit-account-form-name" />
      </div>
      <div>
        <label class="input-label">{{ t('admin.accounts.notes') }}</label>
        <textarea
          v-model="form.notes"
          rows="3"
          class="input"
          :placeholder="t('admin.accounts.notesPlaceholder')"
        ></textarea>
        <p class="input-hint">{{ t('admin.accounts.notesHint') }}</p>
      </div>

      <!-- API Key fields (only for apikey type) -->
      <div v-if="account.type === 'apikey'" class="components-account-edit-account-modal__panel">
        <div>
          <label class="input-label">{{ t('admin.accounts.baseUrl') }}</label>
          <input
            v-model="editBaseUrl"
            type="text"
            class="input"
            :placeholder="
              account.platform === 'openai'
                ? 'https://api.openai.com'
                : account.platform === 'gemini'
                  ? 'https://generativelanguage.googleapis.com'
                  : account.platform === 'antigravity'
                    ? 'https://cloudcode-pa.googleapis.com'
                : account.platform === 'grok'
                  ? 'https://api.x.ai/v1'
                  : 'https://api.anthropic.com'
            "
          />
          <p v-if="baseUrlHint" class="input-hint">{{ baseUrlHint }}</p>
          <GrokBaseUrlPresets
            v-if="account.platform === 'grok'"
            class="components-account-edit-account-modal__grok-base-url-presets"
            @select="editBaseUrl = $event"
          />
          <CnBaseUrlPresets
            v-if="isCNApiKeyAccount"
            class="components-account-edit-account-modal__grok-base-url-presets"
            :platform="cnPresetPlatform"
            :mode="editAccountMode"
            :protocol="editApiProtocol"
            :current-url="editBaseUrl"
            @select="onCnPresetSelect"
          />
        </div>
        <!-- Account Mode Selection (CN providers) -->
        <div v-if="isCNApiKeyAccount">
          <label class="input-label">{{ t('admin.accounts.cnProviders.accountMode.title') }}</label>
          <div class="components-account-edit-account-modal__panel-2">
            <button
              v-for="opt in cnAccountModeOptions"
              :key="opt.value"
              type="button"
              :class="[
                'components-account-edit-account-modal__action-12',
                editAccountMode === opt.value
                  ? 'components-account-edit-account-modal__action-13'
                  : 'components-account-edit-account-modal__action-14'
              ]"
              @click="editAccountMode = opt.value"
            >
              {{ t(`admin.accounts.cnProviders.accountMode.${opt.labelKey}`) }}
            </button>
          </div>
          <p class="input-hint">{{ t(`admin.accounts.cnProviders.accountMode.${editAccountMode}Desc`) }}</p>
        </div>
        <!-- API Protocol Selection (CN providers) -->
        <div v-if="isCNApiKeyAccount">
          <label class="input-label">{{ t('admin.accounts.cnProviders.apiProtocol.title') }}</label>
          <div class="components-account-edit-account-modal__panel-2">
            <button
              v-for="opt in cnProtocolOptions"
              :key="opt.value"
              type="button"
              :class="[
                'components-account-edit-account-modal__action-12',
                editApiProtocol === opt.value
                  ? 'components-account-edit-account-modal__action-13'
                  : 'components-account-edit-account-modal__action-14'
              ]"
              @click="editApiProtocol = opt.value"
            >
              {{ t(`admin.accounts.cnProviders.apiProtocol.${opt.labelKey}`) }}
            </button>
          </div>
          <p class="input-hint">{{ t(`admin.accounts.cnProviders.apiProtocol.${cnProtocolDescKey}Desc`) }}</p>
        </div>
        <div>
          <label class="input-label">{{ t('admin.accounts.apiKey') }}</label>
          <input
            v-model="editApiKey"
            type="password"
            class="components-account-edit-account-modal__field input"
            autocomplete="new-password"
            data-1p-ignore
            data-lpignore="true"
            data-bwignore="true"
            :placeholder="
              account.platform === 'openai'
                ? 'sk-proj-...'
                : account.platform === 'gemini'
                  ? 'AIza...'
                  : account.platform === 'antigravity'
                    ? 'API Key'
                : account.platform === 'grok'
                  ? 'xai-...'
                  : 'sk-ant-...'
            "
          />
          <p class="input-hint">{{ t('admin.accounts.leaveEmptyToKeep') }}</p>
        </div>

        <!-- Model Restriction Section -->
        <div class="components-account-edit-account-modal__panel-3">
          <label class="input-label">{{ t('admin.accounts.modelRestriction') }}</label>
          <AccountModelRuleSelector
            :platform="account?.platform || 'anthropic'"
            :disabled="isOpenAIModelRestrictionDisabled"
            :has-existing-mappings="hasModelRestrictionValues"
            @apply="applyAccountModelRule"
          />

          <div
            v-if="isOpenAIModelRestrictionDisabled"
            class="components-account-edit-account-modal__panel-4"
          >
            <p class="components-account-edit-account-modal__description">
              {{ t('admin.accounts.openai.modelRestrictionDisabledByPassthrough') }}
            </p>
          </div>

          <template v-else>
            <!-- Mode Toggle -->
            <div class="components-account-edit-account-modal__panel-5">
              <button
                type="button"
                @click="modelRestrictionMode = 'whitelist'"
                :class="[
                  'components-account-edit-account-modal__action-15',
                  modelRestrictionMode === 'whitelist'
                    ? 'components-account-edit-account-modal__action-16'
                    : 'components-account-edit-account-modal__action-17'
                ]"
              >
                <svg
                  class="components-account-edit-account-modal__icon"
                  fill="none"
                  viewBox="0 0 24 24"
                  stroke="currentColor"
                >
                  <path
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    stroke-width="2"
                    d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"
                  />
                </svg>
                {{ t('admin.accounts.modelWhitelist') }}
              </button>
              <button
                type="button"
                @click="modelRestrictionMode = 'mapping'"
                :class="[
                  'components-account-edit-account-modal__action-15',
                  modelRestrictionMode === 'mapping'
                    ? 'components-account-edit-account-modal__action-18'
                    : 'components-account-edit-account-modal__action-17'
                ]"
              >
                <svg
                  class="components-account-edit-account-modal__icon"
                  fill="none"
                  viewBox="0 0 24 24"
                  stroke="currentColor"
                >
                  <path
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    stroke-width="2"
                    d="M8 7h12m0 0l-4-4m4 4l-4 4m0 6H4m0 0l4 4m-4-4l4-4"
                  />
                </svg>
                {{ t('admin.accounts.modelMapping') }}
              </button>
            </div>

            <!-- Whitelist Mode -->
            <div v-if="modelRestrictionMode === 'whitelist'">
              <ModelWhitelistSelector v-model="allowedModels" :platform="account?.platform || 'anthropic'" :account-id="account?.id" />
              <p class="components-account-edit-account-modal__description-2">
                {{ t('admin.accounts.selectedModels', { count: allowedModels.length }) }}
                <span v-if="allowedModels.length === 0 && modelMappings.length === 0">{{
                  t('admin.accounts.supportsAllModels')
                }}</span>
              </p>
            </div>

            <!-- Mapping Mode -->
            <div v-else>
              <div class="components-account-edit-account-modal__panel-6">
                <p class="components-account-edit-account-modal__description-3">
                  <svg
                    class="components-account-edit-account-modal__icon-2"
                    fill="none"
                    viewBox="0 0 24 24"
                    stroke="currentColor"
                  >
                    <path
                      stroke-linecap="round"
                      stroke-linejoin="round"
                      stroke-width="2"
                      d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
                    />
                  </svg>
                  {{ t('admin.accounts.mapRequestModels') }}
                </p>
              </div>

            <!-- Model Mapping List -->
            <div v-if="modelMappings.length > 0" class="components-account-edit-account-modal__panel-7">
              <div
                v-for="(mapping, index) in modelMappings"
                :key="getModelMappingKey(mapping)"
                class="components-account-edit-account-modal__panel-8"
              >
                <input
                  v-model="mapping.from"
                  type="text"
                  class="components-account-edit-account-modal__field-2 input"
                  :placeholder="t('admin.accounts.requestModel')"
                />
                <svg
                  class="components-account-edit-account-modal__icon-3"
                  fill="none"
                  viewBox="0 0 24 24"
                  stroke="currentColor"
                >
                  <path
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    stroke-width="2"
                    d="M14 5l7 7m0 0l-7 7m7-7H3"
                  />
                </svg>
                <input
                  v-model="mapping.to"
                  type="text"
                  class="components-account-edit-account-modal__field-2 input"
                  :placeholder="t('admin.accounts.actualModel')"
                />
                <button
                  type="button"
                  @click="removeModelMapping(index)"
                  class="components-account-edit-account-modal__action"
                >
                  <svg class="components-account-edit-account-modal__icon-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path
                      stroke-linecap="round"
                      stroke-linejoin="round"
                      stroke-width="2"
                      d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"
                    />
                  </svg>
                </button>
              </div>
            </div>

            <button
              type="button"
              @click="addModelMapping"
              class="components-account-edit-account-modal__action-2"
            >
              <svg
                class="components-account-edit-account-modal__icon-2"
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
              >
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="2"
                  d="M12 4v16m8-8H4"
                />
              </svg>
              {{ t('admin.accounts.addMapping') }}
            </button>

              <!-- Quick Add Buttons -->
              <div class="components-account-edit-account-modal__panel-9">
                <button
                  v-for="preset in presetMappings"
                  :key="preset.label"
                  type="button"
                  @click="addPresetMapping(preset.from, preset.to)"
                  :class="['components-account-edit-account-modal__action-19', preset.color]"
                >
                  + {{ preset.label }}
                </button>
              </div>
            </div>
          </template>
        </div>

        <!-- Pool Mode Section -->
        <div class="components-account-edit-account-modal__panel-3">
          <div class="components-account-edit-account-modal__panel-10">
            <div>
              <label class="components-account-edit-account-modal__label input-label">{{ t('admin.accounts.poolMode') }}</label>
              <p class="components-account-edit-account-modal__description-4">
                {{ t('admin.accounts.poolModeHint') }}
              </p>
            </div>
            <button
              type="button"
              @click="poolModeEnabled = !poolModeEnabled"
              :class="[
                'components-account-edit-account-modal__action-20',
                poolModeEnabled ? 'components-account-edit-account-modal__action-21' : 'components-account-edit-account-modal__action-22'
              ]"
            >
              <span
                :class="[
                  'components-account-edit-account-modal__text-16',
                  poolModeEnabled ? 'toggle-thumb--on' : 'components-account-edit-account-modal__text-17'
                ]"
              />
            </button>
          </div>
          <div v-if="poolModeEnabled" class="components-account-edit-account-modal__panel-11">
            <p class="components-account-edit-account-modal__description-5">
              <Icon name="exclamationCircle" size="sm" class="components-account-edit-account-modal__icon-5" :stroke-width="2" />
              {{ t('admin.accounts.poolModeInfo') }}
            </p>
          </div>
          <div v-if="poolModeEnabled" class="components-account-edit-account-modal__panel-12">
            <label class="input-label">{{ t('admin.accounts.poolModeRetryCount') }}</label>
            <input
              v-model.number="poolModeRetryCount"
              type="number"
              min="0"
              :max="MAX_POOL_MODE_RETRY_COUNT"
              step="1"
              class="input"
            />
            <p class="components-account-edit-account-modal__description-4">
              {{
                t('admin.accounts.poolModeRetryCountHint', {
                  default: DEFAULT_POOL_MODE_RETRY_COUNT,
                  max: MAX_POOL_MODE_RETRY_COUNT
                })
              }}
            </p>
          </div>
          <div v-if="poolModeEnabled" class="components-account-edit-account-modal__panel-12">
            <label class="input-label">{{ t('admin.accounts.poolModeRetryStatusCodes') }}</label>
            <input
              v-model="poolModeRetryStatusCodesInput"
              type="text"
              class="input"
              :placeholder="DEFAULT_POOL_MODE_RETRY_STATUS_CODES.join(', ')"
            />
            <p class="components-account-edit-account-modal__description-4">
              {{ t('admin.accounts.poolModeRetryStatusCodesHint', { default: DEFAULT_POOL_MODE_RETRY_STATUS_CODES.join(', ') }) }}
            </p>
          </div>
        </div>

        <!-- Custom Error Codes Section -->
        <div class="components-account-edit-account-modal__panel-3">
          <div class="components-account-edit-account-modal__panel-10">
            <div>
              <label class="components-account-edit-account-modal__label input-label">{{ t('admin.accounts.customErrorCodes') }}</label>
              <p class="components-account-edit-account-modal__description-4">
                {{ t('admin.accounts.customErrorCodesHint') }}
              </p>
            </div>
            <button
              type="button"
              @click="customErrorCodesEnabled = !customErrorCodesEnabled"
              :class="[
                'components-account-edit-account-modal__action-20',
                customErrorCodesEnabled ? 'components-account-edit-account-modal__action-21' : 'components-account-edit-account-modal__action-22'
              ]"
            >
              <span
                :class="[
                  'components-account-edit-account-modal__text-16',
                  customErrorCodesEnabled ? 'toggle-thumb--on' : 'components-account-edit-account-modal__text-17'
                ]"
              />
            </button>
          </div>

          <div v-if="customErrorCodesEnabled" class="components-account-edit-account-modal__panel-13">
            <div class="components-account-edit-account-modal__panel-14">
              <p class="components-account-edit-account-modal__description">
                <Icon name="exclamationTriangle" size="sm" class="components-account-edit-account-modal__icon-5" :stroke-width="2" />
                {{ t('admin.accounts.customErrorCodesWarning') }}
              </p>
            </div>

            <!-- Error Code Buttons -->
            <div class="components-account-edit-account-modal__panel-9">
              <button
                v-for="code in commonErrorCodes"
                :key="code.value"
                type="button"
                @click="toggleErrorCode(code.value)"
                :class="[
                  'components-account-edit-account-modal__action-23',
                  selectedErrorCodes.includes(code.value)
                    ? 'components-account-edit-account-modal__action-24'
                    : 'components-account-edit-account-modal__action-17'
                ]"
              >
                {{ code.value }} {{ code.label }}
              </button>
            </div>

            <!-- Manual input -->
            <div class="components-account-edit-account-modal__panel-8">
              <input
                v-model.number="customErrorCodeInput"
                type="number"
                min="100"
                max="599"
                class="components-account-edit-account-modal__field-2 input"
                :placeholder="t('admin.accounts.enterErrorCode')"
                @keyup.enter="addCustomErrorCode"
              />
              <button type="button" @click="addCustomErrorCode" class="components-account-edit-account-modal__action-3 btn btn-secondary">
                <svg class="components-account-edit-account-modal__icon-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    stroke-width="2"
                    d="M12 4v16m8-8H4"
                  />
                </svg>
              </button>
            </div>

            <!-- Selected codes summary -->
            <div class="components-account-edit-account-modal__panel-15">
              <span
                v-for="code in selectedErrorCodes.sort((a, b) => a - b)"
                :key="code"
                class="components-account-edit-account-modal__text"
              >
                {{ code }}
                <button
                  type="button"
                  @click="removeErrorCode(code)"
                  class="components-account-edit-account-modal__action-4"
                >
                  <Icon name="x" size="sm" :stroke-width="2" />
                </button>
              </span>
              <span v-if="selectedErrorCodes.length === 0" class="components-account-edit-account-modal__text-2">
                {{ t('admin.accounts.noneSelectedUsesDefault') }}
              </span>
            </div>
          </div>
        </div>

      </div>

      <!-- Grok OAuth client-tool prompt cache opt-in -->
      <div
        v-if="account.platform === 'grok' && account.type === 'oauth'"
        class="components-account-edit-account-modal__panel-3"
      >
        <div class="components-account-edit-account-modal__panel-16">
          <div class="components-account-edit-account-modal__panel-17">
            <label class="components-account-edit-account-modal__label input-label">{{ t('admin.accounts.grokClientToolCache.title') }}</label>
            <p class="components-account-edit-account-modal__description-4">
              {{ t('admin.accounts.grokClientToolCache.hint') }}
            </p>
          </div>
          <Toggle
            v-model="grokClientToolCacheEnabled"
            data-testid="grok-client-tool-cache-toggle"
            :aria-label="t('admin.accounts.grokClientToolCache.title')"
          />
        </div>
      </div>

      <!-- Grok OAuth Custom Upstream URL (仅改写转发端点，OAuth 授权/刷新不受影响) -->
      <div
        v-if="account.platform === 'grok' && account.type === 'oauth'"
        class="components-account-edit-account-modal__panel-3"
      >
        <div class="components-account-edit-account-modal__panel-10">
          <div>
            <label class="components-account-edit-account-modal__label input-label">{{ t('admin.accounts.grokCustomBaseUrl.title') }}</label>
            <p class="components-account-edit-account-modal__description-4">
              {{ t('admin.accounts.grokCustomBaseUrl.hint') }}
            </p>
          </div>
          <button
            type="button"
            data-testid="grok-custom-base-url-toggle"
            @click="grokOAuthCustomBaseUrlEnabled = !grokOAuthCustomBaseUrlEnabled"
            :class="[
              'components-account-edit-account-modal__action-20',
              grokOAuthCustomBaseUrlEnabled ? 'components-account-edit-account-modal__action-21' : 'components-account-edit-account-modal__action-22'
            ]"
          >
            <span
              :class="[
                'components-account-edit-account-modal__text-16',
                grokOAuthCustomBaseUrlEnabled ? 'toggle-thumb--on' : 'components-account-edit-account-modal__text-17'
              ]"
            />
          </button>
        </div>
        <div v-if="grokOAuthCustomBaseUrlEnabled" class="components-account-edit-account-modal__panel-18">
          <input
            v-model="grokOAuthBaseUrl"
            type="text"
            class="input"
            data-testid="grok-custom-base-url-input"
            :placeholder="t('admin.accounts.grokCustomBaseUrl.placeholder')"
          />
          <GrokBaseUrlPresets @select="grokOAuthBaseUrl = $event" />
        </div>
      </div>

      <!-- Header Override Section (anthropic/openai apikey + grok apikey/oauth) -->
      <div v-if="headerOverrideCapable" class="components-account-edit-account-modal__panel-3">
        <div class="components-account-edit-account-modal__panel-10">
          <div>
            <label class="components-account-edit-account-modal__label input-label">{{ t('admin.accounts.headerOverride.title') }}</label>
            <p class="components-account-edit-account-modal__description-4">
              {{ t('admin.accounts.headerOverride.hint') }}
            </p>
          </div>
          <button
            type="button"
            @click="headerOverrideEnabled = !headerOverrideEnabled"
            :class="[
              'components-account-edit-account-modal__action-20',
              headerOverrideEnabled ? 'components-account-edit-account-modal__action-21' : 'components-account-edit-account-modal__action-22'
            ]"
          >
            <span
              :class="[
                'components-account-edit-account-modal__text-16',
                headerOverrideEnabled ? 'toggle-thumb--on' : 'components-account-edit-account-modal__text-17'
              ]"
            />
          </button>
        </div>

        <div v-if="headerOverrideEnabled" class="components-account-edit-account-modal__panel-13">
          <div class="components-account-edit-account-modal__panel-11">
            <p class="components-account-edit-account-modal__description-5">
              <Icon name="exclamationCircle" size="sm" class="components-account-edit-account-modal__icon-5" :stroke-width="2" />
              {{ t('admin.accounts.headerOverride.info') }}
            </p>
          </div>

          <HeaderOverrideEditor
            :rows="headerOverrideRows"
            @update:rows="headerOverrideRows = $event"
          />
        </div>
      </div>

      <!-- OpenAI/Grok OAuth Model Mapping (OAuth 类型没有 apikey 容器，需要独立的模型映射区域) -->
      <div
        v-if="(account.platform === 'openai' || account.platform === 'gemini' || account.platform === 'antigravity' || account.platform === 'grok') && account.type === 'oauth'"
        class="components-account-edit-account-modal__panel-3"
      >
        <label class="input-label">{{ t('admin.accounts.modelRestriction') }}</label>
        <AccountModelRuleSelector
          :platform="account?.platform || 'anthropic'"
          :disabled="isOpenAIModelRestrictionDisabled"
          :has-existing-mappings="hasModelRestrictionValues"
          @apply="applyAccountModelRule"
        />

        <div
          v-if="isOpenAIModelRestrictionDisabled"
          class="components-account-edit-account-modal__panel-4"
        >
          <p class="components-account-edit-account-modal__description">
            {{ t('admin.accounts.openai.modelRestrictionDisabledByPassthrough') }}
          </p>
        </div>

        <template v-else>
          <!-- Mode Toggle -->
          <div class="components-account-edit-account-modal__panel-5">
            <button
              type="button"
              @click="modelRestrictionMode = 'whitelist'"
              :class="[
                'components-account-edit-account-modal__action-15',
                modelRestrictionMode === 'whitelist'
                  ? 'components-account-edit-account-modal__action-16'
                  : 'components-account-edit-account-modal__action-17'
              ]"
            >
              {{ t('admin.accounts.modelWhitelist') }}
            </button>
            <button
              type="button"
              @click="modelRestrictionMode = 'mapping'"
              :class="[
                'components-account-edit-account-modal__action-15',
                modelRestrictionMode === 'mapping'
                  ? 'components-account-edit-account-modal__action-18'
                  : 'components-account-edit-account-modal__action-17'
              ]"
            >
              {{ t('admin.accounts.modelMapping') }}
            </button>
          </div>

          <!-- Whitelist Mode -->
          <div v-if="modelRestrictionMode === 'whitelist'">
            <ModelWhitelistSelector v-model="allowedModels" :platform="account?.platform || 'anthropic'" :account-id="account?.id" />
            <p class="components-account-edit-account-modal__description-2">
              {{ t('admin.accounts.selectedModels', { count: allowedModels.length }) }}
              <span v-if="allowedModels.length === 0 && modelMappings.length === 0">{{
                t('admin.accounts.supportsAllModels')
              }}</span>
            </p>
          </div>

          <!-- Mapping Mode -->
          <div v-else>
            <div class="components-account-edit-account-modal__panel-6">
              <p class="components-account-edit-account-modal__description-3">
                {{ t('admin.accounts.mapRequestModels') }}
              </p>
            </div>

            <div v-if="modelMappings.length > 0" class="components-account-edit-account-modal__panel-7">
              <div
                v-for="(mapping, index) in modelMappings"
                :key="'oauth-' + getModelMappingKey(mapping)"
                class="components-account-edit-account-modal__panel-8"
              >
                <input
                  v-model="mapping.from"
                  type="text"
                  class="components-account-edit-account-modal__field-2 input"
                  :placeholder="t('admin.accounts.requestModel')"
                />
                <svg
                  class="components-account-edit-account-modal__icon-3"
                  fill="none"
                  viewBox="0 0 24 24"
                  stroke="currentColor"
                >
                  <path
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    stroke-width="2"
                    d="M14 5l7 7m0 0l-7 7m7-7H3"
                  />
                </svg>
                <input
                  v-model="mapping.to"
                  type="text"
                  class="components-account-edit-account-modal__field-2 input"
                  :placeholder="t('admin.accounts.actualModel')"
                />
                <button
                  type="button"
                  @click="removeModelMapping(index)"
                  class="components-account-edit-account-modal__action"
                >
                  <svg class="components-account-edit-account-modal__icon-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path
                      stroke-linecap="round"
                      stroke-linejoin="round"
                      stroke-width="2"
                      d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"
                    />
                  </svg>
                </button>
              </div>
            </div>

            <button
              type="button"
              @click="addModelMapping"
              class="components-account-edit-account-modal__action-2"
            >
              + {{ t('admin.accounts.addMapping') }}
            </button>

            <!-- Quick Add Buttons -->
            <div class="components-account-edit-account-modal__panel-9">
              <button
                v-for="preset in presetMappings"
                :key="'oauth-' + preset.label"
                type="button"
                @click="addPresetMapping(preset.from, preset.to)"
                :class="['components-account-edit-account-modal__action-19', preset.color]"
              >
                + {{ preset.label }}
              </button>
            </div>
          </div>
        </template>
      </div>

      <!-- Upstream fields (only for upstream type) -->
      <div v-if="account.type === 'upstream'" class="components-account-edit-account-modal__panel">
        <div>
          <label class="input-label">{{ t('admin.accounts.upstream.baseUrl') }}</label>
          <input
            v-model="editBaseUrl"
            type="text"
            class="input"
            placeholder="https://cloudcode-pa.googleapis.com"
          />
          <p class="input-hint">{{ t('admin.accounts.upstream.baseUrlHint') }}</p>
        </div>
        <div>
          <label class="input-label">{{ t('admin.accounts.upstream.apiKey') }}</label>
          <input
            v-model="editApiKey"
            type="password"
            class="components-account-edit-account-modal__field input"
            placeholder="sk-..."
          />
          <p class="input-hint">{{ t('admin.accounts.leaveEmptyToKeep') }}</p>
        </div>
      </div>

      <!-- Vertex Service Account -->
      <div v-if="(account.platform === 'anthropic' || account.platform === 'gemini') && account.type === 'service_account'" class="components-account-edit-account-modal__panel">
        <div class="components-account-edit-account-modal__panel-19">
          <div>
            <label class="input-label">Project ID</label>
            <input
              v-model="editVertexProjectId"
              type="text"
              class="components-account-edit-account-modal__field input"
              readonly
              :placeholder="t('admin.accounts.vertexProjectIdPlaceholder')"
            />
            <p class="input-hint">{{ t('admin.accounts.vertexSaJsonEditHint') }}</p>
          </div>
          <div>
            <label class="input-label">Location</label>
            <Select
              v-model="editVertexLocation"
              class="components-account-edit-account-modal__field input"
              :options="VERTEX_LOCATION_SELECT_OPTIONS"
              searchable
            />
            <p class="input-hint">{{ t('admin.accounts.vertexLocationHint') }}</p>
          </div>
        </div>

        <!-- Model Restriction Section for Service Account -->
        <div class="components-account-edit-account-modal__panel-3">
          <label class="input-label">{{ t('admin.accounts.modelRestriction') }}</label>
          <AccountModelRuleSelector
            :platform="account?.platform || 'anthropic'"
            :has-existing-mappings="hasModelRestrictionValues"
            @apply="applyAccountModelRule"
          />

          <!-- Mode Toggle -->
          <div class="components-account-edit-account-modal__panel-5">
            <button
              type="button"
              @click="modelRestrictionMode = 'whitelist'"
              :class="[
                'components-account-edit-account-modal__action-15',
                modelRestrictionMode === 'whitelist'
                  ? 'components-account-edit-account-modal__action-16'
                  : 'components-account-edit-account-modal__action-17'
              ]"
            >
              <svg
                class="components-account-edit-account-modal__icon"
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
              >
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="2"
                  d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"
                />
              </svg>
              {{ t('admin.accounts.modelWhitelist') }}
            </button>
            <button
              type="button"
              @click="modelRestrictionMode = 'mapping'"
              :class="[
                'components-account-edit-account-modal__action-15',
                modelRestrictionMode === 'mapping'
                  ? 'components-account-edit-account-modal__action-18'
                  : 'components-account-edit-account-modal__action-17'
              ]"
            >
              <svg
                class="components-account-edit-account-modal__icon"
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
              >
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="2"
                  d="M8 7h12m0 0l-4-4m4 4l-4 4m0 6H4m0 0l4 4m-4-4l4-4"
                />
              </svg>
              {{ t('admin.accounts.modelMapping') }}
            </button>
          </div>

          <!-- Whitelist Mode -->
          <div v-if="modelRestrictionMode === 'whitelist'">
            <ModelWhitelistSelector v-model="allowedModels" :platform="account?.platform || 'anthropic'" :account-id="account?.id" />
            <p class="components-account-edit-account-modal__description-2">
              {{ t('admin.accounts.selectedModels', { count: allowedModels.length }) }}
              <span v-if="allowedModels.length === 0 && modelMappings.length === 0">{{
                t('admin.accounts.supportsAllModels')
              }}</span>
            </p>
          </div>

          <!-- Mapping Mode -->
          <div v-else>
            <div class="components-account-edit-account-modal__panel-6">
              <p class="components-account-edit-account-modal__description-3">
                <svg
                  class="components-account-edit-account-modal__icon-2"
                  fill="none"
                  viewBox="0 0 24 24"
                  stroke="currentColor"
                >
                  <path
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    stroke-width="2"
                    d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
                  />
                </svg>
                {{ t('admin.accounts.mapRequestModels') }}
              </p>
            </div>

            <!-- Model Mapping List -->
            <div v-if="modelMappings.length > 0" class="components-account-edit-account-modal__panel-7">
              <div
                v-for="(mapping, index) in modelMappings"
                :key="getModelMappingKey(mapping)"
                class="components-account-edit-account-modal__panel-8"
              >
                <input
                  v-model="mapping.from"
                  type="text"
                  class="components-account-edit-account-modal__field-2 input"
                  :placeholder="t('admin.accounts.requestModel')"
                />
                <svg
                  class="components-account-edit-account-modal__icon-3"
                  fill="none"
                  viewBox="0 0 24 24"
                  stroke="currentColor"
                >
                  <path
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    stroke-width="2"
                    d="M14 5l7 7m0 0l-7 7m7-7H3"
                  />
                </svg>
                <input
                  v-model="mapping.to"
                  type="text"
                  class="components-account-edit-account-modal__field-2 input"
                  :placeholder="t('admin.accounts.actualModel')"
                />
                <button
                  type="button"
                  @click="removeModelMapping(index)"
                  class="components-account-edit-account-modal__action"
                >
                  <svg class="components-account-edit-account-modal__icon-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path
                      stroke-linecap="round"
                      stroke-linejoin="round"
                      stroke-width="2"
                      d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"
                    />
                  </svg>
                </button>
              </div>
            </div>

            <button
              type="button"
              @click="addModelMapping"
              class="components-account-edit-account-modal__action-2"
            >
              <svg
                class="components-account-edit-account-modal__icon-2"
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
              >
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="2"
                  d="M12 4v16m8-8H4"
                />
              </svg>
              {{ t('admin.accounts.addMapping') }}
            </button>

            <!-- Quick Add Buttons -->
            <div class="components-account-edit-account-modal__panel-9">
              <button
                v-for="preset in presetMappings"
                :key="preset.label"
                type="button"
                @click="addPresetMapping(preset.from, preset.to)"
                :class="['components-account-edit-account-modal__action-19', preset.color]"
              >
                + {{ preset.label }}
              </button>
            </div>
          </div>
        </div>
      </div>

      <!-- Bedrock fields (for bedrock type, both SigV4 and API Key modes) -->
      <div v-if="account.type === 'bedrock'" class="components-account-edit-account-modal__panel">
        <!-- SigV4 fields -->
        <template v-if="!isBedrockAPIKeyMode">
          <div>
            <label class="input-label">{{ t('admin.accounts.bedrockAccessKeyId') }}</label>
            <input
              v-model="editBedrockAccessKeyId"
              type="text"
              class="components-account-edit-account-modal__field input"
              placeholder="AKIA..."
            />
          </div>
          <div>
            <label class="input-label">{{ t('admin.accounts.bedrockSecretAccessKey') }}</label>
            <input
              v-model="editBedrockSecretAccessKey"
              type="password"
              class="components-account-edit-account-modal__field input"
              :placeholder="t('admin.accounts.bedrockSecretKeyLeaveEmpty')"
            />
            <p class="input-hint">{{ t('admin.accounts.bedrockSecretKeyLeaveEmpty') }}</p>
          </div>
          <div>
            <label class="input-label">{{ t('admin.accounts.bedrockSessionToken') }}</label>
            <input
              v-model="editBedrockSessionToken"
              type="password"
              class="components-account-edit-account-modal__field input"
              :placeholder="t('admin.accounts.bedrockSecretKeyLeaveEmpty')"
            />
            <p class="input-hint">{{ t('admin.accounts.bedrockSessionTokenHint') }}</p>
          </div>
        </template>

        <!-- API Key field -->
        <div v-if="isBedrockAPIKeyMode">
          <label class="input-label">{{ t('admin.accounts.bedrockApiKeyInput') }}</label>
          <input
            v-model="editBedrockApiKeyValue"
            type="password"
            class="components-account-edit-account-modal__field input"
            :placeholder="t('admin.accounts.bedrockApiKeyLeaveEmpty')"
          />
          <p class="input-hint">{{ t('admin.accounts.bedrockApiKeyLeaveEmpty') }}</p>
        </div>

        <!-- Shared: Region -->
        <div>
          <label class="input-label">{{ t('admin.accounts.bedrockRegion') }}</label>
          <input
            v-model="editBedrockRegion"
            type="text"
            class="input"
            placeholder="us-east-1"
          />
          <p class="input-hint">{{ t('admin.accounts.bedrockRegionHint') }}</p>
        </div>

        <!-- Shared: Force Global -->
        <div>
          <label class="components-account-edit-account-modal__label-2">
            <input
              v-model="editBedrockForceGlobal"
              type="checkbox"
              class="components-account-edit-account-modal__field-3"
            />
            <span class="components-account-edit-account-modal__text-3">{{ t('admin.accounts.bedrockForceGlobal') }}</span>
          </label>
          <p class="components-account-edit-account-modal__description-6 input-hint">{{ t('admin.accounts.bedrockForceGlobalHint') }}</p>
        </div>

        <!-- Model Restriction for Bedrock -->
        <div class="components-account-edit-account-modal__panel-3">
          <label class="input-label">{{ t('admin.accounts.modelRestriction') }}</label>
          <AccountModelRuleSelector
            platform="anthropic"
            :has-existing-mappings="hasModelRestrictionValues"
            @apply="applyAccountModelRule"
          />

          <!-- Mode Toggle -->
          <div class="components-account-edit-account-modal__panel-5">
            <button
              type="button"
              @click="modelRestrictionMode = 'whitelist'"
              :class="[
                'components-account-edit-account-modal__action-15',
                modelRestrictionMode === 'whitelist'
                  ? 'components-account-edit-account-modal__action-16'
                  : 'components-account-edit-account-modal__action-17'
              ]"
            >
              {{ t('admin.accounts.modelWhitelist') }}
            </button>
            <button
              type="button"
              @click="modelRestrictionMode = 'mapping'"
              :class="[
                'components-account-edit-account-modal__action-15',
                modelRestrictionMode === 'mapping'
                  ? 'components-account-edit-account-modal__action-18'
                  : 'components-account-edit-account-modal__action-17'
              ]"
            >
              {{ t('admin.accounts.modelMapping') }}
            </button>
          </div>

          <!-- Whitelist Mode -->
          <div v-if="modelRestrictionMode === 'whitelist'">
            <ModelWhitelistSelector v-model="allowedModels" platform="anthropic" />
            <p class="components-account-edit-account-modal__description-2">
              {{ t('admin.accounts.selectedModels', { count: allowedModels.length }) }}
              <span v-if="allowedModels.length === 0 && modelMappings.length === 0">{{ t('admin.accounts.supportsAllModels') }}</span>
            </p>
          </div>

          <!-- Mapping Mode -->
          <div v-else class="components-account-edit-account-modal__panel-13">
            <div v-for="(mapping, index) in modelMappings" :key="getModelMappingKey(mapping)" class="components-account-edit-account-modal__panel-8">
              <input v-model="mapping.from" type="text" class="components-account-edit-account-modal__field-2 input" :placeholder="t('admin.accounts.fromModel')" />
              <span class="components-account-edit-account-modal__text-4">→</span>
              <input v-model="mapping.to" type="text" class="components-account-edit-account-modal__field-2 input" :placeholder="t('admin.accounts.toModel')" />
              <button type="button" @click="modelMappings.splice(index, 1)" class="components-account-edit-account-modal__action-5">
                <Icon name="trash" size="sm" />
              </button>
            </div>
            <button type="button" @click="modelMappings.push({ from: '', to: '' })" class="components-account-edit-account-modal__action-6 btn btn-secondary">
              + {{ t('admin.accounts.addMapping') }}
            </button>
            <!-- Bedrock Preset Mappings -->
            <div class="components-account-edit-account-modal__panel-9">
              <button
                v-for="preset in bedrockPresets"
                :key="preset.from"
                type="button"
                @click="modelMappings.push({ from: preset.from, to: preset.to })"
                :class="['components-account-edit-account-modal__action-19', preset.color]"
              >
                + {{ preset.label }}
              </button>
            </div>
          </div>
        </div>

        <!-- Pool Mode Section for Bedrock -->
        <div class="components-account-edit-account-modal__panel-3">
          <div class="components-account-edit-account-modal__panel-10">
            <div>
              <label class="components-account-edit-account-modal__label input-label">{{ t('admin.accounts.poolMode') }}</label>
              <p class="components-account-edit-account-modal__description-4">
                {{ t('admin.accounts.poolModeHint') }}
              </p>
            </div>
            <button
              type="button"
              @click="poolModeEnabled = !poolModeEnabled"
              :class="[
                'components-account-edit-account-modal__action-20',
                poolModeEnabled ? 'components-account-edit-account-modal__action-21' : 'components-account-edit-account-modal__action-22'
              ]"
            >
              <span
                :class="[
                  'components-account-edit-account-modal__text-16',
                  poolModeEnabled ? 'toggle-thumb--on' : 'components-account-edit-account-modal__text-17'
                ]"
              />
            </button>
          </div>
          <div v-if="poolModeEnabled" class="components-account-edit-account-modal__panel-11">
            <p class="components-account-edit-account-modal__description-5">
              <Icon name="exclamationCircle" size="sm" class="components-account-edit-account-modal__icon-5" :stroke-width="2" />
              {{ t('admin.accounts.poolModeInfo') }}
            </p>
          </div>
          <div v-if="poolModeEnabled" class="components-account-edit-account-modal__panel-12">
            <label class="input-label">{{ t('admin.accounts.poolModeRetryCount') }}</label>
            <input
              v-model.number="poolModeRetryCount"
              type="number"
              min="0"
              :max="MAX_POOL_MODE_RETRY_COUNT"
              step="1"
              class="input"
            />
            <p class="components-account-edit-account-modal__description-4">
              {{
                t('admin.accounts.poolModeRetryCountHint', {
                  default: DEFAULT_POOL_MODE_RETRY_COUNT,
                  max: MAX_POOL_MODE_RETRY_COUNT
                })
              }}
            </p>
          </div>
          <div v-if="poolModeEnabled" class="components-account-edit-account-modal__panel-12">
            <label class="input-label">{{ t('admin.accounts.poolModeRetryStatusCodes') }}</label>
            <input
              v-model="poolModeRetryStatusCodesInput"
              type="text"
              class="input"
              :placeholder="DEFAULT_POOL_MODE_RETRY_STATUS_CODES.join(', ')"
            />
            <p class="components-account-edit-account-modal__description-4">
              {{ t('admin.accounts.poolModeRetryStatusCodesHint', { default: DEFAULT_POOL_MODE_RETRY_STATUS_CODES.join(', ') }) }}
            </p>
          </div>
        </div>
      </div>

      <!-- Temp Unschedulable Rules -->
      <div class="components-account-edit-account-modal__panel-22">
        <div class="components-account-edit-account-modal__panel-10">
          <div>
            <label class="components-account-edit-account-modal__label input-label">{{ t('admin.accounts.tempUnschedulable.title') }}</label>
            <p class="components-account-edit-account-modal__description-4">
              {{ t('admin.accounts.tempUnschedulable.hint') }}
            </p>
          </div>
          <button
            type="button"
            @click="tempUnschedEnabled = !tempUnschedEnabled"
            :class="[
              'components-account-edit-account-modal__action-20',
              tempUnschedEnabled ? 'components-account-edit-account-modal__action-21' : 'components-account-edit-account-modal__action-22'
            ]"
          >
            <span
              :class="[
                'components-account-edit-account-modal__text-16',
                tempUnschedEnabled ? 'toggle-thumb--on' : 'components-account-edit-account-modal__text-17'
              ]"
            />
          </button>
        </div>

        <div v-if="tempUnschedEnabled" class="components-account-edit-account-modal__panel-13">
          <div class="components-account-edit-account-modal__panel-11">
            <p class="components-account-edit-account-modal__description-5">
              <Icon name="exclamationTriangle" size="sm" class="components-account-edit-account-modal__icon-5" :stroke-width="2" />
              {{ t('admin.accounts.tempUnschedulable.notice') }}
            </p>
          </div>

          <div class="components-account-edit-account-modal__panel-9">
            <button
              v-for="preset in tempUnschedPresets"
              :key="preset.label"
              type="button"
              @click="addTempUnschedRule(preset.rule)"
              class="components-account-edit-account-modal__action-8"
            >
              + {{ preset.label }}
            </button>
          </div>

          <div v-if="tempUnschedRules.length > 0" class="components-account-edit-account-modal__panel-13">
            <div
              v-for="(rule, index) in tempUnschedRules"
              :key="getTempUnschedRuleKey(rule)"
              class="components-account-edit-account-modal__panel-23"
            >
              <div class="components-account-edit-account-modal__panel-24">
                <span class="components-account-edit-account-modal__text-5">
                  {{ t('admin.accounts.tempUnschedulable.ruleIndex', { index: index + 1 }) }}
                </span>
                <div class="components-account-edit-account-modal__panel-8">
                  <button
                    type="button"
                    :disabled="index === 0"
                    @click="moveTempUnschedRule(index, -1)"
                    class="components-account-edit-account-modal__action-9"
                  >
                    <Icon name="chevronUp" size="sm" :stroke-width="2" />
                  </button>
                  <button
                    type="button"
                    :disabled="index === tempUnschedRules.length - 1"
                    @click="moveTempUnschedRule(index, 1)"
                    class="components-account-edit-account-modal__action-9"
                  >
                    <svg class="components-account-edit-account-modal__icon-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
                    </svg>
                  </button>
                  <button
                    type="button"
                    @click="removeTempUnschedRule(index)"
                    class="components-account-edit-account-modal__action-10"
                  >
                    <Icon name="x" size="sm" :stroke-width="2" />
                  </button>
                </div>
              </div>

              <div class="components-account-edit-account-modal__panel-25">
                <div>
                  <label class="input-label">{{ t('admin.accounts.tempUnschedulable.errorCode') }}</label>
                  <input
                    v-model.number="rule.error_code"
                    type="number"
                    min="100"
                    max="599"
                    class="input"
                    :placeholder="t('admin.accounts.tempUnschedulable.errorCodePlaceholder')"
                  />
                </div>
                <div>
                  <label class="input-label">{{ t('admin.accounts.tempUnschedulable.durationMinutes') }}</label>
                  <input
                    v-model.number="rule.duration_minutes"
                    type="number"
                    min="1"
                    class="input"
                    :placeholder="t('admin.accounts.tempUnschedulable.durationPlaceholder')"
                  />
                </div>
                <div class="components-account-edit-account-modal__panel-26">
                  <label class="input-label">{{ t('admin.accounts.tempUnschedulable.keywords') }}</label>
                  <input
                    v-model="rule.keywords"
                    type="text"
                    class="input"
                    :placeholder="t('admin.accounts.tempUnschedulable.keywordsPlaceholder')"
                  />
                  <p class="input-hint">{{ t('admin.accounts.tempUnschedulable.keywordsHint') }}</p>
                </div>
                <div class="components-account-edit-account-modal__panel-26">
                  <label class="input-label">{{ t('admin.accounts.tempUnschedulable.description') }}</label>
                  <input
                    v-model="rule.description"
                    type="text"
                    class="input"
                    :placeholder="t('admin.accounts.tempUnschedulable.descriptionPlaceholder')"
                  />
                </div>
              </div>
            </div>
          </div>

          <button
            type="button"
            @click="addTempUnschedRule()"
            class="components-account-edit-account-modal__action-11"
          >
            <svg
              class="components-account-edit-account-modal__icon-2"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
            >
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
            </svg>
            {{ t('admin.accounts.tempUnschedulable.addRule') }}
          </button>
        </div>
      </div>


      <div
        v-if="supportsAccountSchedulingThresholdOverride"
        class="components-account-edit-account-modal__panel-3"
        data-testid="account-scheduling-threshold-section"
      >
        <div class="components-account-edit-account-modal__panel-10">
          <div>
            <label class="components-account-edit-account-modal__label input-label">{{ t('admin.accounts.accountSchedulingThresholdOverride') }}</label>
            <p class="components-account-edit-account-modal__description-4">
              {{ t('admin.accounts.accountSchedulingThresholdOverrideHint') }}
            </p>
          </div>
          <input
            v-model="accountSchedulingThresholdOverrideEnabled"
            data-testid="account-scheduling-threshold-override-enabled"
            type="checkbox"
            class="components-account-edit-account-modal__field-4"
          />
        </div>
        <div v-if="accountSchedulingThresholdOverrideEnabled">
          <label class="input-label">{{ t('admin.accounts.accountSchedulingThresholdOverrideValue') }}</label>
          <input
            v-model.number="accountSchedulingThresholdOverrideValue"
            data-testid="account-scheduling-threshold-override-value"
            type="number"
            min="1"
            max="100"
            class="input"
          />
          <p class="input-hint">{{ t('admin.accounts.accountSchedulingThresholdOverrideDisabledHint') }}</p>
        </div>
      </div>

      <!-- Intercept Warmup Requests -->
      <div
        v-if="account?.platform === 'anthropic'"
        class="components-account-edit-account-modal__panel-3"
      >
        <div class="components-account-edit-account-modal__panel-27">
          <div>
            <label class="components-account-edit-account-modal__label input-label">{{
              t('admin.accounts.interceptWarmupRequests')
            }}</label>
            <p class="components-account-edit-account-modal__description-4">
              {{ t('admin.accounts.interceptWarmupRequestsDesc') }}
            </p>
          </div>
          <button
            type="button"
            @click="interceptWarmupRequests = !interceptWarmupRequests"
            :class="[
              'components-account-edit-account-modal__action-20',
              interceptWarmupRequests ? 'components-account-edit-account-modal__action-21' : 'components-account-edit-account-modal__action-22'
            ]"
          >
            <span
              :class="[
                'components-account-edit-account-modal__text-16',
                interceptWarmupRequests ? 'toggle-thumb--on' : 'components-account-edit-account-modal__text-17'
              ]"
            />
          </button>
        </div>
      </div>

      <div v-if="!isSparkShadow">
        <div class="components-account-edit-account-modal__panel-28">
          <label class="components-account-edit-account-modal__label input-label">{{ t('admin.accounts.proxy') }}</label>
          <ProxyAdBanner />
        </div>
        <ProxySelector v-model="form.proxy_id" :proxies="proxies" />
      </div>

      <div class="components-account-edit-account-modal__panel-29">
        <div>
          <label class="input-label">{{ t('admin.accounts.concurrency') }}</label>
          <input v-model.number="form.concurrency" type="number" min="1" class="input"
            @input="form.concurrency = Math.max(1, form.concurrency || 1)" />
        </div>
        <div>
          <label class="input-label">{{ t('admin.accounts.loadFactor') }}</label>
          <input v-model.number="form.load_factor" type="number" min="1"
            class="input" :placeholder="String(form.concurrency || 1)"
            @input="form.load_factor = (form.load_factor &amp;&amp; form.load_factor >= 1) ? form.load_factor : null" />
          <p class="input-hint">{{ t('admin.accounts.loadFactorHint') }}</p>
        </div>
        <div>
          <label class="input-label">{{ t('admin.accounts.priority') }}</label>
          <input
            v-model.number="form.priority"
            type="number"
            min="1"
            class="input"
            data-tour="account-form-priority"
          />
          <p class="input-hint">{{ t('admin.accounts.priorityHint') }}</p>
        </div>
        <div>
          <label class="input-label">{{ t('admin.accounts.billingRateMultiplier') }}</label>
          <input
            v-model.number="form.rate_multiplier"
            type="number"
            min="0"
            step="0.001"
            class="components-account-edit-account-modal__field-5 input"
            data-testid="account-rate-multiplier"
            :disabled="upstreamBillingRateSyncEnabled"
          />
          <p class="input-hint">
            {{
              t(
                upstreamBillingRateSyncEnabled
                  ? 'admin.accounts.upstreamBilling.syncRateManagedHint'
                  : 'admin.accounts.billingRateMultiplierHint'
              )
            }}
          </p>
          <div
            v-if="account?.type === 'apikey'"
            class="components-account-edit-account-modal__panel-30"
          >
            <div class="components-account-edit-account-modal__panel-17">
              <p class="components-account-edit-account-modal__description-8">
                {{ t('admin.accounts.upstreamBilling.syncRate') }}
              </p>
              <p class="components-account-edit-account-modal__description-4">
                {{ t('admin.accounts.upstreamBilling.syncRateHint') }}
              </p>
            </div>
            <Toggle
              :model-value="upstreamBillingRateSyncEnabled"
              data-testid="upstream-billing-rate-sync"
              :aria-label="t('admin.accounts.upstreamBilling.syncRate')"
              @update:model-value="handleUpstreamBillingRateSyncChange"
            />
          </div>
        </div>
      </div>
      <div class="components-account-edit-account-modal__panel-3">
        <label class="input-label">{{ t('admin.accounts.expiresAt') }}</label>
        <input v-model="expiresAtInput" type="datetime-local" class="input" />
        <p class="input-hint">{{ t('admin.accounts.expiresAtHint') }}</p>
      </div>

      <!-- OpenAI 自动透传开关（OAuth/API Key） -->
      <div
        v-if="account?.platform === 'openai' && (account?.type === 'oauth' || account?.type === 'setup-token' || account?.type === 'apikey')"
        class="components-account-edit-account-modal__panel-3"
      >
        <div class="components-account-edit-account-modal__panel-27">
          <div>
            <label class="components-account-edit-account-modal__label input-label">{{ t('admin.accounts.openai.oauthPassthrough') }}</label>
            <p class="components-account-edit-account-modal__description-4">
              {{ t('admin.accounts.openai.oauthPassthroughDesc') }}
            </p>
          </div>
          <button
            type="button"
            @click="openaiPassthroughEnabled = !openaiPassthroughEnabled"
            :class="[
              'components-account-edit-account-modal__action-20',
              openaiPassthroughEnabled ? 'components-account-edit-account-modal__action-21' : 'components-account-edit-account-modal__action-22'
            ]"
          >
            <span
              :class="[
                'components-account-edit-account-modal__text-16',
                openaiPassthroughEnabled ? 'toggle-thumb--on' : 'components-account-edit-account-modal__text-17'
              ]"
            />
          </button>
        </div>
      </div>

      <!-- OpenAI Codex namespace 工具摊平（兼容开关，仅 OAuth） -->
      <div
        v-if="account?.platform === 'openai' && account?.type === 'oauth'"
        class="components-account-edit-account-modal__panel-3"
      >
        <div class="components-account-edit-account-modal__panel-27">
          <div>
            <label class="components-account-edit-account-modal__label input-label">{{ t('admin.accounts.openai.flattenNamespaces') }}</label>
            <p class="components-account-edit-account-modal__description-4">
              {{ t('admin.accounts.openai.flattenNamespacesDesc') }}
            </p>
          </div>
          <button
            type="button"
            data-testid="edit-openai-flatten-namespaces-toggle"
            @click="openaiFlattenNamespacesEnabled = !openaiFlattenNamespacesEnabled"
            :class="[
              'components-account-edit-account-modal__action-20',
              openaiFlattenNamespacesEnabled ? 'components-account-edit-account-modal__action-21' : 'components-account-edit-account-modal__action-22'
            ]"
          >
            <span
              :class="[
                'components-account-edit-account-modal__text-16',
                openaiFlattenNamespacesEnabled ? 'toggle-thumb--on' : 'components-account-edit-account-modal__text-17'
              ]"
            />
          </button>
        </div>
      </div>

      <!-- OpenAI Codex hosted image_generation bridge policy -->
      <div
        v-if="account?.platform === 'openai' && (account?.type === 'oauth' || account?.type === 'setup-token' || account?.type === 'apikey')"
        class="components-account-edit-account-modal__panel-3"
      >
        <div class="components-account-edit-account-modal__panel-31">
          <div class="components-account-edit-account-modal__panel-32">
            <div class="components-account-edit-account-modal__panel-33">
              <Icon name="sparkles" size="sm" />
            </div>
            <div class="components-account-edit-account-modal__panel-34">
              <div class="components-account-edit-account-modal__panel-35">
                <label class="components-account-edit-account-modal__label input-label">{{ t('admin.accounts.openai.codexImageTool') }}</label>
                <span
                  class="components-account-edit-account-modal__text-6"
                  :class="codexImageToolBadgeClass"
                >
                  {{ codexImageToolBadgeLabel }}
                </span>
              </div>
              <p class="components-account-edit-account-modal__description-9">
                {{ t('admin.accounts.openai.codexImageToolDesc') }}
              </p>
            </div>
          </div>
          <div class="components-account-edit-account-modal__panel-36">
            <div class="components-account-edit-account-modal__panel-37">
              <button
                v-for="option in codexImageToolOptions"
                :key="option.value"
                type="button"
                :data-testid="`codex-image-tool-${option.value}`"
                @click="codexImageToolMode = option.value"
                :class="[
                  'components-account-edit-account-modal__action-25',
                  codexImageToolMode === option.value
                    ? option.selectedCardClass
                    : 'components-account-edit-account-modal__action-26'
                ]"
              >
                <span
                  :class="[
                    'components-account-edit-account-modal__text-18',
                    codexImageToolMode === option.value
                      ? option.selectedDotClass
                      : 'components-account-edit-account-modal__text-19'
                  ]"
                >
                  <Icon name="check" size="xs" :stroke-width="2" />
                </span>
                <span class="components-account-edit-account-modal__panel-17">
                  <span class="components-account-edit-account-modal__text-7">{{ option.label }}</span>
                  <span class="components-account-edit-account-modal__text-8">{{ option.description }}</span>
                </span>
              </button>
            </div>
          </div>
        </div>
      </div>

      <!-- OpenAI WS Mode 三态（off/ctx_pool/passthrough） -->
      <div
        v-if="account?.platform === 'openai' && (account?.type === 'oauth' || account?.type === 'setup-token' || account?.type === 'apikey')"
        class="components-account-edit-account-modal__panel-3"
      >
        <div class="components-account-edit-account-modal__panel-27">
          <div>
            <label class="components-account-edit-account-modal__label input-label">{{ t('admin.accounts.openai.wsMode') }}</label>
            <p class="components-account-edit-account-modal__description-4">
              {{ t('admin.accounts.openai.wsModeDesc') }}
            </p>
            <p class="components-account-edit-account-modal__description-4">
              {{ t(openAIWSModeConcurrencyHintKey) }}
            </p>
          </div>
          <div class="components-account-edit-account-modal__panel-38">
            <Select v-model="openaiResponsesWebSocketV2Mode" data-testid="edit-openai-ws-mode-select" :options="openAIWSModeOptions" />
          </div>
        </div>
      </div>

      <!-- OpenAI APIKey Responses API support mode -->
      <div
        v-if="account?.platform === 'openai' && account?.type === 'apikey'"
        class="components-account-edit-account-modal__panel-39"
      >
        <div class="components-account-edit-account-modal__panel-16">
          <div>
            <label class="components-account-edit-account-modal__label input-label">{{ t('admin.accounts.openai.responsesMode') }}</label>
            <p class="components-account-edit-account-modal__description-4">
              {{ t('admin.accounts.openai.responsesModeDesc') }}
            </p>
          </div>
          <div class="components-account-edit-account-modal__panel-40">
            <Select
              v-model="openAIResponsesMode"
              :options="openAIResponsesModeOptions"
              :disabled="!openAITextGenerationCapabilityEnabled"
              data-testid="openai-responses-mode-select"
            />
          </div>
        </div>
        <div
          v-if="openAITextGenerationCapabilityEnabled"
          class="components-account-edit-account-modal__panel-41"
        >
          <span class="components-account-edit-account-modal__text-9">{{ t(openAIResponsesStatusKey) }}</span>
        </div>
        <div
          v-else
          class="components-account-edit-account-modal__panel-42"
          data-testid="openai-responses-mode-not-applicable"
        >
          {{ t('admin.accounts.openai.responsesModeTextDisabledHint') }}
        </div>
        <div>
          <label class="components-account-edit-account-modal__label-3 input-label">{{ t('admin.accounts.openai.endpointCapabilities') }}</label>
          <div class="components-account-edit-account-modal__panel-37">
            <label
              v-for="option in openAIEndpointCapabilityOptions"
              :key="option.value"
              class="components-account-edit-account-modal__label-4"
            >
              <input
                type="checkbox"
                class="components-account-edit-account-modal__field-3"
                :data-testid="`openai-endpoint-capability-${option.value}`"
                :checked="openAIEndpointCapabilities.includes(option.value)"
                @change="toggleOpenAIEndpointCapability(option.value, $event)"
              />
              <span class="components-account-edit-account-modal__text-10">{{ option.label }}</span>
            </label>
          </div>
          <p class="input-hint">{{ t('admin.accounts.openai.endpointCapabilitiesDesc') }}</p>
        </div>
      </div>

      <div
        v-if="account?.type === 'apikey'"
        class="components-account-edit-account-modal__panel-43"
      >
        <div>
          <label class="components-account-edit-account-modal__label input-label">{{ t('admin.accounts.upstreamBilling.autoProbe') }}</label>
          <p class="components-account-edit-account-modal__description-4">
            {{ t('admin.accounts.upstreamBilling.autoProbeHint') }}
          </p>
        </div>
        <Toggle
          :model-value="upstreamBillingAutoProbeEnabled"
          data-testid="upstream-billing-auto-probe"
          :aria-label="t('admin.accounts.upstreamBilling.autoProbe')"
          @update:model-value="handleUpstreamBillingAutoProbeChange"
        />
      </div>

      <OllamaCloudUsageSettings
        v-if="account?.ollama_cloud_usage?.eligible"
        :account="account"
        @updated="handleOllamaCloudUsageUpdated"
      />

      <!-- Anthropic API Key 自动透传开关 -->
      <div
        v-if="account?.platform === 'anthropic' && account?.type === 'apikey'"
        class="components-account-edit-account-modal__panel-3"
      >
        <div class="components-account-edit-account-modal__panel-27">
          <div>
            <label class="components-account-edit-account-modal__label input-label">{{ t('admin.accounts.anthropic.apiKeyPassthrough') }}</label>
            <p class="components-account-edit-account-modal__description-4">
              {{ t('admin.accounts.anthropic.apiKeyPassthroughDesc') }}
            </p>
          </div>
          <button
            type="button"
            @click="anthropicPassthroughEnabled = !anthropicPassthroughEnabled"
            :class="[
              'components-account-edit-account-modal__action-20',
              anthropicPassthroughEnabled ? 'components-account-edit-account-modal__action-21' : 'components-account-edit-account-modal__action-22'
            ]"
          >
            <span
              :class="[
                'components-account-edit-account-modal__text-16',
                anthropicPassthroughEnabled ? 'toggle-thumb--on' : 'components-account-edit-account-modal__text-17'
              ]"
            />
          </button>
        </div>
      </div>

      <div
        v-if="account?.platform === 'anthropic' && account?.type === 'apikey'"
        class="components-account-edit-account-modal__panel-3"
      >
        <div class="components-account-edit-account-modal__panel-16">
          <div>
            <label class="components-account-edit-account-modal__label input-label">{{ t('admin.accounts.anthropic.apiKeyAuthScheme') }}</label>
            <p class="components-account-edit-account-modal__description-4">
              {{ t('admin.accounts.anthropic.apiKeyAuthSchemeDesc') }}
            </p>
          </div>
          <Select v-model="anthropicAPIKeyAuthScheme" class="components-account-edit-account-modal__field-6" :options="[
            { value: 'x_api_key', label: t('admin.accounts.anthropic.apiKeyAuthSchemeXApiKey') },
            { value: 'authorization_bearer', label: t('admin.accounts.anthropic.apiKeyAuthSchemeBearer') }
          ]" />
        </div>
      </div>

      <!-- Anthropic API Key: Web Search Emulation (hidden when global disabled) -->
      <div
        v-if="account?.platform === 'anthropic' && account?.type === 'apikey' && webSearchGlobalEnabled"
        class="components-account-edit-account-modal__panel-3"
      >
        <div class="components-account-edit-account-modal__panel-27">
          <div>
            <label class="components-account-edit-account-modal__label input-label">{{ t('admin.accounts.anthropic.webSearchEmulation') }}</label>
            <p class="components-account-edit-account-modal__description-4">
              {{ t('admin.accounts.anthropic.webSearchEmulationDesc') }}
            </p>
          </div>
          <Select v-model="webSearchEmulationMode" class="components-account-edit-account-modal__field-7" :options="[
            { value: 'default', label: t('admin.accounts.anthropic.webSearchDefault') },
            { value: 'enabled', label: t('admin.accounts.anthropic.webSearchEnabled') },
            { value: 'disabled', label: t('admin.accounts.anthropic.webSearchDisabled') }
          ]" />
        </div>
      </div>

      <!-- 配额控制 (Anthropic apikey/bedrock: 配额限制 + 亲和) -->
      <div
        v-if="account?.platform === 'anthropic' && (account?.type === 'apikey' || account?.type === 'bedrock')"
        class="components-account-edit-account-modal__panel-22"
      >
        <div class="components-account-edit-account-modal__panel-44">
          <h3 class="components-account-edit-account-modal__heading input-label">{{ t('admin.accounts.quotaControl.title') }}</h3>
          <p class="components-account-edit-account-modal__description-4">
            {{ t('admin.accounts.quotaControl.hint') }}
          </p>
        </div>
        <QuotaLimitCard
          :totalLimit="editQuotaLimit"
          :dailyLimit="editQuotaDailyLimit"
          :weeklyLimit="editQuotaWeeklyLimit"
          :dailyResetMode="editDailyResetMode"
          :dailyResetHour="editDailyResetHour"
          :weeklyResetMode="editWeeklyResetMode"
          :weeklyResetDay="editWeeklyResetDay"
          :weeklyResetHour="editWeeklyResetHour"
          :resetTimezone="editResetTimezone"
          :quotaNotifyGlobalEnabled="quotaNotifyGlobalEnabled"
          :quotaNotifyDailyEnabled="quotaNotifyState.daily.enabled"
          :quotaNotifyDailyThreshold="quotaNotifyState.daily.threshold"
          :quotaNotifyDailyThresholdType="quotaNotifyState.daily.thresholdType"
          :quotaNotifyWeeklyEnabled="quotaNotifyState.weekly.enabled"
          :quotaNotifyWeeklyThreshold="quotaNotifyState.weekly.threshold"
          :quotaNotifyWeeklyThresholdType="quotaNotifyState.weekly.thresholdType"
          :quotaNotifyTotalEnabled="quotaNotifyState.total.enabled"
          :quotaNotifyTotalThreshold="quotaNotifyState.total.threshold"
          :quotaNotifyTotalThresholdType="quotaNotifyState.total.thresholdType"
          @update:totalLimit="editQuotaLimit = $event"
          @update:dailyLimit="editQuotaDailyLimit = $event"
          @update:weeklyLimit="editQuotaWeeklyLimit = $event"
          @update:dailyResetMode="editDailyResetMode = $event"
          @update:dailyResetHour="editDailyResetHour = $event"
          @update:weeklyResetMode="editWeeklyResetMode = $event"
          @update:weeklyResetDay="editWeeklyResetDay = $event"
          @update:weeklyResetHour="editWeeklyResetHour = $event"
          @update:resetTimezone="editResetTimezone = $event"
          @update:quotaNotifyDailyEnabled="quotaNotifyState.daily.enabled = $event"
          @update:quotaNotifyDailyThreshold="quotaNotifyState.daily.threshold = $event"
          @update:quotaNotifyDailyThresholdType="quotaNotifyState.daily.thresholdType = $event"
          @update:quotaNotifyWeeklyEnabled="quotaNotifyState.weekly.enabled = $event"
          @update:quotaNotifyWeeklyThreshold="quotaNotifyState.weekly.threshold = $event"
          @update:quotaNotifyWeeklyThresholdType="quotaNotifyState.weekly.thresholdType = $event"
          @update:quotaNotifyTotalEnabled="quotaNotifyState.total.enabled = $event"
          @update:quotaNotifyTotalThreshold="quotaNotifyState.total.threshold = $event"
          @update:quotaNotifyTotalThresholdType="quotaNotifyState.total.thresholdType = $event"
        />
      </div>
      <!-- 配额控制 (非 Anthropic apikey/bedrock) -->
      <div
        v-else-if="account?.type === 'apikey' || account?.type === 'bedrock'"
        class="components-account-edit-account-modal__panel-22"
      >
        <div class="components-account-edit-account-modal__panel-44">
          <h3 class="components-account-edit-account-modal__heading input-label">{{ t('admin.accounts.quotaControl.title') }}</h3>
          <p class="components-account-edit-account-modal__description-4">
            {{ t('admin.accounts.quotaLimitHint') }}
          </p>
        </div>
        <QuotaLimitCard
          :totalLimit="editQuotaLimit"
          :dailyLimit="editQuotaDailyLimit"
          :weeklyLimit="editQuotaWeeklyLimit"
          :dailyResetMode="editDailyResetMode"
          :dailyResetHour="editDailyResetHour"
          :weeklyResetMode="editWeeklyResetMode"
          :weeklyResetDay="editWeeklyResetDay"
          :weeklyResetHour="editWeeklyResetHour"
          :resetTimezone="editResetTimezone"
          :quotaNotifyGlobalEnabled="quotaNotifyGlobalEnabled"
          :quotaNotifyDailyEnabled="quotaNotifyState.daily.enabled"
          :quotaNotifyDailyThreshold="quotaNotifyState.daily.threshold"
          :quotaNotifyDailyThresholdType="quotaNotifyState.daily.thresholdType"
          :quotaNotifyWeeklyEnabled="quotaNotifyState.weekly.enabled"
          :quotaNotifyWeeklyThreshold="quotaNotifyState.weekly.threshold"
          :quotaNotifyWeeklyThresholdType="quotaNotifyState.weekly.thresholdType"
          :quotaNotifyTotalEnabled="quotaNotifyState.total.enabled"
          :quotaNotifyTotalThreshold="quotaNotifyState.total.threshold"
          :quotaNotifyTotalThresholdType="quotaNotifyState.total.thresholdType"
          @update:totalLimit="editQuotaLimit = $event"
          @update:dailyLimit="editQuotaDailyLimit = $event"
          @update:weeklyLimit="editQuotaWeeklyLimit = $event"
          @update:dailyResetMode="editDailyResetMode = $event"
          @update:dailyResetHour="editDailyResetHour = $event"
          @update:weeklyResetMode="editWeeklyResetMode = $event"
          @update:weeklyResetDay="editWeeklyResetDay = $event"
          @update:weeklyResetHour="editWeeklyResetHour = $event"
          @update:resetTimezone="editResetTimezone = $event"
          @update:quotaNotifyDailyEnabled="quotaNotifyState.daily.enabled = $event"
          @update:quotaNotifyDailyThreshold="quotaNotifyState.daily.threshold = $event"
          @update:quotaNotifyDailyThresholdType="quotaNotifyState.daily.thresholdType = $event"
          @update:quotaNotifyWeeklyEnabled="quotaNotifyState.weekly.enabled = $event"
          @update:quotaNotifyWeeklyThreshold="quotaNotifyState.weekly.threshold = $event"
          @update:quotaNotifyWeeklyThresholdType="quotaNotifyState.weekly.thresholdType = $event"
          @update:quotaNotifyTotalEnabled="quotaNotifyState.total.enabled = $event"
          @update:quotaNotifyTotalThreshold="quotaNotifyState.total.threshold = $event"
          @update:quotaNotifyTotalThresholdType="quotaNotifyState.total.thresholdType = $event"
        />
      </div>

      <!-- OpenAI API 长上下文计费开关 -->
      <div
        v-if="account?.platform === 'openai' && !isSparkShadow && (account?.type === 'oauth' || account?.type === 'setup-token' || account?.type === 'apikey')"
        class="components-account-edit-account-modal__panel-3"
      >
        <div class="components-account-edit-account-modal__panel-16">
          <div>
            <label class="components-account-edit-account-modal__label input-label">{{ t('admin.accounts.openai.longContextBilling') }}</label>
            <p class="components-account-edit-account-modal__description-4">
              {{ t('admin.accounts.openai.longContextBillingDesc') }}
            </p>
          </div>
          <button
            type="button"
            data-testid="openai-long-context-billing-toggle"
            role="switch"
            :aria-checked="openAILongContextBillingEnabled"
            @click="openAILongContextBillingEnabled = !openAILongContextBillingEnabled"
            :class="[
              'components-account-edit-account-modal__action-20',
              openAILongContextBillingEnabled ? 'components-account-edit-account-modal__action-21' : 'components-account-edit-account-modal__action-22'
            ]"
          >
            <span
              :class="[
                'components-account-edit-account-modal__text-16',
                openAILongContextBillingEnabled ? 'toggle-thumb--on' : 'components-account-edit-account-modal__text-17'
              ]"
            />
          </button>
        </div>
      </div>

      <div
        v-if="account?.platform === 'openai' && (account?.type === 'oauth' || account?.type === 'setup-token')"
        class="components-account-edit-account-modal__panel-3"
      >
        <div class="components-account-edit-account-modal__panel-27">
          <div>
            <label class="components-account-edit-account-modal__label input-label">{{ t('admin.accounts.openai.codexCLIOnly') }}</label>
            <p class="components-account-edit-account-modal__description-4">
              {{ t('admin.accounts.openai.codexCLIOnlyDesc') }}
            </p>
          </div>
          <button
            type="button"
            @click="codexCLIOnlyEnabled = !codexCLIOnlyEnabled"
            :class="[
              'components-account-edit-account-modal__action-20',
              codexCLIOnlyEnabled ? 'components-account-edit-account-modal__action-21' : 'components-account-edit-account-modal__action-22'
            ]"
          >
            <span
              :class="[
                'components-account-edit-account-modal__text-16',
                codexCLIOnlyEnabled ? 'toggle-thumb--on' : 'components-account-edit-account-modal__text-17'
              ]"
            />
          </button>
        </div>
        <div
          v-if="codexCLIOnlyEnabled"
          class="components-account-edit-account-modal__panel-45"
        >
          <div>
            <label class="components-account-edit-account-modal__label input-label">{{ t('admin.accounts.openai.codexCLIOnlyAppServer') }}</label>
            <p class="components-account-edit-account-modal__description-4">
              {{ t('admin.accounts.openai.codexCLIOnlyAppServerDesc') }}
            </p>
          </div>
          <button
            type="button"
            @click="codexCLIOnlyAppServerEnabled = !codexCLIOnlyAppServerEnabled"
            :class="[
              'components-account-edit-account-modal__action-20',
              codexCLIOnlyAppServerEnabled ? 'components-account-edit-account-modal__action-21' : 'components-account-edit-account-modal__action-22'
            ]"
          >
            <span
              :class="[
                'components-account-edit-account-modal__text-16',
                codexCLIOnlyAppServerEnabled ? 'toggle-thumb--on' : 'components-account-edit-account-modal__text-17'
              ]"
            />
          </button>
        </div>
      </div>

      <!-- Codex 指纹收敛模式（仅 OpenAI OAuth） -->
      <div
        v-if="account?.platform === 'openai' && account?.type === 'oauth'"
        class="components-account-edit-account-modal__panel-3"
      >
        <div class="components-account-edit-account-modal__panel-16">
          <div class="components-account-edit-account-modal__panel-17">
            <label class="components-account-edit-account-modal__label input-label">{{ t('admin.accounts.openai.codexFingerprintMode') }}</label>
            <p class="components-account-edit-account-modal__description-4">
              {{ t('admin.accounts.openai.codexFingerprintModeDesc') }}
            </p>
          </div>
          <div class="components-account-edit-account-modal__panel-46">
            <Select v-model="codexFingerprintMode" data-testid="edit-codex-fingerprint-mode-select" :options="codexFingerprintModeOptions" />
          </div>
        </div>
      </div>

      <!-- OpenAI 订阅档位手动覆盖（Plus/Pro/Free），仅 OAuth 非影子账号 -->
      <div
        v-if="account?.platform === 'openai' && account?.type === 'oauth' && !isSparkShadow"
        class="components-account-edit-account-modal__panel-3"
      >
        <div class="components-account-edit-account-modal__panel-16">
          <div class="components-account-edit-account-modal__panel-17">
            <label class="components-account-edit-account-modal__label input-label">{{ t('admin.accounts.openai.planType') }}</label>
            <p class="components-account-edit-account-modal__description-4">
              {{ t('admin.accounts.openai.planTypeDesc') }}
            </p>
          </div>
          <div class="components-account-edit-account-modal__panel-47">
            <Select v-model="editPlanType" :options="planTypeOptions" />
          </div>
        </div>
      </div>

      <div
        v-if="account?.platform === 'openai' && (account?.type === 'oauth' || account?.type === 'setup-token' || account?.type === 'apikey')"
        class="components-account-edit-account-modal__panel-22"
      >
        <div class="components-account-edit-account-modal__panel-27">
          <div>
            <label class="components-account-edit-account-modal__label input-label">{{ t('admin.accounts.openai.compactMode') }}</label>
            <p class="components-account-edit-account-modal__description-4">
              {{ t('admin.accounts.openai.compactModeDesc') }}
            </p>
          </div>
          <div class="components-account-edit-account-modal__panel-48">
            <Select v-model="openAICompactMode" :options="openAICompactModeOptions" />
          </div>
        </div>
        <div class="components-account-edit-account-modal__panel-41">
          <span class="components-account-edit-account-modal__text-9">{{ t(openAICompactStatusKey) }}</span>
          <span
            v-if="account?.extra?.openai_compact_checked_at"
            class="components-account-edit-account-modal__text-11"
          >
            {{ t('admin.accounts.openai.compactLastChecked') }}:
            {{ formatDateTime(new Date(String(account.extra.openai_compact_checked_at))) }}
          </span>
        </div>
        <div>
          <label class="input-label">{{ t('admin.accounts.openai.compactModelMapping') }}</label>
          <p class="input-hint">{{ t('admin.accounts.openai.compactModelMappingDesc') }}</p>
          <div v-if="openAICompactModelMappings.length > 0" class="components-account-edit-account-modal__panel-7">
            <div
              v-for="(mapping, index) in openAICompactModelMappings"
              :key="getOpenAICompactModelMappingKey(mapping)"
              class="components-account-edit-account-modal__panel-8"
            >
              <input
                v-model="mapping.from"
                type="text"
                class="components-account-edit-account-modal__field-2 input"
                :placeholder="t('admin.accounts.fromModel')"
              />
              <span class="components-account-edit-account-modal__text-4">→</span>
              <input
                v-model="mapping.to"
                type="text"
                class="components-account-edit-account-modal__field-2 input"
                :placeholder="t('admin.accounts.toModel')"
              />
              <button type="button" @click="removeOpenAICompactModelMapping(index)" class="components-account-edit-account-modal__action-5">
                <Icon name="trash" size="sm" />
              </button>
            </div>
          </div>
          <button type="button" @click="addOpenAICompactModelMapping" class="components-account-edit-account-modal__action-6 btn btn-secondary">
            + {{ t('admin.accounts.addMapping') }}
          </button>
        </div>
      </div>

      <div>
        <div class="components-account-edit-account-modal__panel-27">
          <div>
            <label class="components-account-edit-account-modal__label input-label">{{
              t('admin.accounts.autoPauseOnExpired')
            }}</label>
            <p class="components-account-edit-account-modal__description-4">
              {{ t('admin.accounts.autoPauseOnExpiredDesc') }}
            </p>
          </div>
          <button
            type="button"
            @click="autoPauseOnExpired = !autoPauseOnExpired"
            :class="[
              'components-account-edit-account-modal__action-20',
              autoPauseOnExpired ? 'components-account-edit-account-modal__action-21' : 'components-account-edit-account-modal__action-22'
            ]"
          >
            <span
              :class="[
                'components-account-edit-account-modal__text-16',
                autoPauseOnExpired ? 'toggle-thumb--on' : 'components-account-edit-account-modal__text-17'
              ]"
            />
          </button>
        </div>
      </div>

      <div
        v-if="account?.platform === 'openai'"
        class="components-account-edit-account-modal__panel-22"
      >
        <div class="components-account-edit-account-modal__panel-18">
          <div class="components-account-edit-account-modal__panel-27">
            <label class="components-account-edit-account-modal__label input-label">{{ t('admin.accounts.autoPause5hDisabled') }}</label>
            <button
              type="button"
              @click="autoPause5hDisabled = !autoPause5hDisabled"
              :class="[
                'components-account-edit-account-modal__action-20',
                autoPause5hDisabled ? 'components-account-edit-account-modal__action-21' : 'components-account-edit-account-modal__action-22'
              ]"
              data-testid="auto-pause-5h-disabled"
            >
              <span
                :class="[
                  'components-account-edit-account-modal__text-16',
                  autoPause5hDisabled ? 'toggle-thumb--on' : 'components-account-edit-account-modal__text-17'
                ]"
              />
            </button>
          </div>
          <p class="input-hint">{{ t('admin.accounts.autoPauseDisabledHint') }}</p>
        </div>
        <div>
          <label class="input-label">{{ t('admin.accounts.autoPause5hThreshold') }}</label>
          <input
            v-model.number="autoPause5hThreshold"
            type="number"
            min="0"
            max="100"
            step="0.1"
            class="input"
            :disabled="autoPause5hDisabled"
            data-testid="auto-pause-5h-threshold"
          />
          <p class="input-hint">{{ t('admin.accounts.autoPauseThresholdHint') }}</p>
        </div>
        <div class="components-account-edit-account-modal__panel-18">
          <div class="components-account-edit-account-modal__panel-27">
            <label class="components-account-edit-account-modal__label input-label">{{ t('admin.accounts.autoPause7dDisabled') }}</label>
            <button
              type="button"
              @click="autoPause7dDisabled = !autoPause7dDisabled"
              :class="[
                'components-account-edit-account-modal__action-20',
                autoPause7dDisabled ? 'components-account-edit-account-modal__action-21' : 'components-account-edit-account-modal__action-22'
              ]"
              data-testid="auto-pause-7d-disabled"
            >
              <span
                :class="[
                  'components-account-edit-account-modal__text-16',
                  autoPause7dDisabled ? 'toggle-thumb--on' : 'components-account-edit-account-modal__text-17'
                ]"
              />
            </button>
          </div>
          <p class="input-hint">{{ t('admin.accounts.autoPauseDisabledHint') }}</p>
        </div>
        <div>
          <label class="input-label">{{ t('admin.accounts.autoPause7dThreshold') }}</label>
          <input
            v-model.number="autoPause7dThreshold"
            type="number"
            min="0"
            max="100"
            step="0.1"
            class="input"
            :disabled="autoPause7dDisabled"
            data-testid="auto-pause-7d-threshold"
          />
          <p class="input-hint">{{ t('admin.accounts.autoPauseThresholdHint') }}</p>
        </div>
      </div>

      <!-- 配额控制 (Anthropic OAuth/SetupToken: 亲和 + 窗口费用 + 会话 + RPM 等) -->
      <div
        v-if="account?.platform === 'anthropic' && (account?.type === 'oauth' || account?.type === 'setup-token')"
        class="components-account-edit-account-modal__panel-22"
      >
        <div class="components-account-edit-account-modal__panel-44">
          <h3 class="components-account-edit-account-modal__heading input-label">{{ t('admin.accounts.quotaControl.title') }}</h3>
          <p class="components-account-edit-account-modal__description-4">
            {{ t('admin.accounts.quotaControl.hint') }}
          </p>
        </div>

        <!-- Window Cost Limit -->
        <div class="components-account-edit-account-modal__panel-49">
          <div class="components-account-edit-account-modal__panel-10">
            <div>
              <label class="components-account-edit-account-modal__label input-label">{{ t('admin.accounts.quotaControl.windowCost.label') }}</label>
              <p class="components-account-edit-account-modal__description-4">
                {{ t('admin.accounts.quotaControl.windowCost.hint') }}
              </p>
            </div>
            <button
              type="button"
              @click="windowCostEnabled = !windowCostEnabled"
              :class="[
                'components-account-edit-account-modal__action-20',
                windowCostEnabled ? 'components-account-edit-account-modal__action-21' : 'components-account-edit-account-modal__action-22'
              ]"
            >
              <span
                :class="[
                  'components-account-edit-account-modal__text-16',
                  windowCostEnabled ? 'toggle-thumb--on' : 'components-account-edit-account-modal__text-17'
                ]"
              />
            </button>
          </div>

          <div v-if="windowCostEnabled" class="components-account-edit-account-modal__panel-50">
            <div>
              <label class="input-label">{{ t('admin.accounts.quotaControl.windowCost.limit') }}</label>
              <div class="components-account-edit-account-modal__panel-51">
                <span class="components-account-edit-account-modal__text-12">$</span>
                <input
                  v-model.number="windowCostLimit"
                  type="number"
                  min="0"
                  step="1"
                  class="components-account-edit-account-modal__field-8 input"
                  :placeholder="t('admin.accounts.quotaControl.windowCost.limitPlaceholder')"
                />
              </div>
              <p class="input-hint">{{ t('admin.accounts.quotaControl.windowCost.limitHint') }}</p>
            </div>
            <div>
              <label class="input-label">{{ t('admin.accounts.quotaControl.windowCost.stickyReserve') }}</label>
              <div class="components-account-edit-account-modal__panel-51">
                <span class="components-account-edit-account-modal__text-12">$</span>
                <input
                  v-model.number="windowCostStickyReserve"
                  type="number"
                  min="0"
                  step="1"
                  class="components-account-edit-account-modal__field-8 input"
                  :placeholder="t('admin.accounts.quotaControl.windowCost.stickyReservePlaceholder')"
                />
              </div>
              <p class="input-hint">{{ t('admin.accounts.quotaControl.windowCost.stickyReserveHint') }}</p>
            </div>
          </div>
        </div>

        <!-- Session Limit -->
        <div class="components-account-edit-account-modal__panel-49">
          <div class="components-account-edit-account-modal__panel-10">
            <div>
              <label class="components-account-edit-account-modal__label input-label">{{ t('admin.accounts.quotaControl.sessionLimit.label') }}</label>
              <p class="components-account-edit-account-modal__description-4">
                {{ t('admin.accounts.quotaControl.sessionLimit.hint') }}
              </p>
            </div>
            <button
              type="button"
              @click="sessionLimitEnabled = !sessionLimitEnabled"
              :class="[
                'components-account-edit-account-modal__action-20',
                sessionLimitEnabled ? 'components-account-edit-account-modal__action-21' : 'components-account-edit-account-modal__action-22'
              ]"
            >
              <span
                :class="[
                  'components-account-edit-account-modal__text-16',
                  sessionLimitEnabled ? 'toggle-thumb--on' : 'components-account-edit-account-modal__text-17'
                ]"
              />
            </button>
          </div>

          <div v-if="sessionLimitEnabled" class="components-account-edit-account-modal__panel-50">
            <div>
              <label class="input-label">{{ t('admin.accounts.quotaControl.sessionLimit.maxSessions') }}</label>
              <input
                v-model.number="maxSessions"
                type="number"
                min="1"
                step="1"
                class="input"
                :placeholder="t('admin.accounts.quotaControl.sessionLimit.maxSessionsPlaceholder')"
              />
              <p class="input-hint">{{ t('admin.accounts.quotaControl.sessionLimit.maxSessionsHint') }}</p>
            </div>
            <div>
              <label class="input-label">{{ t('admin.accounts.quotaControl.sessionLimit.idleTimeout') }}</label>
              <div class="components-account-edit-account-modal__panel-51">
                <input
                  v-model.number="sessionIdleTimeout"
                  type="number"
                  min="1"
                  step="1"
                  class="components-account-edit-account-modal__field-9 input"
                  :placeholder="t('admin.accounts.quotaControl.sessionLimit.idleTimeoutPlaceholder')"
                />
                <span class="components-account-edit-account-modal__text-13">{{ t('common.minutes') }}</span>
              </div>
              <p class="input-hint">{{ t('admin.accounts.quotaControl.sessionLimit.idleTimeoutHint') }}</p>
            </div>
          </div>
        </div>

        <!-- RPM Limit -->
        <div class="components-account-edit-account-modal__panel-49">
          <div class="components-account-edit-account-modal__panel-10">
            <div>
              <label class="components-account-edit-account-modal__label input-label">{{ t('admin.accounts.quotaControl.rpmLimit.label') }}</label>
              <p class="components-account-edit-account-modal__description-4">
                {{ t('admin.accounts.quotaControl.rpmLimit.hint') }}
              </p>
            </div>
            <button
              type="button"
              @click="rpmLimitEnabled = !rpmLimitEnabled"
              :class="[
                'components-account-edit-account-modal__action-20',
                rpmLimitEnabled ? 'components-account-edit-account-modal__action-21' : 'components-account-edit-account-modal__action-22'
              ]"
            >
              <span
                :class="[
                  'components-account-edit-account-modal__text-16',
                  rpmLimitEnabled ? 'toggle-thumb--on' : 'components-account-edit-account-modal__text-17'
                ]"
              />
            </button>
          </div>

          <div v-if="rpmLimitEnabled" class="components-account-edit-account-modal__panel">
            <div>
              <label class="input-label">{{ t('admin.accounts.quotaControl.rpmLimit.baseRpm') }}</label>
              <input
                v-model.number="baseRpm"
                type="number"
                min="1"
                max="1000"
                step="1"
                class="input"
                :placeholder="t('admin.accounts.quotaControl.rpmLimit.baseRpmPlaceholder')"
              />
              <p class="input-hint">{{ t('admin.accounts.quotaControl.rpmLimit.baseRpmHint') }}</p>
            </div>

            <div>
              <label class="input-label">{{ t('admin.accounts.quotaControl.rpmLimit.strategy') }}</label>
              <div class="components-account-edit-account-modal__panel-52">
                <button
                  type="button"
                  @click="rpmStrategy = 'tiered'"
                  :class="[
                    'components-account-edit-account-modal__action-27',
                    rpmStrategy === 'tiered'
                      ? 'components-account-edit-account-modal__action-16'
                      : 'components-account-edit-account-modal__action-17'
                  ]"
                >
                  <div class="components-account-edit-account-modal__panel-53">
                    <div>{{ t('admin.accounts.quotaControl.rpmLimit.strategyTiered') }}</div>
                    <div class="components-account-edit-account-modal__panel-54">{{ t('admin.accounts.quotaControl.rpmLimit.strategyTieredHint') }}</div>
                  </div>
                </button>
                <button
                  type="button"
                  @click="rpmStrategy = 'sticky_exempt'"
                  :class="[
                    'components-account-edit-account-modal__action-27',
                    rpmStrategy === 'sticky_exempt'
                      ? 'components-account-edit-account-modal__action-16'
                      : 'components-account-edit-account-modal__action-17'
                  ]"
                >
                  <div class="components-account-edit-account-modal__panel-53">
                    <div>{{ t('admin.accounts.quotaControl.rpmLimit.strategyStickyExempt') }}</div>
                    <div class="components-account-edit-account-modal__panel-54">{{ t('admin.accounts.quotaControl.rpmLimit.strategyStickyExemptHint') }}</div>
                  </div>
                </button>
              </div>
            </div>

            <div v-if="rpmStrategy === 'tiered'">
              <label class="input-label">{{ t('admin.accounts.quotaControl.rpmLimit.stickyBuffer') }}</label>
              <input
                v-model.number="rpmStickyBuffer"
                type="number"
                min="1"
                step="1"
                class="input"
                :placeholder="t('admin.accounts.quotaControl.rpmLimit.stickyBufferPlaceholder')"
              />
              <p class="input-hint">{{ t('admin.accounts.quotaControl.rpmLimit.stickyBufferHint') }}</p>
            </div>

          </div>

          <!-- 用户消息限速模式（独立于 RPM 开关，始终可见） -->
          <div class="components-account-edit-account-modal__panel-55">
            <label class="input-label">{{ t('admin.accounts.quotaControl.rpmLimit.userMsgQueue') }}</label>
            <p class="components-account-edit-account-modal__description-10">
              {{ t('admin.accounts.quotaControl.rpmLimit.userMsgQueueHint') }}
            </p>
            <div class="components-account-edit-account-modal__panel-56">
              <button type="button" v-for="opt in umqModeOptions" :key="opt.value"
                @click="userMsgQueueMode = opt.value"
                :class="[
                  'components-account-edit-account-modal__action-28',
                  userMsgQueueMode === opt.value
                    ? 'components-account-edit-account-modal__action-29'
                    : 'components-account-edit-account-modal__action-30'
                ]">
                {{ opt.label }}
              </button>
            </div>
          </div>
        </div>

        <!-- TLS Fingerprint -->
        <div class="components-account-edit-account-modal__panel-49">
          <div class="components-account-edit-account-modal__panel-27">
            <div>
              <label class="components-account-edit-account-modal__label input-label">{{ t('admin.accounts.quotaControl.tlsFingerprint.label') }}</label>
              <p class="components-account-edit-account-modal__description-4">
                {{ t('admin.accounts.quotaControl.tlsFingerprint.hint') }}
              </p>
            </div>
            <button
              type="button"
              @click="tlsFingerprintEnabled = !tlsFingerprintEnabled"
              :class="[
                'components-account-edit-account-modal__action-20',
                tlsFingerprintEnabled ? 'components-account-edit-account-modal__action-21' : 'components-account-edit-account-modal__action-22'
              ]"
            >
              <span
                :class="[
                  'components-account-edit-account-modal__text-16',
                  tlsFingerprintEnabled ? 'toggle-thumb--on' : 'components-account-edit-account-modal__text-17'
                ]"
              />
            </button>
          </div>
          <!-- Profile selector -->
          <div v-if="tlsFingerprintEnabled" class="components-account-edit-account-modal__panel-12">
            <Select v-model="tlsFingerprintProfileId" :options="[
              { value: null, label: t('admin.accounts.quotaControl.tlsFingerprint.defaultProfile') },
              ...(tlsFingerprintProfiles.length ? [{ value: 'random', label: t('admin.accounts.quotaControl.tlsFingerprint.randomProfile') }] : []),
              ...tlsFingerprintProfiles.map(p => ({ value: p.id, label: p.name }))
            ]" />
          </div>
        </div>

        <!-- Session ID Masking -->
        <div class="components-account-edit-account-modal__panel-49">
          <div class="components-account-edit-account-modal__panel-27">
            <div>
              <label class="components-account-edit-account-modal__label input-label">{{ t('admin.accounts.quotaControl.sessionIdMasking.label') }}</label>
              <p class="components-account-edit-account-modal__description-4">
                {{ t('admin.accounts.quotaControl.sessionIdMasking.hint') }}
              </p>
            </div>
            <button
              type="button"
              @click="sessionIdMaskingEnabled = !sessionIdMaskingEnabled"
              :class="[
                'components-account-edit-account-modal__action-20',
                sessionIdMaskingEnabled ? 'components-account-edit-account-modal__action-21' : 'components-account-edit-account-modal__action-22'
              ]"
            >
              <span
                :class="[
                  'components-account-edit-account-modal__text-16',
                  sessionIdMaskingEnabled ? 'toggle-thumb--on' : 'components-account-edit-account-modal__text-17'
                ]"
              />
            </button>
          </div>
        </div>

        <!-- Cache TTL Override -->
        <div class="components-account-edit-account-modal__panel-49">
          <div class="components-account-edit-account-modal__panel-27">
            <div>
              <label class="components-account-edit-account-modal__label input-label">{{ t('admin.accounts.quotaControl.cacheTTLOverride.label') }}</label>
              <p class="components-account-edit-account-modal__description-4">
                {{ t('admin.accounts.quotaControl.cacheTTLOverride.hint') }}
              </p>
            </div>
            <button
              type="button"
              @click="cacheTTLOverrideEnabled = !cacheTTLOverrideEnabled"
              :class="[
                'components-account-edit-account-modal__action-20',
                cacheTTLOverrideEnabled ? 'components-account-edit-account-modal__action-21' : 'components-account-edit-account-modal__action-22'
              ]"
            >
              <span
                :class="[
                  'components-account-edit-account-modal__text-16',
                  cacheTTLOverrideEnabled ? 'toggle-thumb--on' : 'components-account-edit-account-modal__text-17'
                ]"
              />
            </button>
          </div>
          <div v-if="cacheTTLOverrideEnabled" class="components-account-edit-account-modal__panel-12">
            <label class="components-account-edit-account-modal__label-5 input-label">{{ t('admin.accounts.quotaControl.cacheTTLOverride.target') }}</label>
            <Select v-model="cacheTTLOverrideTarget" class="components-account-edit-account-modal__field-10" :options="[
              { value: '5m', label: '5m' },
              { value: '1h', label: '1h' }
            ]" />
            <p class="components-account-edit-account-modal__description-4">
              {{ t('admin.accounts.quotaControl.cacheTTLOverride.targetHint') }}
            </p>
          </div>
        </div>

        <!-- Custom Base URL Relay -->
        <div class="components-account-edit-account-modal__panel-49">
          <div class="components-account-edit-account-modal__panel-27">
            <div>
              <label class="components-account-edit-account-modal__label input-label">{{ t('admin.accounts.quotaControl.customBaseUrl.label') }}</label>
              <p class="components-account-edit-account-modal__description-4">
                {{ t('admin.accounts.quotaControl.customBaseUrl.hint') }}
              </p>
            </div>
            <button
              type="button"
              @click="customBaseUrlEnabled = !customBaseUrlEnabled"
              :class="[
                'components-account-edit-account-modal__action-20',
                customBaseUrlEnabled ? 'components-account-edit-account-modal__action-21' : 'components-account-edit-account-modal__action-22'
              ]"
            >
              <span
                :class="[
                  'components-account-edit-account-modal__text-16',
                  customBaseUrlEnabled ? 'toggle-thumb--on' : 'components-account-edit-account-modal__text-17'
                ]"
              />
            </button>
          </div>
          <div v-if="customBaseUrlEnabled" class="components-account-edit-account-modal__panel-12">
            <input
              v-model="customBaseUrl"
              type="text"
              class="input"
              :placeholder="t('admin.accounts.quotaControl.customBaseUrl.urlHint')"
            />
          </div>
        </div>
      </div>

      <div class="components-account-edit-account-modal__panel-3">
        <div>
          <label class="input-label">{{ t('common.status') }}</label>
          <Select v-model="form.status" :options="statusOptions" />
        </div>

      </div>

      <!-- Group Selection - 仅标准模式显示 -->
      <GroupSelector
        v-if="!authStore.isSimpleMode"
        v-model="form.group_ids"
        :groups="groups"
        :platform="account?.platform"
        data-tour="account-form-groups"
      />

    </form>

    <template #footer>
      <div v-if="account" class="components-account-edit-account-modal__panel-61">
        <button @click="handleClose" type="button" class="btn btn-secondary">
          {{ t('common.cancel') }}
        </button>
        <button
          type="submit"
          form="edit-account-form"
          :disabled="submitting"
          class="btn btn-primary"
          data-tour="account-form-submit"
        >
          <svg
            v-if="submitting"
            class="components-account-edit-account-modal__icon-6"
            fill="none"
            viewBox="0 0 24 24"
          >
            <circle
              class="components-account-edit-account-modal__circle"
              cx="12"
              cy="12"
              r="10"
              stroke="currentColor"
              stroke-width="4"
            ></circle>
            <path
              class="components-account-edit-account-modal__path"
              fill="currentColor"
              d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
            ></path>
          </svg>
          {{ submitting ? t('admin.accounts.updating') : t('common.update') }}
        </button>
      </div>
    </template>
  </BaseDialog>

</template>

<script setup lang="ts">
import { ref, reactive, computed, watch, nextTick } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import { useAuthStore } from '@/stores/auth'
import { adminAPI } from '@/api/admin'
import { useQuotaNotifyState } from '@/composables/useQuotaNotifyState'
import type {
  Account,
  Proxy,
  AdminGroup,
  OpenAICompactMode,
  OpenAIResponsesMode,
  OpenAIEndpointCapability,
  OllamaCloudUsageState
} from '@/types'
import BaseDialog from '@/components/common/BaseDialog.vue'
import Select from '@/components/common/Select.vue'
import Toggle from '@/components/common/Toggle.vue'
import Icon from '@/components/icons/Icon.vue'
import ProxySelector from '@/components/common/ProxySelector.vue'
import ProxyAdBanner from '@/components/common/ProxyAdBanner.vue'
import GroupSelector from '@/components/common/GroupSelector.vue'
import ModelWhitelistSelector from '@/components/account/ModelWhitelistSelector.vue'
import AccountModelRuleSelector from '@/components/account/AccountModelRuleSelector.vue'
import QuotaLimitCard from '@/components/account/QuotaLimitCard.vue'
import GrokBaseUrlPresets from '@/components/account/GrokBaseUrlPresets.vue'
import CnBaseUrlPresets from '@/components/account/CnBaseUrlPresets.vue'
import HeaderOverrideEditor from '@/components/account/HeaderOverrideEditor.vue'
import OllamaCloudUsageSettings from '@/components/account/OllamaCloudUsageSettings.vue'
import {
  applyHeaderOverride,
  applyInterceptWarmup,
  applyPlanType,
  buildPlanTypeOptions,
  readPlanType,
  isCustomGrokBaseUrl,
  isHeaderOverrideCapable,
  splitHeaderOverridesObject,
  validateHeaderOverrideRows,
  defaultCNBaseUrl,
  HEADER_OVERRIDE_ENABLED_CREDENTIAL_KEY,
  HEADER_OVERRIDES_CREDENTIAL_KEY,
  type CnAccountMode,
  type CnApiProtocol,
  type HeaderOverrideRow
} from '@/components/account/credentialsBuilder'
import { formatDateTime, formatDateTimeLocalInput, parseDateTimeLocalInput } from '@/utils/format'
import { createStableObjectKeyResolver } from '@/utils/stableObjectKey'
import { VERTEX_LOCATION_SELECT_OPTIONS } from '@/constants/account'
import {
  OPENAI_WS_MODE_CTX_POOL,
  OPENAI_WS_MODE_OFF,
  OPENAI_WS_MODE_PASSTHROUGH,
  OPENAI_WS_MODE_HTTP_BRIDGE,
  isOpenAIWSModeEnabled,
  resolveOpenAIWSModeConcurrencyHintKey,
  type OpenAIWSMode,
  resolveOpenAIWSModeFromExtra
} from '@/utils/openaiWsMode'
import {
  getPresetMappingsByPlatform,
  commonErrorCodes,
  buildModelMappingObject,
  splitModelMappingObject
} from '@/composables/useModelWhitelist'

interface Props {
  show: boolean
  account: Account | null
  proxies: Proxy[]
  groups: AdminGroup[]
}

const props = defineProps<Props>()
const emit = defineEmits<{
  close: []
  updated: [account: Account]
}>()

const { t } = useI18n()
const appStore = useAppStore()
const authStore = useAuthStore()

// Spark 影子账号(parent_account_id 非空):代理恒继承母账号,不可独立编辑(外审 B/P1),
// 故隐藏代理选择器。
const isSparkShadow = computed(() => props.account?.parent_account_id != null)

const handleOllamaCloudUsageUpdated = (state: OllamaCloudUsageState) => {
  if (props.account) emit('updated', { ...props.account, ollama_cloud_usage: state })
}

// Platform-specific hint for Base URL
const baseUrlHint = computed(() => {
  if (!props.account) return t('admin.accounts.baseUrlHint')
  if (props.account.platform === 'openai') return t('admin.accounts.openai.baseUrlHint')
  if (props.account.platform === 'grok') return ''
  return t('admin.accounts.baseUrlHint')
})

const bedrockPresets = computed(() => getPresetMappingsByPlatform('bedrock'))

// Model mapping type
interface ModelMapping {
  from: string
  to: string
}

interface TempUnschedRuleForm {
  error_code: number | null
  keywords: string
  duration_minutes: number | null
  description: string
}

// State
const submitting = ref(false)
const editBaseUrl = ref('https://api.anthropic.com')
const editApiKey = ref('')

// ── 国产供应商（Kimi / Zhipu / DeepSeek）account_mode / api_protocol 编辑 ──
// account_mode 决定额度/余额监控路径，api_protocol 决定转发端点与格式；
// 二者均可修正（早期创建的账号可能存错默认值），切换时重置 base_url 预置。
const isCNApiKeyAccount = computed(
  () =>
    props.account?.type === 'apikey' &&
    (props.account.platform === 'kimi' ||
      props.account.platform === 'zhipu' ||
      props.account.platform === 'deepseek')
)
// CnBaseUrlPresets 的 platform prop 是平台字面量联合类型，模板里不能写
// `as` 断言（其中的 `|` 会被 eslint 误判为 Vue2 filter 语法），经此 computed 传递。
const cnPresetPlatform = computed<'kimi' | 'zhipu' | 'deepseek'>(() => {
  const platform = props.account?.platform
  if (platform === 'kimi' || platform === 'zhipu' || platform === 'deepseek') {
    return platform
  }
  return 'kimi'
})
const editApiProtocol = ref<CnApiProtocol>('chat_completions')
const editAccountMode = ref<CnAccountMode>('payg')
// 回填窗口标志：syncFormFromAccount 会同步改写 editAccountMode / editApiProtocol，
// 而 watcher（pre-flush）在同步代码执行完之后才触发——若不抑制，会把刚恢复的
// 存储版 base_url（可能是用户自定义/中转地址）覆盖为官方预设并在下次保存时持久化。
// nextTick 后解除，此后用户主动切换模式/协议仍正常联动重置。
const syncingForm = ref(false)
const cnAccountModeOptions = computed<Array<{ value: CnAccountMode; labelKey: 'payg' | 'coding' }>>(
  () => {
    // DeepSeek 无 coding 套餐（与创建弹窗一致），仅保留按量付费。
    if (props.account?.platform === 'deepseek') {
      return [{ value: 'payg', labelKey: 'payg' }]
    }
    return [
      { value: 'payg', labelKey: 'payg' },
      { value: 'coding', labelKey: 'coding' }
    ]
  }
)
const cnProtocolOptions = computed<Array<{ value: CnApiProtocol; labelKey: string }>>(() => {
  const opts: Array<{ value: CnApiProtocol; labelKey: string }> = [
    { value: 'chat_completions', labelKey: 'chatCompletions' },
    { value: 'anthropic', labelKey: 'anthropic' }
  ]
  if (props.account?.platform === 'deepseek') {
    opts.push({ value: 'responses', labelKey: 'responses' })
  }
  return opts
})
watch(editApiProtocol, (protocol) => {
  if (!isCNApiKeyAccount.value || syncingForm.value) return
  editBaseUrl.value = defaultCNBaseUrl(props.account!.platform, editAccountMode.value, protocol)
})
watch(editAccountMode, (mode) => {
  if (!isCNApiKeyAccount.value || syncingForm.value) return
  // deepseek 无 coding 套餐：防御性回退（UI 已隐藏该选项）。
  const effectiveMode = props.account!.platform === 'deepseek' && mode === 'coding' ? 'payg' : mode
  if (effectiveMode !== mode) {
    editAccountMode.value = effectiveMode
    return
  }
  editBaseUrl.value = defaultCNBaseUrl(props.account!.platform, mode, editApiProtocol.value)
})
const cnProtocolDescKey = computed(
  () => cnProtocolOptions.value.find(o => o.value === editApiProtocol.value)?.labelKey ?? 'chatCompletions'
)
// 点击预设端点：回填 base url 与对应模式/协议。
function onCnPresetSelect(preset: { mode: CnAccountMode; protocol: CnApiProtocol; url: string }) {
  editAccountMode.value = preset.mode
  editApiProtocol.value = preset.protocol
  editBaseUrl.value = preset.url
}
// Bedrock credentials
const editBedrockAccessKeyId = ref('')
const editBedrockSecretAccessKey = ref('')
const editBedrockSessionToken = ref('')
const editBedrockRegion = ref('')
const editBedrockForceGlobal = ref(false)
const editBedrockApiKeyValue = ref('')
const editVertexProjectId = ref('')
const editVertexClientEmail = ref('')
const editVertexLocation = ref('us-central1')
const isBedrockAPIKeyMode = computed(() =>
  props.account?.type === 'bedrock' &&
  (props.account?.credentials as Record<string, unknown>)?.auth_mode === 'apikey'
)
const modelMappings = ref<ModelMapping[]>([])
const openAICompactModelMappings = ref<ModelMapping[]>([])
const modelRestrictionMode = ref<'whitelist' | 'mapping'>('whitelist')
const allowedModels = ref<string[]>([])
const hasModelRestrictionValues = computed(() =>
  modelMappings.value.some(mapping => mapping.from.trim() !== '' || mapping.to.trim() !== '') ||
  allowedModels.value.length > 0
)
const DEFAULT_POOL_MODE_RETRY_COUNT = 3
const MAX_POOL_MODE_RETRY_COUNT = 10
const DEFAULT_POOL_MODE_RETRY_STATUS_CODES = [401, 403, 429]
const GROK_CLIENT_TOOL_CACHE_EXTRA_KEY = 'grok_client_tool_cache_enabled'
const poolModeEnabled = ref(false)
const poolModeRetryCount = ref(DEFAULT_POOL_MODE_RETRY_COUNT)
const poolModeRetryStatusCodesInput = ref('')

function parsePoolModeRetryStatusCodes(input: string): number[] {
  if (!input || !input.trim()) return []
  const seen = new Set<number>()
  const out: number[] = []
  for (const token of input.split(/[,\s]+/)) {
    const trimmed = token.trim()
    if (!trimmed) continue
    const n = Number(trimmed)
    if (!Number.isFinite(n) || !Number.isInteger(n)) continue
    if (n < 100 || n > 599) continue
    if (seen.has(n)) continue
    seen.add(n)
    out.push(n)
  }
  return out.sort((a, b) => a - b)
}

function formatPoolModeRetryStatusCodes(value: unknown): string {
  if (!Array.isArray(value)) return ''
  const out: number[] = []
  const seen = new Set<number>()
  for (const v of value) {
    const n = typeof v === 'string' ? Number(v.trim()) : Number(v)
    if (!Number.isFinite(n) || !Number.isInteger(n)) continue
    if (n < 100 || n > 599) continue
    if (seen.has(n)) continue
    seen.add(n)
    out.push(n)
  }
  return out.sort((a, b) => a - b).join(', ')
}
const customErrorCodesEnabled = ref(false)
const selectedErrorCodes = ref<number[]>([])
const customErrorCodeInput = ref<number | null>(null)
const headerOverrideEnabled = ref(false)
const headerOverrideRows = ref<HeaderOverrideRow[]>([])

const headerOverrideCapable = computed(
  () => !!props.account && isHeaderOverrideCapable(props.account.platform, props.account.type)
)

// Grok OAuth 自定义上游地址（仅转发端点；OAuth 授权/令牌刷新不受影响）
const grokOAuthCustomBaseUrlEnabled = ref(false)
const grokOAuthBaseUrl = ref('')
// Grok Free OAuth accounts use client-tool prompt caching by default. Keep an
// explicit false in the account extra as the opt-out signal.
const grokClientToolCacheEnabled = ref(true)

const interceptWarmupRequests = ref(false)
const autoPauseOnExpired = ref(false)
const autoPause5hThreshold = ref<number | null>(null)
const autoPause7dThreshold = ref<number | null>(null)
const autoPause5hDisabled = ref(false)
const autoPause7dDisabled = ref(false)
const upstreamBillingAutoProbeEnabled = ref(false)
const upstreamBillingRateSyncEnabled = ref(false)
const tempUnschedEnabled = ref(false)
const accountSchedulingThresholdOverrideEnabled = ref(false)
const accountSchedulingThresholdOverrideValue = ref(100)
const ACCOUNT_SCHEDULING_THRESHOLD_CREDENTIAL_KEY = 'account_scheduling_threshold'
const supportsAccountSchedulingThresholdOverride = computed(() =>
  supportsAccountSchedulingThresholdOverridePlatform(props.account?.platform)
)
const tempUnschedRules = ref<TempUnschedRuleForm[]>([])
const getModelMappingKey = createStableObjectKeyResolver<ModelMapping>('edit-model-mapping')
const getOpenAICompactModelMappingKey = createStableObjectKeyResolver<ModelMapping>('edit-openai-compact-model-mapping')
const getTempUnschedRuleKey = createStableObjectKeyResolver<TempUnschedRuleForm>('edit-temp-unsched-rule')

// Quota control state (Anthropic OAuth/SetupToken only)
const windowCostEnabled = ref(false)
const windowCostLimit = ref<number | null>(null)
const windowCostStickyReserve = ref<number | null>(null)
const sessionLimitEnabled = ref(false)
const maxSessions = ref<number | null>(null)
const sessionIdleTimeout = ref<number | null>(null)
const rpmLimitEnabled = ref(false)
const baseRpm = ref<number | null>(null)
const rpmStrategy = ref<'tiered' | 'sticky_exempt'>('tiered')
const rpmStickyBuffer = ref<number | null>(null)
const userMsgQueueMode = ref('')
const umqModeOptions = computed(() => [
  { value: '', label: t('admin.accounts.quotaControl.rpmLimit.umqModeOff') },
  { value: 'throttle', label: t('admin.accounts.quotaControl.rpmLimit.umqModeThrottle') },
  { value: 'serialize', label: t('admin.accounts.quotaControl.rpmLimit.umqModeSerialize') },
])
const tlsFingerprintEnabled = ref(false)
const tlsFingerprintProfileId = ref<string | null>(null)
const tlsFingerprintProfiles = ref<{ id: string; name: string }[]>([])
const sessionIdMaskingEnabled = ref(false)
const cacheTTLOverrideEnabled = ref(false)
const cacheTTLOverrideTarget = ref<string>('5m')
const customBaseUrlEnabled = ref(false)
const customBaseUrl = ref('')

// OpenAI 自动透传开关（OAuth/API Key）
const openaiPassthroughEnabled = ref(false)
// OpenAI Codex namespace 工具摊平兼容开关（仅 OAuth），缺省关闭即原样保留
const openaiFlattenNamespacesEnabled = ref(false)
const openAILongContextBillingEnabled = ref(false)
// OpenAI 订阅档位（Plus/Pro/Free）手动覆盖值,存于 credentials.plan_type;'' 表示清空/自动识别
const editPlanType = ref<string>('')
const openAICompactMode = ref<OpenAICompactMode>('auto')
const openAIResponsesMode = ref<OpenAIResponsesMode>('auto')
const openAIEndpointCapabilities = ref<OpenAIEndpointCapability[]>(['chat_completions', 'embeddings'])
const openaiOAuthResponsesWebSocketV2Mode = ref<OpenAIWSMode>(OPENAI_WS_MODE_OFF)
const openaiAPIKeyResponsesWebSocketV2Mode = ref<OpenAIWSMode>(OPENAI_WS_MODE_OFF)
const codexCLIOnlyEnabled = ref(false)
const codexCLIOnlyAppServerEnabled = ref(false)
type CodexFingerprintMode = 'off' | 'device' | 'session' | 'full'
const codexFingerprintMode = ref<CodexFingerprintMode>('off')
type CodexImageToolMode = 'inherit' | 'enabled' | 'disabled' | 'block'
const codexImageToolMode = ref<CodexImageToolMode>('inherit')
type AnthropicAPIKeyAuthScheme = 'x_api_key' | 'authorization_bearer'
const anthropicPassthroughEnabled = ref(false)
const anthropicAPIKeyAuthScheme = ref<AnthropicAPIKeyAuthScheme>('x_api_key')
const webSearchEmulationMode = ref('default')
const webSearchGlobalEnabled = ref(false)
const {
  globalEnabled: quotaNotifyGlobalEnabled,
  state: quotaNotifyState,
  loadGlobalState: loadQuotaNotifyGlobal,
  loadFromExtra: loadQuotaNotifyFromExtra,
  writeToExtra: writeQuotaNotifyToExtra,
  reset: resetQuotaNotify,
} = useQuotaNotifyState()

// Load global feature states once
adminAPI.settings.getWebSearchEmulationConfig().then(cfg => {
  webSearchGlobalEnabled.value = cfg?.enabled === true && (cfg?.providers?.length ?? 0) > 0
}).catch(() => { webSearchGlobalEnabled.value = false })

loadQuotaNotifyGlobal()
const editQuotaLimit = ref<number | null>(null)
const editQuotaDailyLimit = ref<number | null>(null)
const editQuotaWeeklyLimit = ref<number | null>(null)
const editDailyResetMode = ref<'rolling' | 'fixed' | null>(null)
const editDailyResetHour = ref<number | null>(null)
const editWeeklyResetMode = ref<'rolling' | 'fixed' | null>(null)
const editWeeklyResetDay = ref<number | null>(null)
const editWeeklyResetHour = ref<number | null>(null)
const editResetTimezone = ref<string | null>(null)
const codexFingerprintModeOptions = computed(() => [
  { value: 'off' as CodexFingerprintMode, label: t('admin.accounts.openai.codexFingerprintOff') },
  { value: 'device' as CodexFingerprintMode, label: t('admin.accounts.openai.codexFingerprintDevice') },
  { value: 'session' as CodexFingerprintMode, label: t('admin.accounts.openai.codexFingerprintSession') },
  { value: 'full' as CodexFingerprintMode, label: t('admin.accounts.openai.codexFingerprintFull') },
])

const openAIWSModeOptions = computed(() => [
  { value: OPENAI_WS_MODE_OFF, label: t('admin.accounts.openai.wsModeOff') },
  { value: OPENAI_WS_MODE_CTX_POOL, label: t('admin.accounts.openai.wsModeCtxPool') },
  { value: OPENAI_WS_MODE_PASSTHROUGH, label: t('admin.accounts.openai.wsModePassthrough') },
  { value: OPENAI_WS_MODE_HTTP_BRIDGE, label: t('admin.accounts.openai.wsModeHttpBridge') }
])
const openaiResponsesWebSocketV2Mode = computed({
  get: () => {
    if (props.account?.type === 'apikey') {
      return openaiAPIKeyResponsesWebSocketV2Mode.value
    }
    return openaiOAuthResponsesWebSocketV2Mode.value
  },
  set: (mode: OpenAIWSMode) => {
    if (props.account?.type === 'apikey') {
      openaiAPIKeyResponsesWebSocketV2Mode.value = mode
      return
    }
    openaiOAuthResponsesWebSocketV2Mode.value = mode
  }
})
const openAIWSModeConcurrencyHintKey = computed(() =>
  resolveOpenAIWSModeConcurrencyHintKey(openaiResponsesWebSocketV2Mode.value)
)
const codexImageToolOptions = computed<Array<{
  value: CodexImageToolMode
  label: string
  description: string
  selectedCardClass: string
  selectedDotClass: string
}>>(() => [
  {
    value: 'inherit',
    label: t('admin.accounts.openai.codexImageToolInherit'),
    description: t('admin.accounts.openai.codexImageToolInheritDesc'),
    selectedCardClass: 'components-account-edit-account-modal__state',
    selectedDotClass: 'components-account-edit-account-modal__state-2'
  },
  {
    value: 'enabled',
    label: t('admin.accounts.openai.codexImageToolEnabled'),
    description: t('admin.accounts.openai.codexImageToolEnabledDesc'),
    selectedCardClass: 'components-account-edit-account-modal__state-3',
    selectedDotClass: 'components-account-edit-account-modal__state-4'
  },
  {
    value: 'disabled',
    label: t('admin.accounts.openai.codexImageToolDisabled'),
    description: t('admin.accounts.openai.codexImageToolDisabledDesc'),
    selectedCardClass: 'components-account-edit-account-modal__state-5',
    selectedDotClass: 'components-account-edit-account-modal__state-6'
  },
  {
    value: 'block',
    label: t('admin.accounts.openai.codexImageToolBlock'),
    description: t('admin.accounts.openai.codexImageToolBlockDesc'),
    selectedCardClass: 'components-account-edit-account-modal__state-7',
    selectedDotClass: 'components-account-edit-account-modal__state-8'
  }
])
const codexImageToolBadgeLabel = computed(() => {
  switch (codexImageToolMode.value) {
    case 'enabled':
      return t('admin.accounts.openai.codexImageToolBadgeEnabled')
    case 'disabled':
      return t('admin.accounts.openai.codexImageToolBadgeDisabled')
    case 'block':
      return t('admin.accounts.openai.codexImageToolBadgeBlock')
    default:
      return t('admin.accounts.openai.codexImageToolBadgeInherit')
  }
})
const codexImageToolBadgeClass = computed(() => {
  switch (codexImageToolMode.value) {
    case 'enabled':
      return 'components-account-edit-account-modal__state-9'
    case 'disabled':
      return 'components-account-edit-account-modal__state-10'
    case 'block':
      return 'components-account-edit-account-modal__state-11'
    default:
      return 'components-account-edit-account-modal__state-12'
  }
})
const openAICompactModeOptions = computed(() => [
  { value: 'auto', label: t('admin.accounts.openai.compactModeAuto') },
  { value: 'force_on', label: t('admin.accounts.openai.compactModeForceOn') },
  { value: 'force_off', label: t('admin.accounts.openai.compactModeForceOff') }
])
// OpenAI 订阅档位手动覆盖选项(清空 + Plus/Pro/Free;别名/自定义值友好显示且保留 canonical)
const planTypeOptions = computed(() =>
  buildPlanTypeOptions(editPlanType.value, t('admin.accounts.openai.planTypeClear'))
)
const openAIResponsesModeOptions = computed(() => [
  { value: 'auto', label: t('admin.accounts.openai.responsesModeAuto') },
  { value: 'force_responses', label: t('admin.accounts.openai.responsesModeForceResponses') },
  { value: 'force_chat_completions', label: t('admin.accounts.openai.responsesModeForceChatCompletions') }
])
const openAITextEndpointCapabilityLabel = computed(() => {
  if (openAIResponsesMode.value === 'force_responses') {
    return t('admin.accounts.openai.capabilityResponses')
  }
  if (openAIResponsesMode.value === 'force_chat_completions') {
    return t('admin.accounts.openai.capabilityChatCompletions')
  }
  const extra = props.account?.extra as Record<string, unknown> | undefined
  if (extra?.openai_responses_supported === true) {
    return t('admin.accounts.openai.capabilityResponsesAuto')
  }
  if (extra?.openai_responses_supported === false) {
    return t('admin.accounts.openai.capabilityChatCompletionsAuto')
  }
  return t('admin.accounts.openai.capabilityTextAuto')
})
const openAIEndpointCapabilityOptions = computed<{ value: OpenAIEndpointCapability; label: string }[]>(() => [
  { value: 'chat_completions', label: openAITextEndpointCapabilityLabel.value },
  { value: 'embeddings', label: t('admin.accounts.openai.capabilityEmbeddings') }
])
const openAITextGenerationCapabilityEnabled = computed(() =>
  openAIEndpointCapabilities.value.includes('chat_completions')
)

const normalizeOpenAIEndpointCapabilities = (values: OpenAIEndpointCapability[]) => {
  const allowed: OpenAIEndpointCapability[] = ['chat_completions', 'embeddings']
  const selected = allowed.filter((value) => values.includes(value))
  return selected.length > 0 ? selected : allowed
}

const readOpenAIEndpointCapabilities = (credentials?: Record<string, unknown>): OpenAIEndpointCapability[] => {
  const raw = credentials?.openai_capabilities
  if (Array.isArray(raw)) {
    return normalizeOpenAIEndpointCapabilities(
      raw.filter((value): value is OpenAIEndpointCapability =>
        value === 'chat_completions' || value === 'embeddings'
      )
    )
  }
  if (raw !== null && typeof raw === 'object') {
    const capabilityMap = raw as Record<string, unknown>
    return normalizeOpenAIEndpointCapabilities(
      openAIEndpointCapabilityOptions.value
        .map((option) => option.value)
        .filter((value) => capabilityMap[value] === true)
    )
  }
  return ['chat_completions', 'embeddings']
}

const toggleOpenAIEndpointCapability = (capability: OpenAIEndpointCapability, event?: Event) => {
  if (openAIEndpointCapabilities.value.includes(capability)) {
    if (openAIEndpointCapabilities.value.length <= 1) {
      const input = event?.target as HTMLInputElement | null
      if (input) input.checked = true
      return
    }
    openAIEndpointCapabilities.value = openAIEndpointCapabilities.value.filter(
      (value) => value !== capability
    )
    if (!openAITextGenerationCapabilityEnabled.value) {
      openAIResponsesMode.value = 'auto'
    }
    return
  }
  openAIEndpointCapabilities.value = normalizeOpenAIEndpointCapabilities([
    ...openAIEndpointCapabilities.value,
    capability
  ])
}

const applyOpenAIEndpointCapabilities = (credentials: Record<string, unknown>) => {
  const capabilities = normalizeOpenAIEndpointCapabilities(openAIEndpointCapabilities.value)
  if (capabilities.length === 2) {
    delete credentials.openai_capabilities
    return
  }
  credentials.openai_capabilities = capabilities
}
const normalizeOpenAIResponsesMode = (mode: unknown): OpenAIResponsesMode => {
  if (mode === 'force_responses' || mode === 'force_chat_completions') {
    return mode
  }
  return 'auto'
}
const isOpenAIModelRestrictionDisabled = computed(() =>
  props.account?.platform === 'openai' && openaiPassthroughEnabled.value
)
const openAIResponsesStatusKey = computed(() => {
  if (openAIResponsesMode.value === 'force_responses') {
    return 'admin.accounts.openai.responsesStatusForcedResponses'
  }
  if (openAIResponsesMode.value === 'force_chat_completions') {
    return 'admin.accounts.openai.responsesStatusForcedChatCompletions'
  }
  const extra = props.account?.extra as Record<string, unknown> | undefined
  if (extra?.openai_responses_supported === true) {
    return 'admin.accounts.openai.responsesStatusAutoSupported'
  }
  if (extra?.openai_responses_supported === false) {
    return 'admin.accounts.openai.responsesStatusAutoUnsupported'
  }
  return 'admin.accounts.openai.responsesStatusAutoUnknown'
})
const openAICompactStatusKey = computed(() => {
  const extra = props.account?.extra as Record<string, unknown> | undefined
  if (!props.account || props.account.platform !== 'openai') return ''
  const mode = typeof extra?.openai_compact_mode === 'string' ? extra.openai_compact_mode : 'auto'
  if (mode === 'force_on') return 'admin.accounts.openai.compactSupported'
  if (mode === 'force_off') return 'admin.accounts.openai.compactUnsupported'
  if (typeof extra?.openai_compact_supported === 'boolean') {
    return extra.openai_compact_supported
      ? 'admin.accounts.openai.compactSupported'
      : 'admin.accounts.openai.compactUnsupported'
  }
  return 'admin.accounts.openai.compactAuto'
})

// Computed: current preset mappings based on platform
const presetMappings = computed(() => getPresetMappingsByPlatform(props.account?.platform || 'anthropic'))
const tempUnschedPresets = computed(() => [
  {
    label: t('admin.accounts.tempUnschedulable.presets.overloadLabel'),
    rule: {
      error_code: 529,
      keywords: 'overloaded, too many',
      duration_minutes: 60,
      description: t('admin.accounts.tempUnschedulable.presets.overloadDesc')
    }
  },
  {
    label: t('admin.accounts.tempUnschedulable.presets.rateLimitLabel'),
    rule: {
      error_code: 429,
      keywords: 'rate limit, too many requests',
      duration_minutes: 10,
      description: t('admin.accounts.tempUnschedulable.presets.rateLimitDesc')
    }
  },
  {
    label: t('admin.accounts.tempUnschedulable.presets.unavailableLabel'),
    rule: {
      error_code: 503,
      keywords: 'unavailable, maintenance',
      duration_minutes: 30,
      description: t('admin.accounts.tempUnschedulable.presets.unavailableDesc')
    }
  }
])

// Computed: default base URL based on platform
const defaultBaseUrl = computed(() => {
  if (props.account?.platform === 'openai') return 'https://api.openai.com'
  if (props.account?.platform === 'grok') return 'https://api.x.ai/v1'
  // CN 供应商：按当前模式/协议回落到官方预设（清空输入框提交时使用），
  // 不能落到 anthropic 默认值（会被当 CC base 拼出错误端点）。
  if (
    props.account?.platform === 'kimi' ||
    props.account?.platform === 'zhipu' ||
    props.account?.platform === 'deepseek'
  ) {
    return defaultCNBaseUrl(props.account.platform, editAccountMode.value, editApiProtocol.value)
  }
  return 'https://api.anthropic.com'
})

const form = reactive({
  name: '',
  notes: '',
  proxy_id: null as string | null,
  concurrency: 1,
  load_factor: null as number | null,
  priority: 1,
  rate_multiplier: 1,
  status: 'active' as 'active' | 'inactive' | 'error',
  group_ids: [] as string[],
  expires_at: null as number | null
})

const handleUpstreamBillingRateSyncChange = (enabled: boolean) => {
  upstreamBillingRateSyncEnabled.value = enabled
  if (enabled) {
    upstreamBillingAutoProbeEnabled.value = true
  }
}

const handleUpstreamBillingAutoProbeChange = (enabled: boolean) => {
  upstreamBillingAutoProbeEnabled.value = enabled
  if (!enabled) {
    upstreamBillingRateSyncEnabled.value = false
  }
}

const statusOptions = computed(() => {
  const options = [
    { value: 'active', label: t('common.active') },
    { value: 'inactive', label: t('common.inactive') }
  ]
  if (form.status === 'error') {
    options.push({ value: 'error', label: t('admin.accounts.status.error') })
  }
  return options
})

const expiresAtInput = computed({
  get: () => formatDateTimeLocal(form.expires_at),
  set: (value: string) => {
    form.expires_at = parseDateTimeLocal(value)
  }
})

// Watchers
const normalizePoolModeRetryCount = (value: number) => {
  if (!Number.isFinite(value)) {
    return DEFAULT_POOL_MODE_RETRY_COUNT
  }
  const normalized = Math.trunc(value)
  if (normalized < 0) {
    return 0
  }
  if (normalized > MAX_POOL_MODE_RETRY_COUNT) {
    return MAX_POOL_MODE_RETRY_COUNT
  }
  return normalized
}

const loadModelRestrictionFromMapping = (rawMapping?: Record<string, unknown>) => {
  const parsed = splitModelMappingObject(rawMapping)
  allowedModels.value = parsed.allowedModels
  modelMappings.value = parsed.modelMappings
  modelRestrictionMode.value =
    parsed.modelMappings.length > 0 && parsed.allowedModels.length === 0
      ? 'mapping'
      : 'whitelist'
}

const buildModelRestrictionMapping = () =>
  buildModelMappingObject('combined', allowedModels.value, modelMappings.value)

const applyOpenAIModelMappingCredentials = (credentials: Record<string, unknown>) => {
  const shouldApplyModelMapping = !openaiPassthroughEnabled.value

  if (shouldApplyModelMapping) {
    const modelMapping = buildModelRestrictionMapping()
    if (modelMapping) {
      credentials.model_mapping = modelMapping
    } else {
      delete credentials.model_mapping
    }
  } else if (!credentials.model_mapping) {
    delete credentials.model_mapping
  }

  const compactModelMapping = buildModelMappingObject('mapping', [], openAICompactModelMappings.value)
  if (compactModelMapping) {
    credentials.compact_model_mapping = compactModelMapping
  } else {
    delete credentials.compact_model_mapping
  }
}

const syncFormFromAccount = (newAccount: Account | null) => {
  if (!newAccount) {
    return
  }
  // 进入回填窗口：抑制 CN 模式/协议 watcher 联动重置 base_url（见 syncingForm 注释）。
  syncingForm.value = true
  void nextTick(() => {
    syncingForm.value = false
  })
  form.name = newAccount.name
  form.notes = newAccount.notes || ''
  form.proxy_id = newAccount.proxy_id
  form.concurrency = newAccount.concurrency
  form.load_factor = newAccount.load_factor ?? null
  form.priority = newAccount.priority
  form.rate_multiplier = newAccount.rate_multiplier ?? 1
  form.status = (newAccount.status === 'active' || newAccount.status === 'inactive' || newAccount.status === 'error')
    ? newAccount.status
    : 'active'
  form.group_ids = newAccount.group_ids || []
  form.expires_at = newAccount.expires_at ?? null

  // Load intercept warmup requests setting (applies to all account types)
  const credentials = newAccount.credentials as Record<string, unknown> | undefined
  interceptWarmupRequests.value = credentials?.intercept_warmup_requests === true
  autoPauseOnExpired.value = newAccount.auto_pause_on_expired === true
  editVertexProjectId.value = ''
  editVertexClientEmail.value = ''
  editVertexLocation.value = 'us-central1'
	const extra = newAccount.extra as Record<string, unknown> | undefined
	autoPause5hThreshold.value = typeof extra?.auto_pause_5h_threshold === 'number' ? extra.auto_pause_5h_threshold * 100 : null
	autoPause7dThreshold.value = typeof extra?.auto_pause_7d_threshold === 'number' ? extra.auto_pause_7d_threshold * 100 : null
	autoPause5hDisabled.value = extra?.auto_pause_5h_disabled === true
	autoPause7dDisabled.value = extra?.auto_pause_7d_disabled === true
	upstreamBillingAutoProbeEnabled.value = extra?.upstream_billing_probe_enabled === true
  upstreamBillingRateSyncEnabled.value =
    upstreamBillingAutoProbeEnabled.value && extra?.upstream_billing_rate_sync_enabled === true

  // Load OpenAI passthrough toggle (OpenAI OAuth/SetupToken/API Key)
  openaiPassthroughEnabled.value = false
  openaiFlattenNamespacesEnabled.value = false
  openAILongContextBillingEnabled.value = false
  editPlanType.value = ''
  openAICompactMode.value = 'auto'
  openAIResponsesMode.value = 'auto'
  openAIEndpointCapabilities.value = ['chat_completions', 'embeddings']
  openAICompactModelMappings.value = []
  openaiOAuthResponsesWebSocketV2Mode.value = OPENAI_WS_MODE_OFF
  openaiAPIKeyResponsesWebSocketV2Mode.value = OPENAI_WS_MODE_OFF
  codexCLIOnlyEnabled.value = false
  codexCLIOnlyAppServerEnabled.value = false
  codexFingerprintMode.value = 'off'
  codexImageToolMode.value = 'inherit'
  anthropicPassthroughEnabled.value = false
  anthropicAPIKeyAuthScheme.value = 'x_api_key'
  webSearchEmulationMode.value = 'default'
  if (newAccount.platform === 'openai' && (newAccount.type === 'oauth' || newAccount.type === 'setup-token' || newAccount.type === 'apikey')) {
    openaiPassthroughEnabled.value = extra?.openai_passthrough === true || extra?.openai_oauth_passthrough === true
    openaiFlattenNamespacesEnabled.value =
      newAccount.type === 'oauth' && extra?.openai_responses_flatten_namespaces === true
    const longContextBillingValue = extra?.openai_long_context_billing_enabled
    openAILongContextBillingEnabled.value = longContextBillingValue === true
    // plan_type 手动覆盖仅 OAuth 有实际调度语义(IsOpenAIChatGPTSubscription 要求 oauth),故只对 oauth 回填
    editPlanType.value = newAccount.type === 'oauth'
      ? readPlanType(newAccount.credentials as Record<string, unknown> | undefined)
      : ''
    openAICompactMode.value = (extra?.openai_compact_mode as OpenAICompactMode) || 'auto'
    if (newAccount.type === 'apikey') {
      openAIResponsesMode.value = normalizeOpenAIResponsesMode(extra?.openai_responses_mode)
      openAIEndpointCapabilities.value = readOpenAIEndpointCapabilities(
        newAccount.credentials as Record<string, unknown> | undefined
      )
      if (!openAITextGenerationCapabilityEnabled.value) {
        openAIResponsesMode.value = 'auto'
      }
    }
    const codexImageGenerationBridgeValue = typeof extra?.codex_image_generation_bridge === 'boolean'
      ? extra.codex_image_generation_bridge
      : extra?.codex_image_generation_bridge_enabled
    if (extra?.codex_image_generation_explicit_tool_policy === 'strip') {
      codexImageToolMode.value = 'block'
    } else if (codexImageGenerationBridgeValue === true) {
      codexImageToolMode.value = 'enabled'
    } else if (codexImageGenerationBridgeValue === false) {
      codexImageToolMode.value = 'disabled'
    }
    openaiOAuthResponsesWebSocketV2Mode.value = resolveOpenAIWSModeFromExtra(extra, {
      modeKey: 'openai_oauth_responses_websockets_v2_mode',
      enabledKey: 'openai_oauth_responses_websockets_v2_enabled',
      fallbackEnabledKeys: ['responses_websockets_v2_enabled', 'openai_ws_enabled'],
      defaultMode: OPENAI_WS_MODE_OFF
    })
    openaiAPIKeyResponsesWebSocketV2Mode.value = resolveOpenAIWSModeFromExtra(extra, {
      modeKey: 'openai_apikey_responses_websockets_v2_mode',
      enabledKey: 'openai_apikey_responses_websockets_v2_enabled',
      fallbackEnabledKeys: ['responses_websockets_v2_enabled', 'openai_ws_enabled'],
      defaultMode: OPENAI_WS_MODE_OFF
    })
    if (newAccount.type === 'oauth' || newAccount.type === 'setup-token') {
      codexCLIOnlyEnabled.value = extra?.codex_cli_only === true
      codexCLIOnlyAppServerEnabled.value =
        extra?.codex_cli_only_allow_app_server === true
    }
    if (newAccount.type === 'oauth') {
      const fpMode = extra?.codex_fingerprint_mode as string | undefined
      // 缺省/非法值按 off 呈现，与后端 GetCodexFingerprintMode 的 opt-in 语义一致（#5610）
      codexFingerprintMode.value = (['off', 'device', 'session', 'full'].includes(fpMode || '')
        ? fpMode as CodexFingerprintMode
        : 'off')
    }
    const credentials = newAccount.credentials as Record<string, unknown> | undefined
    const compactMappings = credentials?.compact_model_mapping as Record<string, string> | undefined
    if (compactMappings && typeof compactMappings === 'object') {
      openAICompactModelMappings.value = Object.entries(compactMappings).map(([from, to]) => ({ from, to }))
    }
  }
  if (newAccount.platform === 'anthropic' && newAccount.type === 'apikey') {
    anthropicPassthroughEnabled.value = extra?.anthropic_passthrough === true
    anthropicAPIKeyAuthScheme.value = extra?.anthropic_apikey_auth_scheme === 'authorization_bearer'
      ? 'authorization_bearer'
      : 'x_api_key'
    // 三态：string "default"/"enabled"/"disabled"，向后兼容旧 bool
    const wsVal = extra?.web_search_emulation
    if (wsVal === 'enabled' || wsVal === 'disabled') {
      webSearchEmulationMode.value = wsVal
    } else if (wsVal === true) {
      webSearchEmulationMode.value = 'enabled'
    } else {
      webSearchEmulationMode.value = 'default'
    }
  }

  // Load quota limit for apikey/bedrock accounts (bedrock quota is also loaded in its own branch above)
  if (newAccount.type === 'apikey' || newAccount.type === 'bedrock') {
    const quotaVal = extra?.quota_limit as number | undefined
    editQuotaLimit.value = (quotaVal && quotaVal > 0) ? quotaVal : null
    const dailyVal = extra?.quota_daily_limit as number | undefined
    editQuotaDailyLimit.value = (dailyVal && dailyVal > 0) ? dailyVal : null
    const weeklyVal = extra?.quota_weekly_limit as number | undefined
    editQuotaWeeklyLimit.value = (weeklyVal && weeklyVal > 0) ? weeklyVal : null
    // Load quota reset mode config
    editDailyResetMode.value = (extra?.quota_daily_reset_mode as 'rolling' | 'fixed') || null
    editDailyResetHour.value = (extra?.quota_daily_reset_hour as number) ?? null
    editWeeklyResetMode.value = (extra?.quota_weekly_reset_mode as 'rolling' | 'fixed') || null
    editWeeklyResetDay.value = (extra?.quota_weekly_reset_day as number) ?? null
    editWeeklyResetHour.value = (extra?.quota_weekly_reset_hour as number) ?? null
    editResetTimezone.value = (extra?.quota_reset_timezone as string) || null
    // Load quota notify config
    loadQuotaNotifyFromExtra(extra)
  } else {
    editQuotaLimit.value = null
    editQuotaDailyLimit.value = null
    editQuotaWeeklyLimit.value = null
    editDailyResetMode.value = null
    editDailyResetHour.value = null
    editWeeklyResetMode.value = null
    editWeeklyResetDay.value = null
    editWeeklyResetHour.value = null
    editResetTimezone.value = null
    resetQuotaNotify()
  }

  // Load quota control settings (Anthropic OAuth/SetupToken only)
  loadQuotaControlSettings(newAccount)

  loadTempUnschedRules(credentials)
  loadAccountSchedulingThresholdOverride(newAccount.platform, credentials)

  // Load header override state (anthropic/openai apikey + grok apikey/oauth)
  headerOverrideEnabled.value = false
  headerOverrideRows.value = []
  if (newAccount.credentials && isHeaderOverrideCapable(newAccount.platform, newAccount.type)) {
    const overrideCreds = newAccount.credentials as Record<string, unknown>
    headerOverrideEnabled.value = overrideCreds[HEADER_OVERRIDE_ENABLED_CREDENTIAL_KEY] === true
    headerOverrideRows.value = splitHeaderOverridesObject(
      overrideCreds[HEADER_OVERRIDES_CREDENTIAL_KEY]
    )
  }

  // Load Grok OAuth custom upstream URL state（存储的官方地址视同未定制）
  grokOAuthCustomBaseUrlEnabled.value = false
  grokOAuthBaseUrl.value = ''
  const grokClientToolCacheSetting =
    newAccount.platform === 'grok' && newAccount.type === 'oauth'
      ? newAccount.extra?.[GROK_CLIENT_TOOL_CACHE_EXTRA_KEY]
      : undefined
  grokClientToolCacheEnabled.value =
    newAccount.platform === 'grok' &&
    newAccount.type === 'oauth' &&
    (grokClientToolCacheSetting === undefined || grokClientToolCacheSetting === true)
  if (newAccount.platform === 'grok' && newAccount.type === 'oauth' && newAccount.credentials) {
    const grokCreds = newAccount.credentials as Record<string, unknown>
    if (isCustomGrokBaseUrl(grokCreds.base_url)) {
      grokOAuthCustomBaseUrlEnabled.value = true
      grokOAuthBaseUrl.value = (grokCreds.base_url as string).trim()
    }
  }

  // Initialize API Key fields for apikey type
  if (newAccount.type === 'apikey' && newAccount.credentials) {
    const credentials = newAccount.credentials as Record<string, unknown>
    // 国产供应商：读取 account_mode 与 api_protocol 作为可编辑初始值
    // （编辑弹窗允许修正两者，用于修复早期存错默认值的账号）。
    if (newAccount.platform === 'kimi' || newAccount.platform === 'zhipu' || newAccount.platform === 'deepseek') {
      editAccountMode.value = credentials.account_mode === 'coding' ? 'coding' : 'payg'
      const storedProtocol = credentials.api_protocol
      editApiProtocol.value =
        storedProtocol === 'anthropic' || storedProtocol === 'responses' ? storedProtocol : 'chat_completions'
      if (newAccount.platform !== 'deepseek' && editApiProtocol.value === 'responses') {
        editApiProtocol.value = 'chat_completions'
      }
    }
    const platformDefaultUrl =
      newAccount.platform === 'openai'
        ? 'https://api.openai.com'
        : newAccount.platform === 'grok'
          ? 'https://api.x.ai/v1'
          : newAccount.platform === 'kimi' ||
              newAccount.platform === 'zhipu' ||
              newAccount.platform === 'deepseek'
            ? defaultCNBaseUrl(newAccount.platform, editAccountMode.value, editApiProtocol.value)
            : 'https://api.anthropic.com'
    editBaseUrl.value = (credentials.base_url as string) || platformDefaultUrl

    // Load model mappings and detect mode
    loadModelRestrictionFromMapping(credentials.model_mapping as Record<string, unknown> | undefined)

    // Load pool mode
    poolModeEnabled.value = credentials.pool_mode === true
    poolModeRetryCount.value = normalizePoolModeRetryCount(
      Number(credentials.pool_mode_retry_count ?? DEFAULT_POOL_MODE_RETRY_COUNT)
    )
    poolModeRetryStatusCodesInput.value = formatPoolModeRetryStatusCodes(credentials.pool_mode_retry_status_codes)

    // Load custom error codes
    customErrorCodesEnabled.value = credentials.custom_error_codes_enabled === true
    const existingErrorCodes = credentials.custom_error_codes as number[] | undefined
    if (existingErrorCodes && Array.isArray(existingErrorCodes)) {
      selectedErrorCodes.value = [...existingErrorCodes]
    } else {
      selectedErrorCodes.value = []
    }

  } else if (newAccount.type === 'bedrock' && newAccount.credentials) {
    const bedrockCreds = newAccount.credentials as Record<string, unknown>
    const authMode = (bedrockCreds.auth_mode as string) || 'sigv4'
    editBedrockRegion.value = (bedrockCreds.aws_region as string) || ''
    editBedrockForceGlobal.value = (bedrockCreds.aws_force_global as string) === 'true'

    if (authMode === 'apikey') {
      editBedrockApiKeyValue.value = ''
    } else {
      editBedrockAccessKeyId.value = (bedrockCreds.aws_access_key_id as string) || ''
      editBedrockSecretAccessKey.value = ''
      editBedrockSessionToken.value = ''
    }

    // Load pool mode for bedrock
    poolModeEnabled.value = bedrockCreds.pool_mode === true
    const retryCount = bedrockCreds.pool_mode_retry_count
    poolModeRetryCount.value = (typeof retryCount === 'number' && retryCount >= 0) ? retryCount : DEFAULT_POOL_MODE_RETRY_COUNT
    poolModeRetryStatusCodesInput.value = formatPoolModeRetryStatusCodes(bedrockCreds.pool_mode_retry_status_codes)

    // Load quota limits for bedrock
    const bedrockExtra = (newAccount.extra as Record<string, unknown>) || {}
    editQuotaLimit.value = typeof bedrockExtra.quota_limit === 'number' ? bedrockExtra.quota_limit : null
    editQuotaDailyLimit.value = typeof bedrockExtra.quota_daily_limit === 'number' ? bedrockExtra.quota_daily_limit : null
    editQuotaWeeklyLimit.value = typeof bedrockExtra.quota_weekly_limit === 'number' ? bedrockExtra.quota_weekly_limit : null
    // Load quota notify for bedrock
    loadQuotaNotifyFromExtra(bedrockExtra)

    // Load model mappings for bedrock
    loadModelRestrictionFromMapping(bedrockCreds.model_mapping as Record<string, unknown> | undefined)
  } else if (newAccount.type === 'upstream' && newAccount.credentials) {
    const credentials = newAccount.credentials as Record<string, unknown>
    editBaseUrl.value = (credentials.base_url as string) || ''
  } else if ((newAccount.platform === 'anthropic' || newAccount.platform === 'gemini') && newAccount.type === 'service_account' && newAccount.credentials) {
    const credentials = newAccount.credentials as Record<string, unknown>
    editVertexProjectId.value = (credentials.project_id as string) || ''
    editVertexClientEmail.value = (credentials.client_email as string) || ''
    editVertexLocation.value = (credentials.location as string) || (credentials.vertex_location as string) || 'us-central1'

    // Load model mappings for service_account
    loadModelRestrictionFromMapping(credentials.model_mapping as Record<string, unknown> | undefined)
  } else {
    const platformDefaultUrl =
      newAccount.platform === 'openai'
        ? 'https://api.openai.com'
        : newAccount.platform === 'grok'
          ? 'https://api.x.ai/v1'
          : 'https://api.anthropic.com'
    editBaseUrl.value = platformDefaultUrl

    // Load model mappings for OpenAI/Grok OAuth accounts
    if ((newAccount.platform === 'openai' || newAccount.platform === 'grok') && newAccount.credentials) {
      const oauthCredentials = newAccount.credentials as Record<string, unknown>
      loadModelRestrictionFromMapping(oauthCredentials.model_mapping as Record<string, unknown> | undefined)
    } else {
      modelRestrictionMode.value = 'whitelist'
      modelMappings.value = []
      allowedModels.value = []
    }
    poolModeEnabled.value = false
    poolModeRetryCount.value = DEFAULT_POOL_MODE_RETRY_COUNT
    poolModeRetryStatusCodesInput.value = ''
    customErrorCodesEnabled.value = false
    selectedErrorCodes.value = []
  }
  editApiKey.value = ''
}

async function loadTLSProfiles() {
  try {
    const profiles = await adminAPI.tlsFingerprintProfiles.list()
    tlsFingerprintProfiles.value = profiles.map(p => ({ id: p.id, name: p.name }))
  } catch {
    tlsFingerprintProfiles.value = []
  }
}

watch(
  [() => props.show, () => props.account],
  ([show, newAccount], [wasShow, previousAccount]) => {
    if (!show || !newAccount) {
      return
    }
    if (!wasShow || newAccount !== previousAccount) {
      syncFormFromAccount(newAccount)
      loadTLSProfiles()
    }
  },
  { immediate: true }
)

// Model mapping helpers
const addModelMapping = () => {
  modelMappings.value.push({ from: '', to: '' })
}

const removeModelMapping = (index: number) => {
  modelMappings.value.splice(index, 1)
}

const applyAccountModelRule = (payload: { name: string; allowedModels: string[]; mappings: ModelMapping[] }) => {
  modelRestrictionMode.value = payload.mappings.length > 0 ? 'mapping' : 'whitelist'
  allowedModels.value = payload.allowedModels.map(model => model.trim()).filter(Boolean)
  modelMappings.value = payload.mappings.map(mapping => ({ ...mapping }))
  appStore.showSuccess(t('admin.accounts.modelRules.importSuccess', { name: payload.name }))
}

const addPresetMapping = (from: string, to: string) => {
  const exists = modelMappings.value.some((m) => m.from === from)
  if (exists) {
    appStore.showInfo(t('admin.accounts.mappingExists', { model: from }))
    return
  }
  modelMappings.value.push({ from, to })
}

const addOpenAICompactModelMapping = () => {
  openAICompactModelMappings.value.push({ from: '', to: '' })
}

const removeOpenAICompactModelMapping = (index: number) => {
  openAICompactModelMappings.value.splice(index, 1)
}

// Error code toggle helper
const toggleErrorCode = (code: number) => {
  const index = selectedErrorCodes.value.indexOf(code)
  if (index === -1) {
    // Adding code - check for 429/529 warning
    if (code === 429) {
      if (!confirm(t('admin.accounts.customErrorCodes429Warning'))) {
        return
      }
    } else if (code === 529) {
      if (!confirm(t('admin.accounts.customErrorCodes529Warning'))) {
        return
      }
    }
    selectedErrorCodes.value.push(code)
  } else {
    selectedErrorCodes.value.splice(index, 1)
  }
}

// Add custom error code from input
const addCustomErrorCode = () => {
  const code = customErrorCodeInput.value
  if (code === null || code < 100 || code > 599) {
    appStore.showError(t('admin.accounts.invalidErrorCode'))
    return
  }
  if (selectedErrorCodes.value.includes(code)) {
    appStore.showInfo(t('admin.accounts.errorCodeExists'))
    return
  }
  // Check for 429/529 warning
  if (code === 429) {
    if (!confirm(t('admin.accounts.customErrorCodes429Warning'))) {
      return
    }
  } else if (code === 529) {
    if (!confirm(t('admin.accounts.customErrorCodes529Warning'))) {
      return
    }
  }
  selectedErrorCodes.value.push(code)
  customErrorCodeInput.value = null
}

// Remove error code
const removeErrorCode = (code: number) => {
  const index = selectedErrorCodes.value.indexOf(code)
  if (index !== -1) {
    selectedErrorCodes.value.splice(index, 1)
  }
}

const addTempUnschedRule = (preset?: TempUnschedRuleForm) => {
  if (preset) {
    tempUnschedRules.value.push({ ...preset })
    return
  }
  tempUnschedRules.value.push({
    error_code: null,
    keywords: '',
    duration_minutes: 30,
    description: ''
  })
}

const removeTempUnschedRule = (index: number) => {
  tempUnschedRules.value.splice(index, 1)
}

const moveTempUnschedRule = (index: number, direction: number) => {
  const target = index + direction
  if (target < 0 || target >= tempUnschedRules.value.length) return
  const rules = tempUnschedRules.value
  const current = rules[index]
  rules[index] = rules[target]
  rules[target] = current
}

const buildTempUnschedRules = (rules: TempUnschedRuleForm[]) => {
  const out: Array<{
    error_code: number
    keywords: string[]
    duration_minutes: number
    description: string
  }> = []

  for (const rule of rules) {
    const errorCode = Number(rule.error_code)
    const duration = Number(rule.duration_minutes)
    const keywords = splitTempUnschedKeywords(rule.keywords)
    if (!Number.isFinite(errorCode) || errorCode < 100 || errorCode > 599) {
      continue
    }
    if (!Number.isFinite(duration) || duration <= 0) {
      continue
    }
    if (keywords.length === 0) {
      continue
    }
    out.push({
      error_code: Math.trunc(errorCode),
      keywords,
      duration_minutes: Math.trunc(duration),
      description: rule.description.trim()
    })
  }

  return out
}

const applyTempUnschedConfig = (credentials: Record<string, unknown>) => {
  if (!tempUnschedEnabled.value) {
    delete credentials.temp_unschedulable_enabled
    delete credentials.temp_unschedulable_rules
    return true
  }

  const rules = buildTempUnschedRules(tempUnschedRules.value)
  if (rules.length === 0) {
    appStore.showError(t('admin.accounts.tempUnschedulable.rulesInvalid'))
    return false
  }

  credentials.temp_unschedulable_enabled = true
  credentials.temp_unschedulable_rules = rules
  return true
}


function supportsAccountSchedulingThresholdOverridePlatform(platform: Account['platform'] | undefined) {
  return platform === 'openai' || platform === 'anthropic' || platform === 'grok'
}

function normalizeAccountSchedulingThresholdOverride(value: unknown): number | null {
  if (value === null || value === undefined || value === '') {
    return null
  }
  const numeric = Number(value)
  if (!Number.isFinite(numeric)) {
    return null
  }
  const integer = Math.trunc(numeric)
  if (integer < 1 || integer > 100) {
    return null
  }
  return integer
}

function clampAccountSchedulingThresholdOverride(value: unknown): number {
  return Math.min(100, Math.max(1, Math.trunc(Number(value) || 100)))
}

function loadAccountSchedulingThresholdOverride(
  platform: Account['platform'] | undefined,
  credentials: Record<string, unknown> | undefined
) {
  if (!supportsAccountSchedulingThresholdOverridePlatform(platform)) {
    accountSchedulingThresholdOverrideEnabled.value = false
    accountSchedulingThresholdOverrideValue.value = 100
    return
  }
  const value = normalizeAccountSchedulingThresholdOverride(
    credentials?.[ACCOUNT_SCHEDULING_THRESHOLD_CREDENTIAL_KEY]
  )
  accountSchedulingThresholdOverrideEnabled.value = value !== null
  accountSchedulingThresholdOverrideValue.value = value ?? 100
}

const applyAccountSchedulingThresholdOverridePatch = (
  credentials: Record<string, unknown>,
  currentCredentials: Record<string, unknown>,
  platform: Account['platform'] | undefined = props.account?.platform
) => {
  if (!supportsAccountSchedulingThresholdOverridePlatform(platform)) {
    return
  }
  const current = normalizeAccountSchedulingThresholdOverride(
    currentCredentials[ACCOUNT_SCHEDULING_THRESHOLD_CREDENTIAL_KEY]
  )
  if (!accountSchedulingThresholdOverrideEnabled.value) {
    if (current !== null) {
      credentials[ACCOUNT_SCHEDULING_THRESHOLD_CREDENTIAL_KEY] = null
    }
    return
  }
  const next = clampAccountSchedulingThresholdOverride(accountSchedulingThresholdOverrideValue.value)
  if (current !== next) {
    credentials[ACCOUNT_SCHEDULING_THRESHOLD_CREDENTIAL_KEY] = next
  }
}

function loadTempUnschedRules(credentials?: Record<string, unknown>) {
  tempUnschedEnabled.value = credentials?.temp_unschedulable_enabled === true
  const rawRules = credentials?.temp_unschedulable_rules
  if (!Array.isArray(rawRules)) {
    tempUnschedRules.value = []
    return
  }

  tempUnschedRules.value = rawRules.map((rule) => {
    const entry = rule as Record<string, unknown>
    return {
      error_code: toPositiveNumber(entry.error_code),
      keywords: formatTempUnschedKeywords(entry.keywords),
      duration_minutes: toPositiveNumber(entry.duration_minutes),
      description: typeof entry.description === 'string' ? entry.description : ''
    }
  })
}

// Load quota control settings from account (Anthropic OAuth/SetupToken only)
function loadQuotaControlSettings(account: Account) {
  // Reset all quota control state first
  windowCostEnabled.value = false
  windowCostLimit.value = null
  windowCostStickyReserve.value = null
  sessionLimitEnabled.value = false
  maxSessions.value = null
  sessionIdleTimeout.value = null
  rpmLimitEnabled.value = false
  baseRpm.value = null
  rpmStrategy.value = 'tiered'
  rpmStickyBuffer.value = null
  userMsgQueueMode.value = ''
  tlsFingerprintEnabled.value = false
  tlsFingerprintProfileId.value = null
  sessionIdMaskingEnabled.value = false
  cacheTTLOverrideEnabled.value = false
  cacheTTLOverrideTarget.value = '5m'
  customBaseUrlEnabled.value = false
  customBaseUrl.value = ''

  // Remaining quota control settings only apply to Anthropic accounts
  if (account.platform !== 'anthropic') {
    return
  }

  // Window cost / session limit only apply to Anthropic OAuth/SetupToken accounts
  if (account.type !== 'oauth' && account.type !== 'setup-token') {
    return
  }

  // Load from extra field (via backend DTO fields)
  if (account.window_cost_limit != null && account.window_cost_limit > 0) {
    windowCostEnabled.value = true
    windowCostLimit.value = account.window_cost_limit
    windowCostStickyReserve.value = account.window_cost_sticky_reserve ?? 10
  }

  if (account.max_sessions != null && account.max_sessions > 0) {
    sessionLimitEnabled.value = true
    maxSessions.value = account.max_sessions
    sessionIdleTimeout.value = account.session_idle_timeout_minutes ?? 5
  }

  // RPM limit
  if (account.base_rpm != null && account.base_rpm > 0) {
    rpmLimitEnabled.value = true
    baseRpm.value = account.base_rpm
    rpmStrategy.value = (account.rpm_strategy as 'tiered' | 'sticky_exempt') || 'tiered'
    rpmStickyBuffer.value = account.rpm_sticky_buffer ?? null
  }

  // UMQ mode（独立于 RPM 加载，防止编辑无 RPM 账号时丢失已有配置）
  userMsgQueueMode.value = account.user_msg_queue_mode ?? ''

  // Load TLS fingerprint setting
  if (account.enable_tls_fingerprint === true) {
    tlsFingerprintEnabled.value = true
  }
  tlsFingerprintProfileId.value = account.tls_fingerprint_profile_id ?? null

  // Load session ID masking setting
  if (account.session_id_masking_enabled === true) {
    sessionIdMaskingEnabled.value = true
  }

  // Load cache TTL override setting
  if (account.cache_ttl_override_enabled === true) {
    cacheTTLOverrideEnabled.value = true
    cacheTTLOverrideTarget.value = account.cache_ttl_override_target || '5m'
  }

  // Load custom base URL setting
  if (account.custom_base_url_enabled === true) {
    customBaseUrlEnabled.value = true
    customBaseUrl.value = account.custom_base_url || ''
  }
}

function formatTempUnschedKeywords(value: unknown) {
  if (Array.isArray(value)) {
    return value
      .filter((item): item is string => typeof item === 'string')
      .map((item) => item.trim())
      .filter((item) => item.length > 0)
      .join(', ')
  }
  if (typeof value === 'string') {
    return value
  }
  return ''
}

const splitTempUnschedKeywords = (value: string) => {
  return value
    .split(/[,;]/)
    .map((item) => item.trim())
    .filter((item) => item.length > 0)
}

function toPositiveNumber(value: unknown) {
  const num = Number(value)
  if (!Number.isFinite(num) || num <= 0) {
    return null
  }
  return Math.trunc(num)
}

const formatDateTimeLocal = formatDateTimeLocalInput
const parseDateTimeLocal = parseDateTimeLocalInput

// Methods
const handleClose = () => {
  emit('close')
}

const submitUpdateAccount = async (accountID: string, updatePayload: Record<string, unknown>) => {
  submitting.value = true
  try {
    const updatedAccount = await adminAPI.accounts.update(accountID, updatePayload)
    appStore.showSuccess(t('admin.accounts.accountUpdated'))
    emit('updated', updatedAccount)
    handleClose()
  } catch (error: any) {
    appStore.showError(error.message || t('admin.accounts.failedToUpdate'))
  } finally {
    submitting.value = false
  }
}

const handleSubmit = async () => {
  if (!props.account) return
  const accountID = props.account.id

  if (form.status !== 'active' && form.status !== 'inactive' && form.status !== 'error') {
    appStore.showError(t('admin.accounts.pleaseSelectStatus'))
    return
  }

  const updatePayload: Record<string, unknown> = { ...form }
  try {
    // UUID 外键以 null 表示清除代理。
    if (form.expires_at === null) {
      updatePayload.expires_at = 0
    }
    // load_factor: 空值/NaN/0/负数 时发送 0（后端约定 <= 0 = 清除）
    const lf = form.load_factor
    if (lf == null || Number.isNaN(lf) || lf <= 0) {
      updatePayload.load_factor = 0
    }
    updatePayload.auto_pause_on_expired = autoPauseOnExpired.value
    if (props.account.type === 'apikey') {
      updatePayload.upstream_billing_probe_enabled = upstreamBillingAutoProbeEnabled.value
      updatePayload.upstream_billing_rate_sync_enabled = upstreamBillingRateSyncEnabled.value
      if (upstreamBillingRateSyncEnabled.value) {
        delete updatePayload.rate_multiplier
      }
    }

    // For apikey type, handle credentials update
    if (props.account.type === 'apikey') {
      const currentCredentials = (props.account.credentials as Record<string, unknown>) || {}
      const newBaseUrl = editBaseUrl.value.trim() || defaultBaseUrl.value
      const shouldApplyModelMapping = !(props.account.platform === 'openai' && openaiPassthroughEnabled.value)

      // Always update credentials for apikey type to handle model mapping changes
      const newCredentials: Record<string, unknown> = {
        ...currentCredentials,
        base_url: newBaseUrl
      }

      // 国产供应商：模式与协议写入凭据（决定额度/余额探测与转发端点/格式）。
      if (isCNApiKeyAccount.value) {
        newCredentials.account_mode = editAccountMode.value
        newCredentials.api_protocol = editApiProtocol.value
      }

      // Handle API key
      // 后端响应已脱敏：currentCredentials 不会再包含 api_key 原文。
      // 用户填入新值则覆盖；留空时优先看 credentials_status.has_api_key；
      // 若后端尚未升级（无 credentials_status），回退读旧结构 currentCredentials.api_key。
      // 两者都无才报错。
      const hasExistingApiKey =
        props.account.credentials_status?.has_api_key ?? Boolean(currentCredentials.api_key)
      if (editApiKey.value.trim()) {
        newCredentials.api_key = editApiKey.value.trim()
      } else if (!hasExistingApiKey) {
        appStore.showError(t('admin.accounts.apiKeyIsRequired'))
        return
      }

      // Add model mapping if configured（OpenAI 开启自动透传时保留现有映射，不再编辑）
      if (shouldApplyModelMapping) {
        const modelMapping = buildModelRestrictionMapping()
        if (modelMapping) {
          newCredentials.model_mapping = modelMapping
        } else {
          delete newCredentials.model_mapping
        }
      } else if (currentCredentials.model_mapping) {
        newCredentials.model_mapping = currentCredentials.model_mapping
      }
      if (props.account.platform === 'openai') {
        applyOpenAIEndpointCapabilities(newCredentials)
        const compactModelMapping = buildModelMappingObject('mapping', [], openAICompactModelMappings.value)
        if (compactModelMapping) {
          newCredentials.compact_model_mapping = compactModelMapping
        } else {
          delete newCredentials.compact_model_mapping
        }
      }

      // Add pool mode if enabled
      if (poolModeEnabled.value) {
        newCredentials.pool_mode = true
        newCredentials.pool_mode_retry_count = normalizePoolModeRetryCount(poolModeRetryCount.value)
        const parsedRetryStatusCodes = parsePoolModeRetryStatusCodes(poolModeRetryStatusCodesInput.value)
        if (parsedRetryStatusCodes.length > 0) {
          newCredentials.pool_mode_retry_status_codes = parsedRetryStatusCodes
        } else {
          delete newCredentials.pool_mode_retry_status_codes
        }
      } else {
        delete newCredentials.pool_mode
        delete newCredentials.pool_mode_retry_count
        delete newCredentials.pool_mode_retry_status_codes
      }

      // Add custom error codes if enabled
      if (customErrorCodesEnabled.value) {
        newCredentials.custom_error_codes_enabled = true
        newCredentials.custom_error_codes = [...selectedErrorCodes.value]
      } else {
        delete newCredentials.custom_error_codes_enabled
        delete newCredentials.custom_error_codes
      }

      // Add header override if enabled (anthropic/openai/grok apikey)
      if (isHeaderOverrideCapable(props.account.platform, 'apikey')) {
        if (headerOverrideEnabled.value) {
          const headerError = validateHeaderOverrideRows(headerOverrideRows.value)
          if (headerError) {
            appStore.showError(t(`admin.accounts.headerOverride.${headerError}`))
            return
          }
        }
        applyHeaderOverride(newCredentials, headerOverrideEnabled.value, headerOverrideRows.value, 'edit')
      }

      // Add intercept warmup requests setting
      applyInterceptWarmup(newCredentials, interceptWarmupRequests.value, 'edit')
      applyAccountSchedulingThresholdOverridePatch(newCredentials, currentCredentials)
      if (!applyTempUnschedConfig(newCredentials)) {
        return
      }

      updatePayload.credentials = newCredentials
    } else if (props.account.type === 'upstream') {
      const currentCredentials = (props.account.credentials as Record<string, unknown>) || {}
      const newCredentials: Record<string, unknown> = { ...currentCredentials }

      newCredentials.base_url = editBaseUrl.value.trim()

      if (editApiKey.value.trim()) {
        newCredentials.api_key = editApiKey.value.trim()
      }

      // Add intercept warmup requests setting
      applyInterceptWarmup(newCredentials, interceptWarmupRequests.value, 'edit')

      applyAccountSchedulingThresholdOverridePatch(newCredentials, currentCredentials)
      if (!applyTempUnschedConfig(newCredentials)) {
        return
      }

      updatePayload.credentials = newCredentials
    } else if ((props.account.platform === 'anthropic' || props.account.platform === 'gemini') && props.account.type === 'service_account') {
      const currentCredentials = (props.account.credentials as Record<string, unknown>) || {}
      const newCredentials: Record<string, unknown> = { ...currentCredentials }

      if (!editVertexProjectId.value.trim()) {
        appStore.showError(t('admin.accounts.vertexSaJsonMissingProjectId'))
        return
      }
      if (!editVertexClientEmail.value.trim()) {
        appStore.showError(t('admin.accounts.vertexSaJsonMissingClientEmail'))
        return
      }
      if (!editVertexLocation.value.trim()) {
        appStore.showError(t('admin.accounts.vertexLocationRequired'))
        return
      }

      // SA JSON 已脱敏不再随 credentials 返回，存在性优先读 credentials_status。
      // 若后端尚未升级（无 credentials_status），回退读旧结构 service_account_json / service_account。
      const credentialsStatus = props.account.credentials_status
      const hasExistingServiceAccountJson = credentialsStatus
        ? Boolean(
            credentialsStatus.has_service_account_json || credentialsStatus.has_service_account
          )
        : Boolean(currentCredentials.service_account_json || currentCredentials.service_account)
      if (!hasExistingServiceAccountJson) {
        appStore.showError(t('admin.accounts.vertexSaJsonRequired'))
        return
      }
      newCredentials.project_id = editVertexProjectId.value.trim()
      newCredentials.client_email = editVertexClientEmail.value.trim()
      newCredentials.location = editVertexLocation.value.trim()
      newCredentials.tier_id = 'vertex'

      // Add model mapping if configured
      const modelMapping = buildModelRestrictionMapping()
      if (modelMapping) {
        newCredentials.model_mapping = modelMapping
      } else {
        delete newCredentials.model_mapping
      }

      applyInterceptWarmup(newCredentials, interceptWarmupRequests.value, 'edit')
      applyAccountSchedulingThresholdOverridePatch(newCredentials, currentCredentials)
      if (!applyTempUnschedConfig(newCredentials)) {
        return
      }

      updatePayload.credentials = newCredentials
    } else if (props.account.type === 'bedrock') {
      const currentCredentials = (props.account.credentials as Record<string, unknown>) || {}
      const newCredentials: Record<string, unknown> = { ...currentCredentials }

      newCredentials.aws_region = editBedrockRegion.value.trim()
      if (editBedrockForceGlobal.value) {
        newCredentials.aws_force_global = 'true'
      } else {
        delete newCredentials.aws_force_global
      }

      if (isBedrockAPIKeyMode.value) {
        // API Key mode: only update api_key if user provided new value
        if (editBedrockApiKeyValue.value.trim()) {
          newCredentials.api_key = editBedrockApiKeyValue.value.trim()
        }
      } else {
        // SigV4 mode
        newCredentials.aws_access_key_id = editBedrockAccessKeyId.value.trim()
        if (editBedrockSecretAccessKey.value.trim()) {
          newCredentials.aws_secret_access_key = editBedrockSecretAccessKey.value.trim()
        }
        if (editBedrockSessionToken.value.trim()) {
          newCredentials.aws_session_token = editBedrockSessionToken.value.trim()
        }
      }

      // Pool mode
      if (poolModeEnabled.value) {
        newCredentials.pool_mode = true
        newCredentials.pool_mode_retry_count = normalizePoolModeRetryCount(poolModeRetryCount.value)
        const parsedRetryStatusCodes = parsePoolModeRetryStatusCodes(poolModeRetryStatusCodesInput.value)
        if (parsedRetryStatusCodes.length > 0) {
          newCredentials.pool_mode_retry_status_codes = parsedRetryStatusCodes
        } else {
          delete newCredentials.pool_mode_retry_status_codes
        }
      } else {
        delete newCredentials.pool_mode
        delete newCredentials.pool_mode_retry_count
        delete newCredentials.pool_mode_retry_status_codes
      }

      // Model mapping
      const modelMapping = buildModelRestrictionMapping()
      if (modelMapping) {
        newCredentials.model_mapping = modelMapping
      } else {
        delete newCredentials.model_mapping
      }

      applyInterceptWarmup(newCredentials, interceptWarmupRequests.value, 'edit')
      applyAccountSchedulingThresholdOverridePatch(newCredentials, currentCredentials)
      if (!applyTempUnschedConfig(newCredentials)) {
        return
      }

      updatePayload.credentials = newCredentials
    } else {
      // For oauth/setup-token types, only update intercept_warmup_requests if changed
      const currentCredentials = (props.account.credentials as Record<string, unknown>) || {}
      const newCredentials: Record<string, unknown> = { ...currentCredentials }

      applyInterceptWarmup(newCredentials, interceptWarmupRequests.value, 'edit')
      applyAccountSchedulingThresholdOverridePatch(newCredentials, currentCredentials)
      if (!applyTempUnschedConfig(newCredentials)) {
        return
      }

      updatePayload.credentials = newCredentials
    }

    // OpenAI/Grok OAuth: persist model mapping to credentials
    if ((props.account.platform === 'openai' || props.account.platform === 'gemini' || props.account.platform === 'antigravity' || props.account.platform === 'grok') && props.account.type === 'oauth') {
      const currentCredentials = isSparkShadow.value
        ? {}
        : (updatePayload.credentials as Record<string, unknown>) ||
          ((props.account.credentials as Record<string, unknown>) || {})
      const newCredentials: Record<string, unknown> = { ...currentCredentials }
      if (props.account.platform === 'openai') {
        applyOpenAIModelMappingCredentials(newCredentials)
      } else {
        const modelMapping = buildModelRestrictionMapping()
        if (modelMapping) {
          newCredentials.model_mapping = modelMapping
        } else {
          delete newCredentials.model_mapping
        }
      }

      updatePayload.credentials = newCredentials
    }

    // Grok OAuth: 自定义上游地址 + 请求头覆写。base_url 仅改写转发端点，
    // OAuth 授权与令牌刷新链路不读取该值；关闭开关即恢复默认官方网关。
    if (props.account.platform === 'grok' && props.account.type === 'oauth') {
      const currentCredentials =
        (updatePayload.credentials as Record<string, unknown>) ||
        ((props.account.credentials as Record<string, unknown>) || {})
      const newCredentials: Record<string, unknown> = { ...currentCredentials }

      if (grokOAuthCustomBaseUrlEnabled.value) {
        const trimmedBaseUrl = grokOAuthBaseUrl.value.trim()
        if (!trimmedBaseUrl) {
          appStore.showError(t('admin.accounts.grokCustomBaseUrl.required'))
          return
        }
        if (!/^https?:\/\//i.test(trimmedBaseUrl)) {
          appStore.showError(t('admin.accounts.grokCustomBaseUrl.invalid'))
          return
        }
        newCredentials.base_url = trimmedBaseUrl
      } else {
        delete newCredentials.base_url
      }

      if (headerOverrideEnabled.value) {
        const headerError = validateHeaderOverrideRows(headerOverrideRows.value)
        if (headerError) {
          appStore.showError(t(`admin.accounts.headerOverride.${headerError}`))
          return
        }
      }
      applyHeaderOverride(newCredentials, headerOverrideEnabled.value, headerOverrideRows.value, 'edit')

      updatePayload.credentials = newCredentials

      const newExtra: Record<string, unknown> = {
        ...((props.account.extra as Record<string, unknown>) || {})
      }
      // Persist both states so a disabled account remains opted out when the
      // backend applies the default-enabled policy to missing values.
      newExtra[GROK_CLIENT_TOOL_CACHE_EXTRA_KEY] = grokClientToolCacheEnabled.value
      updatePayload.extra = newExtra
    }

    // OpenAI: 手动覆盖订阅档位 plan_type（Plus/Pro/Free）。仅 OAuth 非影子账号：
    // 影子账号凭据由母账号管理(且后端会 sanitize),setup-token 无订阅调度语义。
    if (props.account.platform === 'openai' && props.account.type === 'oauth' && !isSparkShadow.value) {
      const currentCredentials = (updatePayload.credentials as Record<string, unknown>) ||
        ((props.account.credentials as Record<string, unknown>) || {})
      updatePayload.credentials = applyPlanType({ ...currentCredentials }, editPlanType.value)
    }

    // For Anthropic OAuth/SetupToken accounts, handle quota control settings in extra
    if (props.account.platform === 'anthropic' && (props.account.type === 'oauth' || props.account.type === 'setup-token')) {
      const currentExtra = (updatePayload.extra as Record<string, unknown>) || (props.account.extra as Record<string, unknown>) || {}
      const newExtra: Record<string, unknown> = { ...currentExtra }

      // Window cost limit settings
      if (windowCostEnabled.value && windowCostLimit.value != null && windowCostLimit.value > 0) {
        newExtra.window_cost_limit = windowCostLimit.value
        newExtra.window_cost_sticky_reserve = windowCostStickyReserve.value ?? 10
      } else {
        delete newExtra.window_cost_limit
        delete newExtra.window_cost_sticky_reserve
      }

      // Session limit settings
      if (sessionLimitEnabled.value && maxSessions.value != null && maxSessions.value > 0) {
        newExtra.max_sessions = maxSessions.value
        newExtra.session_idle_timeout_minutes = sessionIdleTimeout.value ?? 5
      } else {
        delete newExtra.max_sessions
        delete newExtra.session_idle_timeout_minutes
      }

      // RPM limit settings
      if (rpmLimitEnabled.value) {
        const DEFAULT_BASE_RPM = 15
        newExtra.base_rpm = (baseRpm.value != null && baseRpm.value > 0)
          ? baseRpm.value
          : DEFAULT_BASE_RPM
        newExtra.rpm_strategy = rpmStrategy.value
        if (rpmStickyBuffer.value != null && rpmStickyBuffer.value > 0) {
          newExtra.rpm_sticky_buffer = rpmStickyBuffer.value
        } else {
          delete newExtra.rpm_sticky_buffer
        }
      } else {
        delete newExtra.base_rpm
        delete newExtra.rpm_strategy
        delete newExtra.rpm_sticky_buffer
      }

      // UMQ mode（独立于 RPM 保存）
      if (userMsgQueueMode.value) {
        newExtra.user_msg_queue_mode = userMsgQueueMode.value
      } else {
        delete newExtra.user_msg_queue_mode
      }
      delete newExtra.user_msg_queue_enabled  // 清理旧字段

      // TLS fingerprint setting
      if (tlsFingerprintEnabled.value) {
        newExtra.enable_tls_fingerprint = true
        if (tlsFingerprintProfileId.value) {
          newExtra.tls_fingerprint_profile_id = tlsFingerprintProfileId.value
        } else {
          delete newExtra.tls_fingerprint_profile_id
        }
      } else {
        delete newExtra.enable_tls_fingerprint
        delete newExtra.tls_fingerprint_profile_id
      }

      // Session ID masking setting
      if (sessionIdMaskingEnabled.value) {
        newExtra.session_id_masking_enabled = true
      } else {
        delete newExtra.session_id_masking_enabled
      }

      // Cache TTL override setting
      if (cacheTTLOverrideEnabled.value) {
        newExtra.cache_ttl_override_enabled = true
        newExtra.cache_ttl_override_target = cacheTTLOverrideTarget.value
      } else {
        delete newExtra.cache_ttl_override_enabled
        delete newExtra.cache_ttl_override_target
      }

      // Custom base URL relay setting
      if (customBaseUrlEnabled.value && customBaseUrl.value.trim()) {
        newExtra.custom_base_url_enabled = true
        newExtra.custom_base_url = customBaseUrl.value.trim()
      } else {
        delete newExtra.custom_base_url_enabled
        delete newExtra.custom_base_url
      }

      updatePayload.extra = newExtra
    }

    // For Anthropic API Key accounts, handle passthrough mode + web search emulation in extra
    if (props.account.platform === 'anthropic' && props.account.type === 'apikey') {
      const currentExtra = (updatePayload.extra as Record<string, unknown>) || (props.account.extra as Record<string, unknown>) || {}
      const newExtra: Record<string, unknown> = { ...currentExtra }
      if (anthropicPassthroughEnabled.value) {
        newExtra.anthropic_passthrough = true
      } else {
        delete newExtra.anthropic_passthrough
      }
      if (anthropicAPIKeyAuthScheme.value === 'authorization_bearer') {
        newExtra.anthropic_apikey_auth_scheme = 'authorization_bearer'
      } else {
        delete newExtra.anthropic_apikey_auth_scheme
      }
      if (webSearchEmulationMode.value === 'default') {
        delete newExtra.web_search_emulation
      } else {
        newExtra.web_search_emulation = webSearchEmulationMode.value
      }
      updatePayload.extra = newExtra
    }

    // For OpenAI OAuth/SetupToken/API Key accounts, handle passthrough mode in extra
    if (props.account.platform === 'openai' && (props.account.type === 'oauth' || props.account.type === 'setup-token' || props.account.type === 'apikey')) {
      const currentExtra = (props.account.extra as Record<string, unknown>) || {}
      const newExtra: Record<string, unknown> = { ...currentExtra }
      const hadCodexCLIOnlyEnabled = currentExtra.codex_cli_only === true
      if (props.account.type === 'oauth' || props.account.type === 'setup-token') {
        newExtra.openai_oauth_responses_websockets_v2_mode = openaiOAuthResponsesWebSocketV2Mode.value
        newExtra.openai_oauth_responses_websockets_v2_enabled = isOpenAIWSModeEnabled(openaiOAuthResponsesWebSocketV2Mode.value)
      } else if (props.account.type === 'apikey') {
        newExtra.openai_apikey_responses_websockets_v2_mode = openaiAPIKeyResponsesWebSocketV2Mode.value
        newExtra.openai_apikey_responses_websockets_v2_enabled = isOpenAIWSModeEnabled(openaiAPIKeyResponsesWebSocketV2Mode.value)
      }
      delete newExtra.responses_websockets_v2_enabled
      delete newExtra.openai_ws_enabled
      if (openaiPassthroughEnabled.value) {
        newExtra.openai_passthrough = true
      } else {
        delete newExtra.openai_passthrough
        delete newExtra.openai_oauth_passthrough
      }
      // 缺省即保留 namespace，不写空值，避免 extra 里堆积默认项
      if (props.account.type === 'oauth' && openaiFlattenNamespacesEnabled.value) {
        newExtra.openai_responses_flatten_namespaces = true
      } else {
        delete newExtra.openai_responses_flatten_namespaces
      }
      if (isSparkShadow.value) {
        delete newExtra.openai_long_context_billing_enabled
      } else {
        newExtra.openai_long_context_billing_enabled = openAILongContextBillingEnabled.value
      }
      if (openAICompactMode.value === 'auto') {
        delete newExtra.openai_compact_mode
      } else {
        newExtra.openai_compact_mode = openAICompactMode.value
      }
		if (props.account.type === 'apikey') {
        if (!openAITextGenerationCapabilityEnabled.value || openAIResponsesMode.value === 'auto') {
          delete newExtra.openai_responses_mode
        } else {
          newExtra.openai_responses_mode = openAIResponsesMode.value
        }
		}
		if (autoPause5hThreshold.value != null && autoPause5hThreshold.value > 0) {
			newExtra.auto_pause_5h_threshold = autoPause5hThreshold.value / 100
		} else {
			delete newExtra.auto_pause_5h_threshold
		}
		if (autoPause7dThreshold.value != null && autoPause7dThreshold.value > 0) {
			newExtra.auto_pause_7d_threshold = autoPause7dThreshold.value / 100
		} else {
			delete newExtra.auto_pause_7d_threshold
		}
		if (autoPause5hDisabled.value) {
			newExtra.auto_pause_5h_disabled = true
		} else {
			delete newExtra.auto_pause_5h_disabled
		}
		if (autoPause7dDisabled.value) {
			newExtra.auto_pause_7d_disabled = true
		} else {
			delete newExtra.auto_pause_7d_disabled
		}

		delete newExtra.codex_image_generation_bridge_enabled
      switch (codexImageToolMode.value) {
        case 'enabled':
        case 'disabled':
          newExtra.codex_image_generation_bridge = codexImageToolMode.value === 'enabled'
          delete newExtra.codex_image_generation_explicit_tool_policy
          break
        case 'block':
          newExtra.codex_image_generation_explicit_tool_policy = 'strip'
          delete newExtra.codex_image_generation_bridge
          break
        default:
          delete newExtra.codex_image_generation_bridge
          delete newExtra.codex_image_generation_explicit_tool_policy
      }

      if (props.account.type === 'oauth' || props.account.type === 'setup-token') {
        if (codexCLIOnlyEnabled.value) {
          newExtra.codex_cli_only = true
        } else if (hadCodexCLIOnlyEnabled) {
          // 关闭时显式写 false，避免 extra 为空被后端忽略导致旧值无法清除
          newExtra.codex_cli_only = false
        } else {
          delete newExtra.codex_cli_only
        }
        // Claude Code 插件放行已迁移到全局 codex_cli_only_whitelist，编辑时清理废弃账号级快捷字段。
        delete newExtra.codex_cli_only_allowed_clients
        if (codexCLIOnlyEnabled.value && codexCLIOnlyAppServerEnabled.value) {
          newExtra.codex_cli_only_allow_app_server = true
        } else {
          delete newExtra.codex_cli_only_allow_app_server
        }
      }

      // 指纹收敛模式：默认 off（不写入）；device/session/full 是显式 opt-in，
      // 必须落键，否则管理员的选择会被后端当作"未设置"而回落到 off（#5610）。
      if (props.account.type === 'oauth') {
        if (codexFingerprintMode.value !== 'off') {
          newExtra.codex_fingerprint_mode = codexFingerprintMode.value
        } else {
          delete newExtra.codex_fingerprint_mode
        }
      }

      updatePayload.extra = newExtra
    }

    // For apikey/bedrock accounts, handle quota_limit in extra
    if (props.account.type === 'apikey' || props.account.type === 'bedrock') {
      const currentExtra = (updatePayload.extra as Record<string, unknown>) ||
        (props.account.extra as Record<string, unknown>) || {}
      const newExtra: Record<string, unknown> = { ...currentExtra }
      // 上游倍率自动探测对全部 API-key 平台开放（EasySub2api 上游即可应答），
      // Bedrock 凭证无静态 Key 不参与。
      if (props.account.type === 'apikey') {
        delete newExtra.upstream_billing_probe_enabled
        delete newExtra.upstream_billing_rate_sync_enabled
      }
      // Total quota
      if (editQuotaLimit.value != null && editQuotaLimit.value > 0) {
        newExtra.quota_limit = editQuotaLimit.value
      } else {
        delete newExtra.quota_limit
      }
      // Daily quota
      if (editQuotaDailyLimit.value != null && editQuotaDailyLimit.value > 0) {
        newExtra.quota_daily_limit = editQuotaDailyLimit.value
      } else {
        delete newExtra.quota_daily_limit
        delete newExtra.quota_daily_used
        delete newExtra.quota_daily_start
      }
      // Weekly quota
      if (editQuotaWeeklyLimit.value != null && editQuotaWeeklyLimit.value > 0) {
        newExtra.quota_weekly_limit = editQuotaWeeklyLimit.value
      } else {
        delete newExtra.quota_weekly_limit
        delete newExtra.quota_weekly_used
        delete newExtra.quota_weekly_start
      }
      // Quota reset mode config
      if (editDailyResetMode.value === 'fixed') {
        newExtra.quota_daily_reset_mode = 'fixed'
        newExtra.quota_daily_reset_hour = editDailyResetHour.value ?? 0
      } else {
        delete newExtra.quota_daily_reset_mode
        delete newExtra.quota_daily_reset_hour
      }
      if (editWeeklyResetMode.value === 'fixed') {
        newExtra.quota_weekly_reset_mode = 'fixed'
        newExtra.quota_weekly_reset_day = editWeeklyResetDay.value ?? 1
        newExtra.quota_weekly_reset_hour = editWeeklyResetHour.value ?? 0
      } else {
        delete newExtra.quota_weekly_reset_mode
        delete newExtra.quota_weekly_reset_day
        delete newExtra.quota_weekly_reset_hour
      }
      if (editDailyResetMode.value === 'fixed' || editWeeklyResetMode.value === 'fixed') {
        newExtra.quota_reset_timezone = editResetTimezone.value || 'UTC'
      } else {
        delete newExtra.quota_reset_timezone
      }
      // Quota notify config
      writeQuotaNotifyToExtra(newExtra, 'update')
      updatePayload.extra = newExtra
    }

    await submitUpdateAccount(accountID, updatePayload)
  } catch (error: any) {
    appStore.showError(error.message || t('admin.accounts.failedToUpdate'))
  }
}

</script>
