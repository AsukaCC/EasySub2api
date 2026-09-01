<template>
  <div class="table-page-layout" :class="{ 'mobile-mode': isMobile }">
    <!-- 固定区域：操作按钮 -->
    <div v-if="$slots.actions" class="layout-section-fixed layout-section-actions">
      <slot name="actions" />
    </div>

    <!-- 固定区域：搜索和过滤器 -->
    <div v-if="$slots.filters" class="layout-section-fixed layout-section-filters">
      <slot name="filters" />
    </div>

    <!-- 滚动区域：表格 -->
    <div class="layout-section-scrollable">
      <div class="card table-scroll-container">
        <slot name="table" />
      </div>
    </div>

    <!-- 固定区域：分页器 -->
    <div v-if="$slots.pagination" class="layout-section-fixed layout-section-pagination">
      <slot name="pagination" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'

const isMobile = ref(false)

const checkMobile = () => {
  isMobile.value = window.innerWidth < 1024
}

onMounted(() => {
  checkMobile()
  window.addEventListener('resize', checkMobile)
})

onUnmounted(() => {
  window.removeEventListener('resize', checkMobile)
})
</script>

<style scoped>
/* 桌面端：Flexbox 布局 */
.table-page-layout {
  position: relative;
  isolation: isolate;
  height: 100%;
}

/* 层级只在当前页面布局内生效:工具栏下拉高于表格,但低于应用导航和浮层。 */
.layout-section-fixed {
  position: relative;
  z-index: 2;
  flex-shrink: 0;
}

.layout-section-actions {
  z-index: 3;
}

.layout-section-filters {
  z-index: 2;
}

.layout-section-pagination {
  z-index: 2;
}

.layout-section-scrollable {
  position: relative;
  z-index: 1;
  display: flex;
  flex: 1;
  min-height: 0;
  flex-direction: column;
}

/* 表格滚动容器 - 增强版表体滚动方案
   height:auto + max-height:100%:数据少时卡片收缩到内容高度,
   不再强制拉满视口留大片空白;数据多时仍在容器内滚动。 */
.table-scroll-container {
  display: flex;
  flex-direction: column;
  height: auto;
  max-height: 100%;
  overflow: hidden;
  border: 1px solid var(--glass-border);
  border-radius: var(--radius-xl);
  background: var(--glass-bg);
  /* 磨砂玻璃:与顶部导航壳同款 backdrop 模糊。容器本身成为 backdrop root,
     内部玻璃表头/固定列的 blur 恰好以容器内容为采样边界。 */
  -webkit-backdrop-filter: blur(var(--glass-blur-thin)) saturate(var(--glass-saturate));
  backdrop-filter: blur(var(--glass-blur-thin)) saturate(var(--glass-saturate));
  box-shadow:
    var(--glass-shadow),
    0 1px 0 var(--glass-highlight) inset;
}

.table-scroll-container :deep(.table-wrapper) {
  flex: 1;
  overflow-x: auto;
  overflow-y: auto;
}

.table-scroll-container :deep(table) {
  width: 100%;
  min-width: max-content; /* 关键：确保表格宽度根据内容撑开，从而触发横向滚动 */
  display: table; /* 使用标准 table 布局以支持 sticky 列 */
}

/* 表头底色交给单元格自身的玻璃材质(DataTable 表头带 backdrop 模糊),
   容器不再涂不透明底色,避免玻璃卡内出现实色条带 */
.table-scroll-container :deep(thead) {
  background: transparent;
}

.table-scroll-container :deep(tbody) {
  /* 保持默认 table-row-group 显示，不使用 block */
}

.table-scroll-container :deep(th) {
  padding: 1rem 1.25rem;
  border-bottom: 1px solid var(--color-border);
  color: var(--color-text-secondary);
  font-size: var(--font-size-sm);
  font-weight: 500;
  text-align: left;
}

.table-scroll-container :deep(td) {
  padding: 1rem 1.25rem;
  border-bottom: 1px solid var(--color-border-subtle);
  color: var(--color-text-secondary);
  font-size: var(--font-size-sm);
}

/* 移动端：恢复正常滚动 */
.table-page-layout.mobile-mode .table-scroll-container {
  height: auto;
  overflow: visible;
  border: none;
  background: transparent;
  box-shadow: none;
}

.table-page-layout.mobile-mode .layout-section-scrollable {
  flex: none;
  min-height: fit-content;
}

.table-page-layout.mobile-mode .table-scroll-container :deep(table) {
  display: table;
  min-width: 100%;
}
</style>
