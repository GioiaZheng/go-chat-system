import { reactive } from 'vue'

const state = reactive({
  byId: {},
})

function normalizeId(meta = {}) {
  return String(meta.id || meta.conversationId || meta.conversation_id || '')
}

export function upsertConversationMeta(meta = {}) {
  const id = normalizeId(meta)
  if (!id) return
  const existing = state.byId[id] || {}
  state.byId[id] = {
    ...existing,
    ...meta,
    id,
  }
}

export function hydrateConversationList(list = []) {
  if (!Array.isArray(list)) return
  list.forEach(item => upsertConversationMeta(item))
}

export function removeConversation(id) {
  const key = String(id || '')
  if (!key) return
  delete state.byId[key]
}

export function getConversationMeta(id) {
  return state.byId[String(id || '')] || null
}

export function useConversationStore() {
  return state
}

export default state
