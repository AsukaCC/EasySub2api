<template>
  <div ref="containerRef" class="views-admin-settings-open-aifast-policy-user-selector__panel">
    <div v-if="selectedUserIds.length > 0" class="views-admin-settings-open-aifast-policy-user-selector__panel-2">
      <span
        v-for="userId in selectedUserIds"
        :key="userId"
        class="views-admin-settings-open-aifast-policy-user-selector__text"
      >
        <span class="views-admin-settings-open-aifast-policy-user-selector__text-2" :title="selectedUserLabel(userId)">
          {{ selectedUserLabel(userId) }}
        </span>
        <span class="views-admin-settings-open-aifast-policy-user-selector__text-3">#{{ userId }}</span>
        <span
          v-if="selectedUsers[userId]?.deleted"
          class="views-admin-settings-open-aifast-policy-user-selector__text-3"
        >
          {{ t("admin.settings.openaiFastPolicy.userDeleted") }}
        </span>
        <button
          type="button"
          class="views-admin-settings-open-aifast-policy-user-selector__action"
          :aria-label="t('admin.settings.openaiFastPolicy.removeUser')"
          :title="t('admin.settings.openaiFastPolicy.removeUser')"
          @click="removeUser(userId)"
        >
          <Icon name="x" size="xs" :stroke-width="2" />
        </button>
      </span>
    </div>

    <div class="views-admin-settings-open-aifast-policy-user-selector__panel">
      <Icon
        name="search"
        size="sm"
        class="views-admin-settings-open-aifast-policy-user-selector__icon"
      />
      <input
        v-model="searchQuery"
        type="text"
        autocomplete="off"
        class="views-admin-settings-open-aifast-policy-user-selector__field input input-sm"
        :placeholder="t('admin.settings.openaiFastPolicy.userSearchPlaceholder')"
        @input="debounceSearch"
        @focus="showDropdown = true"
      />
    </div>

    <div
      v-if="showDropdown && searchQuery.trim()"
      class="views-admin-settings-open-aifast-policy-user-selector__panel-3"
    >
      <div v-if="searchLoading" class="views-admin-settings-open-aifast-policy-user-selector__panel-4">
        {{ t("common.loading") }}
      </div>
      <div
        v-else-if="availableResults.length === 0"
        class="views-admin-settings-open-aifast-policy-user-selector__panel-4"
      >
        {{ t("admin.settings.openaiFastPolicy.userSearchEmpty") }}
      </div>
      <template v-else>
        <button
          v-for="user in availableResults"
          :key="user.id"
          type="button"
          class="views-admin-settings-open-aifast-policy-user-selector__action-2"
          @click="selectUser(user)"
        >
          <span class="views-admin-settings-open-aifast-policy-user-selector__text-4">
            {{ user.email }}
            <span v-if="user.deleted" class="views-admin-settings-open-aifast-policy-user-selector__text-5">
              {{ t("admin.settings.openaiFastPolicy.userDeleted") }}
            </span>
          </span>
          <span class="views-admin-settings-open-aifast-policy-user-selector__text-6">#{{ user.id }}</span>
        </button>
      </template>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref, watch } from "vue";
import { useI18n } from "vue-i18n";
import { adminAPI } from "@/api/admin";
import type { SimpleUser } from "@/api/admin/usage";
import Icon from "@/components/icons/Icon.vue";

const props = defineProps<{
  modelValue: string[];
}>();

const emit = defineEmits<{
  "update:modelValue": [value: string[]];
}>();

const { t } = useI18n();
const containerRef = ref<HTMLElement | null>(null);
const searchQuery = ref("");
const searchResults = ref<SimpleUser[]>([]);
const searchLoading = ref(false);
const showDropdown = ref(false);
const selectedUsers = ref<Record<string, SimpleUser>>({});
let searchTimer: ReturnType<typeof setTimeout> | null = null;
let searchSequence = 0;

const selectedUserIds = computed(() =>
  Array.from(new Set(props.modelValue.filter(Boolean))),
);

const availableResults = computed(() => {
  const selected = new Set(selectedUserIds.value);
  return searchResults.value
    .filter((user) => !selected.has(user.id))
    .sort((a, b) => Number(a.deleted) - Number(b.deleted));
});

function selectedUserLabel(userId: string): string {
  return selectedUsers.value[userId]?.email ||
    t("admin.settings.openaiFastPolicy.userIdFallback", { id: userId });
}

function clearPendingSearch(): void {
  if (searchTimer) {
    clearTimeout(searchTimer);
    searchTimer = null;
  }
  searchSequence += 1;
}

function debounceSearch(): void {
  clearPendingSearch();
  const query = searchQuery.value.trim();
  showDropdown.value = true;
  if (!query) {
    searchResults.value = [];
    searchLoading.value = false;
    return;
  }

  const sequence = searchSequence;
  searchTimer = setTimeout(async () => {
    searchLoading.value = true;
    try {
      const results = await adminAPI.usage.searchUsers(query);
      if (sequence === searchSequence) {
        searchResults.value = results;
      }
    } catch {
      if (sequence === searchSequence) {
        searchResults.value = [];
      }
    } finally {
      if (sequence === searchSequence) {
        searchLoading.value = false;
      }
    }
  }, 300);
}

function selectUser(user: SimpleUser): void {
  selectedUsers.value = { ...selectedUsers.value, [user.id]: user };
  emit("update:modelValue", [...selectedUserIds.value, user.id]);
  clearPendingSearch();
  searchQuery.value = "";
  searchResults.value = [];
  searchLoading.value = false;
  showDropdown.value = false;
}

function removeUser(userId: string): void {
  emit(
    "update:modelValue",
    selectedUserIds.value.filter((id) => id !== userId),
  );
}

async function hydrateSelectedUsers(userIds: string[]): Promise<void> {
  const missing = userIds.filter((id) => !selectedUsers.value[id]);
  if (missing.length === 0) return;

  const users = await Promise.all(
    missing.map(async (id) => {
      try {
        const user = await adminAPI.users.getById(id, true);
        return {
          id: user.id,
          email: user.email,
          deleted: Boolean(user.deleted_at),
        } satisfies SimpleUser;
      } catch {
        return null;
      }
    }),
  );

  const next = { ...selectedUsers.value };
  for (const user of users) {
    if (user && props.modelValue.includes(user.id)) {
      next[user.id] = user;
    }
  }
  selectedUsers.value = next;
}

function handleDocumentClick(event: MouseEvent): void {
  const target = event.target as Node | null;
  if (target && !containerRef.value?.contains(target)) {
    showDropdown.value = false;
  }
}

watch(
  selectedUserIds,
  (userIds) => {
    void hydrateSelectedUsers(userIds);
  },
  { immediate: true },
);

onMounted(() => {
  document.addEventListener("click", handleDocumentClick);
});

onUnmounted(() => {
  clearPendingSearch();
  document.removeEventListener("click", handleDocumentClick);
});
</script>
