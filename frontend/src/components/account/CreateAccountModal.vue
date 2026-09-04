<template>
  <BaseDialog
    :show="show"
    :title="t('admin.accounts.createAccount')"
    width="wide"
    @close="handleClose"
  >
    <!-- Step Indicator for OAuth accounts -->
    <div v-if="isOAuthFlow" class="components-account-create-account-modal__panel">
      <div class="components-account-create-account-modal__panel-2">
        <div class="components-account-create-account-modal__panel-3">
          <div
            :class="[
              'components-account-create-account-modal__panel-80',
              step >= 1 ? 'components-account-create-account-modal__panel-81' : 'components-account-create-account-modal__panel-82'
            ]"
          >
            1
          </div>
          <span class="components-account-create-account-modal__text">{{
            t('admin.accounts.oauth.authMethod')
          }}</span>
        </div>
        <div class="components-account-create-account-modal__panel-4" />
        <div class="components-account-create-account-modal__panel-3">
          <div
            :class="[
              'components-account-create-account-modal__panel-80',
              step >= 2 ? 'components-account-create-account-modal__panel-81' : 'components-account-create-account-modal__panel-82'
            ]"
          >
            2
          </div>
          <span class="components-account-create-account-modal__text">{{
            oauthStepTitle
          }}</span>
        </div>
      </div>
    </div>

    <!-- Step 1: Basic Info -->
    <form
      v-if="step === 1"
      id="create-account-form"
      @submit.prevent="handleSubmit"
      class="components-account-create-account-modal__form"
    >
      <div>
        <label class="input-label">{{ t('admin.accounts.accountName') }}</label>
        <input
          v-model="form.name"
          type="text"
          :required="!isGrokSSOInputMethod"
          class="input"
          :placeholder="t('admin.accounts.enterAccountName')"
          data-tour="account-form-name"
        />
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

      <!-- Platform Selection - Segmented Control Style -->
      <div>
        <label class="input-label">{{ t('admin.accounts.platform') }}</label>
        <div class="components-account-create-account-modal__panel-5" data-tour="account-form-platform">
          <button
            type="button"
            @click="form.platform = 'anthropic'"
            :class="[
              'components-account-create-account-modal__action-14',
              form.platform === 'anthropic'
                ? 'components-account-create-account-modal__action-15'
                : 'components-account-create-account-modal__action-16'
            ]"
          >
            <Icon name="sparkles" size="sm" />
            Anthropic
          </button>
          <button
            type="button"
            @click="form.platform = 'openai'"
            :class="[
              'components-account-create-account-modal__action-14',
              form.platform === 'openai'
                ? 'components-account-create-account-modal__action-17'
                : 'components-account-create-account-modal__action-16'
            ]"
          >
            <svg
              class="components-account-create-account-modal__icon"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
              stroke-width="1.5"
            >
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                d="M3.75 13.5l10.5-11.25L12 10.5h8.25L9.75 21.75 12 13.5H3.75z"
              />
            </svg>
            OpenAI
          </button>
          <button
            type="button"
            @click="form.platform = 'grok'"
            :class="[
              'components-account-create-account-modal__action-14',
              form.platform === 'grok'
                ? 'components-account-create-account-modal__action-20'
                : 'components-account-create-account-modal__action-16'
            ]"
          >
            <PlatformIcon platform="grok" size="sm" />
            Grok
          </button>
        </div>
        <!-- Google platforms row: Gemini / Antigravity -->
        <div class="components-account-create-account-modal__panel-5">
          <button
            type="button"
            @click="form.platform = 'gemini'"
            :class="[
              'components-account-create-account-modal__action-14',
              form.platform === 'gemini'
                ? 'components-account-create-account-modal__action-17'
                : 'components-account-create-account-modal__action-16'
            ]"
          >
            <PlatformIcon platform="gemini" size="sm" />
            Gemini
          </button>
          <button
            type="button"
            @click="form.platform = 'antigravity'"
            :class="[
              'components-account-create-account-modal__action-14',
              form.platform === 'antigravity'
                ? 'components-account-create-account-modal__action-22'
                : 'components-account-create-account-modal__action-16'
            ]"
          >
            <PlatformIcon platform="antigravity" size="sm" />
            Antigravity
          </button>
        </div>
        <!-- CN providers row: Kimi / Zhipu GLM / DeepSeek -->
        <div class="components-account-create-account-modal__panel-5">
          <button
            type="button"
            @click="selectCNPlatform('kimi')"
            :class="[
              'components-account-create-account-modal__action-14',
              form.platform === 'kimi'
                ? 'components-account-create-account-modal__action-21'
                : 'components-account-create-account-modal__action-16'
            ]"
          >
            <PlatformIcon platform="kimi" size="sm" />
            Kimi
          </button>
          <button
            type="button"
            @click="selectCNPlatform('zhipu')"
            :class="[
              'components-account-create-account-modal__action-14',
              form.platform === 'zhipu'
                ? 'components-account-create-account-modal__action-22'
                : 'components-account-create-account-modal__action-16'
            ]"
          >
            <PlatformIcon platform="zhipu" size="sm" />
            Zhipu GLM
          </button>
          <button
            type="button"
            @click="selectCNPlatform('deepseek')"
            :class="[
              'components-account-create-account-modal__action-14',
              form.platform === 'deepseek'
                ? 'components-account-create-account-modal__action-23'
                : 'components-account-create-account-modal__action-16'
            ]"
          >
            <PlatformIcon platform="deepseek" size="sm" />
            DeepSeek
          </button>
        </div>
      </div>

      <!-- Account Type Selection (Gemini / Antigravity) -->
      <div v-if="form.platform === 'gemini' || form.platform === 'antigravity'">
        <label class="input-label">{{ t('admin.accounts.accountType') }}</label>
        <div class="components-account-create-account-modal__panel-9" data-tour="account-form-type">
          <button
            type="button"
            @click="accountCategory = 'oauth-based'"
            :class="[
              'components-account-create-account-modal__action-24',
              accountCategory === 'oauth-based'
                ? 'components-account-create-account-modal__action-35'
                : 'components-account-create-account-modal__action-36'
            ]"
          >
            <div :class="['components-account-create-account-modal__panel-83', accountCategory === 'oauth-based' ? 'components-account-create-account-modal__panel-90' : 'components-account-create-account-modal__panel-85']">
              <Icon name="key" size="sm" />
            </div>
            <div>
              <span class="components-account-create-account-modal__text-2">OAuth</span>
              <span class="components-account-create-account-modal__text-3">{{ form.platform === 'gemini' ? t('admin.accounts.types.geminiOauth') : t('admin.accounts.types.antigravityOauth') }}</span>
            </div>
          </button>
          <button
            type="button"
            @click="accountCategory = 'apikey'"
            :class="[
              'components-account-create-account-modal__action-24',
              accountCategory === 'apikey'
                ? 'components-account-create-account-modal__action-27'
                : 'components-account-create-account-modal__action-28'
            ]"
          >
            <div :class="['components-account-create-account-modal__panel-83', accountCategory === 'apikey' ? 'components-account-create-account-modal__panel-86' : 'components-account-create-account-modal__panel-85']">
              <Icon name="key" size="sm" />
            </div>
            <div>
              <span class="components-account-create-account-modal__text-2">API Key</span>
              <span class="components-account-create-account-modal__text-3">{{ t('admin.accounts.apiKey') }}</span>
            </div>
          </button>
          <button
            v-if="form.platform === 'gemini'"
            type="button"
            @click="accountCategory = 'service_account'"
            :class="[
              'components-account-create-account-modal__action-24',
              accountCategory === 'service_account'
                ? 'components-account-create-account-modal__action-31'
                : 'components-account-create-account-modal__action-32'
            ]"
          >
            <div :class="['components-account-create-account-modal__panel-83', accountCategory === 'service_account' ? 'components-account-create-account-modal__panel-88' : 'components-account-create-account-modal__panel-85']">
              <Icon name="cloud" size="sm" />
            </div>
            <div>
              <span class="components-account-create-account-modal__text-2">Vertex</span>
              <span class="components-account-create-account-modal__text-3">Service Account</span>
            </div>
          </button>
        </div>
      </div>

      <!-- Gemini OAuth provider and fallback tier -->
      <div v-if="form.platform === 'gemini' && accountCategory === 'oauth-based'" class="components-account-create-account-modal__panel-23">
        <div>
          <label class="input-label">OAuth Provider</label>
          <Select v-model="geminiOAuthType" :options="[
            { value: 'google_one', label: 'Google One' },
            { value: 'code_assist', label: 'GCP Code Assist' },
            ...(geminiAIStudioOAuthEnabled ? [{ value: 'ai_studio', label: 'AI Studio' }] : [])
          ]" />
        </div>
        <div>
          <label class="input-label">Tier</label>
          <Select v-if="geminiOAuthType === 'google_one'" v-model="geminiTierGoogleOne" :options="[
            { value: 'google_one_free', label: 'Google One Free' },
            { value: 'google_ai_pro', label: 'Google AI Pro' },
            { value: 'google_ai_ultra', label: 'Google AI Ultra' }
          ]" />
          <Select v-else-if="geminiOAuthType === 'code_assist'" v-model="geminiTierGcp" :options="[
            { value: 'gcp_standard', label: 'GCP Standard' },
            { value: 'gcp_enterprise', label: 'GCP Enterprise' }
          ]" />
          <Select v-else v-model="geminiTierAIStudio" :options="[
            { value: 'aistudio_free', label: 'AI Studio Free' },
            { value: 'aistudio_paid', label: 'AI Studio Paid' }
          ]" />
        </div>
      </div>

      <!-- Account Type Selection (Anthropic) -->
      <div v-if="form.platform === 'anthropic'">
        <label class="input-label">{{ t('admin.accounts.accountType') }}</label>
        <div class="components-account-create-account-modal__panel-6" data-tour="account-form-type">
          <button
            type="button"
            @click="accountCategory = 'oauth-based'"
            :class="[
              'components-account-create-account-modal__action-24',
              accountCategory === 'oauth-based'
                ? 'components-account-create-account-modal__action-25'
                : 'components-account-create-account-modal__action-26'
            ]"
          >
            <div
              :class="[
                'components-account-create-account-modal__panel-83',
                accountCategory === 'oauth-based'
                  ? 'components-account-create-account-modal__panel-84'
                  : 'components-account-create-account-modal__panel-85'
              ]"
            >
              <Icon name="sparkles" size="sm" />
            </div>
            <div>
              <span class="components-account-create-account-modal__text-2">{{
                t('admin.accounts.claudeCode')
              }}</span>
              <span class="components-account-create-account-modal__text-3">{{
                t('admin.accounts.oauthSetupToken')
              }}</span>
            </div>
          </button>

          <button
            type="button"
            @click="accountCategory = 'apikey'"
            :class="[
              'components-account-create-account-modal__action-24',
              accountCategory === 'apikey'
                ? 'components-account-create-account-modal__action-27'
                : 'components-account-create-account-modal__action-28'
            ]"
          >
            <div
              :class="[
                'components-account-create-account-modal__panel-83',
                accountCategory === 'apikey'
                  ? 'components-account-create-account-modal__panel-86'
                  : 'components-account-create-account-modal__panel-85'
              ]"
            >
              <Icon name="key" size="sm" />
            </div>
            <div>
              <span class="components-account-create-account-modal__text-2">{{
                t('admin.accounts.claudeConsole')
              }}</span>
              <span class="components-account-create-account-modal__text-3">{{
                t('admin.accounts.apiKey')
              }}</span>
            </div>
          </button>

          <button
            type="button"
            @click="accountCategory = 'bedrock'"
            :class="[
              'components-account-create-account-modal__action-24',
              accountCategory === 'bedrock'
                ? 'components-account-create-account-modal__action-29'
                : 'components-account-create-account-modal__action-30'
            ]"
          >
            <div
              :class="[
                'components-account-create-account-modal__panel-83',
                accountCategory === 'bedrock'
                  ? 'components-account-create-account-modal__panel-87'
                  : 'components-account-create-account-modal__panel-85'
              ]"
            >
              <Icon name="cloud" size="sm" />
            </div>
            <div>
              <span class="components-account-create-account-modal__text-2">{{
                t('admin.accounts.bedrockLabel')
              }}</span>
              <span class="components-account-create-account-modal__text-3">{{
                t('admin.accounts.bedrockDesc')
              }}</span>
            </div>
          </button>

          <button
            type="button"
            @click="accountCategory = 'service_account'"
            :class="[
              'components-account-create-account-modal__action-24',
              accountCategory === 'service_account'
                ? 'components-account-create-account-modal__action-31'
                : 'components-account-create-account-modal__action-32'
            ]"
          >
            <div
              :class="[
                'components-account-create-account-modal__panel-83',
                accountCategory === 'service_account'
                  ? 'components-account-create-account-modal__panel-88'
                  : 'components-account-create-account-modal__panel-85'
              ]"
            >
              <Icon name="cloud" size="sm" />
            </div>
            <div>
              <span class="components-account-create-account-modal__text-2">Vertex</span>
              <span class="components-account-create-account-modal__text-3">Service Account</span>
            </div>
          </button>

        </div>

        <div
          v-if="accountCategory === 'service_account'"
          class="components-account-create-account-modal__panel-7"
        >
          <p>{{ t('admin.accounts.vertexAnthropicHint') }}</p>
        </div>
      </div>

      <!-- Account Type Selection (OpenAI) -->
      <div v-if="form.platform === 'openai'">
        <label class="input-label">{{ t('admin.accounts.accountType') }}</label>
        <div class="components-account-create-account-modal__panel-8" data-tour="account-form-type">
          <button
            type="button"
            @click="accountCategory = 'oauth-based'"
            :class="[
              'components-account-create-account-modal__action-24',
              accountCategory === 'oauth-based'
                ? 'components-account-create-account-modal__action-33'
                : 'components-account-create-account-modal__action-34'
            ]"
          >
            <div
              :class="[
                'components-account-create-account-modal__panel-83',
                accountCategory === 'oauth-based'
                  ? 'components-account-create-account-modal__panel-89'
                  : 'components-account-create-account-modal__panel-85'
              ]"
            >
              <Icon name="key" size="sm" />
            </div>
            <div>
              <span class="components-account-create-account-modal__text-2">OAuth</span>
              <span class="components-account-create-account-modal__text-3">{{ t('admin.accounts.types.chatgptOauth') }}</span>
            </div>
          </button>

          <button
            type="button"
            @click="accountCategory = 'apikey'"
            :class="[
              'components-account-create-account-modal__action-24',
              accountCategory === 'apikey'
                ? 'components-account-create-account-modal__action-27'
                : 'components-account-create-account-modal__action-28'
            ]"
          >
            <div
              :class="[
                'components-account-create-account-modal__panel-83',
                accountCategory === 'apikey'
                  ? 'components-account-create-account-modal__panel-86'
                  : 'components-account-create-account-modal__panel-85'
              ]"
            >
              <Icon name="key" size="sm" />
            </div>
            <div>
              <span class="components-account-create-account-modal__text-2">API Key</span>
              <span class="components-account-create-account-modal__text-3">{{ t('admin.accounts.types.responsesApi') }}</span>
            </div>
          </button>

        </div>
      </div>

      <!-- Account Type Selection (Grok) -->
      <div v-if="form.platform === 'grok'">
        <label class="input-label">{{ t('admin.accounts.accountType') }}</label>
        <div class="components-account-create-account-modal__panel-9" data-tour="account-form-type">
          <button
            type="button"
            @click="accountCategory = 'oauth-based'"
            :class="[
              'components-account-create-account-modal__action-24',
              accountCategory === 'oauth-based'
                ? 'components-account-create-account-modal__action-35'
                : 'components-account-create-account-modal__action-36'
            ]"
          >
            <div
              :class="[
                'components-account-create-account-modal__panel-83',
                accountCategory === 'oauth-based'
                  ? 'components-account-create-account-modal__panel-90'
                  : 'components-account-create-account-modal__panel-85'
              ]"
            >
              <PlatformIcon platform="grok" size="sm" />
            </div>
            <div>
              <span class="components-account-create-account-modal__text-2">OAuth</span>
              <span class="components-account-create-account-modal__text-3">{{ t('admin.accounts.types.grokOauth') }}</span>
            </div>
          </button>

          <button
            type="button"
            data-testid="grok-account-type-api-key"
            @click="accountCategory = 'apikey'"
            :class="[
              'components-account-create-account-modal__action-24',
              accountCategory === 'apikey'
                ? 'components-account-create-account-modal__action-27'
                : 'components-account-create-account-modal__action-28'
            ]"
          >
            <div
              :class="[
                'components-account-create-account-modal__panel-83',
                accountCategory === 'apikey'
                  ? 'components-account-create-account-modal__panel-86'
                  : 'components-account-create-account-modal__panel-85'
              ]"
            >
              <Icon name="key" size="sm" />
            </div>
            <div>
              <span class="components-account-create-account-modal__text-2">API Key</span>
              <span class="components-account-create-account-modal__text-3">{{ t('admin.accounts.types.responsesApi') }}</span>
            </div>
          </button>
        </div>
      </div>

      <!-- Account Mode Selection (Kimi / Zhipu / DeepSeek) -->
      <div v-if="isCNPlatform">
        <label class="input-label">{{ t('admin.accounts.cnProviders.accountMode.title') }}</label>
        <div class="components-account-create-account-modal__panel-9" data-tour="account-form-mode">
          <!-- Pay-as-you-go (token balance) -->
          <button
            type="button"
            @click="accountMode = 'payg'"
            :class="[
              'components-account-create-account-modal__action-24',
              accountMode === 'payg'
                ? cnAccentActiveClass
                : 'components-account-create-account-modal__action-37'
            ]"
          >
            <div
              :class="[
                'components-account-create-account-modal__panel-83',
                accountMode === 'payg'
                  ? cnAccentIconClass
                  : 'components-account-create-account-modal__panel-85'
              ]"
            >
              <Icon name="creditCard" size="sm" />
            </div>
            <div>
              <span class="components-account-create-account-modal__text-2">{{ t('admin.accounts.cnProviders.accountMode.payg') }}</span>
              <span class="components-account-create-account-modal__text-3">{{ t('admin.accounts.cnProviders.accountMode.paygDesc') }}</span>
            </div>
          </button>
          <!-- Coding Plan (kimi / zhipu only — DeepSeek has no coding plan) -->
          <button
            v-if="form.platform !== 'deepseek'"
            type="button"
            @click="accountMode = 'coding'"
            :class="[
              'components-account-create-account-modal__action-24',
              accountMode === 'coding'
                ? cnAccentActiveClass
                : 'components-account-create-account-modal__action-37'
            ]"
          >
            <div
              :class="[
                'components-account-create-account-modal__panel-83',
                accountMode === 'coding'
                  ? cnAccentIconClass
                  : 'components-account-create-account-modal__panel-85'
              ]"
            >
              <Icon name="bolt" size="sm" />
            </div>
            <div>
              <span class="components-account-create-account-modal__text-2">{{ t('admin.accounts.cnProviders.accountMode.coding') }}</span>
              <span class="components-account-create-account-modal__text-3">{{ t('admin.accounts.cnProviders.accountMode.codingDesc') }}</span>
            </div>
          </button>
        </div>
      </div>

      <!-- API Protocol Selection (Kimi / Zhipu / DeepSeek) -->
      <div v-if="isCNPlatform" class="components-account-create-account-modal__panel-10">
        <label class="input-label">{{ t('admin.accounts.cnProviders.apiProtocol.title') }}</label>
        <div class="components-account-create-account-modal__panel-11">
          <button
            v-for="opt in cnProtocolOptions"
            :key="opt.value"
            type="button"
            @click="apiProtocol = opt.value"
            :class="[
              'components-account-create-account-modal__action-24',
              apiProtocol === opt.value
                ? cnAccentActiveClass
                : 'components-account-create-account-modal__action-37'
            ]"
          >
            <div
              :class="[
                'components-account-create-account-modal__panel-83',
                apiProtocol === opt.value
                  ? cnAccentIconClass
                  : 'components-account-create-account-modal__panel-85'
              ]"
            >
              <Icon :name="opt.value === 'anthropic' ? 'sparkles' : opt.value === 'responses' ? 'terminal' : 'chat'" size="sm" />
            </div>
            <div>
              <span class="components-account-create-account-modal__text-2">{{ t(`admin.accounts.cnProviders.apiProtocol.${opt.labelKey}`) }}</span>
              <span class="components-account-create-account-modal__text-3">{{ t(`admin.accounts.cnProviders.apiProtocol.${opt.labelKey}Desc`) }}</span>
            </div>
          </button>
        </div>
      </div>

      <!-- Zhipu team Coding Plan context. Empty values keep the personal-plan endpoint. -->
      <div
        v-if="form.platform === 'zhipu' && accountMode === 'coding'"
        class="components-account-create-account-modal__panel-23"
      >
        <div>
          <label class="input-label">{{ t('admin.accounts.cnProviders.zhipuTeam.organization') }}</label>
          <input
            v-model="zhipuOrganization"
            type="text"
            class="input"
            :placeholder="t('admin.accounts.cnProviders.zhipuTeam.organizationPlaceholder')"
          />
        </div>
        <div>
          <label class="input-label">{{ t('admin.accounts.cnProviders.zhipuTeam.project') }}</label>
          <input
            v-model="zhipuProject"
            type="text"
            class="input"
            :placeholder="t('admin.accounts.cnProviders.zhipuTeam.projectPlaceholder')"
          />
        </div>
        <p class="input-hint">{{ t('admin.accounts.cnProviders.zhipuTeam.hint') }}</p>
      </div>

      <!-- Vertex Service Account -->
      <div v-if="(form.platform === 'anthropic' || form.platform === 'gemini') && accountCategory === 'service_account'" class="components-account-create-account-modal__panel-23">
        <div>
          <label class="input-label">Service Account JSON</label>
          <input
            ref="vertexServiceAccountFileInput"
            type="file"
            accept="application/json,.json"
            class="components-account-create-account-modal__field-2"
            @change="handleVertexServiceAccountFile"
          />
          <div
            :class="[
              'components-account-create-account-modal__panel-92',
              vertexServiceAccountDragActive
                ? 'components-account-create-account-modal__panel-93'
                : 'components-account-create-account-modal__panel-94'
            ]"
            @dragenter.prevent="vertexServiceAccountDragActive = true"
            @dragover.prevent="vertexServiceAccountDragActive = true"
            @dragleave.prevent="vertexServiceAccountDragActive = false"
            @drop.prevent="handleVertexServiceAccountDrop"
          >
            <div class="components-account-create-account-modal__panel-25">
              <div class="components-account-create-account-modal__panel-16">
                <div class="components-account-create-account-modal__panel-26">
                  <Icon name="upload" size="sm" />
                  <span>{{ vertexClientEmail ? t('admin.accounts.vertexSaJsonLoaded') : t('admin.accounts.vertexSaJsonDrop') }}</span>
                </div>
                <p class="components-account-create-account-modal__panel-18">
                  {{ vertexClientEmail ? t('admin.accounts.vertexSaJsonKeyHidden') : t('admin.accounts.vertexSaJsonDropHint') }}
                </p>
              </div>
              <button
                type="button"
                class="components-account-create-account-modal__action-3 btn btn-secondary"
                @click="vertexServiceAccountFileInput?.click()"
              >
                <Icon name="upload" size="sm" />
                {{ t('admin.accounts.vertexSaJsonSelectBtn') }}
              </button>
            </div>
            <div
              v-if="vertexClientEmail"
              class="components-account-create-account-modal__panel-27"
            >
              <div class="components-account-create-account-modal__panel-28">Project ID: <span class="components-account-create-account-modal__field">{{ vertexProjectId }}</span></div>
              <div class="components-account-create-account-modal__panel-28">Client Email: <span class="components-account-create-account-modal__field">{{ vertexClientEmail }}</span></div>
            </div>
          </div>
          <p class="input-hint">{{ t('admin.accounts.vertexSaJsonUploadHint') }}</p>
        </div>

        <div class="components-account-create-account-modal__panel-29">
          <div>
            <label class="input-label">Project ID</label>
            <input
              v-model="vertexProjectId"
              type="text"
              class="components-account-create-account-modal__field input"
              readonly
              :placeholder="t('admin.accounts.vertexProjectIdPlaceholder')"
            />
          </div>
          <div>
            <label class="input-label">Location</label>
            <Select
              v-model="vertexLocation"
              class="components-account-create-account-modal__field input"
              :options="VERTEX_LOCATION_SELECT_OPTIONS"
              searchable
            />
            <p class="input-hint">{{ t('admin.accounts.vertexLocationHint') }}</p>
          </div>
        </div>
      </div>

      <!-- Add Method (only for Anthropic OAuth-based type) -->
      <div v-if="form.platform === 'anthropic' && isOAuthFlow">
        <label class="input-label">{{ t('admin.accounts.addMethod') }}</label>
        <div class="components-account-create-account-modal__panel-36">
          <label class="components-account-create-account-modal__label-2">
            <input
              v-model="addMethod"
              type="radio"
              value="oauth"
              class="components-account-create-account-modal__field-3"
            />
            <span class="components-account-create-account-modal__text-9">{{ t('admin.accounts.types.oauth') }}</span>
          </label>
          <label class="components-account-create-account-modal__label-2">
            <input
              v-model="addMethod"
              type="radio"
              value="setup-token"
              class="components-account-create-account-modal__field-3"
            />
            <span class="components-account-create-account-modal__text-9">{{
              t('admin.accounts.setupTokenLongLived')
            }}</span>
          </label>
        </div>
      </div>

      <!-- API Key input -->
      <div v-if="form.type === 'apikey'" class="components-account-create-account-modal__panel-23">
        <div>
          <label class="input-label">{{ t('admin.accounts.baseUrl') }}</label>
          <input
            v-model="apiKeyBaseUrl"
            type="text"
            class="input"
            :placeholder="apiKeyBaseUrlPlaceholder"
          />
          <p v-if="baseUrlHint" class="input-hint">{{ baseUrlHint }}</p>
          <GrokBaseUrlPresets
            v-if="form.platform === 'grok'"
            class="components-account-create-account-modal__panel-22"
            @select="apiKeyBaseUrl = $event"
          />
          <CnBaseUrlPresets
            v-if="isCNPlatform"
            class="components-account-create-account-modal__panel-22"
            :platform="cnPresetPlatform"
            :mode="accountMode"
            :protocol="apiProtocol"
            :current-url="apiKeyBaseUrl"
            @select="onCnPresetSelect"
          />
        </div>
        <div>
          <label class="input-label">{{ t('admin.accounts.apiKeyRequired') }}</label>
          <input
            v-model="apiKeyValue"
            type="password"
            required
            class="components-account-create-account-modal__field input"
            :placeholder="apiKeyValuePlaceholder"
          />
          <p v-if="apiKeyHint" class="input-hint">{{ apiKeyHint }}</p>
        </div>

        <!-- 上游倍率自动探测：全部 API-key 平台可用（所在区块已限定 apikey 类型） -->
        <div
          class="components-account-create-account-modal__panel-24"
        >
          <div>
            <label class="components-account-create-account-modal__label input-label">{{ t('admin.accounts.upstreamBilling.autoProbe') }}</label>
            <p class="components-account-create-account-modal__panel-18">
              {{ t('admin.accounts.upstreamBilling.autoProbeHint') }}
            </p>
          </div>
          <Toggle
            v-model="upstreamBillingAutoProbeEnabled"
            data-testid="upstream-billing-auto-probe"
            :aria-label="t('admin.accounts.upstreamBilling.autoProbe')"
          />
        </div>

        <!-- Model Restriction Section -->
        <div class="components-account-create-account-modal__panel-30">
          <label class="input-label">{{ t('admin.accounts.modelRestriction') }}</label>
          <AccountModelRuleSelector
            :platform="form.platform"
            :disabled="isOpenAIModelRestrictionDisabled"
            :has-existing-mappings="hasModelRestrictionValues"
            @apply="applyAccountModelRule"
          />

          <div
            v-if="isOpenAIModelRestrictionDisabled"
            class="components-account-create-account-modal__panel-37"
          >
            <p class="components-account-create-account-modal__description-3">
              {{ t('admin.accounts.openai.modelRestrictionDisabledByPassthrough') }}
            </p>
          </div>

          <template v-else>
            <!-- Mode Toggle -->
            <div class="components-account-create-account-modal__panel-38">
              <button
                type="button"
                @click="modelRestrictionMode = 'whitelist'"
                :class="[
                  'components-account-create-account-modal__action-43',
                  modelRestrictionMode === 'whitelist'
                    ? 'components-account-create-account-modal__action-44'
                    : 'components-account-create-account-modal__action-45'
                ]"
              >
                <svg
                  class="components-account-create-account-modal__icon-4"
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
                  'components-account-create-account-modal__action-43',
                  modelRestrictionMode === 'mapping'
                    ? 'components-account-create-account-modal__action-46'
                    : 'components-account-create-account-modal__action-45'
                ]"
              >
                <svg
                  class="components-account-create-account-modal__icon-4"
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
              <ModelWhitelistSelector v-model="allowedModels" :platform="form.platform" :sync-credentials="syncPreviewCredentials" @upstream-synced="upstreamModelsPreviewed = true" />
              <p class="components-account-create-account-modal__text-3">
                {{ t('admin.accounts.selectedModels', { count: allowedModels.length }) }}
                <span v-if="allowedModels.length === 0">{{
                  t('admin.accounts.supportsAllModels')
                }}</span>
              </p>
            </div>

            <!-- Mapping Mode -->
            <div v-else>
              <div class="components-account-create-account-modal__panel-31">
                <p class="components-account-create-account-modal__description">
                  <svg
                    class="components-account-create-account-modal__icon-3"
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
            <div v-if="modelMappings.length > 0" class="components-account-create-account-modal__panel-32">
              <div
                v-for="(mapping, index) in modelMappings"
                :key="getModelMappingKey(mapping)"
                class="components-account-create-account-modal__panel-34"
              >
                <input
                  v-model="mapping.from"
                  type="text"
                  class="components-account-create-account-modal__field-4 input"
                  :placeholder="t('admin.accounts.requestModel')"
                />
                <svg
                  class="components-account-create-account-modal__icon-2"
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
                  class="components-account-create-account-modal__field-4 input"
                  :placeholder="t('admin.accounts.actualModel')"
                />
                <button
                  type="button"
                  @click="removeModelMapping(index)"
                  class="components-account-create-account-modal__action-4"
                >
                  <svg class="components-account-create-account-modal__icon" fill="none" viewBox="0 0 24 24" stroke="currentColor">
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
              class="components-account-create-account-modal__action-5"
            >
              <svg
                class="components-account-create-account-modal__icon-3"
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
              <div class="components-account-create-account-modal__panel-35">
                <button
                  v-for="preset in presetMappings"
                  :key="preset.label"
                  type="button"
                  @click="addPresetMapping(preset.from, preset.to)"
                  :class="['components-account-create-account-modal__action-42', preset.color]"
                >
                  + {{ preset.label }}
                </button>
              </div>
            </div>
          </template>
        </div>

        <!-- Pool Mode Section -->
        <div class="components-account-create-account-modal__panel-30">
          <div class="components-account-create-account-modal__panel-39">
            <div>
              <label class="components-account-create-account-modal__label input-label">{{ t('admin.accounts.poolMode') }}</label>
              <p class="components-account-create-account-modal__panel-18">
                {{ t('admin.accounts.poolModeHint') }}
              </p>
            </div>
            <button
              type="button"
              @click="poolModeEnabled = !poolModeEnabled"
              :class="[
                'components-account-create-account-modal__action-47',
                poolModeEnabled ? 'components-account-create-account-modal__action-48' : 'components-account-create-account-modal__action-49'
              ]"
            >
              <span
                :class="[
                  'components-account-create-account-modal__text-19',
                  poolModeEnabled ? 'toggle-thumb--on' : 'components-account-create-account-modal__text-20'
                ]"
              />
            </button>
          </div>
          <div v-if="poolModeEnabled" class="components-account-create-account-modal__panel-40">
            <p class="components-account-create-account-modal__description-4">
              <Icon name="exclamationCircle" size="sm" class="components-account-create-account-modal__icon-5" :stroke-width="2" />
              {{ t('admin.accounts.poolModeInfo') }}
            </p>
          </div>
          <div v-if="poolModeEnabled" class="components-account-create-account-modal__panel-19">
            <label class="input-label">{{ t('admin.accounts.poolModeRetryCount') }}</label>
            <input
              v-model.number="poolModeRetryCount"
              type="number"
              min="0"
              :max="MAX_POOL_MODE_RETRY_COUNT"
              step="1"
              class="input"
            />
            <p class="components-account-create-account-modal__panel-18">
              {{
                t('admin.accounts.poolModeRetryCountHint', {
                  default: DEFAULT_POOL_MODE_RETRY_COUNT,
                  max: MAX_POOL_MODE_RETRY_COUNT
                })
              }}
            </p>
          </div>
          <div v-if="poolModeEnabled" class="components-account-create-account-modal__panel-19">
            <label class="input-label">{{ t('admin.accounts.poolModeRetryStatusCodes') }}</label>
            <input
              v-model="poolModeRetryStatusCodesInput"
              type="text"
              class="input"
              :placeholder="DEFAULT_POOL_MODE_RETRY_STATUS_CODES.join(', ')"
            />
            <p class="components-account-create-account-modal__panel-18">
              {{ t('admin.accounts.poolModeRetryStatusCodesHint', { default: DEFAULT_POOL_MODE_RETRY_STATUS_CODES.join(', ') }) }}
            </p>
          </div>
        </div>

        <!-- Custom Error Codes Section -->
        <div class="components-account-create-account-modal__panel-30">
          <div class="components-account-create-account-modal__panel-39">
            <div>
              <label class="components-account-create-account-modal__label input-label">{{ t('admin.accounts.customErrorCodes') }}</label>
              <p class="components-account-create-account-modal__panel-18">
                {{ t('admin.accounts.customErrorCodesHint') }}
              </p>
            </div>
            <button
              type="button"
              @click="customErrorCodesEnabled = !customErrorCodesEnabled"
              :class="[
                'components-account-create-account-modal__action-47',
                customErrorCodesEnabled ? 'components-account-create-account-modal__action-48' : 'components-account-create-account-modal__action-49'
              ]"
            >
              <span
                :class="[
                  'components-account-create-account-modal__text-19',
                  customErrorCodesEnabled ? 'toggle-thumb--on' : 'components-account-create-account-modal__text-20'
                ]"
              />
            </button>
          </div>

          <div v-if="customErrorCodesEnabled" class="components-account-create-account-modal__panel-41">
            <div class="components-account-create-account-modal__panel-42">
              <p class="components-account-create-account-modal__description-3">
                <Icon name="exclamationTriangle" size="sm" class="components-account-create-account-modal__icon-5" :stroke-width="2" />
                {{ t('admin.accounts.customErrorCodesWarning') }}
              </p>
            </div>

            <!-- Error Code Buttons -->
            <div class="components-account-create-account-modal__panel-35">
              <button
                v-for="code in commonErrorCodes"
                :key="code.value"
                type="button"
                @click="toggleErrorCode(code.value)"
                :class="[
                  'components-account-create-account-modal__action-50',
                  selectedErrorCodes.includes(code.value)
                    ? 'components-account-create-account-modal__action-51'
                    : 'components-account-create-account-modal__action-45'
                ]"
              >
                {{ code.value }} {{ code.label }}
              </button>
            </div>

            <!-- Manual input -->
            <div class="components-account-create-account-modal__panel-34">
              <input
                v-model.number="customErrorCodeInput"
                type="number"
                min="100"
                max="599"
                class="components-account-create-account-modal__field-4 input"
                :placeholder="t('admin.accounts.enterErrorCode')"
                @keyup.enter="addCustomErrorCode"
              />
              <button type="button" @click="addCustomErrorCode" class="components-account-create-account-modal__action-6 btn btn-secondary">
                <svg class="components-account-create-account-modal__icon" fill="none" viewBox="0 0 24 24" stroke="currentColor">
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
            <div class="components-account-create-account-modal__panel-43">
              <span
                v-for="code in selectedErrorCodes.sort((a, b) => a - b)"
                :key="code"
                class="components-account-create-account-modal__text-10"
              >
                {{ code }}
                <button
                  type="button"
                  @click="removeErrorCode(code)"
                  class="components-account-create-account-modal__action-7"
                >
                  <Icon name="x" size="sm" :stroke-width="2" />
                </button>
              </span>
              <span v-if="selectedErrorCodes.length === 0" class="components-account-create-account-modal__text-11">
                {{ t('admin.accounts.noneSelectedUsesDefault') }}
              </span>
            </div>
          </div>
        </div>

        <!-- Header Override Section (anthropic/openai apikey only) -->
        <div
          v-if="isHeaderOverrideCapable(form.platform, 'apikey')"
          class="components-account-create-account-modal__panel-30"
        >
          <div class="components-account-create-account-modal__panel-39">
            <div>
              <label class="components-account-create-account-modal__label input-label">{{ t('admin.accounts.headerOverride.title') }}</label>
              <p class="components-account-create-account-modal__panel-18">
                {{ t('admin.accounts.headerOverride.hint') }}
              </p>
            </div>
            <button
              type="button"
              @click="headerOverrideEnabled = !headerOverrideEnabled"
              :class="[
                'components-account-create-account-modal__action-47',
                headerOverrideEnabled ? 'components-account-create-account-modal__action-48' : 'components-account-create-account-modal__action-49'
              ]"
            >
              <span
                :class="[
                  'components-account-create-account-modal__text-19',
                  headerOverrideEnabled ? 'toggle-thumb--on' : 'components-account-create-account-modal__text-20'
                ]"
              />
            </button>
          </div>

          <div v-if="headerOverrideEnabled" class="components-account-create-account-modal__panel-41">
            <div class="components-account-create-account-modal__panel-40">
              <p class="components-account-create-account-modal__description-4">
                <Icon name="exclamationCircle" size="sm" class="components-account-create-account-modal__icon-5" :stroke-width="2" />
                {{ t('admin.accounts.headerOverride.info') }}
              </p>
            </div>

            <HeaderOverrideEditor
              :rows="headerOverrideRows"
              @update:rows="headerOverrideRows = $event"
            />
          </div>
        </div>

      </div>

      <!-- Bedrock credentials (only for Anthropic Bedrock type) -->
      <div v-if="form.platform === 'anthropic' && accountCategory === 'bedrock'" class="components-account-create-account-modal__panel-23">
        <!-- Auth Mode Radio -->
        <div>
          <label class="input-label">{{ t('admin.accounts.bedrockAuthMode') }}</label>
          <div class="components-account-create-account-modal__panel-36">
            <label class="components-account-create-account-modal__label-2">
              <input
                v-model="bedrockAuthMode"
                type="radio"
                value="sigv4"
                class="components-account-create-account-modal__field-3"
              />
              <span class="components-account-create-account-modal__text-9">{{ t('admin.accounts.bedrockAuthModeSigv4') }}</span>
            </label>
            <label class="components-account-create-account-modal__label-2">
              <input
                v-model="bedrockAuthMode"
                type="radio"
                value="apikey"
                class="components-account-create-account-modal__field-3"
              />
              <span class="components-account-create-account-modal__text-9">{{ t('admin.accounts.bedrockAuthModeApikey') }}</span>
            </label>
          </div>
        </div>

        <!-- SigV4 fields -->
        <template v-if="bedrockAuthMode === 'sigv4'">
          <div>
            <label class="input-label">{{ t('admin.accounts.bedrockAccessKeyId') }}</label>
            <input
              v-model="bedrockAccessKeyId"
              type="text"
              required
              class="components-account-create-account-modal__field input"
              placeholder="AKIA..."
            />
          </div>
          <div>
            <label class="input-label">{{ t('admin.accounts.bedrockSecretAccessKey') }}</label>
            <input
              v-model="bedrockSecretAccessKey"
              type="password"
              required
              class="components-account-create-account-modal__field input"
            />
          </div>
          <div>
            <label class="input-label">{{ t('admin.accounts.bedrockSessionToken') }}</label>
            <input
              v-model="bedrockSessionToken"
              type="password"
              class="components-account-create-account-modal__field input"
            />
            <p class="input-hint">{{ t('admin.accounts.bedrockSessionTokenHint') }}</p>
          </div>
        </template>

        <!-- API Key field -->
        <div v-if="bedrockAuthMode === 'apikey'">
          <label class="input-label">{{ t('admin.accounts.bedrockApiKeyInput') }}</label>
          <input
            v-model="bedrockApiKeyValue"
            type="password"
            required
            class="components-account-create-account-modal__field input"
          />
        </div>

        <!-- Shared: Region -->
        <div>
          <label class="input-label">{{ t('admin.accounts.bedrockRegion') }}</label>
          <Select v-model="bedrockRegion" :options="BEDROCK_REGION_OPTIONS" searchable />
          <p class="input-hint">{{ t('admin.accounts.bedrockRegionHint') }}</p>
        </div>

        <!-- Shared: Force Global -->
        <div>
          <label class="components-account-create-account-modal__label-3">
            <input
              v-model="bedrockForceGlobal"
              type="checkbox"
              class="components-account-create-account-modal__field-5"
            />
            <span class="components-account-create-account-modal__text-9">{{ t('admin.accounts.bedrockForceGlobal') }}</span>
          </label>
          <p class="components-account-create-account-modal__description-5 input-hint">{{ t('admin.accounts.bedrockForceGlobalHint') }}</p>
        </div>

        <!-- Model Restriction Section for Bedrock -->
        <div class="components-account-create-account-modal__panel-30">
          <label class="input-label">{{ t('admin.accounts.modelRestriction') }}</label>
          <AccountModelRuleSelector
            platform="anthropic"
            :has-existing-mappings="hasModelRestrictionValues"
            @apply="applyAccountModelRule"
          />

          <!-- Mode Toggle -->
          <div class="components-account-create-account-modal__panel-38">
            <button
              type="button"
              @click="modelRestrictionMode = 'whitelist'"
              :class="[
                'components-account-create-account-modal__action-43',
                modelRestrictionMode === 'whitelist'
                  ? 'components-account-create-account-modal__action-44'
                  : 'components-account-create-account-modal__action-45'
              ]"
            >
              {{ t('admin.accounts.modelWhitelist') }}
            </button>
            <button
              type="button"
              @click="modelRestrictionMode = 'mapping'"
              :class="[
                'components-account-create-account-modal__action-43',
                modelRestrictionMode === 'mapping'
                  ? 'components-account-create-account-modal__action-46'
                  : 'components-account-create-account-modal__action-45'
              ]"
            >
              {{ t('admin.accounts.modelMapping') }}
            </button>
          </div>

          <!-- Whitelist Mode -->
          <div v-if="modelRestrictionMode === 'whitelist'">
            <ModelWhitelistSelector v-model="allowedModels" platform="anthropic" :sync-credentials="syncPreviewCredentials" @upstream-synced="upstreamModelsPreviewed = true" />
            <p class="components-account-create-account-modal__text-3">
              {{ t('admin.accounts.selectedModels', { count: allowedModels.length }) }}
              <span v-if="allowedModels.length === 0">{{ t('admin.accounts.supportsAllModels') }}</span>
            </p>
          </div>

          <!-- Mapping Mode -->
          <div v-else class="components-account-create-account-modal__panel-41">
            <div v-for="(mapping, index) in modelMappings" :key="index" class="components-account-create-account-modal__panel-34">
              <input v-model="mapping.from" type="text" class="components-account-create-account-modal__field-4 input" :placeholder="t('admin.accounts.fromModel')" />
              <span class="components-account-create-account-modal__text-12">→</span>
              <input v-model="mapping.to" type="text" class="components-account-create-account-modal__field-4 input" :placeholder="t('admin.accounts.toModel')" />
              <button type="button" @click="modelMappings.splice(index, 1)" class="components-account-create-account-modal__action-8">
                <Icon name="trash" size="sm" />
              </button>
            </div>
            <button type="button" @click="modelMappings.push({ from: '', to: '' })" class="components-account-create-account-modal__action-9 btn btn-secondary">
              + {{ t('admin.accounts.addMapping') }}
            </button>
            <!-- Bedrock Preset Mappings -->
            <div class="components-account-create-account-modal__panel-35">
              <button
                v-for="preset in bedrockPresets"
                :key="preset.from"
                type="button"
                @click="addPresetMapping(preset.from, preset.to)"
                :class="['components-account-create-account-modal__action-42', preset.color]"
              >
                + {{ preset.label }}
              </button>
            </div>
          </div>
        </div>

        <!-- Pool Mode Section for Bedrock -->
        <div class="components-account-create-account-modal__panel-30">
          <div class="components-account-create-account-modal__panel-39">
            <div>
              <label class="components-account-create-account-modal__label input-label">{{ t('admin.accounts.poolMode') }}</label>
              <p class="components-account-create-account-modal__panel-18">
                {{ t('admin.accounts.poolModeHint') }}
              </p>
            </div>
            <button
              type="button"
              @click="poolModeEnabled = !poolModeEnabled"
              :class="[
                'components-account-create-account-modal__action-47',
                poolModeEnabled ? 'components-account-create-account-modal__action-48' : 'components-account-create-account-modal__action-49'
              ]"
            >
              <span
                :class="[
                  'components-account-create-account-modal__text-19',
                  poolModeEnabled ? 'toggle-thumb--on' : 'components-account-create-account-modal__text-20'
                ]"
              />
            </button>
          </div>
          <div v-if="poolModeEnabled" class="components-account-create-account-modal__panel-40">
            <p class="components-account-create-account-modal__description-4">
              <Icon name="exclamationCircle" size="sm" class="components-account-create-account-modal__icon-5" :stroke-width="2" />
              {{ t('admin.accounts.poolModeInfo') }}
            </p>
          </div>
          <div v-if="poolModeEnabled" class="components-account-create-account-modal__panel-19">
            <label class="input-label">{{ t('admin.accounts.poolModeRetryCount') }}</label>
            <input
              v-model.number="poolModeRetryCount"
              type="number"
              min="0"
              :max="MAX_POOL_MODE_RETRY_COUNT"
              step="1"
              class="input"
            />
            <p class="components-account-create-account-modal__panel-18">
              {{
                t('admin.accounts.poolModeRetryCountHint', {
                  default: DEFAULT_POOL_MODE_RETRY_COUNT,
                  max: MAX_POOL_MODE_RETRY_COUNT
                })
              }}
            </p>
          </div>
          <div v-if="poolModeEnabled" class="components-account-create-account-modal__panel-19">
            <label class="input-label">{{ t('admin.accounts.poolModeRetryStatusCodes') }}</label>
            <input
              v-model="poolModeRetryStatusCodesInput"
              type="text"
              class="input"
              :placeholder="DEFAULT_POOL_MODE_RETRY_STATUS_CODES.join(', ')"
            />
            <p class="components-account-create-account-modal__panel-18">
              {{ t('admin.accounts.poolModeRetryStatusCodesHint', { default: DEFAULT_POOL_MODE_RETRY_STATUS_CODES.join(', ') }) }}
            </p>
          </div>
        </div>
      </div>

      <!-- 配额控制 (Anthropic apikey/bedrock: 配额限制 + 亲和) -->
      <div
        v-if="form.platform === 'anthropic' && (form.type === 'apikey' || form.type === 'bedrock')"
        class="components-account-create-account-modal__panel-44"
      >
        <div class="components-account-create-account-modal__panel-45">
          <h3 class="components-account-create-account-modal__heading input-label">{{ t('admin.accounts.quotaControl.title') }}</h3>
          <p class="components-account-create-account-modal__panel-18">
            {{ t('admin.accounts.quotaControl.hint') }}
          </p>
        </div>
        <QuotaLimitCard
          :totalLimit="editQuotaLimit"
          :dailyLimit="editQuotaDailyLimit"
          :weeklyLimit="editQuotaWeeklyLimit"
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
          :dailyResetMode="editDailyResetMode"
          :dailyResetHour="editDailyResetHour"
          :weeklyResetMode="editWeeklyResetMode"
          :weeklyResetDay="editWeeklyResetDay"
          :weeklyResetHour="editWeeklyResetHour"
          :resetTimezone="editResetTimezone"
          @update:totalLimit="editQuotaLimit = $event"
          @update:dailyLimit="editQuotaDailyLimit = $event"
          @update:weeklyLimit="editQuotaWeeklyLimit = $event"
          @update:quotaNotifyDailyEnabled="quotaNotifyState.daily.enabled = $event"
          @update:quotaNotifyDailyThreshold="quotaNotifyState.daily.threshold = $event"
          @update:quotaNotifyDailyThresholdType="quotaNotifyState.daily.thresholdType = $event"
          @update:quotaNotifyWeeklyEnabled="quotaNotifyState.weekly.enabled = $event"
          @update:quotaNotifyWeeklyThreshold="quotaNotifyState.weekly.threshold = $event"
          @update:quotaNotifyWeeklyThresholdType="quotaNotifyState.weekly.thresholdType = $event"
          @update:quotaNotifyTotalEnabled="quotaNotifyState.total.enabled = $event"
          @update:quotaNotifyTotalThreshold="quotaNotifyState.total.threshold = $event"
          @update:quotaNotifyTotalThresholdType="quotaNotifyState.total.thresholdType = $event"
          @update:dailyResetMode="editDailyResetMode = $event"
          @update:dailyResetHour="editDailyResetHour = $event"
          @update:weeklyResetMode="editWeeklyResetMode = $event"
          @update:weeklyResetDay="editWeeklyResetDay = $event"
          @update:weeklyResetHour="editWeeklyResetHour = $event"
          @update:resetTimezone="editResetTimezone = $event"
        />
      </div>

      <!-- 配额控制 (非 Anthropic apikey/bedrock) -->
      <div
        v-else-if="form.type === 'apikey' || form.type === 'bedrock'"
        class="components-account-create-account-modal__panel-44"
      >
        <div class="components-account-create-account-modal__panel-45">
          <h3 class="components-account-create-account-modal__heading input-label">{{ t('admin.accounts.quotaControl.title') }}</h3>
          <p class="components-account-create-account-modal__panel-18">
            {{ t('admin.accounts.quotaLimitHint') }}
          </p>
        </div>
        <QuotaLimitCard
          :totalLimit="editQuotaLimit"
          :dailyLimit="editQuotaDailyLimit"
          :weeklyLimit="editQuotaWeeklyLimit"
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
          :dailyResetMode="editDailyResetMode"
          :dailyResetHour="editDailyResetHour"
          :weeklyResetMode="editWeeklyResetMode"
          :weeklyResetDay="editWeeklyResetDay"
          :weeklyResetHour="editWeeklyResetHour"
          :resetTimezone="editResetTimezone"
          @update:totalLimit="editQuotaLimit = $event"
          @update:dailyLimit="editQuotaDailyLimit = $event"
          @update:weeklyLimit="editQuotaWeeklyLimit = $event"
          @update:quotaNotifyDailyEnabled="quotaNotifyState.daily.enabled = $event"
          @update:quotaNotifyDailyThreshold="quotaNotifyState.daily.threshold = $event"
          @update:quotaNotifyDailyThresholdType="quotaNotifyState.daily.thresholdType = $event"
          @update:quotaNotifyWeeklyEnabled="quotaNotifyState.weekly.enabled = $event"
          @update:quotaNotifyWeeklyThreshold="quotaNotifyState.weekly.threshold = $event"
          @update:quotaNotifyWeeklyThresholdType="quotaNotifyState.weekly.thresholdType = $event"
          @update:quotaNotifyTotalEnabled="quotaNotifyState.total.enabled = $event"
          @update:quotaNotifyTotalThreshold="quotaNotifyState.total.threshold = $event"
          @update:quotaNotifyTotalThresholdType="quotaNotifyState.total.thresholdType = $event"
          @update:dailyResetMode="editDailyResetMode = $event"
          @update:dailyResetHour="editDailyResetHour = $event"
          @update:weeklyResetMode="editWeeklyResetMode = $event"
          @update:weeklyResetDay="editWeeklyResetDay = $event"
          @update:weeklyResetHour="editWeeklyResetHour = $event"
          @update:resetTimezone="editResetTimezone = $event"
        />
      </div>

      <!-- Grok OAuth Custom Upstream URL (仅改写转发端点，OAuth 授权/刷新不受影响) -->
      <div
        v-if="form.platform === 'grok' && isOAuthFlow"
        class="components-account-create-account-modal__panel-30"
      >
        <div class="components-account-create-account-modal__panel-39">
          <div>
            <label class="components-account-create-account-modal__label input-label">{{ t('admin.accounts.grokCustomBaseUrl.title') }}</label>
            <p class="components-account-create-account-modal__panel-18">
              {{ t('admin.accounts.grokCustomBaseUrl.hint') }}
            </p>
          </div>
          <button
            type="button"
            data-testid="grok-custom-base-url-toggle"
            @click="grokOAuthCustomBaseUrlEnabled = !grokOAuthCustomBaseUrlEnabled"
            :class="[
              'components-account-create-account-modal__action-47',
              grokOAuthCustomBaseUrlEnabled ? 'components-account-create-account-modal__action-48' : 'components-account-create-account-modal__action-49'
            ]"
          >
            <span
              :class="[
                'components-account-create-account-modal__text-19',
                grokOAuthCustomBaseUrlEnabled ? 'toggle-thumb--on' : 'components-account-create-account-modal__text-20'
              ]"
            />
          </button>
        </div>
        <div v-if="grokOAuthCustomBaseUrlEnabled" class="components-account-create-account-modal__panel-46">
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

      <!-- Grok OAuth Header Override (OAuth 类型没有 apikey 容器，需要独立区域) -->
      <div
        v-if="form.platform === 'grok' && isOAuthFlow"
        class="components-account-create-account-modal__panel-30"
      >
        <div class="components-account-create-account-modal__panel-39">
          <div>
            <label class="components-account-create-account-modal__label input-label">{{ t('admin.accounts.headerOverride.title') }}</label>
            <p class="components-account-create-account-modal__panel-18">
              {{ t('admin.accounts.headerOverride.hint') }}
            </p>
          </div>
          <button
            type="button"
            @click="headerOverrideEnabled = !headerOverrideEnabled"
            :class="[
              'components-account-create-account-modal__action-47',
              headerOverrideEnabled ? 'components-account-create-account-modal__action-48' : 'components-account-create-account-modal__action-49'
            ]"
          >
            <span
              :class="[
                'components-account-create-account-modal__text-19',
                headerOverrideEnabled ? 'toggle-thumb--on' : 'components-account-create-account-modal__text-20'
              ]"
            />
          </button>
        </div>

        <div v-if="headerOverrideEnabled" class="components-account-create-account-modal__panel-41">
          <div class="components-account-create-account-modal__panel-40">
            <p class="components-account-create-account-modal__description-4">
              <Icon name="exclamationCircle" size="sm" class="components-account-create-account-modal__icon-5" :stroke-width="2" />
              {{ t('admin.accounts.headerOverride.info') }}
            </p>
          </div>

          <HeaderOverrideEditor
            :rows="headerOverrideRows"
            @update:rows="headerOverrideRows = $event"
          />
        </div>
      </div>

      <!-- OpenAI OAuth Model Mapping (OAuth 类型没有 apikey 容器，需要独立的模型映射区域) -->
      <div
        v-if="(form.platform === 'openai' || form.platform === 'gemini' || form.platform === 'antigravity' || form.platform === 'grok') && isOAuthFlow"
        class="components-account-create-account-modal__panel-30"
      >
        <label class="input-label">{{ t('admin.accounts.modelRestriction') }}</label>
        <AccountModelRuleSelector
          :platform="form.platform"
          :disabled="isOpenAIModelRestrictionDisabled"
          :has-existing-mappings="hasModelRestrictionValues"
          @apply="applyAccountModelRule"
        />

        <div
          v-if="isOpenAIModelRestrictionDisabled"
          class="components-account-create-account-modal__panel-37"
        >
          <p class="components-account-create-account-modal__description-3">
            {{ t('admin.accounts.openai.modelRestrictionDisabledByPassthrough') }}
          </p>
        </div>

        <template v-else>
          <!-- Mode Toggle -->
          <div class="components-account-create-account-modal__panel-38">
            <button
              type="button"
              @click="modelRestrictionMode = 'whitelist'"
              :class="[
                'components-account-create-account-modal__action-43',
                modelRestrictionMode === 'whitelist'
                  ? 'components-account-create-account-modal__action-44'
                  : 'components-account-create-account-modal__action-45'
              ]"
            >
              {{ t('admin.accounts.modelWhitelist') }}
            </button>
            <button
              type="button"
              @click="modelRestrictionMode = 'mapping'"
              :class="[
                'components-account-create-account-modal__action-43',
                modelRestrictionMode === 'mapping'
                  ? 'components-account-create-account-modal__action-46'
                  : 'components-account-create-account-modal__action-45'
              ]"
            >
              {{ t('admin.accounts.modelMapping') }}
            </button>
          </div>

          <!-- Whitelist Mode -->
          <div v-if="modelRestrictionMode === 'whitelist'">
            <ModelWhitelistSelector v-model="allowedModels" :platform="form.platform" :sync-credentials="syncPreviewCredentials" @upstream-synced="upstreamModelsPreviewed = true" />
            <p class="components-account-create-account-modal__text-3">
              {{ t('admin.accounts.selectedModels', { count: allowedModels.length }) }}
              <span v-if="allowedModels.length === 0">{{
                t('admin.accounts.supportsAllModels')
              }}</span>
            </p>
          </div>

          <!-- Mapping Mode -->
          <div v-else>
            <div class="components-account-create-account-modal__panel-31">
              <p class="components-account-create-account-modal__description">
                {{ t('admin.accounts.mapRequestModels') }}
              </p>
            </div>

            <div v-if="modelMappings.length > 0" class="components-account-create-account-modal__panel-32">
              <div
                v-for="(mapping, index) in modelMappings"
                :key="'oauth-' + getModelMappingKey(mapping)"
                class="components-account-create-account-modal__panel-34"
              >
                <input
                  v-model="mapping.from"
                  type="text"
                  class="components-account-create-account-modal__field-4 input"
                  :placeholder="t('admin.accounts.requestModel')"
                />
                <svg
                  class="components-account-create-account-modal__icon-2"
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
                  class="components-account-create-account-modal__field-4 input"
                  :placeholder="t('admin.accounts.actualModel')"
                />
                <button
                  type="button"
                  @click="removeModelMapping(index)"
                  class="components-account-create-account-modal__action-4"
                >
                  <svg class="components-account-create-account-modal__icon" fill="none" viewBox="0 0 24 24" stroke="currentColor">
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
              class="components-account-create-account-modal__action-5"
            >
              + {{ t('admin.accounts.addMapping') }}
            </button>

            <!-- Quick Add Buttons -->
            <div class="components-account-create-account-modal__panel-35">
              <button
                v-for="preset in presetMappings"
                :key="'oauth-' + preset.label"
                type="button"
                @click="addPresetMapping(preset.from, preset.to)"
                :class="['components-account-create-account-modal__action-42', preset.color]"
              >
                + {{ preset.label }}
              </button>
            </div>
          </div>
        </template>
      </div>

      <!-- Temp Unschedulable Rules -->
      <div class="components-account-create-account-modal__panel-44">
        <div class="components-account-create-account-modal__panel-39">
          <div>
            <label class="components-account-create-account-modal__label input-label">{{ t('admin.accounts.tempUnschedulable.title') }}</label>
            <p class="components-account-create-account-modal__panel-18">
              {{ t('admin.accounts.tempUnschedulable.hint') }}
            </p>
          </div>
          <button
            type="button"
            @click="tempUnschedEnabled = !tempUnschedEnabled"
            :class="[
              'components-account-create-account-modal__action-47',
              tempUnschedEnabled ? 'components-account-create-account-modal__action-48' : 'components-account-create-account-modal__action-49'
            ]"
          >
            <span
              :class="[
                'components-account-create-account-modal__text-19',
                tempUnschedEnabled ? 'toggle-thumb--on' : 'components-account-create-account-modal__text-20'
              ]"
            />
          </button>
        </div>

        <div v-if="tempUnschedEnabled" class="components-account-create-account-modal__panel-41">
          <div class="components-account-create-account-modal__panel-40">
              <p class="components-account-create-account-modal__description-4">
                <Icon name="exclamationTriangle" size="sm" class="components-account-create-account-modal__icon-5" :stroke-width="2" />
                {{ t('admin.accounts.tempUnschedulable.notice') }}
              </p>
            </div>

          <div class="components-account-create-account-modal__panel-35">
            <button
              v-for="preset in tempUnschedPresets"
              :key="preset.label"
              type="button"
              @click="addTempUnschedRule(preset.rule)"
              class="components-account-create-account-modal__action-10"
            >
              + {{ preset.label }}
            </button>
          </div>

          <div v-if="tempUnschedRules.length > 0" class="components-account-create-account-modal__panel-41">
            <div
              v-for="(rule, index) in tempUnschedRules"
              :key="getTempUnschedRuleKey(rule)"
              class="components-account-create-account-modal__panel-47"
            >
              <div class="components-account-create-account-modal__panel-48">
                <span class="components-account-create-account-modal__text-13">
                  {{ t('admin.accounts.tempUnschedulable.ruleIndex', { index: index + 1 }) }}
                </span>
                <div class="components-account-create-account-modal__panel-34">
                  <button
                    type="button"
                    :disabled="index === 0"
                    @click="moveTempUnschedRule(index, -1)"
                    class="components-account-create-account-modal__action-11"
                  >
                    <Icon name="chevronUp" size="sm" :stroke-width="2" />
                  </button>
                  <button
                    type="button"
                    :disabled="index === tempUnschedRules.length - 1"
                    @click="moveTempUnschedRule(index, 1)"
                    class="components-account-create-account-modal__action-11"
                  >
                    <svg class="components-account-create-account-modal__icon" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
                    </svg>
                  </button>
                  <button
                    type="button"
                    @click="removeTempUnschedRule(index)"
                    class="components-account-create-account-modal__action-12"
                  >
                    <Icon name="x" size="sm" :stroke-width="2" />
                  </button>
                </div>
              </div>

              <div class="components-account-create-account-modal__panel-49">
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
                <div class="components-account-create-account-modal__panel-50">
                  <label class="input-label">{{ t('admin.accounts.tempUnschedulable.keywords') }}</label>
                  <input
                    v-model="rule.keywords"
                    type="text"
                    class="input"
                    :placeholder="t('admin.accounts.tempUnschedulable.keywordsPlaceholder')"
                  />
                  <p class="input-hint">{{ t('admin.accounts.tempUnschedulable.keywordsHint') }}</p>
                </div>
                <div class="components-account-create-account-modal__panel-50">
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
            class="components-account-create-account-modal__action-13"
          >
            <svg
              class="components-account-create-account-modal__icon-3"
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

      <!-- Intercept Warmup Requests -->
      <div
        v-if="form.platform === 'anthropic'"
        class="components-account-create-account-modal__panel-30"
      >
        <div class="components-account-create-account-modal__panel-12">
          <div>
            <label class="components-account-create-account-modal__label input-label">{{
              t('admin.accounts.interceptWarmupRequests')
            }}</label>
            <p class="components-account-create-account-modal__panel-18">
              {{ t('admin.accounts.interceptWarmupRequestsDesc') }}
            </p>
          </div>
          <button
            type="button"
            @click="interceptWarmupRequests = !interceptWarmupRequests"
            :class="[
              'components-account-create-account-modal__action-47',
              interceptWarmupRequests ? 'components-account-create-account-modal__action-48' : 'components-account-create-account-modal__action-49'
            ]"
          >
            <span
              :class="[
                'components-account-create-account-modal__text-19',
                interceptWarmupRequests ? 'toggle-thumb--on' : 'components-account-create-account-modal__text-20'
              ]"
            />
          </button>
        </div>
      </div>

      <!-- 配额控制 (Anthropic OAuth/SetupToken: 亲和 + 窗口费用 + 会话 + RPM 等) -->
      <div
        v-if="form.platform === 'anthropic' && accountCategory === 'oauth-based'"
        class="components-account-create-account-modal__panel-44"
      >
        <div class="components-account-create-account-modal__panel-45">
          <h3 class="components-account-create-account-modal__heading input-label">{{ t('admin.accounts.quotaControl.title') }}</h3>
          <p class="components-account-create-account-modal__panel-18">
            {{ t('admin.accounts.quotaControl.hint') }}
          </p>
        </div>

        <!-- Window Cost Limit -->
        <div class="components-account-create-account-modal__panel-51">
          <div class="components-account-create-account-modal__panel-39">
            <div>
              <label class="components-account-create-account-modal__label input-label">{{ t('admin.accounts.quotaControl.windowCost.label') }}</label>
              <p class="components-account-create-account-modal__panel-18">
                {{ t('admin.accounts.quotaControl.windowCost.hint') }}
              </p>
            </div>
            <button
              type="button"
              @click="windowCostEnabled = !windowCostEnabled"
              :class="[
                'components-account-create-account-modal__action-47',
                windowCostEnabled ? 'components-account-create-account-modal__action-48' : 'components-account-create-account-modal__action-49'
              ]"
            >
              <span
                :class="[
                  'components-account-create-account-modal__text-19',
                  windowCostEnabled ? 'toggle-thumb--on' : 'components-account-create-account-modal__text-20'
                ]"
              />
            </button>
          </div>

          <div v-if="windowCostEnabled" class="components-account-create-account-modal__panel-52">
            <div>
              <label class="input-label">{{ t('admin.accounts.quotaControl.windowCost.limit') }}</label>
              <div class="components-account-create-account-modal__panel-53">
                <span class="components-account-create-account-modal__text-14">$</span>
                <input
                  v-model.number="windowCostLimit"
                  type="number"
                  min="0"
                  step="1"
                  class="components-account-create-account-modal__field-6 input"
                  :placeholder="t('admin.accounts.quotaControl.windowCost.limitPlaceholder')"
                />
              </div>
              <p class="input-hint">{{ t('admin.accounts.quotaControl.windowCost.limitHint') }}</p>
            </div>
            <div>
              <label class="input-label">{{ t('admin.accounts.quotaControl.windowCost.stickyReserve') }}</label>
              <div class="components-account-create-account-modal__panel-53">
                <span class="components-account-create-account-modal__text-14">$</span>
                <input
                  v-model.number="windowCostStickyReserve"
                  type="number"
                  min="0"
                  step="1"
                  class="components-account-create-account-modal__field-6 input"
                  :placeholder="t('admin.accounts.quotaControl.windowCost.stickyReservePlaceholder')"
                />
              </div>
              <p class="input-hint">{{ t('admin.accounts.quotaControl.windowCost.stickyReserveHint') }}</p>
            </div>
          </div>
        </div>

        <!-- Session Limit -->
        <div class="components-account-create-account-modal__panel-51">
          <div class="components-account-create-account-modal__panel-39">
            <div>
              <label class="components-account-create-account-modal__label input-label">{{ t('admin.accounts.quotaControl.sessionLimit.label') }}</label>
              <p class="components-account-create-account-modal__panel-18">
                {{ t('admin.accounts.quotaControl.sessionLimit.hint') }}
              </p>
            </div>
            <button
              type="button"
              @click="sessionLimitEnabled = !sessionLimitEnabled"
              :class="[
                'components-account-create-account-modal__action-47',
                sessionLimitEnabled ? 'components-account-create-account-modal__action-48' : 'components-account-create-account-modal__action-49'
              ]"
            >
              <span
                :class="[
                  'components-account-create-account-modal__text-19',
                  sessionLimitEnabled ? 'toggle-thumb--on' : 'components-account-create-account-modal__text-20'
                ]"
              />
            </button>
          </div>

          <div v-if="sessionLimitEnabled" class="components-account-create-account-modal__panel-52">
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
              <div class="components-account-create-account-modal__panel-53">
                <input
                  v-model.number="sessionIdleTimeout"
                  type="number"
                  min="1"
                  step="1"
                  class="components-account-create-account-modal__field-7 input"
                  :placeholder="t('admin.accounts.quotaControl.sessionLimit.idleTimeoutPlaceholder')"
                />
                <span class="components-account-create-account-modal__text-15">{{ t('common.minutes') }}</span>
              </div>
              <p class="input-hint">{{ t('admin.accounts.quotaControl.sessionLimit.idleTimeoutHint') }}</p>
            </div>
          </div>
        </div>

        <!-- RPM Limit -->
        <div class="components-account-create-account-modal__panel-51">
          <div class="components-account-create-account-modal__panel-39">
            <div>
              <label class="components-account-create-account-modal__label input-label">{{ t('admin.accounts.quotaControl.rpmLimit.label') }}</label>
              <p class="components-account-create-account-modal__panel-18">
                {{ t('admin.accounts.quotaControl.rpmLimit.hint') }}
              </p>
            </div>
            <button
              type="button"
              @click="rpmLimitEnabled = !rpmLimitEnabled"
              :class="[
                'components-account-create-account-modal__action-47',
                rpmLimitEnabled ? 'components-account-create-account-modal__action-48' : 'components-account-create-account-modal__action-49'
              ]"
            >
              <span
                :class="[
                  'components-account-create-account-modal__text-19',
                  rpmLimitEnabled ? 'toggle-thumb--on' : 'components-account-create-account-modal__text-20'
                ]"
              />
            </button>
          </div>

          <div v-if="rpmLimitEnabled" class="components-account-create-account-modal__panel-23">
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
              <div class="components-account-create-account-modal__panel-54">
                <button
                  type="button"
                  @click="rpmStrategy = 'tiered'"
                  :class="[
                    'components-account-create-account-modal__action-52',
                    rpmStrategy === 'tiered'
                      ? 'components-account-create-account-modal__action-44'
                      : 'components-account-create-account-modal__action-45'
                  ]"
                >
                  <div class="components-account-create-account-modal__panel-55">
                    <div>{{ t('admin.accounts.quotaControl.rpmLimit.strategyTiered') }}</div>
                    <div class="components-account-create-account-modal__panel-56">{{ t('admin.accounts.quotaControl.rpmLimit.strategyTieredHint') }}</div>
                  </div>
                </button>
                <button
                  type="button"
                  @click="rpmStrategy = 'sticky_exempt'"
                  :class="[
                    'components-account-create-account-modal__action-52',
                    rpmStrategy === 'sticky_exempt'
                      ? 'components-account-create-account-modal__action-44'
                      : 'components-account-create-account-modal__action-45'
                  ]"
                >
                  <div class="components-account-create-account-modal__panel-55">
                    <div>{{ t('admin.accounts.quotaControl.rpmLimit.strategyStickyExempt') }}</div>
                    <div class="components-account-create-account-modal__panel-56">{{ t('admin.accounts.quotaControl.rpmLimit.strategyStickyExemptHint') }}</div>
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
          <div class="components-account-create-account-modal__panel-10">
            <label class="input-label">{{ t('admin.accounts.quotaControl.rpmLimit.userMsgQueue') }}</label>
            <p class="components-account-create-account-modal__description-6">
              {{ t('admin.accounts.quotaControl.rpmLimit.userMsgQueueHint') }}
            </p>
            <div class="components-account-create-account-modal__panel-57">
              <button type="button" v-for="opt in umqModeOptions" :key="opt.value"
                @click="userMsgQueueMode = opt.value"
                :class="[
                  'components-account-create-account-modal__action-53',
                  userMsgQueueMode === opt.value
                    ? 'components-account-create-account-modal__action-54'
                    : 'components-account-create-account-modal__action-55'
                ]">
                {{ opt.label }}
              </button>
            </div>
          </div>
        </div>

        <!-- TLS Fingerprint -->
        <div class="components-account-create-account-modal__panel-51">
          <div class="components-account-create-account-modal__panel-12">
            <div>
              <label class="components-account-create-account-modal__label input-label">{{ t('admin.accounts.quotaControl.tlsFingerprint.label') }}</label>
              <p class="components-account-create-account-modal__panel-18">
                {{ t('admin.accounts.quotaControl.tlsFingerprint.hint') }}
              </p>
            </div>
            <button
              type="button"
              @click="tlsFingerprintEnabled = !tlsFingerprintEnabled"
              :class="[
                'components-account-create-account-modal__action-47',
                tlsFingerprintEnabled ? 'components-account-create-account-modal__action-48' : 'components-account-create-account-modal__action-49'
              ]"
            >
              <span
                :class="[
                  'components-account-create-account-modal__text-19',
                  tlsFingerprintEnabled ? 'toggle-thumb--on' : 'components-account-create-account-modal__text-20'
                ]"
              />
            </button>
          </div>
          <!-- Profile selector -->
          <div v-if="tlsFingerprintEnabled" class="components-account-create-account-modal__panel-19">
            <Select v-model="tlsFingerprintProfileId" :options="[
              { value: null, label: t('admin.accounts.quotaControl.tlsFingerprint.defaultProfile') },
              ...(tlsFingerprintProfiles.length ? [{ value: 'random', label: t('admin.accounts.quotaControl.tlsFingerprint.randomProfile') }] : []),
              ...tlsFingerprintProfiles.map(p => ({ value: p.id, label: p.name }))
            ]" />
          </div>
        </div>

        <!-- Session ID Masking -->
        <div class="components-account-create-account-modal__panel-51">
          <div class="components-account-create-account-modal__panel-12">
            <div>
              <label class="components-account-create-account-modal__label input-label">{{ t('admin.accounts.quotaControl.sessionIdMasking.label') }}</label>
              <p class="components-account-create-account-modal__panel-18">
                {{ t('admin.accounts.quotaControl.sessionIdMasking.hint') }}
              </p>
            </div>
            <button
              type="button"
              @click="sessionIdMaskingEnabled = !sessionIdMaskingEnabled"
              :class="[
                'components-account-create-account-modal__action-47',
                sessionIdMaskingEnabled ? 'components-account-create-account-modal__action-48' : 'components-account-create-account-modal__action-49'
              ]"
            >
              <span
                :class="[
                  'components-account-create-account-modal__text-19',
                  sessionIdMaskingEnabled ? 'toggle-thumb--on' : 'components-account-create-account-modal__text-20'
                ]"
              />
            </button>
          </div>
        </div>

        <!-- Cache TTL Override -->
        <div class="components-account-create-account-modal__panel-51">
          <div class="components-account-create-account-modal__panel-12">
            <div>
              <label class="components-account-create-account-modal__label input-label">{{ t('admin.accounts.quotaControl.cacheTTLOverride.label') }}</label>
              <p class="components-account-create-account-modal__panel-18">
                {{ t('admin.accounts.quotaControl.cacheTTLOverride.hint') }}
              </p>
            </div>
            <button
              type="button"
              @click="cacheTTLOverrideEnabled = !cacheTTLOverrideEnabled"
              :class="[
                'components-account-create-account-modal__action-47',
                cacheTTLOverrideEnabled ? 'components-account-create-account-modal__action-48' : 'components-account-create-account-modal__action-49'
              ]"
            >
              <span
                :class="[
                  'components-account-create-account-modal__text-19',
                  cacheTTLOverrideEnabled ? 'toggle-thumb--on' : 'components-account-create-account-modal__text-20'
                ]"
              />
            </button>
          </div>
          <div v-if="cacheTTLOverrideEnabled" class="components-account-create-account-modal__panel-19">
            <label class="components-account-create-account-modal__label-4 input-label">{{ t('admin.accounts.quotaControl.cacheTTLOverride.target') }}</label>
            <Select v-model="cacheTTLOverrideTarget" class="components-account-create-account-modal__field-8" :options="[
              { value: '5m', label: '5m' },
              { value: '1h', label: '1h' }
            ]" />
            <p class="components-account-create-account-modal__panel-18">
              {{ t('admin.accounts.quotaControl.cacheTTLOverride.targetHint') }}
            </p>
          </div>
        </div>

        <!-- Custom Base URL Relay -->
        <div class="components-account-create-account-modal__panel-51">
          <div class="components-account-create-account-modal__panel-12">
            <div>
              <label class="components-account-create-account-modal__label input-label">{{ t('admin.accounts.quotaControl.customBaseUrl.label') }}</label>
              <p class="components-account-create-account-modal__panel-18">
                {{ t('admin.accounts.quotaControl.customBaseUrl.hint') }}
              </p>
            </div>
            <button
              type="button"
              @click="customBaseUrlEnabled = !customBaseUrlEnabled"
              :class="[
                'components-account-create-account-modal__action-47',
                customBaseUrlEnabled ? 'components-account-create-account-modal__action-48' : 'components-account-create-account-modal__action-49'
              ]"
            >
              <span
                :class="[
                  'components-account-create-account-modal__text-19',
                  customBaseUrlEnabled ? 'toggle-thumb--on' : 'components-account-create-account-modal__text-20'
                ]"
              />
            </button>
          </div>
          <div v-if="customBaseUrlEnabled" class="components-account-create-account-modal__panel-19">
            <input
              v-model="customBaseUrl"
              type="text"
              class="input"
              :placeholder="t('admin.accounts.quotaControl.customBaseUrl.urlHint')"
            />
          </div>
        </div>
      </div>

      <div>
        <div class="components-account-create-account-modal__panel-58">
          <label class="components-account-create-account-modal__label input-label">{{ t('admin.accounts.proxy') }}</label>
          <ProxyAdBanner />
        </div>
        <ProxySelector v-model="form.proxy_id" :proxies="proxies" />
      </div>

      <div class="components-account-create-account-modal__panel-59">
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
          <input v-model.number="form.rate_multiplier" type="number" min="0" step="0.001" class="input" />
          <p class="input-hint">{{ t('admin.accounts.billingRateMultiplierHint') }}</p>
        </div>
      </div>
      <div class="components-account-create-account-modal__panel-30">
        <label class="input-label">{{ t('admin.accounts.expiresAt') }}</label>
        <input v-model="expiresAtInput" type="datetime-local" class="input" />
        <p class="input-hint">
          {{ t('admin.accounts.expiresAtHint') }}
          {{ t('admin.accounts.expiresAtTimezoneHint', { timezone: browserTimeZone }) }}
        </p>
      </div>

      <!-- OpenAI 自动透传开关（OAuth/API Key） -->
      <div
        v-if="form.platform === 'openai'"
        class="components-account-create-account-modal__panel-30"
      >
        <div class="components-account-create-account-modal__panel-12">
          <div>
            <label class="components-account-create-account-modal__label input-label">{{ t('admin.accounts.openai.oauthPassthrough') }}</label>
            <p class="components-account-create-account-modal__panel-18">
              {{ t('admin.accounts.openai.oauthPassthroughDesc') }}
            </p>
          </div>
          <button
            type="button"
            @click="openaiPassthroughEnabled = !openaiPassthroughEnabled"
            :class="[
              'components-account-create-account-modal__action-47',
              openaiPassthroughEnabled ? 'components-account-create-account-modal__action-48' : 'components-account-create-account-modal__action-49'
            ]"
          >
            <span
              :class="[
                'components-account-create-account-modal__text-19',
                openaiPassthroughEnabled ? 'toggle-thumb--on' : 'components-account-create-account-modal__text-20'
              ]"
            />
          </button>
        </div>
      </div>

      <!-- OpenAI Codex namespace 工具摊平（兼容开关，仅 OAuth） -->
      <div
        v-if="form.platform === 'openai' && form.type === 'oauth'"
        class="components-account-create-account-modal__panel-30"
      >
        <div class="components-account-create-account-modal__panel-12">
          <div>
            <label class="components-account-create-account-modal__label input-label">{{ t('admin.accounts.openai.flattenNamespaces') }}</label>
            <p class="components-account-create-account-modal__panel-18">
              {{ t('admin.accounts.openai.flattenNamespacesDesc') }}
            </p>
          </div>
          <button
            type="button"
            data-testid="create-openai-flatten-namespaces-toggle"
            @click="openaiFlattenNamespacesEnabled = !openaiFlattenNamespacesEnabled"
            :class="[
              'components-account-create-account-modal__action-47',
              openaiFlattenNamespacesEnabled ? 'components-account-create-account-modal__action-48' : 'components-account-create-account-modal__action-49'
            ]"
          >
            <span
              :class="[
                'components-account-create-account-modal__text-19',
                openaiFlattenNamespacesEnabled ? 'toggle-thumb--on' : 'components-account-create-account-modal__text-20'
              ]"
            />
          </button>
        </div>
      </div>

      <!-- OpenAI WS Mode 三态（off/ctx_pool/passthrough） -->
      <div
        v-if="form.platform === 'openai' && (accountCategory === 'oauth-based' || accountCategory === 'apikey')"
        class="components-account-create-account-modal__panel-30"
      >
        <div class="components-account-create-account-modal__panel-12">
          <div>
            <label class="components-account-create-account-modal__label input-label">{{ t('admin.accounts.openai.wsMode') }}</label>
            <p class="components-account-create-account-modal__panel-18">
              {{ t('admin.accounts.openai.wsModeDesc') }}
            </p>
            <p class="components-account-create-account-modal__panel-18">
              {{ t(openAIWSModeConcurrencyHintKey) }}
            </p>
          </div>
          <div class="components-account-create-account-modal__panel-60">
            <Select v-model="openaiResponsesWebSocketV2Mode" :options="openAIWSModeOptions" />
          </div>
        </div>
      </div>

      <!-- Anthropic API Key 自动透传开关 -->
      <div
        v-if="form.platform === 'anthropic' && accountCategory === 'apikey'"
        class="components-account-create-account-modal__panel-30"
      >
        <div class="components-account-create-account-modal__panel-12">
          <div>
            <label class="components-account-create-account-modal__label input-label">{{ t('admin.accounts.anthropic.apiKeyPassthrough') }}</label>
            <p class="components-account-create-account-modal__panel-18">
              {{ t('admin.accounts.anthropic.apiKeyPassthroughDesc') }}
            </p>
          </div>
          <button
            type="button"
            @click="anthropicPassthroughEnabled = !anthropicPassthroughEnabled"
            :class="[
              'components-account-create-account-modal__action-47',
              anthropicPassthroughEnabled ? 'components-account-create-account-modal__action-48' : 'components-account-create-account-modal__action-49'
            ]"
          >
            <span
              :class="[
                'components-account-create-account-modal__text-19',
                anthropicPassthroughEnabled ? 'toggle-thumb--on' : 'components-account-create-account-modal__text-20'
              ]"
            />
          </button>
        </div>
      </div>

      <div
        v-if="form.platform === 'anthropic' && accountCategory === 'apikey'"
        class="components-account-create-account-modal__panel-30"
      >
        <div class="components-account-create-account-modal__panel-61">
          <div>
            <label class="components-account-create-account-modal__label input-label">{{ t('admin.accounts.anthropic.apiKeyAuthScheme') }}</label>
            <p class="components-account-create-account-modal__panel-18">
              {{ t('admin.accounts.anthropic.apiKeyAuthSchemeDesc') }}
            </p>
          </div>
          <Select v-model="anthropicAPIKeyAuthScheme" class="components-account-create-account-modal__field-9" :options="[
            { value: 'x_api_key', label: t('admin.accounts.anthropic.apiKeyAuthSchemeXApiKey') },
            { value: 'authorization_bearer', label: t('admin.accounts.anthropic.apiKeyAuthSchemeBearer') }
          ]" />
        </div>
      </div>

      <!-- Anthropic API Key: Web Search Emulation (hidden when global disabled) -->
      <div
        v-if="form.platform === 'anthropic' && accountCategory === 'apikey' && webSearchGlobalEnabled"
        class="components-account-create-account-modal__panel-30"
      >
        <div class="components-account-create-account-modal__panel-12">
          <div>
            <label class="components-account-create-account-modal__label input-label">{{ t('admin.accounts.anthropic.webSearchEmulation') }}</label>
            <p class="components-account-create-account-modal__panel-18">
              {{ t('admin.accounts.anthropic.webSearchEmulationDesc') }}
            </p>
          </div>
          <Select v-model="webSearchEmulationMode" class="components-account-create-account-modal__field-10" :options="[
            { value: 'default', label: t('admin.accounts.anthropic.webSearchDefault') },
            { value: 'enabled', label: t('admin.accounts.anthropic.webSearchEnabled') },
            { value: 'disabled', label: t('admin.accounts.anthropic.webSearchDisabled') }
          ]" />
        </div>
      </div>

      <!-- OpenAI OAuth Codex 官方客户端限制开关 -->
      <div
        v-if="form.platform === 'openai' && (accountCategory === 'oauth-based' || accountCategory === 'apikey')"
        class="components-account-create-account-modal__panel-30"
      >
        <div class="components-account-create-account-modal__panel-61">
          <div>
            <label class="components-account-create-account-modal__label input-label">{{ t('admin.accounts.openai.longContextBilling') }}</label>
            <p class="components-account-create-account-modal__panel-18">
              {{ t('admin.accounts.openai.longContextBillingDesc') }}
            </p>
          </div>
          <button
            type="button"
            data-testid="openai-long-context-billing-toggle"
            role="switch"
            :aria-checked="openAILongContextBillingEnabled"
            @click="toggleOpenAILongContextBilling"
            :class="[
              'components-account-create-account-modal__action-47',
              openAILongContextBillingEnabled ? 'components-account-create-account-modal__action-48' : 'components-account-create-account-modal__action-49'
            ]"
          >
            <span
              :class="[
                'components-account-create-account-modal__text-19',
                openAILongContextBillingEnabled ? 'toggle-thumb--on' : 'components-account-create-account-modal__text-20'
              ]"
            />
          </button>
        </div>
      </div>

      <div
        v-if="form.platform === 'openai' && accountCategory === 'oauth-based'"
        class="components-account-create-account-modal__panel-30"
      >
        <div class="components-account-create-account-modal__panel-12">
          <div>
            <label class="components-account-create-account-modal__label input-label">{{ t('admin.accounts.openai.codexCLIOnly') }}</label>
            <p class="components-account-create-account-modal__panel-18">
              {{ t('admin.accounts.openai.codexCLIOnlyDesc') }}
            </p>
          </div>
          <button
            type="button"
            @click="codexCLIOnlyEnabled = !codexCLIOnlyEnabled"
            :class="[
              'components-account-create-account-modal__action-47',
              codexCLIOnlyEnabled ? 'components-account-create-account-modal__action-48' : 'components-account-create-account-modal__action-49'
            ]"
          >
            <span
              :class="[
                'components-account-create-account-modal__text-19',
                codexCLIOnlyEnabled ? 'toggle-thumb--on' : 'components-account-create-account-modal__text-20'
              ]"
            />
          </button>
        </div>
        <div
          v-if="codexCLIOnlyEnabled"
          class="components-account-create-account-modal__panel-62"
        >
          <div>
            <label class="components-account-create-account-modal__label input-label">{{ t('admin.accounts.openai.codexCLIOnlyAppServer') }}</label>
            <p class="components-account-create-account-modal__panel-18">
              {{ t('admin.accounts.openai.codexCLIOnlyAppServerDesc') }}
            </p>
          </div>
          <button
            type="button"
            @click="codexCLIOnlyAppServerEnabled = !codexCLIOnlyAppServerEnabled"
            :class="[
              'components-account-create-account-modal__action-47',
              codexCLIOnlyAppServerEnabled ? 'components-account-create-account-modal__action-48' : 'components-account-create-account-modal__action-49'
            ]"
          >
            <span
              :class="[
                'components-account-create-account-modal__text-19',
                codexCLIOnlyAppServerEnabled ? 'toggle-thumb--on' : 'components-account-create-account-modal__text-20'
              ]"
            />
          </button>
        </div>
      </div>

      <!-- Codex 指纹收敛模式（仅 OpenAI OAuth） -->
      <div
        v-if="form.platform === 'openai' && accountCategory === 'oauth-based'"
        class="components-account-create-account-modal__panel-30"
      >
        <div class="components-account-create-account-modal__panel-61">
          <div class="components-account-create-account-modal__panel-16">
            <label class="components-account-create-account-modal__label input-label">{{ t('admin.accounts.openai.codexFingerprintMode') }}</label>
            <p class="components-account-create-account-modal__panel-18">
              {{ t('admin.accounts.openai.codexFingerprintModeDesc') }}
            </p>
          </div>
          <div class="components-account-create-account-modal__panel-63">
            <Select v-model="codexFingerprintMode" data-testid="create-codex-fingerprint-mode-select" :options="codexFingerprintModeOptions" />
          </div>
        </div>
      </div>

      <!-- OpenAI Compact 能力配置 -->
      <div
        v-if="form.platform === 'openai' && (accountCategory === 'oauth-based' || accountCategory === 'apikey')"
        class="components-account-create-account-modal__panel-44"
      >
        <div class="components-account-create-account-modal__panel-12">
          <div>
            <label class="components-account-create-account-modal__label input-label">{{ t('admin.accounts.openai.compactMode') }}</label>
            <p class="components-account-create-account-modal__panel-18">
              {{ t('admin.accounts.openai.compactModeDesc') }}
            </p>
          </div>
          <div class="components-account-create-account-modal__panel-64">
            <Select v-model="openAICompactMode" :options="openAICompactModeOptions" />
          </div>
        </div>
        <div>
          <label class="input-label">{{ t('admin.accounts.openai.compactModelMapping') }}</label>
          <p class="input-hint">{{ t('admin.accounts.openai.compactModelMappingDesc') }}</p>
          <div v-if="openAICompactModelMappings.length > 0" class="components-account-create-account-modal__panel-32">
            <div
              v-for="(mapping, index) in openAICompactModelMappings"
              :key="getOpenAICompactModelMappingKey(mapping)"
              class="components-account-create-account-modal__panel-34"
            >
              <input v-model="mapping.from" type="text" class="components-account-create-account-modal__field-4 input" :placeholder="t('admin.accounts.fromModel')" />
              <span class="components-account-create-account-modal__text-12">→</span>
              <input v-model="mapping.to" type="text" class="components-account-create-account-modal__field-4 input" :placeholder="t('admin.accounts.toModel')" />
              <button type="button" @click="removeOpenAICompactModelMapping(index)" class="components-account-create-account-modal__action-8">
                <Icon name="trash" size="sm" />
              </button>
            </div>
          </div>
          <button type="button" @click="addOpenAICompactModelMapping" class="components-account-create-account-modal__action-9 btn btn-secondary">
            + {{ t('admin.accounts.addMapping') }}
          </button>
        </div>
      </div>

      <!-- OpenAI APIKey Responses API support mode -->
      <div
        v-if="form.platform === 'openai' && accountCategory === 'apikey'"
        class="components-account-create-account-modal__panel-65"
      >
        <div class="components-account-create-account-modal__panel-61">
          <div>
            <label class="components-account-create-account-modal__label input-label">{{ t('admin.accounts.openai.responsesMode') }}</label>
            <p class="components-account-create-account-modal__panel-18">
              {{ t('admin.accounts.openai.responsesModeDesc') }}
            </p>
          </div>
          <div class="components-account-create-account-modal__panel-66">
            <Select
              v-model="openAIResponsesMode"
              :options="openAIResponsesModeOptions"
              :disabled="!openAITextGenerationCapabilityEnabled"
              data-testid="openai-responses-mode-select"
            />
          </div>
        </div>
        <p
          v-if="!openAITextGenerationCapabilityEnabled"
          class="components-account-create-account-modal__description-7"
          data-testid="openai-responses-mode-not-applicable"
        >
          {{ t('admin.accounts.openai.responsesModeTextDisabledHint') }}
        </p>
        <div>
          <label class="components-account-create-account-modal__label-5 input-label">{{ t('admin.accounts.openai.endpointCapabilities') }}</label>
          <div class="components-account-create-account-modal__panel-67">
            <label
              v-for="option in openAIEndpointCapabilityOptions"
              :key="option.value"
              class="components-account-create-account-modal__label-6"
            >
              <input
                type="checkbox"
                class="components-account-create-account-modal__field-5"
                :data-testid="`openai-endpoint-capability-${option.value}`"
                :checked="openAIEndpointCapabilities.includes(option.value)"
                @change="toggleOpenAIEndpointCapability(option.value, $event)"
              />
              <span class="components-account-create-account-modal__text-16">{{ option.label }}</span>
            </label>
          </div>
          <p class="input-hint">{{ t('admin.accounts.openai.endpointCapabilitiesDesc') }}</p>
        </div>
      </div>

      <div>
        <div class="components-account-create-account-modal__panel-12">
          <div>
            <label class="components-account-create-account-modal__label input-label">{{
              t('admin.accounts.autoPauseOnExpired')
            }}</label>
            <p class="components-account-create-account-modal__panel-18">
              {{ t('admin.accounts.autoPauseOnExpiredDesc') }}
            </p>
          </div>
          <button
            type="button"
            @click="autoPauseOnExpired = !autoPauseOnExpired"
            :class="[
              'components-account-create-account-modal__action-47',
              autoPauseOnExpired ? 'components-account-create-account-modal__action-48' : 'components-account-create-account-modal__action-49'
            ]"
          >
            <span
              :class="[
                'components-account-create-account-modal__text-19',
                autoPauseOnExpired ? 'toggle-thumb--on' : 'components-account-create-account-modal__text-20'
              ]"
            />
          </button>
        </div>
      </div>

      <div class="components-account-create-account-modal__panel-30">
        <!-- Group Selection - 仅标准模式显示 -->
        <GroupSelector
          v-if="!authStore.isSimpleMode"
          v-model="form.group_ids"
          :groups="groups"
          :platform="form.platform"
          data-tour="account-form-groups"
        />
      </div>

    </form>

    <!-- Step 2: OAuth Authorization -->
    <div v-else class="components-account-create-account-modal__form">
      <OAuthAuthorizationFlow
        ref="oauthFlowRef"
        :add-method="form.platform === 'anthropic' ? addMethod : 'oauth'"
        :auth-url="currentAuthUrl"
        :session-id="currentSessionId"
        :loading="currentOAuthLoading"
        :error="currentOAuthError"
        :show-help="form.platform === 'anthropic'"
        :show-proxy-warning="form.platform !== 'openai' && form.platform !== 'grok' && !!form.proxy_id"
        :allow-multiple="form.platform === 'anthropic'"
        :show-cookie-option="form.platform === 'anthropic'"
        :show-refresh-token-option="form.platform === 'openai' || form.platform === 'antigravity' || form.platform === 'grok'"
        :show-mobile-refresh-token-option="form.platform === 'openai'"
        :show-session-token-option="false"
        :show-access-token-option="false"
        :show-codex-session-import-option="form.platform === 'openai'"
        :show-agent-identity-option="form.platform === 'openai'"
        :show-codex-pat-option="form.platform === 'openai'"
        :show-sso-option="form.platform === 'grok'"
        :show-email-password-option="false"
        :show-manual-option="true"
        :initial-input-method="'manual'"
        :platform="form.platform"
        :show-project-id="form.platform === 'gemini' && geminiOAuthType === 'code_assist'"
        @generate-url="handleGenerateUrl"
        @cookie-auth="handleCookieAuth"
        @validate-refresh-token="handleValidateRefreshToken"
        @validate-mobile-refresh-token="handleOpenAIValidateMobileRT"
        @validate-session-token="handleValidateSessionToken"
        @import-codex-session="handleOpenAIImportCodexSession"
        @import-codex-pat="handleOpenAIImportCodexPAT"
        @import-sso="handleGrokImportSSO"
        @authorize-password="handleGrokAuthorizePassword"
      />

    </div>

    <template #footer>
      <div v-if="step === 1" class="components-account-create-account-modal__panel-72">
        <button @click="handleClose" type="button" class="btn btn-secondary">
          {{ t('common.cancel') }}
        </button>
        <button
          type="submit"
          form="create-account-form"
          :disabled="submitting"
          class="btn btn-primary"
          data-tour="account-form-submit"
        >
          <svg
            v-if="submitting"
            class="components-account-create-account-modal__icon-6"
            fill="none"
            viewBox="0 0 24 24"
          >
            <circle
              class="components-account-create-account-modal__circle"
              cx="12"
              cy="12"
              r="10"
              stroke="currentColor"
              stroke-width="4"
            ></circle>
            <path
              class="components-account-create-account-modal__path"
              fill="currentColor"
              d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
            ></path>
          </svg>
          {{
            isOAuthFlow
              ? t('common.next')
              : submitting
                ? t('admin.accounts.creating')
                : t('common.create')
          }}
        </button>
      </div>
      <div v-else class="components-account-create-account-modal__panel-73">
        <button type="button" class="btn btn-secondary" @click="goBackToBasicInfo">
          {{ t('common.back') }}
        </button>
        <button
          v-if="isManualInputMethod"
          type="button"
          :disabled="!canExchangeCode"
          class="btn btn-primary"
          @click="handleExchangeCode"
        >
          <svg
            v-if="currentOAuthLoading"
            class="components-account-create-account-modal__icon-6"
            fill="none"
            viewBox="0 0 24 24"
          >
            <circle
              class="components-account-create-account-modal__circle"
              cx="12"
              cy="12"
              r="10"
              stroke="currentColor"
              stroke-width="4"
            ></circle>
            <path
              class="components-account-create-account-modal__path"
              fill="currentColor"
              d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
            ></path>
          </svg>
          {{
            currentOAuthLoading
              ? t('admin.accounts.oauth.verifying')
              : t('admin.accounts.oauth.completeAuth')
          }}
        </button>
      </div>
    </template>
  </BaseDialog>

</template>

<script setup lang="ts">
import { ref, reactive, computed, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import {
  claudeModels,
  getPresetMappingsByPlatform,
  getModelsByPlatform,
  commonErrorCodes,
  buildModelMappingObject,
  buildModelReasoningEffortsObject
} from '@/composables/useModelWhitelist'
import { useAuthStore } from '@/stores/auth'
import { adminAPI } from '@/api/admin'
import { useQuotaNotifyState } from '@/composables/useQuotaNotifyState'
import {
  useAccountOAuth,
  type AddMethod,
  type AuthInputMethod
} from '@/composables/useAccountOAuth'
import { useOpenAIOAuth } from '@/composables/useOpenAIOAuth'
import { useGeminiOAuth } from '@/composables/useGeminiOAuth'
import { useAntigravityOAuth } from '@/composables/useAntigravityOAuth'
import { useGrokOAuth } from '@/composables/useGrokOAuth'
import type {
  Proxy,
  AdminGroup,
  AccountPlatform,
  AccountType,
  CreateAccountRequest,
  CodexSessionImportMessage,
  OpenAICompactMode,
  OpenAIResponsesMode,
  OpenAIEndpointCapability
} from '@/types'
import BaseDialog from '@/components/common/BaseDialog.vue'
import Select from '@/components/common/Select.vue'
import PlatformIcon from '@/components/common/PlatformIcon.vue'
import Icon from '@/components/icons/Icon.vue'
import ProxySelector from '@/components/common/ProxySelector.vue'
import ProxyAdBanner from '@/components/common/ProxyAdBanner.vue'
import GroupSelector from '@/components/common/GroupSelector.vue'
import ModelWhitelistSelector from '@/components/account/ModelWhitelistSelector.vue'
import AccountModelRuleSelector from '@/components/account/AccountModelRuleSelector.vue'
import QuotaLimitCard from '@/components/account/QuotaLimitCard.vue'
import Toggle from '@/components/common/Toggle.vue'
import GrokBaseUrlPresets from '@/components/account/GrokBaseUrlPresets.vue'
import CnBaseUrlPresets from '@/components/account/CnBaseUrlPresets.vue'
import HeaderOverrideEditor from '@/components/account/HeaderOverrideEditor.vue'
import {
  applyHeaderOverride,
  applyInterceptWarmup,
  cnSupportsNativeResponses,
  defaultCNBaseUrl,
  isHeaderOverrideCapable,
  validateHeaderOverrideRows,
  type CnAccountMode,
  type CnApiProtocol,
  type HeaderOverrideRow
} from '@/components/account/credentialsBuilder'
import {
  formatDateTimeLocalInput,
  getBrowserTimeZone,
  parseDateTimeLocalInput
} from '@/utils/format'
import { createStableObjectKeyResolver } from '@/utils/stableObjectKey'
import { VERTEX_LOCATION_SELECT_OPTIONS } from '@/constants/account'
import {
  OPENAI_WS_MODE_CTX_POOL,
  OPENAI_WS_MODE_OFF,
  OPENAI_WS_MODE_PASSTHROUGH,
  OPENAI_WS_MODE_HTTP_BRIDGE,
  isOpenAIWSModeEnabled,
  resolveOpenAIWSModeConcurrencyHintKey,
  type OpenAIWSMode
} from '@/utils/openaiWsMode'
import OAuthAuthorizationFlow from './OAuthAuthorizationFlow.vue'

// Type for exposed OAuthAuthorizationFlow component
// Note: defineExpose automatically unwraps refs, so we use the unwrapped types
interface OAuthFlowExposed {
  authCode: string
  oauthState: string
  projectId: string
  sessionKey: string
  refreshToken: string
  sessionToken: string
  codexSession: string
  codexPAT: string
  ssoCookie: string
  inputMethod: AuthInputMethod
  reset: () => void
}

const { t } = useI18n()
const authStore = useAuthStore()
const browserTimeZone = getBrowserTimeZone()

const oauthStepTitle = computed(() => {
  if (form.platform === 'openai') return t('admin.accounts.oauth.openai.title')
  if (form.platform === 'gemini') return t('admin.accounts.oauth.gemini.title')
  if (form.platform === 'antigravity') return t('admin.accounts.oauth.antigravity.title')
  if (form.platform === 'grok') return t('admin.accounts.oauth.grok.title')
  return t('admin.accounts.oauth.title')
})

// Platform-specific hints for API Key type
const baseUrlHint = computed(() => {
  if (form.platform === 'openai') return t('admin.accounts.openai.baseUrlHint')
  if (form.platform === 'grok') return ''
  return t('admin.accounts.baseUrlHint')
})

const apiKeyHint = computed(() => {
  if (form.platform === 'openai') return t('admin.accounts.openai.apiKeyHint')
  if (form.platform === 'grok') return ''
  return t('admin.accounts.apiKeyHint')
})

// Base URL / API Key 占位符：国产供应商随账号类型变化。
const apiKeyBaseUrlPlaceholder = computed(() => {
  if (isCNPlatform.value) {
    return defaultCNBaseUrl(form.platform, accountMode.value, apiProtocol.value) || 'https://api.example.com'
  }
  switch (form.platform) {
    case 'openai':
      return 'https://api.openai.com'
    case 'gemini':
      return 'https://generativelanguage.googleapis.com'
    case 'antigravity':
      return 'https://cloudcode-pa.googleapis.com'
    case 'grok':
      return 'https://api.x.ai/v1'
    default:
      return 'https://api.anthropic.com'
  }
})

const apiKeyValuePlaceholder = computed(() => {
  switch (form.platform) {
    case 'openai':
      return 'sk-proj-...'
    case 'gemini':
      return 'AIza...'
    case 'antigravity':
      return 'API Key'
    case 'grok':
      return 'xai-...'
    case 'kimi':
      return 'sk-...'
    case 'zhipu':
      return '<api-key>.<secret>'
    case 'deepseek':
      return 'sk-...'
    default:
      return 'sk-ant-...'
  }
})

interface Props {
  show: boolean
  proxies: Proxy[]
  groups: AdminGroup[]
}

const props = defineProps<Props>()
const emit = defineEmits<{
  close: []
  created: []
}>()

const appStore = useAppStore()

// OAuth composables
const oauth = useAccountOAuth() // For Anthropic OAuth
const openaiOAuth = useOpenAIOAuth() // For OpenAI OAuth
const geminiOAuth = useGeminiOAuth() // For Gemini OAuth
const antigravityOAuth = useAntigravityOAuth() // For Antigravity OAuth
const grokOAuth = useGrokOAuth() // For Grok OAuth

// Computed: current OAuth state for template binding
const currentAuthUrl = computed(() => {
  if (form.platform === 'openai') return openaiOAuth.authUrl.value
  if (form.platform === 'gemini') return geminiOAuth.authUrl.value
  if (form.platform === 'antigravity') return antigravityOAuth.authUrl.value
  if (form.platform === 'grok') return grokOAuth.authUrl.value
  return oauth.authUrl.value
})

const currentSessionId = computed(() => {
  if (form.platform === 'openai') return openaiOAuth.sessionId.value
  if (form.platform === 'gemini') return geminiOAuth.sessionId.value
  if (form.platform === 'antigravity') return antigravityOAuth.sessionId.value
  if (form.platform === 'grok') return grokOAuth.sessionId.value
  return oauth.sessionId.value
})

const currentOAuthLoading = computed(() => {
  if (form.platform === 'openai') return openaiOAuth.loading.value
  if (form.platform === 'gemini') return geminiOAuth.loading.value
  if (form.platform === 'antigravity') return antigravityOAuth.loading.value
  if (form.platform === 'grok') return grokOAuth.loading.value
  return oauth.loading.value
})

const currentOAuthError = computed(() => {
  if (form.platform === 'openai') return openaiOAuth.error.value
  if (form.platform === 'gemini') return geminiOAuth.error.value
  if (form.platform === 'antigravity') return antigravityOAuth.error.value
  if (form.platform === 'grok') return grokOAuth.error.value
  return oauth.error.value
})

// Refs
const oauthFlowRef = ref<OAuthFlowExposed | null>(null)

// Model mapping type
interface ModelMapping {
  from: string
  to: string
  reasoning_effort?: string
}

interface TempUnschedRuleForm {
  error_code: number | null
  keywords: string
  duration_minutes: number | null
  description: string
}

// State
const step = ref(1)
const submitting = ref(false)
const accountCategory = ref<'oauth-based' | 'apikey' | 'bedrock' | 'service_account'>('oauth-based') // UI selection for account category
const addMethod = ref<AddMethod>('oauth') // For oauth-based: 'oauth' or 'setup-token'
const apiKeyBaseUrl = ref('https://api.anthropic.com')
const apiKeyValue = ref('')
const upstreamBillingAutoProbeEnabled = ref(true)

// ── 国产供应商（Kimi / Zhipu / DeepSeek）账号类型、API 协议与端点 ──
const accountMode = ref<CnAccountMode>('payg')
// API 协议决定转发端点与格式：cc=现有转换链，anthropic=原生直通（Claude Code），
// responses=deepseek / kimi 原生 Responses 端点（Codex）。与账号类型正交。
const apiProtocol = ref<CnApiProtocol>('chat_completions')
const zhipuOrganization = ref('')
const zhipuProject = ref('')
const isCNPlatform = computed(
  () => form.platform === 'kimi' || form.platform === 'zhipu' || form.platform === 'deepseek'
)
// CnBaseUrlPresets 的 platform prop 是平台字面量联合类型，模板里不能写
// `as` 断言（其中的 `|` 会被 eslint 误判为 Vue2 filter 语法），经此 computed 传递。
const cnPresetPlatform = computed<'kimi' | 'zhipu' | 'deepseek'>(() => {
  if (form.platform === 'kimi' || form.platform === 'zhipu' || form.platform === 'deepseek') {
    return form.platform
  }
  return 'kimi'
})
// 当前平台可选的协议档（responses 仅 deepseek / kimi）。
const cnProtocolOptions = computed<Array<{ value: CnApiProtocol; labelKey: string }>>(() => {
  const opts: Array<{ value: CnApiProtocol; labelKey: string }> = [
    { value: 'chat_completions', labelKey: 'chatCompletions' },
    { value: 'anthropic', labelKey: 'anthropic' }
  ]
  if (cnSupportsNativeResponses(form.platform)) {
    opts.push({ value: 'responses', labelKey: 'responses' })
  }
  return opts
})
// 当前选中平台的品牌色（选中卡片描边 / 图标底色），与 platformColors 取色一致。
const cnAccentActiveClass = computed(() => {
  switch (form.platform) {
    case 'kimi':
      return 'components-account-create-account-modal__state'
    case 'zhipu':
      return 'components-account-create-account-modal__state-2'
    case 'deepseek':
      return 'components-account-create-account-modal__state-3'
    default:
      return 'components-account-create-account-modal__state-4'
  }
})
const cnAccentIconClass = computed(() => {
  switch (form.platform) {
    case 'kimi':
      return 'components-account-create-account-modal__state-5'
    case 'zhipu':
      return 'components-account-create-account-modal__state-6'
    case 'deepseek':
      return 'components-account-create-account-modal__state-7'
    default:
      return 'components-account-create-account-modal__panel-81'
  }
})
// 切换国产供应商平台：强制 apikey 类型，deepseek 无 coding 套餐故锁定 payg，
// 协议回落 chat_completions，并把 base url 重置为该平台默认端点。
function selectCNPlatform(platform: 'kimi' | 'zhipu' | 'deepseek') {
  form.platform = platform
  form.type = 'apikey'
  accountCategory.value = 'apikey'
  apiProtocol.value = 'chat_completions'
  if (platform === 'deepseek') {
    accountMode.value = 'payg'
  }
  apiKeyBaseUrl.value = defaultCNBaseUrl(platform, accountMode.value, apiProtocol.value)
}
// 账号类型 / 协议变更时同步默认 base url。
watch(accountMode, (mode) => {
  if (!isCNPlatform.value) return
  apiKeyBaseUrl.value = defaultCNBaseUrl(form.platform, mode, apiProtocol.value)
})
watch(apiProtocol, (protocol) => {
  if (!isCNPlatform.value) return
  apiKeyBaseUrl.value = defaultCNBaseUrl(form.platform, accountMode.value, protocol)
})
// 点击预设端点：同时回填 base url、账号类型与协议。
function onCnPresetSelect(preset: { mode: CnAccountMode; protocol: CnApiProtocol; url: string }) {
  accountMode.value = preset.mode
  apiProtocol.value = preset.protocol
  apiKeyBaseUrl.value = preset.url
}

const upstreamModelsPreviewed = ref(false)
const syncPreviewCredentials = computed(() => {
  if (!apiKeyValue.value) return undefined
  return {
    platform: form.platform,
    type: form.type,
    base_url: apiKeyBaseUrl.value || undefined,
    api_key: apiKeyValue.value,
    model_mapping: buildCurrentModelRestrictionMapping() || undefined
  }
})

const editQuotaLimit = ref<number | null>(null)
const editQuotaDailyLimit = ref<number | null>(null)
const editQuotaWeeklyLimit = ref<number | null>(null)
const editDailyResetMode = ref<'rolling' | 'fixed' | null>(null)
const editDailyResetHour = ref<number | null>(null)
const editWeeklyResetMode = ref<'rolling' | 'fixed' | null>(null)
const editWeeklyResetDay = ref<number | null>(null)
const editWeeklyResetHour = ref<number | null>(null)
const editResetTimezone = ref<string | null>(null)
const modelMappings = ref<ModelMapping[]>([])
const openAICompactModelMappings = ref<ModelMapping[]>([])
const modelRestrictionMode = ref<'whitelist' | 'mapping'>('whitelist')
const allowedModels = ref<string[]>([])
const hasNonDefaultWhitelist = computed(() => {
  if (modelRestrictionMode.value !== 'whitelist' || allowedModels.value.length === 0) return false
  const defaults = getModelsByPlatform(form.platform)
  if (allowedModels.value.length !== defaults.length) return true
  const selected = new Set(allowedModels.value)
  return defaults.some(model => !selected.has(model))
})
const hasModelRestrictionValues = computed(() =>
  modelMappings.value.some(mapping => mapping.from.trim() !== '' || mapping.to.trim() !== '') ||
  hasNonDefaultWhitelist.value
)
const DEFAULT_POOL_MODE_RETRY_COUNT = 3
const MAX_POOL_MODE_RETRY_COUNT = 10
const DEFAULT_POOL_MODE_RETRY_STATUS_CODES = [401, 403, 429]
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
const customErrorCodesEnabled = ref(false)
const selectedErrorCodes = ref<number[]>([])
const customErrorCodeInput = ref<number | null>(null)
const headerOverrideEnabled = ref(false)
const headerOverrideRows = ref<HeaderOverrideRow[]>([])

// Grok OAuth：自定义上游地址（base_url 仅改写转发端点，OAuth 授权/刷新不受影响）
const grokOAuthCustomBaseUrlEnabled = ref(false)
const grokOAuthBaseUrl = ref('')

// Grok OAuth 三条创建路径（授权码/RT 批量/SSO 批量）共用的前置校验。
// 授权码路径必须在兑换 code 之前调用，避免校验失败时白白消耗一次性授权码。
const validateGrokOAuthUpstreamConfig = (): boolean => {
  if (grokOAuthCustomBaseUrlEnabled.value) {
    const trimmed = grokOAuthBaseUrl.value.trim()
    if (!trimmed) {
      appStore.showError(t('admin.accounts.grokCustomBaseUrl.required'))
      return false
    }
    if (!/^https?:\/\//i.test(trimmed)) {
      appStore.showError(t('admin.accounts.grokCustomBaseUrl.invalid'))
      return false
    }
  }
  if (headerOverrideEnabled.value) {
    const headerError = validateHeaderOverrideRows(headerOverrideRows.value)
    if (headerError) {
      appStore.showError(t(`admin.accounts.headerOverride.${headerError}`))
      return false
    }
  }
  return true
}

// 把已通过校验的自定义上游地址与请求头覆写写入 credentials
const applyGrokOAuthUpstreamConfig = (credentials: Record<string, unknown>) => {
  if (grokOAuthCustomBaseUrlEnabled.value) {
    credentials.base_url = grokOAuthBaseUrl.value.trim()
  }
  applyHeaderOverride(credentials, headerOverrideEnabled.value, headerOverrideRows.value, 'create')
}
const interceptWarmupRequests = ref(false)
const autoPauseOnExpired = ref(true)
const openaiPassthroughEnabled = ref(false)
// OpenAI Codex namespace 工具摊平兼容开关（仅 OAuth），缺省关闭即原样保留
const openaiFlattenNamespacesEnabled = ref(false)
const openAILongContextBillingEnabled = ref(false)
const openAILongContextBillingTouched = ref(false)
const openAICompactMode = ref<OpenAICompactMode>('auto')
const openAIResponsesMode = ref<OpenAIResponsesMode>('auto')
const openAIEndpointCapabilities = ref<OpenAIEndpointCapability[]>(['chat_completions', 'embeddings'])
const openaiOAuthResponsesWebSocketV2Mode = ref<OpenAIWSMode>(OPENAI_WS_MODE_OFF)
const openaiAPIKeyResponsesWebSocketV2Mode = ref<OpenAIWSMode>(OPENAI_WS_MODE_OFF)
const codexCLIOnlyEnabled = ref(false)
const codexCLIOnlyAppServerEnabled = ref(false)
type CodexFingerprintMode = 'off' | 'device' | 'session' | 'full'
const codexFingerprintMode = ref<CodexFingerprintMode>('off')
const codexFingerprintModeOptions = computed(() => [
  { value: 'off' as CodexFingerprintMode, label: t('admin.accounts.openai.codexFingerprintOff') },
  { value: 'device' as CodexFingerprintMode, label: t('admin.accounts.openai.codexFingerprintDevice') },
  { value: 'session' as CodexFingerprintMode, label: t('admin.accounts.openai.codexFingerprintSession') },
  { value: 'full' as CodexFingerprintMode, label: t('admin.accounts.openai.codexFingerprintFull') },
])
type AnthropicAPIKeyAuthScheme = 'x_api_key' | 'authorization_bearer'
const anthropicPassthroughEnabled = ref(false)
const anthropicAPIKeyAuthScheme = ref<AnthropicAPIKeyAuthScheme>('x_api_key')
const webSearchEmulationMode = ref('default')
const webSearchGlobalEnabled = ref(false)

const toggleOpenAILongContextBilling = () => {
  openAILongContextBillingEnabled.value = !openAILongContextBillingEnabled.value
  openAILongContextBillingTouched.value = true
}
const {
  globalEnabled: quotaNotifyGlobalEnabled,
  state: quotaNotifyState,
  loadGlobalState: loadQuotaNotifyGlobal,
  writeToExtra: writeQuotaNotifyToExtra,
} = useQuotaNotifyState()

// Load global feature states once
adminAPI.settings.getWebSearchEmulationConfig().then(cfg => {
  webSearchGlobalEnabled.value = cfg?.enabled === true && (cfg?.providers?.length ?? 0) > 0
}).catch(() => { webSearchGlobalEnabled.value = false })

loadQuotaNotifyGlobal()
const bedrockPresets = computed(() => getPresetMappingsByPlatform('bedrock'))

// Bedrock credentials
const bedrockAuthMode = ref<'sigv4' | 'apikey'>('sigv4')
const bedrockAccessKeyId = ref('')
const bedrockSecretAccessKey = ref('')
const bedrockSessionToken = ref('')
const bedrockRegion = ref('us-east-1')
const BEDROCK_REGION_OPTIONS = [
  ['us-east-1', 'us-east-1 (N. Virginia)'], ['us-east-2', 'us-east-2 (Ohio)'],
  ['us-west-1', 'us-west-1 (N. California)'], ['us-west-2', 'us-west-2 (Oregon)'],
  ['us-gov-east-1', 'us-gov-east-1 (GovCloud US-East)'], ['us-gov-west-1', 'us-gov-west-1 (GovCloud US-West)'],
  ['eu-west-1', 'eu-west-1 (Ireland)'], ['eu-west-2', 'eu-west-2 (London)'],
  ['eu-west-3', 'eu-west-3 (Paris)'], ['eu-central-1', 'eu-central-1 (Frankfurt)'],
  ['eu-central-2', 'eu-central-2 (Zurich)'], ['eu-south-1', 'eu-south-1 (Milan)'],
  ['eu-south-2', 'eu-south-2 (Spain)'], ['eu-north-1', 'eu-north-1 (Stockholm)'],
  ['ap-northeast-1', 'ap-northeast-1 (Tokyo)'], ['ap-northeast-2', 'ap-northeast-2 (Seoul)'],
  ['ap-northeast-3', 'ap-northeast-3 (Osaka)'], ['ap-south-1', 'ap-south-1 (Mumbai)'],
  ['ap-south-2', 'ap-south-2 (Hyderabad)'], ['ap-southeast-1', 'ap-southeast-1 (Singapore)'],
  ['ap-southeast-2', 'ap-southeast-2 (Sydney)'], ['ca-central-1', 'ca-central-1 (Canada)'],
  ['sa-east-1', 'sa-east-1 (São Paulo)'],
].map(([value, label]) => ({ value, label }))
const bedrockForceGlobal = ref(false)
const bedrockApiKeyValue = ref('')
const vertexServiceAccountFileInput = ref<HTMLInputElement | null>(null)
const vertexServiceAccountJson = ref('')
const vertexProjectId = ref('')
const vertexClientEmail = ref('')
const vertexLocation = ref('global')
const vertexServiceAccountDragActive = ref(false)
const tempUnschedEnabled = ref(false)
const tempUnschedRules = ref<TempUnschedRuleForm[]>([])
const getModelMappingKey = createStableObjectKeyResolver<ModelMapping>('create-model-mapping')
const getOpenAICompactModelMappingKey = createStableObjectKeyResolver<ModelMapping>('create-openai-compact-model-mapping')
const getTempUnschedRuleKey = createStableObjectKeyResolver<TempUnschedRuleForm>('create-temp-unsched-rule')
const geminiOAuthType = ref<'code_assist' | 'google_one' | 'ai_studio'>('google_one')
const geminiAIStudioOAuthEnabled = ref(false)
const geminiTierGoogleOne = ref<'google_one_free' | 'google_ai_pro' | 'google_ai_ultra'>('google_one_free')
const geminiTierGcp = ref<'gcp_standard' | 'gcp_enterprise'>('gcp_standard')
const geminiTierAIStudio = ref<'aistudio_free' | 'aistudio_paid'>('aistudio_free')
const geminiSelectedTier = computed(() => {
  if (accountCategory.value === 'apikey') return geminiTierAIStudio.value
  if (geminiOAuthType.value === 'google_one') return geminiTierGoogleOne.value
  if (geminiOAuthType.value === 'code_assist') return geminiTierGcp.value
  return geminiTierAIStudio.value
})
const openAICompactModeOptions = computed(() => [
  { value: 'auto', label: t('admin.accounts.openai.compactModeAuto') },
  { value: 'force_on', label: t('admin.accounts.openai.compactModeForceOn') },
  { value: 'force_off', label: t('admin.accounts.openai.compactModeForceOff') }
])
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

const buildOpenAICompactModelMapping = () =>
  buildModelMappingObject('mapping', [], openAICompactModelMappings.value)


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

const openAIWSModeOptions = computed(() => [
  { value: OPENAI_WS_MODE_OFF, label: t('admin.accounts.openai.wsModeOff') },
  { value: OPENAI_WS_MODE_CTX_POOL, label: t('admin.accounts.openai.wsModeCtxPool') },
  { value: OPENAI_WS_MODE_PASSTHROUGH, label: t('admin.accounts.openai.wsModePassthrough') },
  { value: OPENAI_WS_MODE_HTTP_BRIDGE, label: t('admin.accounts.openai.wsModeHttpBridge') }
])

const openaiResponsesWebSocketV2Mode = computed({
  get: () => {
    if (form.platform === 'openai' && accountCategory.value === 'apikey') {
      return openaiAPIKeyResponsesWebSocketV2Mode.value
    }
    return openaiOAuthResponsesWebSocketV2Mode.value
  },
  set: (mode: OpenAIWSMode) => {
    if (form.platform === 'openai' && accountCategory.value === 'apikey') {
      openaiAPIKeyResponsesWebSocketV2Mode.value = mode
      return
    }
    openaiOAuthResponsesWebSocketV2Mode.value = mode
  }
})

const openAIWSModeConcurrencyHintKey = computed(() =>
  resolveOpenAIWSModeConcurrencyHintKey(openaiResponsesWebSocketV2Mode.value)
)

const isOpenAIModelRestrictionDisabled = computed(() =>
  form.platform === 'openai' && openaiPassthroughEnabled.value
)

// Computed: current preset mappings based on platform
const presetMappings = computed(() => getPresetMappingsByPlatform(form.platform))
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

const form = reactive({
  name: '',
  notes: '',
  platform: 'anthropic' as AccountPlatform,
  type: 'oauth' as AccountType, // Will be 'oauth', 'setup-token', or 'apikey'
  credentials: {} as Record<string, unknown>,
  proxy_id: null as string | null,
  concurrency: 10,
  load_factor: null as number | null,
  priority: 1,
  rate_multiplier: 1,
  group_ids: [] as string[],
  expires_at: null as number | null
})

// Helper to check if current type needs OAuth flow
const isOAuthFlow = computed(() => {
  // Bedrock 类型不需要 OAuth 流程
  if (form.platform === 'anthropic' && accountCategory.value === 'bedrock') {
    return false
  }
  return accountCategory.value === 'oauth-based'
})

const isGrokSSOInputMethod = computed(() => form.platform === 'grok' && oauthFlowRef.value?.inputMethod === 'sso_cookie')

const isManualInputMethod = computed(() => {
  return oauthFlowRef.value?.inputMethod === 'manual'
})

const expiresAtInput = computed({
  get: () => formatDateTimeLocal(form.expires_at),
  set: (value: string) => {
    form.expires_at = parseDateTimeLocal(value)
  }
})

const canExchangeCode = computed(() => {
  const authCode = oauthFlowRef.value?.authCode || ''
  if (form.platform === 'openai') {
    return authCode.trim() && openaiOAuth.sessionId.value && !openaiOAuth.loading.value
  }
  if (form.platform === 'gemini') {
    return authCode.trim() && geminiOAuth.sessionId.value && !geminiOAuth.loading.value
  }
  if (form.platform === 'antigravity') {
    return authCode.trim() && antigravityOAuth.sessionId.value && !antigravityOAuth.loading.value
  }
  if (form.platform === 'grok') {
    return authCode.trim() && grokOAuth.sessionId.value && !grokOAuth.loading.value
  }
  return authCode.trim() && oauth.sessionId.value && !oauth.loading.value
})

// Watchers
watch(
  () => props.show,
  (newVal) => {
    if (newVal) {
      // Load TLS fingerprint profiles
      adminAPI.tlsFingerprintProfiles.list()
        .then(profiles => { tlsFingerprintProfiles.value = profiles.map(p => ({ id: p.id, name: p.name })) })
        .catch(() => { tlsFingerprintProfiles.value = [] })
      // Modal opened - fill related models
      allowedModels.value = [...getModelsByPlatform(form.platform)]
    } else {
      resetForm()
    }
  }
)

// Sync form.type based on accountCategory and addMethod.
watch(
  [accountCategory, addMethod, () => form.platform],
  ([category, method]) => {
    // Bedrock 类型
    if (form.platform === 'anthropic' && category === 'bedrock') {
      form.type = 'bedrock' as AccountType
      return
    }
    if ((form.platform === 'anthropic' || form.platform === 'gemini') && category === 'service_account') {
      form.type = 'service_account' as AccountType
    } else if (category === 'oauth-based') {
      form.type = form.platform === 'anthropic' ? method as AccountType : 'oauth'
    } else {
      form.type = 'apikey'
    }
  },
  { immediate: true }
)

// Reset platform-specific settings when platform changes
watch(
  () => form.platform,
  (newPlatform) => {
    // Reset base URL based on platform
    if (newPlatform === 'kimi' || newPlatform === 'zhipu' || newPlatform === 'deepseek') {
      apiKeyBaseUrl.value = defaultCNBaseUrl(newPlatform, accountMode.value, apiProtocol.value)
    } else {
      apiKeyBaseUrl.value =
        (newPlatform === 'openai')
          ? 'https://api.openai.com'
          : newPlatform === 'gemini'
            ? 'https://generativelanguage.googleapis.com'
            : newPlatform === 'antigravity'
              ? 'https://cloudcode-pa.googleapis.com'
          : newPlatform === 'grok'
              ? 'https://api.x.ai/v1'
              : 'https://api.anthropic.com'
    }
    // Clear model-related settings
    allowedModels.value = []
    modelMappings.value = []
    if (newPlatform === 'grok') {
      accountCategory.value = 'oauth-based'
      addMethod.value = 'oauth'
      modelRestrictionMode.value = 'mapping'
      form.concurrency = 1
      form.load_factor = null
    }
    if (newPlatform !== 'anthropic' && newPlatform !== 'gemini' && accountCategory.value === 'service_account') {
      accountCategory.value = 'oauth-based'
    }
    if (newPlatform !== 'anthropic' && accountCategory.value === 'bedrock') {
      accountCategory.value = 'oauth-based'
    }
    // Reset Bedrock fields when switching platforms
    bedrockAccessKeyId.value = ''
    bedrockSecretAccessKey.value = ''
    bedrockSessionToken.value = ''
    bedrockRegion.value = 'us-east-1'
    bedrockForceGlobal.value = false
    bedrockAuthMode.value = 'sigv4'
    bedrockApiKeyValue.value = ''
    vertexServiceAccountJson.value = ''
    vertexProjectId.value = ''
    vertexClientEmail.value = ''
    vertexLocation.value = 'global'
    // Reset Anthropic-specific settings when switching to other platforms.
    if (newPlatform !== 'anthropic') {
      interceptWarmupRequests.value = false
    }
    if (newPlatform !== 'openai') {
      openaiPassthroughEnabled.value = false
      openaiFlattenNamespacesEnabled.value = false
      openAIEndpointCapabilities.value = ['chat_completions', 'embeddings']
      openaiOAuthResponsesWebSocketV2Mode.value = OPENAI_WS_MODE_OFF
      openaiAPIKeyResponsesWebSocketV2Mode.value = OPENAI_WS_MODE_OFF
      codexCLIOnlyEnabled.value = false
      codexCLIOnlyAppServerEnabled.value = false
    }
    if (newPlatform !== 'anthropic') {
      anthropicPassthroughEnabled.value = false
      anthropicAPIKeyAuthScheme.value = 'x_api_key'
      webSearchEmulationMode.value = 'default'
    }
    // 请求头覆写为平台相关配置（常用头集合不同），切换平台时清空，
    // 避免上一平台的配置行被提交到新平台账号
    headerOverrideEnabled.value = false
    headerOverrideRows.value = []
    grokOAuthCustomBaseUrlEnabled.value = false
    grokOAuthBaseUrl.value = ''
    // Reset OAuth states
    oauth.resetState()
    openaiOAuth.resetState()
    geminiOAuth.resetState()
    antigravityOAuth.resetState()
    grokOAuth.resetState()
  }
)

watch(
  [() => props.show, () => form.platform, accountCategory],
  async ([show, platform, category]) => {
    if (!show || platform !== 'gemini' || category !== 'oauth-based') {
      geminiAIStudioOAuthEnabled.value = false
      return
    }
    const capabilities = await geminiOAuth.getCapabilities()
    geminiAIStudioOAuthEnabled.value = capabilities?.ai_studio_oauth_enabled === true
    if (!geminiAIStudioOAuthEnabled.value && geminiOAuthType.value === 'ai_studio') {
      geminiOAuthType.value = 'code_assist'
    }
  },
  { immediate: true }
)

watch(
  [accountCategory, () => form.platform],
  ([category, platform]) => {
    if (platform === 'openai' && category !== 'oauth-based') {
      codexCLIOnlyEnabled.value = false
      codexCLIOnlyAppServerEnabled.value = false
    }
    if (platform !== 'anthropic' || category !== 'apikey') {
      anthropicPassthroughEnabled.value = false
      anthropicAPIKeyAuthScheme.value = 'x_api_key'
      webSearchEmulationMode.value = 'default'
    }
  }
)

// Auto-fill related models when switching to whitelist mode or changing platform
watch(
  [modelRestrictionMode, () => form.platform],
  ([newMode]) => {
    if (newMode === 'whitelist') {
      allowedModels.value = [...getModelsByPlatform(form.platform)]
    }
  }
)

// Model mapping helpers
const addModelMapping = () => {
  modelMappings.value.push({ from: '', to: '' })
}

const addOpenAICompactModelMapping = () => {
  openAICompactModelMappings.value.push({ from: '', to: '' })
}

const removeOpenAICompactModelMapping = (index: number) => {
  openAICompactModelMappings.value.splice(index, 1)
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

const buildCurrentModelRestrictionMapping = () => {
  const mode = allowedModels.value.length > 0 && modelMappings.value.length > 0
    ? 'combined'
    : modelRestrictionMode.value
  return buildModelMappingObject(mode, allowedModels.value, modelMappings.value)
}

const applyCurrentModelReasoningEfforts = (credentials: Record<string, unknown>) => {
  const rawMapping = credentials.model_mapping
  if (!rawMapping || typeof rawMapping !== 'object' || Array.isArray(rawMapping)) {
    delete credentials.model_reasoning_efforts
    return
  }

  const modelMapping = rawMapping as Record<string, unknown>
  const configuredEfforts = buildModelReasoningEffortsObject(modelMappings.value)
  const reasoningEfforts = Object.fromEntries(
    Object.entries(configuredEfforts || {}).filter(([model]) => model in modelMapping)
  )
  if (Object.keys(reasoningEfforts).length > 0) {
    credentials.model_reasoning_efforts = reasoningEfforts
  } else {
    delete credentials.model_reasoning_efforts
  }
}

const addPresetMapping = (from: string, to: string) => {
  if (modelMappings.value.some((m) => m.from === from)) {
    appStore.showInfo(t('admin.accounts.mappingExists', { model: from }))
    return
  }
  modelMappings.value.push({ from, to })
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

const splitTempUnschedKeywords = (value: string) => {
  return value
    .split(/[,;]/)
    .map((item) => item.trim())
    .filter((item) => item.length > 0)
}

const submitCreateAccount = async (payload: CreateAccountRequest) => {
  submitting.value = true
  try {
    if (payload.platform === 'openai') {
      applyCurrentModelReasoningEfforts(payload.credentials)
    }
    const account = await adminAPI.accounts.create(payload)
    if (payload.type === 'apikey' && upstreamModelsPreviewed.value) {
      try {
        await adminAPI.accounts.syncUpstreamModels(account.id)
      } catch {
        appStore.showWarning(t('admin.accounts.syncUpstreamModelsFailed'))
      }
    }
    if (
      payload.type === 'apikey' &&
      payload.upstream_billing_probe_enabled === true
    ) {
      try {
        await adminAPI.accounts.probeUpstreamBilling(account.id)
      } catch {
        appStore.showWarning(t('admin.accounts.upstreamBilling.probeFailed'))
      }
    }
    appStore.showSuccess(t('admin.accounts.accountCreated'))
    emit('created')
    handleClose()
  } catch (error: any) {
    appStore.showError(error.response?.data?.message || error.response?.data?.detail || t('admin.accounts.failedToCreate'))
  } finally {
    submitting.value = false
  }
}

// Methods
const resetForm = () => {
  step.value = 1
  form.name = ''
  form.notes = ''
  form.platform = 'anthropic'
  form.type = 'oauth'
  upstreamModelsPreviewed.value = false
  form.credentials = {}
  form.proxy_id = null
  form.concurrency = 10
  form.load_factor = null
  form.priority = 1
  form.rate_multiplier = 1
  form.group_ids = []
  form.expires_at = null
  accountCategory.value = 'oauth-based'
  addMethod.value = 'oauth'
  accountMode.value = 'payg'
  apiProtocol.value = 'chat_completions'
  zhipuOrganization.value = ''
  zhipuProject.value = ''
  apiKeyBaseUrl.value = 'https://api.anthropic.com'
  apiKeyValue.value = ''
  upstreamBillingAutoProbeEnabled.value = true
  editQuotaLimit.value = null
  editQuotaDailyLimit.value = null
  editQuotaWeeklyLimit.value = null
  editDailyResetMode.value = null
  editDailyResetHour.value = null
  editWeeklyResetMode.value = null
  editWeeklyResetDay.value = null
  editWeeklyResetHour.value = null
  editResetTimezone.value = null
  modelMappings.value = []
  openAICompactModelMappings.value = []
  modelRestrictionMode.value = 'whitelist'
  allowedModels.value = [...claudeModels] // Default fill related models

  poolModeEnabled.value = false
  poolModeRetryCount.value = DEFAULT_POOL_MODE_RETRY_COUNT
  poolModeRetryStatusCodesInput.value = ''
  customErrorCodesEnabled.value = false
  selectedErrorCodes.value = []
  customErrorCodeInput.value = null
  headerOverrideEnabled.value = false
  headerOverrideRows.value = []
  grokOAuthCustomBaseUrlEnabled.value = false
  grokOAuthBaseUrl.value = ''
  interceptWarmupRequests.value = false
  autoPauseOnExpired.value = true
  openaiPassthroughEnabled.value = false
  openaiFlattenNamespacesEnabled.value = false
  openAILongContextBillingEnabled.value = false
  openAILongContextBillingTouched.value = false
  openAICompactMode.value = 'auto'
  openAIResponsesMode.value = 'auto'
  openAIEndpointCapabilities.value = ['chat_completions', 'embeddings']
  openaiOAuthResponsesWebSocketV2Mode.value = OPENAI_WS_MODE_OFF
  openaiAPIKeyResponsesWebSocketV2Mode.value = OPENAI_WS_MODE_OFF
  codexCLIOnlyEnabled.value = false
  codexCLIOnlyAppServerEnabled.value = false
  codexFingerprintMode.value = 'off'
  anthropicPassthroughEnabled.value = false
  anthropicAPIKeyAuthScheme.value = 'x_api_key'
  webSearchEmulationMode.value = 'default'
  // Reset quota control state
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
  vertexServiceAccountJson.value = ''
  vertexProjectId.value = ''
  vertexClientEmail.value = ''
  vertexLocation.value = 'global'
  tempUnschedEnabled.value = false
  tempUnschedRules.value = []
  geminiOAuthType.value = 'google_one'
  geminiTierGoogleOne.value = 'google_one_free'
  geminiTierGcp.value = 'gcp_standard'
  geminiTierAIStudio.value = 'aistudio_free'
  oauth.resetState()
  openaiOAuth.resetState()
  geminiOAuth.resetState()
  antigravityOAuth.resetState()
  grokOAuth.resetState()
  oauthFlowRef.value?.reset()
}

const handleClose = () => {
  emit('close')
}

const buildOpenAIExtra = (base?: Record<string, unknown>): Record<string, unknown> | undefined => {
  if (form.platform !== 'openai') {
    return base
  }

  const extra: Record<string, unknown> = { ...(base || {}) }
  if (accountCategory.value === 'oauth-based') {
    extra.openai_oauth_responses_websockets_v2_mode = openaiOAuthResponsesWebSocketV2Mode.value
    extra.openai_oauth_responses_websockets_v2_enabled = isOpenAIWSModeEnabled(openaiOAuthResponsesWebSocketV2Mode.value)
  } else if (accountCategory.value === 'apikey') {
    extra.openai_apikey_responses_websockets_v2_mode = openaiAPIKeyResponsesWebSocketV2Mode.value
    extra.openai_apikey_responses_websockets_v2_enabled = isOpenAIWSModeEnabled(openaiAPIKeyResponsesWebSocketV2Mode.value)
  }
  // 清理兼容旧键，统一改用分类型开关。
  delete extra.responses_websockets_v2_enabled
  delete extra.openai_ws_enabled
  if (openaiPassthroughEnabled.value) {
    extra.openai_passthrough = true
  } else {
    delete extra.openai_passthrough
    delete extra.openai_oauth_passthrough
  }
  // 缺省即保留 namespace，不写空值，避免 extra 里堆积默认项
  if (form.type === 'oauth' && openaiFlattenNamespacesEnabled.value) {
    extra.openai_responses_flatten_namespaces = true
  } else {
    delete extra.openai_responses_flatten_namespaces
  }
  extra.openai_long_context_billing_enabled = openAILongContextBillingEnabled.value

  if (accountCategory.value === 'oauth-based' && codexCLIOnlyEnabled.value) {
    extra.codex_cli_only = true
  } else {
    delete extra.codex_cli_only
  }
  delete extra.codex_cli_only_allowed_clients
  if (
    accountCategory.value === 'oauth-based' &&
    codexCLIOnlyEnabled.value &&
    codexCLIOnlyAppServerEnabled.value
  ) {
    extra.codex_cli_only_allow_app_server = true
  } else {
    delete extra.codex_cli_only_allow_app_server
  }
  // 收敛是显式 opt-in：off 即默认值，不落键；device/session/full 必须显式写入，
  // 否则管理员的选择会被当成默认而丢失（#5610）。
  if (codexFingerprintMode.value !== 'off') {
    extra.codex_fingerprint_mode = codexFingerprintMode.value
  } else {
    delete extra.codex_fingerprint_mode
  }
  if (openAICompactMode.value !== 'auto') {
    extra.openai_compact_mode = openAICompactMode.value
  } else {
    delete extra.openai_compact_mode
  }

  if (
    accountCategory.value === 'apikey' &&
    openAITextGenerationCapabilityEnabled.value &&
    openAIResponsesMode.value !== 'auto'
  ) {
    extra.openai_responses_mode = openAIResponsesMode.value
  } else {
    delete extra.openai_responses_mode
  }

  return Object.keys(extra).length > 0 ? extra : undefined
}

const buildOpenAICodexImportExtra = (): Record<string, unknown> | undefined => {
  const extra = buildOpenAIExtra()
  if (!extra) {
    return undefined
  }
  if (!openAILongContextBillingTouched.value) {
    delete extra.openai_long_context_billing_enabled
  }
  return Object.keys(extra).length > 0 ? extra : undefined
}

const buildAnthropicExtra = (base?: Record<string, unknown>): Record<string, unknown> | undefined => {
  if (form.platform !== 'anthropic' || accountCategory.value !== 'apikey') {
    return base
  }

  const extra: Record<string, unknown> = { ...(base || {}) }
  if (anthropicPassthroughEnabled.value) {
    extra.anthropic_passthrough = true
  } else {
    delete extra.anthropic_passthrough
  }
  if (anthropicAPIKeyAuthScheme.value === 'authorization_bearer') {
    extra.anthropic_apikey_auth_scheme = 'authorization_bearer'
  } else {
    delete extra.anthropic_apikey_auth_scheme
  }
  if (webSearchEmulationMode.value === 'default') {
    delete extra.web_search_emulation
  } else {
    extra.web_search_emulation = webSearchEmulationMode.value
  }

  return Object.keys(extra).length > 0 ? extra : undefined
}

const doCreateAccount = async (payload: CreateAccountRequest) => {
  await submitCreateAccount(payload)
}

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

const applyVertexServiceAccountJson = (value: string) => {
  const raw = value.trim()
  if (!raw) {
    vertexProjectId.value = ''
    vertexClientEmail.value = ''
    return false
  }
  try {
    const parsed = JSON.parse(raw) as Record<string, unknown>
    const projectId = typeof parsed.project_id === 'string' ? parsed.project_id.trim() : ''
    const clientEmail = typeof parsed.client_email === 'string' ? parsed.client_email.trim() : ''
    const privateKey = typeof parsed.private_key === 'string' ? parsed.private_key.trim() : ''
    if (!projectId || !clientEmail || !privateKey) {
      appStore.showError(t('admin.accounts.vertexSaJsonMissingFields'))
      return false
    }
    vertexProjectId.value = projectId
    vertexClientEmail.value = clientEmail
    vertexServiceAccountJson.value = JSON.stringify(parsed)
    return true
  } catch {
    appStore.showError(t('admin.accounts.vertexSaJsonInvalid'))
    return false
  }
}

const parseVertexServiceAccountJson = () => applyVertexServiceAccountJson(vertexServiceAccountJson.value)

const handleVertexServiceAccountFile = async (event: Event) => {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  if (!file) return
  try {
    applyVertexServiceAccountJson(await file.text())
  } finally {
    input.value = ''
  }
}

const handleVertexServiceAccountDrop = async (event: DragEvent) => {
  vertexServiceAccountDragActive.value = false
  const file = event.dataTransfer?.files?.[0]
  if (!file) return
  applyVertexServiceAccountJson(await file.text())
}

const handleSubmit = async () => {
  // For OAuth-based type, handle OAuth flow (goes to step 2)
  if (isOAuthFlow.value) {
    if (!isGrokSSOInputMethod.value && !form.name.trim()) {
      appStore.showError(t('admin.accounts.pleaseEnterAccountName'))
      return
    }
    step.value = 2
    return
  }

  // For Bedrock type, create directly
  if (form.platform === 'anthropic' && accountCategory.value === 'bedrock') {
    if (!form.name.trim()) {
      appStore.showError(t('admin.accounts.pleaseEnterAccountName'))
      return
    }

    const credentials: Record<string, unknown> = {
      auth_mode: bedrockAuthMode.value,
      aws_region: bedrockRegion.value.trim() || 'us-east-1',
    }

    if (bedrockAuthMode.value === 'sigv4') {
      if (!bedrockAccessKeyId.value.trim()) {
        appStore.showError(t('admin.accounts.bedrockAccessKeyIdRequired'))
        return
      }
      if (!bedrockSecretAccessKey.value.trim()) {
        appStore.showError(t('admin.accounts.bedrockSecretAccessKeyRequired'))
        return
      }
      credentials.aws_access_key_id = bedrockAccessKeyId.value.trim()
      credentials.aws_secret_access_key = bedrockSecretAccessKey.value.trim()
      if (bedrockSessionToken.value.trim()) {
        credentials.aws_session_token = bedrockSessionToken.value.trim()
      }
    } else {
      if (!bedrockApiKeyValue.value.trim()) {
        appStore.showError(t('admin.accounts.bedrockApiKeyRequired'))
        return
      }
      credentials.api_key = bedrockApiKeyValue.value.trim()
    }

    if (bedrockForceGlobal.value) {
      credentials.aws_force_global = 'true'
    }

    // Model mapping
    const modelMapping = buildCurrentModelRestrictionMapping()
    if (modelMapping) {
      credentials.model_mapping = modelMapping
    }

    // Pool mode
    if (poolModeEnabled.value) {
      credentials.pool_mode = true
      credentials.pool_mode_retry_count = normalizePoolModeRetryCount(poolModeRetryCount.value)
      const parsedRetryStatusCodes = parsePoolModeRetryStatusCodes(poolModeRetryStatusCodesInput.value)
      if (parsedRetryStatusCodes.length > 0) {
        credentials.pool_mode_retry_status_codes = parsedRetryStatusCodes
      }
    }

    applyInterceptWarmup(credentials, interceptWarmupRequests.value, 'create')

    await createAccountAndFinish('anthropic', 'bedrock' as AccountType, credentials)
    return
  }

  if ((form.platform === 'anthropic' || form.platform === 'gemini') && accountCategory.value === 'service_account') {
    if (!form.name.trim()) {
      appStore.showError(t('admin.accounts.pleaseEnterAccountName'))
      return
    }
    if (!parseVertexServiceAccountJson()) {
      return
    }
    if (!vertexLocation.value.trim()) {
      appStore.showError(t('admin.accounts.vertexLocationRequired'))
      return
    }
    const credentials: Record<string, unknown> = {
      service_account_json: vertexServiceAccountJson.value.trim(),
      project_id: vertexProjectId.value.trim(),
      client_email: vertexClientEmail.value.trim(),
      location: vertexLocation.value.trim(),
      tier_id: 'vertex'
    }
    await createAccountAndFinish(form.platform, 'service_account' as AccountType, credentials)
    return
  }

  // For apikey type, create directly
  if (!apiKeyValue.value.trim()) {
    appStore.showError(t('admin.accounts.pleaseEnterApiKey'))
    return
  }

  // Determine default base URL based on platform
  const defaultBaseUrl =
    form.platform === 'openai'
      ? 'https://api.openai.com'
      : form.platform === 'gemini'
        ? 'https://generativelanguage.googleapis.com'
        : form.platform === 'antigravity'
          ? 'https://cloudcode-pa.googleapis.com'
      : form.platform === 'grok'
          ? 'https://api.x.ai/v1'
          : 'https://api.anthropic.com'

  // Build credentials with optional model mapping
  const credentials: Record<string, unknown> = {
    base_url: apiKeyBaseUrl.value.trim() || defaultBaseUrl,
    api_key: apiKeyValue.value.trim()
  }
  if (form.platform === 'gemini') {
    credentials.tier_id = geminiSelectedTier.value
  }
  // 国产供应商：账号模式 + 协议 + 对应端点写入凭据；后端按 account_mode 路由
  // 额度/余额探测，按 api_protocol 路由转发端点与格式。注意 CN apikey 走本函数
  // 的通用路径（直接 doCreateAccount），不经过 createAccountAndFinish。
  if (form.platform === 'kimi' || form.platform === 'zhipu' || form.platform === 'deepseek') {
    credentials.account_mode = accountMode.value
    credentials.api_protocol = apiProtocol.value
    const resolvedCNBase = (
      apiKeyBaseUrl.value.trim() || defaultCNBaseUrl(form.platform, accountMode.value, apiProtocol.value)
    ).trim()
    if (resolvedCNBase) {
      credentials.base_url = resolvedCNBase
    }
    if (form.platform === 'zhipu' && accountMode.value === 'coding') {
      const organization = zhipuOrganization.value.trim()
      const project = zhipuProject.value.trim()
      if (organization) {
        credentials.zhipu_organization = organization
        if (project) credentials.zhipu_project = project
      }
    }
  }

  // Add model mapping if configured（OpenAI 开启自动透传时不应用）
  if (!isOpenAIModelRestrictionDisabled.value) {
    const modelMapping = buildCurrentModelRestrictionMapping()
    if (modelMapping) {
      credentials.model_mapping = modelMapping
    }
  }
  if (form.platform === 'openai') {
    applyOpenAIEndpointCapabilities(credentials)
    const compactModelMapping = buildOpenAICompactModelMapping()
    if (compactModelMapping) {
      credentials.compact_model_mapping = compactModelMapping
    }
  }

  // Add pool mode if enabled
  if (poolModeEnabled.value) {
    credentials.pool_mode = true
    credentials.pool_mode_retry_count = normalizePoolModeRetryCount(poolModeRetryCount.value)
    const parsedRetryStatusCodes = parsePoolModeRetryStatusCodes(poolModeRetryStatusCodesInput.value)
    if (parsedRetryStatusCodes.length > 0) {
      credentials.pool_mode_retry_status_codes = parsedRetryStatusCodes
    }
  }

  // Add custom error codes if enabled
  if (customErrorCodesEnabled.value) {
    credentials.custom_error_codes_enabled = true
    credentials.custom_error_codes = [...selectedErrorCodes.value]
  }

  // Add header override if enabled (anthropic/openai/grok apikey)
  if (isHeaderOverrideCapable(form.platform, 'apikey')) {
    if (headerOverrideEnabled.value) {
      const headerError = validateHeaderOverrideRows(headerOverrideRows.value)
      if (headerError) {
        appStore.showError(t(`admin.accounts.headerOverride.${headerError}`))
        return
      }
    }
    applyHeaderOverride(credentials, headerOverrideEnabled.value, headerOverrideRows.value, 'create')
  }

  applyInterceptWarmup(credentials, interceptWarmupRequests.value, 'create')
  if (!applyTempUnschedConfig(credentials)) {
    return
  }

  form.credentials = credentials
  const extra = buildAnthropicExtra(buildOpenAIExtra())

  await doCreateAccount({
    ...form,
    group_ids: form.group_ids,
    extra,
    upstream_billing_probe_enabled: upstreamBillingAutoProbeEnabled.value,
    auto_pause_on_expired: autoPauseOnExpired.value
  })
}

const goBackToBasicInfo = () => {
  step.value = 1
  oauth.resetState()
  openaiOAuth.resetState()
  geminiOAuth.resetState()
  antigravityOAuth.resetState()
  grokOAuth.resetState()
  oauthFlowRef.value?.reset()
}

const handleGenerateUrl = async () => {
  if (form.platform === 'openai') {
    await openaiOAuth.generateAuthUrl(form.proxy_id)
  } else if (form.platform === 'gemini') {
    await geminiOAuth.generateAuthUrl(
      form.proxy_id,
      oauthFlowRef.value?.projectId,
      geminiOAuthType.value,
      geminiSelectedTier.value
    )
  } else if (form.platform === 'antigravity') {
    await antigravityOAuth.generateAuthUrl(form.proxy_id)
  } else if (form.platform === 'grok') {
    await grokOAuth.generateAuthUrl(form.proxy_id)
  } else {
    await oauth.generateAuthUrl(addMethod.value, form.proxy_id)
  }
}

const handleValidateRefreshToken = (rt: string) => {
  if (form.platform === 'openai') {
    handleOpenAIValidateRT(rt)
  } else if (form.platform === 'antigravity') {
    handleAntigravityValidateRT(rt)
  } else if (form.platform === 'grok') {
    handleGrokValidateRT(rt)
  }
}

const handleValidateSessionToken = (_sessionToken: string) => {
  // Session token validation removed
}

const formatDateTimeLocal = formatDateTimeLocalInput
const parseDateTimeLocal = parseDateTimeLocalInput

// Create account and handle success/failure
const createAccountAndFinish = async (
  platform: AccountPlatform,
  type: AccountType,
  credentials: Record<string, unknown>,
  extra?: Record<string, unknown>
) => {
  if (!applyTempUnschedConfig(credentials)) {
    return
  }
  // Inject quota limits for apikey/bedrock accounts
  let finalExtra = extra
  if (type === 'apikey' || type === 'bedrock') {
    const quotaExtra: Record<string, unknown> = { ...(extra || {}) }
    if (editQuotaLimit.value != null && editQuotaLimit.value > 0) {
      quotaExtra.quota_limit = editQuotaLimit.value
    }
    if (editQuotaDailyLimit.value != null && editQuotaDailyLimit.value > 0) {
      quotaExtra.quota_daily_limit = editQuotaDailyLimit.value
    }
    if (editQuotaWeeklyLimit.value != null && editQuotaWeeklyLimit.value > 0) {
      quotaExtra.quota_weekly_limit = editQuotaWeeklyLimit.value
    }
    // Quota reset mode config
    if (editDailyResetMode.value === 'fixed') {
      quotaExtra.quota_daily_reset_mode = 'fixed'
      quotaExtra.quota_daily_reset_hour = editDailyResetHour.value ?? 0
    }
    if (editWeeklyResetMode.value === 'fixed') {
      quotaExtra.quota_weekly_reset_mode = 'fixed'
      quotaExtra.quota_weekly_reset_day = editWeeklyResetDay.value ?? 1
      quotaExtra.quota_weekly_reset_hour = editWeeklyResetHour.value ?? 0
    }
    if (editDailyResetMode.value === 'fixed' || editWeeklyResetMode.value === 'fixed') {
      quotaExtra.quota_reset_timezone = editResetTimezone.value || 'UTC'
    }
    // Quota notify config
    writeQuotaNotifyToExtra(quotaExtra, 'create')
    if (Object.keys(quotaExtra).length > 0) {
      finalExtra = quotaExtra
    }
  }
  if (platform === 'openai') {
    if (type === 'apikey') {
      applyOpenAIEndpointCapabilities(credentials)
    }
    const compactModelMapping = buildOpenAICompactModelMapping()
    if (compactModelMapping) {
      credentials.compact_model_mapping = compactModelMapping
    } else {
      delete credentials.compact_model_mapping
    }
  }
  if (platform === 'grok') {
    if (!credentials.base_url) {
      credentials.base_url = apiKeyBaseUrl.value.trim() || 'https://api.x.ai/v1'
    }
    const modelMapping = buildCurrentModelRestrictionMapping()
    if (modelMapping) {
      credentials.model_mapping = modelMapping
    } else {
      delete credentials.model_mapping
    }
  }
  await doCreateAccount({
    name: form.name,
    notes: form.notes,
    platform,
    type,
    credentials,
    extra: finalExtra,
    proxy_id: form.proxy_id,
    concurrency: form.concurrency,
    load_factor: form.load_factor ?? undefined,
    priority: form.priority,
    rate_multiplier: form.rate_multiplier,
    group_ids: form.group_ids,
    expires_at: form.expires_at,
    // 上游倍率探测对全部 API-key 平台开放；
    // 非 apikey 类型（bedrock/oauth）不传，后端不动作。
    upstream_billing_probe_enabled: type === 'apikey' ? upstreamBillingAutoProbeEnabled.value : undefined,
    auto_pause_on_expired: autoPauseOnExpired.value
  })
}

// Grok 手动 RT 批量验证和创建
const handleGrokValidateRT = async (refreshTokenInput: string) => {
  if (!refreshTokenInput.trim()) return

  const refreshTokens = refreshTokenInput
    .split('\n')
    .map((rt) => rt.trim())
    .filter((rt) => rt)

  if (refreshTokens.length === 0) {
    grokOAuth.error.value = t('admin.accounts.oauth.grok.pleaseEnterRefreshToken')
    return
  }
  if (!validateGrokOAuthUpstreamConfig()) return

  grokOAuth.loading.value = true
  grokOAuth.error.value = ''

  let successCount = 0
  let failedCount = 0
  const errors: string[] = []

  try {
    for (let i = 0; i < refreshTokens.length; i++) {
      try {
        const tokenInfo = await grokOAuth.validateRefreshToken(refreshTokens[i], form.proxy_id)
        if (!tokenInfo) {
          failedCount++
          errors.push(`#${i + 1}: ${grokOAuth.error.value || 'Validation failed'}`)
          grokOAuth.error.value = ''
          continue
        }

        const credentials = grokOAuth.buildCredentials(tokenInfo)
        applyGrokOAuthUpstreamConfig(credentials)
        const extra = grokOAuth.buildExtraInfo(tokenInfo)
        const accountName = refreshTokens.length > 1 ? `${form.name || tokenInfo.email || 'Grok OAuth Account'} #${i + 1}` : (form.name || tokenInfo.email || 'Grok OAuth Account')

        const modelMapping = buildCurrentModelRestrictionMapping()
        if (modelMapping) {
          credentials.model_mapping = modelMapping
        }
        if (!applyTempUnschedConfig(credentials)) {
          return
        }

        await adminAPI.accounts.create({
          name: accountName,
          notes: form.notes,
          platform: 'grok',
          type: 'oauth',
          credentials,
          extra,
          proxy_id: form.proxy_id,
          concurrency: form.concurrency,
          load_factor: form.load_factor ?? undefined,
          priority: form.priority,
          rate_multiplier: form.rate_multiplier,
          group_ids: form.group_ids,
          expires_at: form.expires_at,
          auto_pause_on_expired: autoPauseOnExpired.value
        })
        successCount++
      } catch (error: any) {
        failedCount++
        const errMsg = error.response?.data?.detail || error.message || 'Unknown error'
        errors.push(`#${i + 1}: ${errMsg}`)
      }
    }

    if (successCount > 0 && failedCount === 0) {
      appStore.showSuccess(
        refreshTokens.length > 1
          ? t('admin.accounts.oauth.batchSuccess', { count: successCount })
          : t('admin.accounts.accountCreated')
      )
      emit('created')
      handleClose()
    } else if (successCount > 0) {
      appStore.showWarning(t('admin.accounts.oauth.batchPartialSuccess', { success: successCount, failed: failedCount }))
      grokOAuth.error.value = errors.join('\n')
      emit('created')
    } else {
      grokOAuth.error.value = errors.join('\n')
      appStore.showError(t('admin.accounts.oauth.batchFailed'))
    }
  } finally {
    grokOAuth.loading.value = false
  }
}

const handleGrokImportSSO = async (ssoInput: string) => {
  // Align with OpenAI/Grok RT batch import: one token per line, no client-side dedupe.
  const ssoTokens = ssoInput
    .split('\n')
    .map((token) => token.trim())
    .filter((token) => token)
  if (ssoTokens.length === 0) return
  if (!validateGrokOAuthUpstreamConfig()) return

  grokOAuth.loading.value = true
  grokOAuth.error.value = ''

  const credentials: Record<string, unknown> = {}
  applyGrokOAuthUpstreamConfig(credentials)
  const modelMapping = buildCurrentModelRestrictionMapping()
  if (modelMapping) {
    credentials.model_mapping = modelMapping
  }
  if (!applyTempUnschedConfig(credentials)) {
    grokOAuth.loading.value = false
    return
  }

  try {
    const result = await adminAPI.grok.createFromSSO({
      sso_tokens: ssoTokens,
      name: form.name || undefined,
      notes: form.notes || undefined,
      proxy_id: form.proxy_id,
      group_ids: form.group_ids,
      credentials,
      concurrency: form.concurrency,
      load_factor: form.load_factor ?? undefined,
      priority: form.priority,
      rate_multiplier: form.rate_multiplier,
      expires_at: form.expires_at,
      auto_pause_on_expired: autoPauseOnExpired.value
    })

    const successCount = result.created?.length || 0
    const failedCount = result.failed?.length || 0
    if (successCount > 0 && failedCount === 0) {
      appStore.showSuccess(
        ssoTokens.length > 1
          ? t('admin.accounts.oauth.batchSuccess', { count: successCount })
          : t('admin.accounts.accountCreated')
      )
      emit('created')
      handleClose()
    } else if (successCount > 0 && failedCount > 0) {
      // Same as OpenAI/Grok RT: keep input, show failures, refresh list.
      appStore.showWarning(
        t('admin.accounts.oauth.batchPartialSuccess', { success: successCount, failed: failedCount })
      )
      grokOAuth.error.value = (result.failed || [])
        .map((item) => `#${item.index}: ${item.error || 'Unknown error'}`)
        .join('\n')
      emit('created')
    } else {
      grokOAuth.error.value = (result.failed || [])
        .map((item) => `#${item.index}: ${item.error || 'Unknown error'}`)
        .join('\n') || t('admin.accounts.oauth.grok.failedToConvertSSO')
      appStore.showError(t('admin.accounts.oauth.batchFailed'))
    }
  } catch (error: any) {
    grokOAuth.error.value = error.response?.data?.detail || error.message || t('admin.accounts.oauth.grok.failedToConvertSSO')
    appStore.showError(grokOAuth.error.value)
  } finally {
    grokOAuth.loading.value = false
  }
}

/**
 * Grok password login: each line is email----password.
 * Password is only used for the authorize API call; buildCredentials never stores it.
 */
const handleGrokAuthorizePassword = async (emailPasswordInput: string) => {
  if (!emailPasswordInput.trim()) return
  if (!validateGrokOAuthUpstreamConfig()) return

  const lines = emailPasswordInput
    .split('\n')
    // Keep the password portion byte-for-byte; trim is only for determining
    // whether this textarea line is blank.
    .filter((line) => line.trim() && line.includes('----'))

  if (lines.length === 0) {
    grokOAuth.error.value = t(
      'admin.accounts.oauth.grok.pleaseEnterPassword',
      'Please enter email----password (one per line)'
    )
    return
  }

  grokOAuth.loading.value = true
  grokOAuth.error.value = ''

  let successCount = 0
  let failedCount = 0
  const errors: string[] = []

  try {
    for (let i = 0; i < lines.length; i++) {
      try {
        const tokenInfo = await grokOAuth.authorizePassword(lines[i], form.proxy_id)
        if (!tokenInfo) {
          failedCount++
          errors.push(`#${i + 1}: ${grokOAuth.error.value || 'Authorization failed'}`)
          grokOAuth.error.value = ''
          continue
        }

        const credentials = grokOAuth.buildCredentials(tokenInfo)
        applyGrokOAuthUpstreamConfig(credentials)
        const extra = grokOAuth.buildExtraInfo(tokenInfo)
        const accountName =
          lines.length > 1
            ? `${form.name || tokenInfo.email || 'Grok OAuth Account'} #${i + 1}`
            : form.name || tokenInfo.email || 'Grok OAuth Account'

        const modelMapping = buildCurrentModelRestrictionMapping()
        if (modelMapping) {
          credentials.model_mapping = modelMapping
        }
        if (!applyTempUnschedConfig(credentials)) {
          return
        }

        await adminAPI.accounts.create({
          name: accountName,
          notes: form.notes,
          platform: 'grok',
          type: 'oauth',
          credentials,
          extra,
          proxy_id: form.proxy_id,
          concurrency: form.concurrency,
          load_factor: form.load_factor ?? undefined,
          priority: form.priority,
          rate_multiplier: form.rate_multiplier,
          group_ids: form.group_ids,
          expires_at: form.expires_at,
          auto_pause_on_expired: autoPauseOnExpired.value
        })
        successCount++
      } catch (error: any) {
        failedCount++
        const errMsg = error.response?.data?.detail || error.message || 'Unknown error'
        errors.push(`#${i + 1}: ${errMsg}`)
      }
    }

    if (successCount > 0 && failedCount === 0) {
      appStore.showSuccess(
        lines.length > 1
          ? t('admin.accounts.oauth.batchSuccess', { count: successCount })
          : t('admin.accounts.accountCreated')
      )
      emit('created')
      handleClose()
    } else if (successCount > 0) {
      appStore.showWarning(
        t('admin.accounts.oauth.batchPartialSuccess', {
          success: successCount,
          failed: failedCount
        })
      )
      grokOAuth.error.value = errors.join('\n')
      emit('created')
    } else {
      grokOAuth.error.value = errors.join('\n')
      appStore.showError(t('admin.accounts.oauth.batchFailed'))
    }
  } finally {
    grokOAuth.loading.value = false
  }
}

// OpenAI OAuth 授权码兑换
const handleOpenAIExchange = async (authCode: string) => {
  const oauthClient = openaiOAuth
  if (!authCode.trim() || !oauthClient.sessionId.value) return

  oauthClient.loading.value = true
  oauthClient.error.value = ''

  try {
    const stateToUse = (oauthFlowRef.value?.oauthState || oauthClient.oauthState.value || '').trim()
    if (!stateToUse) {
      oauthClient.error.value = t('admin.accounts.oauth.authFailed')
      appStore.showError(oauthClient.error.value)
      return
    }

    const tokenInfo = await oauthClient.exchangeAuthCode(
      authCode.trim(),
      oauthClient.sessionId.value,
      stateToUse,
      form.proxy_id
    )
    if (!tokenInfo) return

    const credentials = oauthClient.buildCredentials(tokenInfo)
    const oauthExtra = oauthClient.buildExtraInfo(tokenInfo) as Record<string, unknown> | undefined
    const extra = buildOpenAIExtra(oauthExtra)
    const shouldCreateOpenAI = form.platform === 'openai'

    // Add model mapping for OpenAI OAuth accounts（透传模式下不应用）
    if (shouldCreateOpenAI && !isOpenAIModelRestrictionDisabled.value) {
    const modelMapping = buildCurrentModelRestrictionMapping()
      if (modelMapping) {
        credentials.model_mapping = modelMapping
      }
    }
    if (shouldCreateOpenAI) {
      const compactModelMapping = buildOpenAICompactModelMapping()
      if (compactModelMapping) {
        credentials.compact_model_mapping = compactModelMapping
      }
    }

    // 应用临时不可调度配置
    if (!applyTempUnschedConfig(credentials)) {
      return
    }

    if (shouldCreateOpenAI) {
      applyCurrentModelReasoningEfforts(credentials)
      await adminAPI.accounts.create({
        name: form.name,
        notes: form.notes,
        platform: 'openai',
        type: 'oauth',
        credentials,
        extra,
        proxy_id: form.proxy_id,
        concurrency: form.concurrency,
        load_factor: form.load_factor ?? undefined,
        priority: form.priority,
        rate_multiplier: form.rate_multiplier,
        group_ids: form.group_ids,
        expires_at: form.expires_at,
        auto_pause_on_expired: autoPauseOnExpired.value
      })
      appStore.showSuccess(t('admin.accounts.accountCreated'))
    }

    emit('created')
    handleClose()
  } catch (error: any) {
    oauthClient.error.value = error.response?.data?.detail || t('admin.accounts.oauth.authFailed')
    appStore.showError(oauthClient.error.value)
  } finally {
    oauthClient.loading.value = false
  }
}

// OpenAI 手动 RT 批量验证和创建
// OpenAI Mobile RT client_id
const OPENAI_MOBILE_RT_CLIENT_ID = 'app_LlGpXReQgckcGGUo2JrYvtJK'

const buildOpenAICodexImportCredentialExtras = (): Record<string, unknown> | null => {
  const credentials: Record<string, unknown> = {}
  if (!isOpenAIModelRestrictionDisabled.value) {
      const modelMapping = buildCurrentModelRestrictionMapping()
    if (modelMapping) {
      credentials.model_mapping = modelMapping
    }
  }

  const compactModelMapping = buildOpenAICompactModelMapping()
  if (compactModelMapping) {
    credentials.compact_model_mapping = compactModelMapping
  }
  applyCurrentModelReasoningEfforts(credentials)

  if (!applyTempUnschedConfig(credentials)) {
    return null
  }
  return credentials
}

const formatCodexImportMessages = (messages?: CodexSessionImportMessage[]) => {
  return (messages || [])
    .map((item) => {
      const name = item.name ? ` ${item.name}` : ''
      return `#${item.index}${name}: ${item.message}`
    })
    .join('\n')
}

const isAgentIdentityImportContent = (content: string) => {
  const isAgentIdentityValue = (value: unknown): boolean => {
    if (Array.isArray(value)) return value.length > 0 && value.every(isAgentIdentityValue)
    if (!value || typeof value !== 'object') return false
    const record = value as Record<string, unknown>
    const authMode = record.auth_mode ?? record.authMode
    const agentIdentity = record.agent_identity ?? record.agentIdentity
    return (typeof authMode === 'string' && authMode.toLowerCase() === 'agentidentity')
      || (!!agentIdentity && typeof agentIdentity === 'object')
  }

  try {
    return isAgentIdentityValue(JSON.parse(content))
  } catch {
    const lines = content.split('\n').map((line) => line.trim()).filter(Boolean)
    if (lines.length === 0) return false
    try {
      return lines.every((line) => isAgentIdentityValue(JSON.parse(line)))
    } catch {
      return false
    }
  }
}

const handleOpenAIImportCodexSession = async (content: string) => {
  const oauthClient = openaiOAuth
  const trimmed = content.trim()
  if (!trimmed) {
    oauthClient.error.value = t('admin.accounts.oauth.openai.codexSessionEmpty')
    return
  }
  if (oauthFlowRef.value?.inputMethod === 'agent_identity' && !isAgentIdentityImportContent(trimmed)) {
    oauthClient.error.value = t('admin.accounts.oauth.openai.agentIdentityInvalid')
    return
  }

  const credentialExtras = buildOpenAICodexImportCredentialExtras()
  if (credentialExtras === null) {
    return
  }

  oauthClient.loading.value = true
  oauthClient.error.value = ''

  try {
    const extra = buildOpenAICodexImportExtra()
    const result = await adminAPI.accounts.importCodexSession({
      content: trimmed,
      name: form.name,
      notes: form.notes || null,
      proxy_id: form.proxy_id,
      concurrency: form.concurrency,
      load_factor: form.load_factor ?? undefined,
      priority: form.priority,
      rate_multiplier: form.rate_multiplier,
      group_ids: form.group_ids,
      expires_at: form.expires_at,
      auto_pause_on_expired: autoPauseOnExpired.value,
      credential_extras: Object.keys(credentialExtras).length > 0 ? credentialExtras : undefined,
      extra,
      update_existing: true
    })

    const successCount = result.created + result.updated
    const params = {
      created: result.created,
      updated: result.updated,
      skipped: result.skipped,
      failed: result.failed
    }

    if (successCount > 0 && result.failed === 0) {
      appStore.showSuccess(t('admin.accounts.oauth.openai.codexSessionImportSuccess', params))
      emit('created')
      handleClose()
      return
    }

    const errorText = formatCodexImportMessages(result.errors)
    const warningText = formatCodexImportMessages(result.warnings)
    oauthClient.error.value = [errorText, warningText].filter(Boolean).join('\n')

    if (result.failed === 0) {
      appStore.showWarning(t('admin.accounts.oauth.openai.codexSessionImportSuccess', params))
      return
    }

    if (successCount > 0) {
      appStore.showWarning(t('admin.accounts.oauth.openai.codexSessionImportPartial', params))
      emit('created')
      return
    }

    appStore.showError(t('admin.accounts.oauth.openai.codexSessionImportFailed'))
  } catch (error: any) {
    oauthClient.error.value =
      error.response?.data?.detail ||
      error.response?.data?.message ||
      error.message ||
      t('admin.accounts.oauth.openai.codexSessionImportFailed')
    appStore.showError(oauthClient.error.value)
  } finally {
    oauthClient.loading.value = false
  }
}

const handleOpenAIImportCodexPAT = async (accessToken: string) => {
  const oauthClient = openaiOAuth
  const trimmed = accessToken.trim()
  if (!trimmed) {
    oauthClient.error.value = t('admin.accounts.oauth.openai.codexPatEmpty')
    return
  }

  const credentialExtras = buildOpenAICodexImportCredentialExtras()
  if (credentialExtras === null) {
    return
  }

  oauthClient.loading.value = true
  oauthClient.error.value = ''

  try {
    const extra = buildOpenAICodexImportExtra()
    await adminAPI.accounts.createOpenAICodexPAT({
      access_token: trimmed,
      name: form.name,
      notes: form.notes || null,
      proxy_id: form.proxy_id,
      concurrency: form.concurrency,
      load_factor: form.load_factor ?? undefined,
      priority: form.priority,
      rate_multiplier: form.rate_multiplier,
      group_ids: form.group_ids,
      expires_at: form.expires_at,
      auto_pause_on_expired: autoPauseOnExpired.value,
      credential_extras: Object.keys(credentialExtras).length > 0 ? credentialExtras : undefined,
      extra
    })

    appStore.showSuccess(t('admin.accounts.messages.accountCreated'))
    emit('created')
    handleClose()
  } catch (error: any) {
    oauthClient.error.value =
      error.response?.data?.detail ||
      error.response?.data?.message ||
      error.message ||
      t('admin.accounts.oauth.openai.codexPatImportFailed')
    appStore.showError(oauthClient.error.value)
  } finally {
    oauthClient.loading.value = false
  }
}

// OpenAI RT 批量验证和创建（共享逻辑）
const handleOpenAIBatchRT = async (refreshTokenInput: string, clientId?: string) => {
  const oauthClient = openaiOAuth
  if (!refreshTokenInput.trim()) return

  const refreshTokens = refreshTokenInput
    .split('\n')
    .map((rt) => rt.trim())
    .filter((rt) => rt)

  if (refreshTokens.length === 0) {
    oauthClient.error.value = t('admin.accounts.oauth.openai.pleaseEnterRefreshToken')
    return
  }

  oauthClient.loading.value = true
  oauthClient.error.value = ''

  let successCount = 0
  let failedCount = 0
  const errors: string[] = []
  const shouldCreateOpenAI = form.platform === 'openai'

  try {
    for (let i = 0; i < refreshTokens.length; i++) {
      try {
        const tokenInfo = await oauthClient.validateRefreshToken(
          refreshTokens[i],
          form.proxy_id,
          clientId
        )
        if (!tokenInfo) {
          failedCount++
          errors.push(`#${i + 1}: ${oauthClient.error.value || 'Validation failed'}`)
          oauthClient.error.value = ''
          continue
        }

        const credentials = oauthClient.buildCredentials(tokenInfo)
        if (clientId) {
          credentials.client_id = clientId
        }
        const oauthExtra = oauthClient.buildExtraInfo(tokenInfo) as Record<string, unknown> | undefined
        const extra = buildOpenAIExtra(oauthExtra)

        // Add model mapping for OpenAI OAuth accounts（透传模式下不应用）
        if (shouldCreateOpenAI && !isOpenAIModelRestrictionDisabled.value) {
          const modelMapping = buildCurrentModelRestrictionMapping()
          if (modelMapping) {
            credentials.model_mapping = modelMapping
          }
        }
        if (shouldCreateOpenAI) {
          const compactModelMapping = buildOpenAICompactModelMapping()
          if (compactModelMapping) {
            credentials.compact_model_mapping = compactModelMapping
          }
        }

        // Generate account name; fallback to email if name is empty (ent schema requires NotEmpty)
        const baseName = form.name || tokenInfo.email || 'OpenAI OAuth Account'
        const accountName = refreshTokens.length > 1 ? `${baseName} #${i + 1}` : baseName

        if (shouldCreateOpenAI) {
          applyCurrentModelReasoningEfforts(credentials)
          await adminAPI.accounts.create({
            name: accountName,
            notes: form.notes,
            platform: 'openai',
            type: 'oauth',
            credentials,
            extra,
            proxy_id: form.proxy_id,
            concurrency: form.concurrency,
            load_factor: form.load_factor ?? undefined,
            priority: form.priority,
            rate_multiplier: form.rate_multiplier,
            group_ids: form.group_ids,
            expires_at: form.expires_at,
            auto_pause_on_expired: autoPauseOnExpired.value
          })
        }

        successCount++
      } catch (error: any) {
        failedCount++
        const errMsg = error.response?.data?.detail || error.message || 'Unknown error'
        errors.push(`#${i + 1}: ${errMsg}`)
      }
    }

    // Show results
    if (successCount > 0 && failedCount === 0) {
      appStore.showSuccess(
        refreshTokens.length > 1
          ? t('admin.accounts.oauth.batchSuccess', { count: successCount })
          : t('admin.accounts.accountCreated')
      )
      emit('created')
      handleClose()
    } else if (successCount > 0 && failedCount > 0) {
      appStore.showWarning(
        t('admin.accounts.oauth.batchPartialSuccess', { success: successCount, failed: failedCount })
      )
      oauthClient.error.value = errors.join('\n')
      emit('created')
    } else {
      oauthClient.error.value = errors.join('\n')
      appStore.showError(t('admin.accounts.oauth.batchFailed'))
    }
  } finally {
    oauthClient.loading.value = false
  }
}

// 手动输入 RT（Codex CLI client_id，默认）
const handleOpenAIValidateRT = (rt: string) => handleOpenAIBatchRT(rt)

// 手动输入 Mobile RT
const handleOpenAIValidateMobileRT = (rt: string) => handleOpenAIBatchRT(rt, OPENAI_MOBILE_RT_CLIENT_ID)

// Antigravity RT batch validation and account creation.
const handleAntigravityValidateRT = async (refreshTokenInput: string) => {
  const refreshTokens = refreshTokenInput.split('\n').map((value) => value.trim()).filter(Boolean)
  if (refreshTokens.length === 0) {
    antigravityOAuth.error.value = t('admin.accounts.oauth.antigravity.pleaseEnterRefreshToken')
    return
  }

  antigravityOAuth.loading.value = true
  antigravityOAuth.error.value = ''
  let successCount = 0
  const errors: string[] = []
  try {
    for (let index = 0; index < refreshTokens.length; index++) {
      const tokenInfo = await antigravityOAuth.validateRefreshToken(refreshTokens[index], form.proxy_id)
      if (!tokenInfo) {
        errors.push(`#${index + 1}: ${antigravityOAuth.error.value || 'Validation failed'}`)
        antigravityOAuth.error.value = ''
        continue
      }
      try {
        const credentials = antigravityOAuth.buildCredentials(tokenInfo, refreshTokens[index])
        const modelMapping = buildCurrentModelRestrictionMapping()
        if (modelMapping) credentials.model_mapping = modelMapping
        applyInterceptWarmup(credentials, interceptWarmupRequests.value, 'create')
        await adminAPI.accounts.create({
          name: refreshTokens.length > 1 ? `${form.name || tokenInfo.email || 'Antigravity OAuth Account'} #${index + 1}` : (form.name || tokenInfo.email || 'Antigravity OAuth Account'),
          notes: form.notes,
          platform: 'antigravity',
          type: 'oauth',
          credentials,
          proxy_id: form.proxy_id,
          concurrency: form.concurrency,
          load_factor: form.load_factor ?? undefined,
          priority: form.priority,
          rate_multiplier: form.rate_multiplier,
          group_ids: form.group_ids,
          expires_at: form.expires_at,
          auto_pause_on_expired: autoPauseOnExpired.value
        })
        successCount++
      } catch (error: any) {
        errors.push(`#${index + 1}: ${error.response?.data?.detail || error.message || 'Unknown error'}`)
      }
    }
    if (successCount > 0) {
      appStore.showSuccess(t('admin.accounts.oauth.batchSuccess', { count: successCount }))
      emit('created')
      if (errors.length === 0) handleClose()
    }
    if (errors.length > 0) antigravityOAuth.error.value = errors.join('\n')
  } finally {
    antigravityOAuth.loading.value = false
  }
}

const handleGeminiExchange = async (authCode: string) => {
  if (!authCode.trim() || !geminiOAuth.sessionId.value) return
  const state = oauthFlowRef.value?.oauthState || geminiOAuth.state.value
  if (!state) return
  const tokenInfo = await geminiOAuth.exchangeAuthCode({
    code: authCode.trim(),
    sessionId: geminiOAuth.sessionId.value,
    state,
    proxyId: form.proxy_id,
    oauthType: geminiOAuthType.value,
    tierId: geminiSelectedTier.value
  })
  if (!tokenInfo) return
  const credentials = geminiOAuth.buildCredentials(tokenInfo)
  const modelMapping = buildCurrentModelRestrictionMapping()
  if (modelMapping) credentials.model_mapping = modelMapping
  await createAccountAndFinish('gemini', 'oauth', credentials, geminiOAuth.buildExtraInfo(tokenInfo))
}

const handleAntigravityExchange = async (authCode: string) => {
  if (!authCode.trim() || !antigravityOAuth.sessionId.value) return
  const state = oauthFlowRef.value?.oauthState || antigravityOAuth.state.value
  if (!state) return
  const tokenInfo = await antigravityOAuth.exchangeAuthCode({
    code: authCode.trim(),
    sessionId: antigravityOAuth.sessionId.value,
    state,
    proxyId: form.proxy_id
  })
  if (!tokenInfo) return
  const credentials = antigravityOAuth.buildCredentials(tokenInfo)
  const modelMapping = buildCurrentModelRestrictionMapping()
  if (modelMapping) credentials.model_mapping = modelMapping
  applyInterceptWarmup(credentials, interceptWarmupRequests.value, 'create')
  await createAccountAndFinish('antigravity', 'oauth', credentials)
}

// Grok OAuth 授权码兑换
const handleGrokExchange = async (authCode: string) => {
  if (!authCode.trim() || !grokOAuth.sessionId.value) return
  if (!validateGrokOAuthUpstreamConfig()) return

  grokOAuth.loading.value = true
  grokOAuth.error.value = ''

  try {
    const stateFromInput = oauthFlowRef.value?.oauthState || ''
    const stateToUse = stateFromInput || grokOAuth.state.value
    if (!stateToUse) {
      grokOAuth.error.value = t('admin.accounts.oauth.authFailed')
      appStore.showError(grokOAuth.error.value)
      return
    }

    const tokenInfo = await grokOAuth.exchangeAuthCode({
      code: authCode.trim(),
      sessionId: grokOAuth.sessionId.value,
      state: stateToUse,
      proxyId: form.proxy_id
    })
    if (!tokenInfo) return

    const credentials = grokOAuth.buildCredentials(tokenInfo)
    applyGrokOAuthUpstreamConfig(credentials)
    const extra = grokOAuth.buildExtraInfo(tokenInfo)
    await createAccountAndFinish('grok', 'oauth', credentials, extra)
  } catch (error: any) {
    grokOAuth.error.value = error.response?.data?.detail || t('admin.accounts.oauth.authFailed')
    appStore.showError(grokOAuth.error.value)
  } finally {
    grokOAuth.loading.value = false
  }
}

// Anthropic OAuth 授权码兑换
const handleAnthropicExchange = async (authCode: string) => {
  if (!authCode.trim() || !oauth.sessionId.value) return

  oauth.loading.value = true
  oauth.error.value = ''

  try {
    const proxyConfig = form.proxy_id ? { proxy_id: form.proxy_id } : {}
    const endpoint =
      addMethod.value === 'oauth'
        ? '/admin/accounts/exchange-code'
        : '/admin/accounts/exchange-setup-token-code'

    const tokenInfo = await adminAPI.accounts.exchangeCode(endpoint, {
      session_id: oauth.sessionId.value,
      code: authCode.trim(),
      ...proxyConfig
    })

    // Build extra with quota control settings
    const baseExtra = oauth.buildExtraInfo(tokenInfo) || {}
    const extra: Record<string, unknown> = { ...baseExtra }

    // Add window cost limit settings
    if (windowCostEnabled.value && windowCostLimit.value != null && windowCostLimit.value > 0) {
      extra.window_cost_limit = windowCostLimit.value
      extra.window_cost_sticky_reserve = windowCostStickyReserve.value ?? 10
    }

    // Add session limit settings
    if (sessionLimitEnabled.value && maxSessions.value != null && maxSessions.value > 0) {
      extra.max_sessions = maxSessions.value
      extra.session_idle_timeout_minutes = sessionIdleTimeout.value ?? 5
    }

    // Add RPM limit settings
    if (rpmLimitEnabled.value) {
      const DEFAULT_BASE_RPM = 15
      extra.base_rpm = (baseRpm.value != null && baseRpm.value > 0)
        ? baseRpm.value
        : DEFAULT_BASE_RPM
      extra.rpm_strategy = rpmStrategy.value
      if (rpmStickyBuffer.value != null && rpmStickyBuffer.value > 0) {
        extra.rpm_sticky_buffer = rpmStickyBuffer.value
      }
    }

    // UMQ mode（独立于 RPM）
    if (userMsgQueueMode.value) {
      extra.user_msg_queue_mode = userMsgQueueMode.value
    }

    // Add TLS fingerprint settings
    if (tlsFingerprintEnabled.value) {
      extra.enable_tls_fingerprint = true
      if (tlsFingerprintProfileId.value) {
        extra.tls_fingerprint_profile_id = tlsFingerprintProfileId.value
      }
    }

    // Add session ID masking settings
    if (sessionIdMaskingEnabled.value) {
      extra.session_id_masking_enabled = true
    }

    // Add cache TTL override settings
    if (cacheTTLOverrideEnabled.value) {
      extra.cache_ttl_override_enabled = true
      extra.cache_ttl_override_target = cacheTTLOverrideTarget.value
    }

    // Add custom base URL settings
    if (customBaseUrlEnabled.value && customBaseUrl.value.trim()) {
      extra.custom_base_url_enabled = true
      extra.custom_base_url = customBaseUrl.value.trim()
    }

    const credentials: Record<string, unknown> = { ...tokenInfo }
    applyInterceptWarmup(credentials, interceptWarmupRequests.value, 'create')
    await createAccountAndFinish(form.platform, addMethod.value as AccountType, credentials, extra)
  } catch (error: any) {
    oauth.error.value = error.response?.data?.detail || t('admin.accounts.oauth.authFailed')
    appStore.showError(oauth.error.value)
  } finally {
    oauth.loading.value = false
  }
}

// 主入口：根据平台路由到对应处理函数
const handleExchangeCode = async () => {
  const authCode = oauthFlowRef.value?.authCode || ''

  switch (form.platform) {
    case 'openai':
      return handleOpenAIExchange(authCode)
    case 'gemini':
      return handleGeminiExchange(authCode)
    case 'antigravity':
      return handleAntigravityExchange(authCode)
    case 'grok':
      return handleGrokExchange(authCode)
    default:
      return handleAnthropicExchange(authCode)
  }
}

const handleCookieAuth = async (sessionKey: string) => {
  oauth.loading.value = true
  oauth.error.value = ''

  try {
    const proxyConfig = form.proxy_id ? { proxy_id: form.proxy_id } : {}
    const keys = oauth.parseSessionKeys(sessionKey)

    if (keys.length === 0) {
      oauth.error.value = t('admin.accounts.oauth.pleaseEnterSessionKey')
      return
    }

    const tempUnschedPayload = tempUnschedEnabled.value
      ? buildTempUnschedRules(tempUnschedRules.value)
      : []
    if (tempUnschedEnabled.value && tempUnschedPayload.length === 0) {
      appStore.showError(t('admin.accounts.tempUnschedulable.rulesInvalid'))
      return
    }

    const endpoint =
      addMethod.value === 'oauth'
        ? '/admin/accounts/cookie-auth'
        : '/admin/accounts/setup-token-cookie-auth'

    let successCount = 0
    let failedCount = 0
    const errors: string[] = []

    for (let i = 0; i < keys.length; i++) {
      try {
        const tokenInfo = await adminAPI.accounts.exchangeCode(endpoint, {
          session_id: '',
          code: keys[i],
          ...proxyConfig
        })

        // Build extra with quota control settings
        const baseExtra = oauth.buildExtraInfo(tokenInfo) || {}
        const extra: Record<string, unknown> = { ...baseExtra }

        // Add window cost limit settings
        if (windowCostEnabled.value && windowCostLimit.value != null && windowCostLimit.value > 0) {
          extra.window_cost_limit = windowCostLimit.value
          extra.window_cost_sticky_reserve = windowCostStickyReserve.value ?? 10
        }

        // Add session limit settings
        if (sessionLimitEnabled.value && maxSessions.value != null && maxSessions.value > 0) {
          extra.max_sessions = maxSessions.value
          extra.session_idle_timeout_minutes = sessionIdleTimeout.value ?? 5
        }

        // Add RPM limit settings
        if (rpmLimitEnabled.value) {
          const DEFAULT_BASE_RPM = 15
          extra.base_rpm = (baseRpm.value != null && baseRpm.value > 0)
            ? baseRpm.value
            : DEFAULT_BASE_RPM
          extra.rpm_strategy = rpmStrategy.value
          if (rpmStickyBuffer.value != null && rpmStickyBuffer.value > 0) {
            extra.rpm_sticky_buffer = rpmStickyBuffer.value
          }
        }

        // UMQ mode（独立于 RPM）
        if (userMsgQueueMode.value) {
          extra.user_msg_queue_mode = userMsgQueueMode.value
        }

        // Add TLS fingerprint settings
        if (tlsFingerprintEnabled.value) {
          extra.enable_tls_fingerprint = true
          if (tlsFingerprintProfileId.value) {
            extra.tls_fingerprint_profile_id = tlsFingerprintProfileId.value
          }
        }

        // Add session ID masking settings
        if (sessionIdMaskingEnabled.value) {
          extra.session_id_masking_enabled = true
        }

        // Add cache TTL override settings
        if (cacheTTLOverrideEnabled.value) {
          extra.cache_ttl_override_enabled = true
          extra.cache_ttl_override_target = cacheTTLOverrideTarget.value
        }

        // Add custom base URL settings
        if (customBaseUrlEnabled.value && customBaseUrl.value.trim()) {
          extra.custom_base_url_enabled = true
          extra.custom_base_url = customBaseUrl.value.trim()
        }

        const accountName = keys.length > 1 ? `${form.name} #${i + 1}` : form.name

        const credentials: Record<string, unknown> = { ...tokenInfo }
        applyInterceptWarmup(credentials, interceptWarmupRequests.value, 'create')
        if (tempUnschedEnabled.value) {
          credentials.temp_unschedulable_enabled = true
          credentials.temp_unschedulable_rules = tempUnschedPayload
        }

        await adminAPI.accounts.create({
          name: accountName,
          notes: form.notes,
          platform: form.platform,
          type: addMethod.value, // Use addMethod as type: 'oauth' or 'setup-token'
          credentials,
          extra,
          proxy_id: form.proxy_id,
          concurrency: form.concurrency,
          load_factor: form.load_factor ?? undefined,
          priority: form.priority,
          rate_multiplier: form.rate_multiplier,
          group_ids: form.group_ids,
          expires_at: form.expires_at,
          auto_pause_on_expired: autoPauseOnExpired.value
        })

        successCount++
      } catch (error: any) {
        failedCount++
        errors.push(
          t('admin.accounts.oauth.keyAuthFailed', {
            index: i + 1,
            error: error.response?.data?.detail || t('admin.accounts.oauth.authFailed')
          })
        )
      }
    }

    if (successCount > 0) {
      appStore.showSuccess(t('admin.accounts.oauth.successCreated', { count: successCount }))
      if (failedCount === 0) {
        emit('created')
        handleClose()
      } else {
        emit('created')
      }
    }

    if (failedCount > 0) {
      oauth.error.value = errors.join('\n')
    }
  } catch (error: any) {
    oauth.error.value = error.response?.data?.detail || t('admin.accounts.oauth.cookieAuthFailed')
  } finally {
    oauth.loading.value = false
  }
}
</script>
