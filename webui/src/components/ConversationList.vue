<template>
  <aside :class="paneClass">
    <div class="list-header">
      <div class="list-title">{{ title }}</div>
      <span class="badge">{{ totalCount }}</span>
    </div>

    <div class="list-search">
      <input
        :value="search"
        type="search"
        class="search"
        :placeholder="searchPlaceholder"
        aria-label="Search conversations"
        @input="onSearchInput"
      />
    </div>

    <div class="section">
      <div class="section-head">
        <h3 class="section-title text-secondary">Direct</h3>
        <span class="badge">{{ privateConvs.length }}</span>
      </div>

      <ul class="list" role="list">
        <li
          v-for="c in privateConvs"
          :key="c.id"
          class="item"
          :class="{ active: isActive(c) }"
          @click="emitSelect(c)"
        >
          <div class="left">
            <span v-if="!avatarFor(c)" class="avatar-fallback avatar-circle">{{ initials(c) }}</span>
            <img v-else class="avatar avatar-circle" :src="avatarFor(c)" alt="avatar" />
          </div>

          <div class="info">
            <div class="top">
              <div class="name">{{ displayName(c) }}</div>
              <div class="time">{{ fmtTime(c.last_time) }}</div>
            </div>
            <div class="bottom">
              <div class="preview">{{ c.last_preview || 'No messages yet' }}</div>
            </div>
          </div>

          <button v-if="showDelete" class="del" @click.stop="emitDelete(c)">Delete</button>
        </li>

        <li v-if="!privateConvs.length" class="empty">{{ emptyDirectText }}</li>
      </ul>
    </div>

    <div class="section">
      <div class="section-head second">
        <h3 class="section-title text-secondary">Groups</h3>
        <span class="badge badge--secondary">{{ groupConvs.length }}</span>
      </div>

      <ul class="list" role="list">
        <li
          v-for="c in groupConvs"
          :key="c.id"
          class="item"
          :class="{ active: isActive(c) }"
          @click="emitSelect(c)"
        >
          <div class="left">
            <span v-if="!avatarFor(c)" class="avatar-fallback avatar-circle">{{ initials(c) }}</span>
            <img v-else class="avatar avatar-circle" :src="avatarFor(c)" alt="avatar" />
          </div>
          <div class="info">
            <div class="top">
              <div class="name">{{ displayName(c) }}</div>
              <div class="time">{{ fmtTime(c.last_time) }}</div>
            </div>
            <div class="bottom">
              <div class="preview">{{ c.last_preview || 'No messages yet' }}</div>
            </div>
          </div>
          <button v-if="showDelete" class="del" @click.stop="emitDelete(c)">Delete</button>
        </li>
        <li v-if="!groupConvs.length" class="empty">{{ emptyGroupText }}</li>
      </ul>
    </div>

    <p v-if="loading" class="muted">Loading conversations…</p>
    <ErrorMsg v-else-if="error" :text="error" />
  </aside>
</template>

<script setup>
import { computed } from 'vue'
import ErrorMsg from '@/components/ErrorMsg.vue'

const props = defineProps({
  title: { type: String, default: 'Conversations' },
  variant: { type: String, default: 'default' },
  search: { type: String, default: '' },
  searchPlaceholder: { type: String, default: 'search conversations' },
  privateConvs: { type: Array, default: () => [] },
  groupConvs: { type: Array, default: () => [] },
  selectedId: { type: String, default: '' },
  displayName: { type: Function, required: true },
  avatarFor: { type: Function, required: true },
  initials: { type: Function, required: true },
  fmtTime: { type: Function, required: true },
  showDelete: { type: Boolean, default: false },
  loading: { type: Boolean, default: false },
  error: { type: String, default: '' },
  emptyDirectText: { type: String, default: 'No direct chats yet.' },
  emptyGroupText: { type: String, default: 'No group chats yet.' },
})

const emit = defineEmits(['update:search', 'select', 'delete'])

const totalCount = computed(() => props.privateConvs.length + props.groupConvs.length)
const paneClass = computed(() => [
  'list-pane',
  props.variant === 'split' ? 'list-pane--split' : '',
])

function onSearchInput(event) {
  emit('update:search', event.target.value)
}

function normalizeId(c) {
  return String(c?.id || '')
}

function isActive(c) {
  return normalizeId(c) === String(props.selectedId || '')
}

function emitSelect(c) {
  emit('select', c)
}

function emitDelete(c) {
  emit('delete', c)
}
</script>

<style scoped>
.list-pane {
  flex: 0 0 360px;
  background: var(--panel);
  border: 1px solid var(--border);
  display: flex;
  flex-direction: column;
  padding: 16px;
  gap: 10px;
  min-height: 0;
  overflow: auto;
}

.list-pane--split {
  border: 0;
  border-right: 1px solid var(--border);
  border-radius: 0;
}

.list-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
  padding: 4px 0 2px;
}

.list-title {
  font-weight: 800;
  color: #0f172a;
}

.list-search {
  padding: 0 0 8px;
}

.search {
  width: 100%;
  border: 1px solid #e5e7eb;
  border-radius: var(--radius-control);
  padding: 10px 12px;
  background: #ffffff;
  font-size: var(--font-primary);
  transition: border-color 0.15s ease, box-shadow 0.15s ease;
}

.search:focus {
  outline: none;
  border-color: #32d583;
  box-shadow: 0 0 0 3px rgba(50, 213, 131, 0.2);
}

.section {
  display: flex;
  flex-direction: column;
  gap: 8px;
  min-height: 0;
}

.section-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
  padding: 4px 4px 2px;
}

.section-head.second {
  margin-top: 4px;
  border-top: 1px solid #e5e7eb;
  padding-top: 10px;
}

.section-title {
  margin: 0;
  font-weight: 700;
  color: #0f172a;
  font-size: var(--font-secondary);
}

.badge {
  display: inline-flex;
  min-width: 28px;
  height: 22px;
  border-radius: 0;
  padding: 0 8px;
  align-items: center;
  justify-content: center;
  border: 1px solid #cbd5e1;
  background: #fff;
  font-size: var(--font-secondary);
  color: #475569;
  font-weight: 600;
}

.badge--secondary {
  background: #f1f5f9;
  border-color: #d8e0e8;
}

.list {
  list-style: none;
  margin: 0;
  padding: 0;
  display: flex;
  flex-direction: column;
  gap: 10px;
  flex: 1 1 auto;
  overflow-y: auto;
  max-height: 100%;
  min-height: 0;
}

.item {
  background: #ffffff;
  border: 1px solid var(--border);
  border-radius: var(--radius-control);
  padding: 14px 16px;
  display: grid;
  grid-template-columns: auto 1fr auto;
  align-items: center;
  gap: 14px;
  transition: background 0.15s ease, border-color 0.15s ease, box-shadow 0.15s ease;
  position: relative;
  cursor: pointer;
}

.item:hover {
  background: #c6f6d5;
  border-color: rgba(50, 213, 131, 0.6);
  box-shadow: 0 6px 16px rgba(15, 23, 42, 0.08);
}

.item.active {
  background: #32d583;
  border-color: #32d583;
}

.item.active::before {
  content: '';
  position: absolute;
  left: 0;
  top: 10px;
  bottom: 10px;
  width: 3px;
  background: #16a34a;
}

.item.active .name {
  color: #0f172a;
}

.item.active .time,
.item.active .preview {
  color: #065f46;
}

.left {
  display: flex;
  align-items: center;
  justify-content: center;
}

.info {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.top {
  display: grid;
  grid-template-columns: 1fr auto;
  align-items: center;
  gap: 12px;
}

.name {
  font-weight: 600;
  color: #0f172a;
  font-size: var(--font-primary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.time {
  font-size: 0.92rem;
  color: #64748b;
  white-space: nowrap;
}

.preview {
  color: #64748b;
  font-size: 0.92rem;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.del {
  border: 0;
  border-radius: var(--radius-control);
  padding: 0.35rem 0.7rem;
  font-size: var(--font-secondary);
  color: #fff;
  background: linear-gradient(135deg, #ef4444, #dc2626);
  transition: 0.2s;
  opacity: 0;
  pointer-events: none;
}

.item:hover .del,
.item:focus-within .del {
  opacity: 1;
  pointer-events: auto;
}

.del:hover {
  opacity: 0.9;
}

.empty {
  text-align: center;
  color: #6b7280;
  padding: 16px 0;
}
</style>
