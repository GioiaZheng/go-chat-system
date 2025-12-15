<template>
  <div class="msg" :class="{ mine }">

    <!-- Main message content -->
    <div class="bubble">

      <!-- Image message -->
      <template v-if="type === 'image' && imageUrl">
        <img 
          :src="imageUrl" 
          alt="image message"
          class="msg-img"
        />
      </template>

      <!-- Text message -->
      <template v-else>
        {{ text }}
      </template>

      <!-- inline reply -->
      <div v-if="replyTo" class="reply-preview">
        ↪ {{ replyTo }}
      </div>

      <!-- reactions -->
      <div v-if="reactions && reactions.length" class="reactions">
        <span v-for="(r,i) in reactions" :key="i" class="reaction">{{ r }}</span>
      </div>

    </div>

    <!-- Meta info (timestamp + checkmarks) -->
    <div class="meta">
      {{ meta }}
      <span v-if="ticks === 1">✓</span>
      <span v-else-if="ticks === 2">✓✓</span>
      <span v-else-if="ticks === 3">✓✓</span>
    </div>

  </div>
</template>

<script setup>
const props = defineProps({
  mine: { type: Boolean, default: false },

  /* Content */
  text: { type: String, default: '' },
  type: { type: String, default: 'text' },      
  imageUrl: { type: String, default: '' }, 

  /* Meta */
  meta: { type: String, default: '' },
  ticks: { type: Number, default: -1 },  

  /* Inline reply */
  replyTo: { type: String, default: '' },

  /* Reactions (emoji array) */
  reactions: { type: Array, default: () => [] },
})
</script>

<style scoped>
.msg {
  margin: 8px 0;
  display: flex;
  flex-direction: column;
  max-width: 70%;
}
.msg.mine {
  align-self: flex-end;
}
.bubble {
  background: #fff;
  padding: 8px 12px;
  border-radius: 12px;
  font-size: 14px;
  position: relative;
}
.msg.mine .bubble {
  background: #d1f1ff;
}
.meta {
  font-size: 12px;
  color: #666;
  margin-top: 2px;
  display: flex;
  align-items: center;
  gap: 4px;
}
.msg-img {
  max-width: 200px;
  border-radius: 10px;
}
.reply-preview {
  margin-top: 4px;
  padding: 4px 6px;
  font-size: 12px;
  background: #eef2f7;
  border-left: 3px solid #3b82f6;
  border-radius: 6px;
}
.reactions {
  margin-top: 6px;
  display: flex;
  gap: 4px;
}
.reaction {
  font-size: 14px;
  padding: 2px 4px;
  background: #eee;
  border-radius: 6px;
}
</style>
