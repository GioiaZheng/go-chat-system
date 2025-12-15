<template>
  <img
    :src="srcFinal"
    :alt="alt || 'avatar'"
    class="rounded-circle object-fit-cover"
    :style="{ width: size + 'px', height: size + 'px', border: '1px solid #ddd' }"
    @error="onErr"
  />
</template>

<script>
const PLACEHOLDER =
  'data:image/svg+xml;utf8,' +
  encodeURIComponent(`<svg xmlns="http://www.w3.org/2000/svg" width="96" height="96">
  <rect width="100%" height="100%" fill="#e0f7ee"/>
  <circle cx="48" cy="36" r="18" fill="#a7f3d0"/>
  <rect x="18" y="58" width="60" height="22" rx="11" fill="#34d399" opacity="0.6"/>
</svg>`)

export default {
  name: 'AvatarImg',
  props: {
    src: String,
    alt: String,
    size: { type: Number, default: 40 }
  },
  data() {
    return { broken: false }
  },
  computed: {
    srcFinal() {
      if (this.broken || !this.src) return PLACEHOLDER
      if (this.src.startsWith('http')) return this.src
      return `${window.__API_URL__ || ''}${this.src}`
    }
  },
  methods: {
    onErr() { this.broken = true }
  }
}
</script>
