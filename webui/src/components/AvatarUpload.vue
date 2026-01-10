<template>
  <div
    class="avatar-upload"
    :class="{ disabled }"
    :style="{ '--avatar-upload-size': sizeStyle }"
  >
    <button
      class="avatar-upload__button"
      type="button"
      :disabled="disabled"
      @click="triggerPick"
    >
      <img
        v-if="src"
        :src="src"
        class="avatar-upload__image"
        :alt="alt || 'avatar preview'"
      />
      <span v-else class="avatar-upload__fallback">{{ fallbackText }}</span>
      <span class="avatar-upload__overlay">
        <span class="avatar-upload__icon" aria-hidden="true">📷</span>
        <span class="avatar-upload__text">{{ overlayText }}</span>
      </span>
    </button>
    <input
      ref="fileInput"
      class="avatar-upload__input"
      type="file"
      accept="image/*"
      :disabled="disabled"
      @change="onPick"
    />
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'

const props = defineProps({
  src: { type: String, default: '' },
  fallbackText: { type: String, default: '+' },
  overlayText: { type: String, default: 'Upload' },
  alt: { type: String, default: '' },
  size: { type: [String, Number], default: 48 },
  disabled: { type: Boolean, default: false },
})

const emit = defineEmits(['select'])

const fileInput = ref(null)
const sizeStyle = computed(() => {
  const value = Number(props.size)
  return Number.isNaN(value) ? String(props.size) : `${value}px`
})

function triggerPick() {
  if (props.disabled) return
  fileInput.value?.click()
}

function onPick(event) {
  const file = event?.target?.files?.[0] || null
  if (file) {
    emit('select', file)
  }
  if (event?.target) {
    event.target.value = ''
  }
}
</script>

<style scoped>
.avatar-upload {
  --avatar-upload-size: 48px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
}

.avatar-upload.disabled {
  opacity: 0.6;
}

.avatar-upload__button {
  width: var(--avatar-upload-size);
  height: var(--avatar-upload-size);
  border-radius: 50%;
  border: 1px solid var(--border);
  padding: 0;
  background: transparent;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  position: relative;
  cursor: pointer;
  overflow: hidden;
  transition: transform 0.2s ease, box-shadow 0.2s ease, border-color 0.2s ease;
}

.avatar-upload__button:focus-visible {
  outline: 2px solid #2563eb;
  outline-offset: 2px;
}

.avatar-upload__button:disabled {
  cursor: not-allowed;
}

.avatar-upload__image,
.avatar-upload__fallback {
  width: 100%;
  height: 100%;
  border-radius: 50%;
  object-fit: cover;
}

.avatar-upload__fallback {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  font-weight: 700;
  color: #0f766e;
  background: #e0f7ee;
  border: 1px solid #a7f3d0;
}

.avatar-upload__overlay {
  position: absolute;
  inset: 0;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 2px;
  background: rgba(15, 23, 42, 0.6);
  color: #f8fafc;
  opacity: 0;
  transition: opacity 0.2s ease;
}

.avatar-upload__button:hover .avatar-upload__overlay,
.avatar-upload__button:focus-visible .avatar-upload__overlay {
  opacity: 1;
}

.avatar-upload__icon {
  font-size: 0.9rem;
}

.avatar-upload__text {
  font-size: 0.72rem;
  font-weight: 600;
}

.avatar-upload__input {
  display: none;
}
</style>
