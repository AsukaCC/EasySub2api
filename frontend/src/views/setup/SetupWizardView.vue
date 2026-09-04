<template>
  <div
    class="views-setup-setup-wizard-view__panel"
  >
    <div class="views-setup-setup-wizard-view__panel-2">
      <!-- Logo & Title -->
      <div class="views-setup-setup-wizard-view__panel-3">
        <div
          class="views-setup-setup-wizard-view__panel-4"
        >
          <Icon name="cog" size="xl" class="views-setup-setup-wizard-view__icon" />
        </div>
        <h1 class="views-setup-setup-wizard-view__heading">{{ t('setup.title') }}</h1>
        <p class="views-setup-setup-wizard-view__description">{{ t('setup.description') }}</p>
      </div>

      <!-- Progress Steps -->
      <div class="views-setup-setup-wizard-view__panel-5">
        <div class="views-setup-setup-wizard-view__panel-6">
          <template v-for="(step, index) in steps" :key="step.id">
            <div class="views-setup-setup-wizard-view__panel-7">
              <div
                :class="[
                  'views-setup-setup-wizard-view__panel-19',
                  currentStep > index
                    ? 'views-setup-setup-wizard-view__panel-20'
                    : currentStep === index
                      ? 'views-setup-setup-wizard-view__panel-21'
                      : 'views-setup-setup-wizard-view__panel-22'
                ]"
              >
                <Icon
                  v-if="currentStep > index"
                  name="check"
                  size="md"
                  :stroke-width="2"
                />
                <span v-else>{{ index + 1 }}</span>
              </div>
              <span
                class="views-setup-setup-wizard-view__text"
                :class="
                  currentStep >= index
                    ? 'views-setup-setup-wizard-view__description-5'
                    : 'views-setup-setup-wizard-view__text-2'
                "
              >
                {{ step.title }}
              </span>
            </div>
            <div
              v-if="index < steps.length - 1"
              class="views-setup-setup-wizard-view__steps-length"
              :class="currentStep > index ? 'views-setup-setup-wizard-view__steps-length-2' : 'views-setup-setup-wizard-view__steps-length-3'"
            ></div>
          </template>
        </div>
      </div>

      <!-- Step Content -->
      <div class="views-setup-setup-wizard-view__panel-8">
        <!-- Step 1: Database -->
        <div v-if="currentStep === 0" class="views-setup-setup-wizard-view__panel-9">
          <div class="views-setup-setup-wizard-view__panel-10">
            <h2 class="views-setup-setup-wizard-view__heading-2">
              {{ t('setup.database.title') }}
            </h2>
            <p class="views-setup-setup-wizard-view__description-2">
              {{ t('setup.database.description') }}
            </p>
          </div>

          <div class="views-setup-setup-wizard-view__panel-11">
            <div>
              <label class="input-label">{{ t('setup.database.host') }}</label>
              <input
                v-model="formData.database.host"
                type="text"
                class="input"
                placeholder="localhost"
              />
            </div>
            <div>
              <label class="input-label">{{ t('setup.database.port') }}</label>
              <input
                v-model.number="formData.database.port"
                type="number"
                class="input"
                placeholder="5432"
              />
            </div>
          </div>

          <div class="views-setup-setup-wizard-view__panel-12">
            <div>
              <p class="views-setup-setup-wizard-view__description-3">
                {{ t("setup.redis.enableTls") }}
              </p>
              <p class="views-setup-setup-wizard-view__description-4">
                {{ t("setup.redis.enableTlsHint") }}
              </p>
            </div>
            <Toggle v-model="formData.redis.enable_tls" />
          </div>

          <div class="views-setup-setup-wizard-view__panel-11">
            <div>
              <label class="input-label">{{ t('setup.database.username') }}</label>
              <input
                v-model="formData.database.user"
                type="text"
                class="input"
                placeholder="postgres"
              />
            </div>
            <div>
              <label class="input-label">{{ t('setup.database.password') }}</label>
              <input
                v-model="formData.database.password"
                type="password"
                class="input"
                :placeholder="t('setup.database.passwordPlaceholder')"
              />
            </div>
          </div>

          <div class="views-setup-setup-wizard-view__panel-11">
            <div>
              <label class="input-label">{{ t('setup.database.databaseName') }}</label>
              <input
                v-model="formData.database.dbname"
                type="text"
                class="input"
                placeholder="easysub2api"
              />
            </div>
            <div>
              <label class="input-label">{{ t('setup.database.sslMode') }}</label>
              <Select
                v-model="formData.database.sslmode"
                :options="[
                  { value: 'disable', label: t('setup.database.ssl.disable') },
                  { value: 'require', label: t('setup.database.ssl.require') },
                  { value: 'verify-ca', label: t('setup.database.ssl.verifyCa') },
                  { value: 'verify-full', label: t('setup.database.ssl.verifyFull') }
                ]"
              />
            </div>
          </div>

          <button :aria-busy="testingDb"
            @click="testDatabaseConnection"
            :disabled="testingDb"
            class="views-setup-setup-wizard-view__action btn btn-secondary"
          >
            <LoadingButtonContent :loading="testingDb" :loading-text="t('setup.status.testing')">
            <Icon v-if="dbConnected" name="check" size="md" class="views-setup-setup-wizard-view__icon-3" :stroke-width="2" />
                        {{ dbConnected
                              ? t('setup.status.success')
                              : t('setup.status.testConnection') }}
            </LoadingButtonContent>
          </button>
        </div>

        <!-- Step 2: Redis -->
        <div v-if="currentStep === 1" class="views-setup-setup-wizard-view__panel-9">
          <div class="views-setup-setup-wizard-view__panel-10">
            <h2 class="views-setup-setup-wizard-view__heading-2">
              {{ t('setup.redis.title') }}
            </h2>
            <p class="views-setup-setup-wizard-view__description-2">
              {{ t('setup.redis.description') }}
            </p>
          </div>

          <div class="views-setup-setup-wizard-view__panel-11">
            <div>
              <label class="input-label">{{ t('setup.redis.host') }}</label>
              <input
                v-model="formData.redis.host"
                type="text"
                class="input"
                placeholder="localhost"
              />
            </div>
            <div>
              <label class="input-label">{{ t('setup.redis.port') }}</label>
              <input
                v-model.number="formData.redis.port"
                type="number"
                class="input"
                placeholder="6379"
              />
            </div>
          </div>

          <div class="views-setup-setup-wizard-view__panel-11">
            <div>
              <label class="input-label">{{ t('setup.redis.username') }}</label>
              <input
                v-model="formData.redis.username"
                type="text"
                class="input"
                :placeholder="t('setup.redis.usernamePlaceholder')"
              />
            </div>
            <div>
              <label class="input-label">{{ t('setup.redis.password') }}</label>
              <input
                v-model="formData.redis.password"
                type="password"
                class="input"
                :placeholder="t('setup.redis.passwordPlaceholder')"
              />
            </div>
            <div>
              <label class="input-label">{{ t('setup.redis.database') }}</label>
              <input
                v-model.number="formData.redis.db"
                type="number"
                class="input"
                placeholder="0"
              />
            </div>
          </div>

          <div class="views-setup-setup-wizard-view__panel-12">
            <div>
              <p class="views-setup-setup-wizard-view__description-3">
                {{ t("setup.redis.enableTls") }}
              </p>
              <p class="views-setup-setup-wizard-view__description-4">
                {{ t("setup.redis.enableTlsHint") }}
              </p>
            </div>
            <Toggle v-model="formData.redis.enable_tls" />
          </div>

          <button :aria-busy="testingRedis"
            @click="testRedisConnection"
            :disabled="testingRedis"
            class="views-setup-setup-wizard-view__action btn btn-secondary"
          >
            <LoadingButtonContent :loading="testingRedis" :loading-text="t('setup.status.testing')">
            <Icon
                          v-if="redisConnected"
                          name="check"
                          size="md"
                          class="views-setup-setup-wizard-view__icon-3"
                          :stroke-width="2"
                        />
                        {{ redisConnected
                              ? t('setup.status.success')
                              : t('setup.status.testConnection') }}
            </LoadingButtonContent>
          </button>
        </div>

        <!-- Step 3: Admin -->
        <div v-if="currentStep === 2" class="views-setup-setup-wizard-view__panel-9">
          <div class="views-setup-setup-wizard-view__panel-10">
            <h2 class="views-setup-setup-wizard-view__heading-2">
              {{ t('setup.admin.title') }}
            </h2>
            <p class="views-setup-setup-wizard-view__description-2">
              {{ t('setup.admin.description') }}
            </p>
          </div>

          <div>
            <label class="input-label">{{ t('setup.admin.email') }}</label>
            <input
              v-model="formData.admin.email"
              type="email"
              class="input"
              placeholder="admin@example.com"
            />
          </div>

          <div>
            <label class="input-label">{{ t('setup.admin.password') }}</label>
            <input
              v-model="formData.admin.password"
              type="password"
              class="input"
              :placeholder="t('setup.admin.passwordPlaceholder')"
            />
          </div>

          <div>
            <label class="input-label">{{ t('setup.admin.confirmPassword') }}</label>
            <input
              v-model="confirmPassword"
              type="password"
              class="input"
              :placeholder="t('setup.admin.confirmPasswordPlaceholder')"
            />
            <p
              v-if="confirmPassword && formData.admin.password !== confirmPassword"
              class="input-error-text"
            >
              {{ t('setup.admin.passwordMismatch') }}
            </p>
          </div>
        </div>

        <!-- Step 4: Complete -->
        <div v-if="currentStep === 3" class="views-setup-setup-wizard-view__panel-9">
          <div class="views-setup-setup-wizard-view__panel-10">
            <h2 class="views-setup-setup-wizard-view__heading-2">
              {{ t('setup.ready.title') }}
            </h2>
            <p class="views-setup-setup-wizard-view__description-2">
              {{ t('setup.ready.description') }}
            </p>
          </div>

          <div class="views-setup-setup-wizard-view__panel-13">
            <div class="views-setup-setup-wizard-view__panel-14">
              <h3 class="views-setup-setup-wizard-view__heading-3">
                {{ t('setup.ready.database') }}
              </h3>
              <p class="views-setup-setup-wizard-view__description-5">
                {{ formData.database.user }}@{{ formData.database.host }}:{{
                  formData.database.port
                }}/{{ formData.database.dbname }}
              </p>
            </div>

            <div class="views-setup-setup-wizard-view__panel-14">
              <h3 class="views-setup-setup-wizard-view__heading-3">
                {{ t('setup.ready.redis') }}
              </h3>
              <p class="views-setup-setup-wizard-view__description-5">
                {{ formData.redis.host }}:{{ formData.redis.port }}
              </p>
            </div>

            <div class="views-setup-setup-wizard-view__panel-14">
              <h3 class="views-setup-setup-wizard-view__heading-3">
                {{ t('setup.ready.adminEmail') }}
              </h3>
              <p class="views-setup-setup-wizard-view__description-5">{{ formData.admin.email }}</p>
            </div>
          </div>
        </div>

        <!-- Error Message -->
        <div
          v-if="errorMessage"
          class="views-setup-setup-wizard-view__panel-15"
        >
          <div class="views-setup-setup-wizard-view__panel-16">
            <Icon name="exclamationCircle" size="md" class="views-setup-setup-wizard-view__icon-4" />
            <p class="views-setup-setup-wizard-view__description-6">{{ errorMessage }}</p>
          </div>
        </div>

        <!-- Success Message -->
        <div
          v-if="installSuccess"
          class="views-setup-setup-wizard-view__panel-17"
        >
          <div class="views-setup-setup-wizard-view__panel-16">
            <LoadingSpinner v-if="!serviceReady" size="sm" color="inherit" decorative />
            <Icon v-else name="checkCircle" size="md" class="views-setup-setup-wizard-view__icon-6" />
            <div>
              <p class="views-setup-setup-wizard-view__description-7">
                {{ t('setup.status.completed') }}
              </p>
              <p class="views-setup-setup-wizard-view__description-8">
                {{
                  serviceReady
                    ? t('setup.status.redirecting')
                    : t('setup.status.restarting')
                }}
              </p>
            </div>
          </div>
        </div>

        <!-- Navigation Buttons -->
        <div class="views-setup-setup-wizard-view__panel-18">
          <button
            v-if="currentStep > 0 && !installSuccess"
            @click="currentStep--"
            class="btn btn-secondary"
          >
            <Icon name="chevronLeft" size="sm" class="views-setup-setup-wizard-view__icon-7" :stroke-width="2" />
            {{ t('common.back') }}
          </button>
          <div v-else></div>

          <button
            v-if="currentStep < 3"
            @click="nextStep"
            :disabled="!canProceed"
            class="btn btn-primary"
          >
            {{ t('common.next') }}
            <Icon name="chevronRight" size="sm" class="views-setup-setup-wizard-view__icon-8" :stroke-width="2" />
          </button>

          <button :aria-busy="installing"
            v-else-if="!installSuccess"
            @click="performInstall"
            :disabled="installing"
            class="btn btn-primary"
          >
            <LoadingButtonContent :loading="installing" :loading-text="t('setup.status.installing')">
            {{ t('setup.status.completeInstallation') }}
            </LoadingButtonContent>
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import LoadingButtonContent from '@/components/common/LoadingButtonContent.vue'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import { ref, reactive, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { testDatabase, testRedis, install, type InstallRequest } from '@/api/setup'
import { buildGatewayUrl } from '@/api/client'
import Select from '@/components/common/Select.vue'
import Toggle from '@/components/common/Toggle.vue'
import Icon from '@/components/icons/Icon.vue'

const { t } = useI18n()

const steps = computed(() => [
  { id: 'database', title: t('setup.database.title') },
  { id: 'redis', title: t('setup.redis.title') },
  { id: 'admin', title: t('setup.admin.title') },
  { id: 'complete', title: t('setup.ready.title') }
])

const currentStep = ref(0)
const errorMessage = ref('')
const installSuccess = ref(false)

// Connection test states
const testingDb = ref(false)
const testingRedis = ref(false)
const dbConnected = ref(false)
const redisConnected = ref(false)
const installing = ref(false)
const confirmPassword = ref('')
const serviceReady = ref(false)

// Default server port
const getCurrentPort = (): number => {
  const port = window.location.port
  if (port) {
    return parseInt(port, 10)
  }

  return window.location.protocol === 'https:' ? 443 : 80
}

const formData = reactive<InstallRequest>({
  database: {
    host: 'localhost',
    port: 5432,
    user: 'postgres',
    password: '',
    dbname: 'easysub2api',
    sslmode: 'disable'
  },
  redis: {
    host: 'localhost',
    port: 6379,
    username: '',
    password: '',
    db: 0,
    enable_tls: false
  },
  admin: {
    email: '',
    password: ''
  },
  server: {
    host: '0.0.0.0',
    port: getCurrentPort(), // Use current port from browser
    mode: 'release'
  }
})

const canProceed = computed(() => {
  switch (currentStep.value) {
    case 0:
      return dbConnected.value
    case 1:
      return redisConnected.value
    case 2:
      return (
        formData.admin.email &&
        formData.admin.password.length >= 8 &&
        formData.admin.password === confirmPassword.value
      )
    default:
      return true
  }
})

async function testDatabaseConnection() {
  testingDb.value = true
  errorMessage.value = ''
  dbConnected.value = false

  try {
    await testDatabase(formData.database)
    dbConnected.value = true
  } catch (error: unknown) {
    const err = error as { response?: { data?: { detail?: string; message?: string } }; message?: string }
    errorMessage.value =
      err.response?.data?.detail || err.response?.data?.message || err.message || 'Connection failed'
  } finally {
    testingDb.value = false
  }
}

async function testRedisConnection() {
  testingRedis.value = true
  errorMessage.value = ''
  redisConnected.value = false

  try {
    await testRedis(formData.redis)
    redisConnected.value = true
  } catch (error: unknown) {
    const err = error as { response?: { data?: { detail?: string; message?: string } }; message?: string }
    errorMessage.value =
      err.response?.data?.detail || err.response?.data?.message || err.message || 'Connection failed'
  } finally {
    testingRedis.value = false
  }
}

function nextStep() {
  if (canProceed.value) {
    errorMessage.value = ''
    currentStep.value++
  }
}

async function performInstall() {
  installing.value = true
  errorMessage.value = ''

  try {
    await install(formData)
    installSuccess.value = true
    // Start polling for service restart
    waitForServiceRestart()
  } catch (error: unknown) {
    const err = error as { response?: { data?: { detail?: string; message?: string } }; message?: string }
    errorMessage.value =
      err.response?.data?.detail || err.response?.data?.message || err.message || 'Installation failed'
  } finally {
    installing.value = false
  }
}

// Wait for service to restart and become available
async function waitForServiceRestart() {
  const maxAttempts = 60 // Increase to 60 attempts, ~60 seconds max
  const interval = 1000 // 1 second between attempts

  // Wait a moment for the service to start restarting
  await new Promise((resolve) => setTimeout(resolve, 3000))

  for (let attempt = 0; attempt < maxAttempts; attempt++) {
    try {
      // Use setup status endpoint as it tells us the real mode
      // Service might return 404 or connection refused while restarting
      const response = await fetch(buildGatewayUrl('/setup/status'), {
        method: 'GET',
        cache: 'no-store'
      })

      if (response.ok) {
        const data = await response.json()
        // If needs_setup is false, service has restarted in normal mode
        if (data.data && !data.data.needs_setup) {
          serviceReady.value = true
          // Redirect to login page after a short delay
          setTimeout(() => {
            window.location.href = '/login'
          }, 1500)
          return
        }
      }
    } catch {
      // Service not ready or network error during restart, continue polling
    }

    await new Promise((resolve) => setTimeout(resolve, interval))
  }

  // If we reach here, service didn't restart in time
  // Show a message to refresh manually
  errorMessage.value = t('setup.status.timeout')
}
</script>
