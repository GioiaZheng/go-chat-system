<template>
  <div :class="['bubble', mine ? 'mine' : 'theirs']">
    <div class="meta">
      <span v-if="!mine" class="sender">{{ msg.senderName || msg.senderId }}</span>
      <span class="time">{{ timeText }}</span>
    </div>

    <div class="content">{{ msg.content }}</div>

    <!-- Status check marks for my messages -->
    <div v-if="mine" class="status">
      {{ msg.read ? "✔✔" : "✔" }}
    </div>

    <!-- Actions -->
    <div class="actions">
      <!-- Toggle comments (lazy load via parent) -->
      <button @click="$emit('toggle-comments', msg)">
        {{ showComments ? "Hide comments" : "Comments" }}
      </button>

      <button @click="$emit('comment', msg)">Add comment</button>

      <!-- uncomment only makes sense for my comment; we keep it simple here -->
      <button @click="$emit('uncomment', msg)">Uncomment</button>

      <!-- Forward (to user or group) -->
      <button @click="$emit('forward', msg)">Forward</button>

      <!-- Delete my own message -->
      <button v-if="mine" @click="$emit('delete', msg)">Delete</button>
    </div>

    <!-- Comments list (if toggled on) -->
    <ul v-if="showComments" class="comments">
      <li v-if="!comments?.length" class="empty">No comments</li>
      <li v-for="c in comments" :key="c.id" class="comment">
        <span class="author">{{ c.authorName || c.userId }}</span>:
        <span class="text">{{ c.text || c.comment || c.content }}</span>
      </li>
    </ul>
  </div>
</template>

<script>
/**
 * MessageBubble
 * - Renders message bubble + actions.
 * - Shows comments (when parent passes them & toggled).
 * - Emits events so the parent ChatView can call APIs (comment/uncomment/forward/delete).
 */
export default {
  name: "MessageBubble",
  props: {
    msg: { type: Object, required: true },
    meId: { type: [String, Number], required: true },
    // Parent controls whether to show comments and provides list (lazy loading)
    showComments: { type: Boolean, default: false },
    comments: { type: Array, default: () => [] },
  },
  computed: {
    mine() {
      return String(this.msg.senderId) === String(this.meId);
    },
    timeText() {
      const t = new Date(this.msg.timestamp || this.msg.time || Date.now());
      return t.toTimeString().slice(0, 5);
    },
  },
};
</script>

<style scoped>
.bubble { max-width: 70%; padding: 8px 12px; border-radius: 8px; position: relative; }
.mine   { align-self: flex-end; background: #e7f5ff; }
.theirs { align-self: flex-start; background: #f5f5f5; }
.meta { font-size: 11px; opacity: 0.65; margin-bottom: 2px; display: flex; justify-content: space-between; }
.content { font-size: 14px; }
.status { font-size: 11px; opacity: 0.65; margin-top: 2px; text-align: right; }
.actions { margin-top: 6px; display: flex; gap: 8px; flex-wrap: wrap; }
.actions button { border: 0; background: transparent; color: #555; cursor: pointer; font-size: 12px; }
.comments { margin: 6px 0 0; padding: 0 0 0 12px; display: grid; gap: 4px; }
.comment .author { font-weight: 600; margin-right: 4px; }
.comment .text { opacity: .9; }
.empty { opacity: .6; font-size: 12px; }
</style>
