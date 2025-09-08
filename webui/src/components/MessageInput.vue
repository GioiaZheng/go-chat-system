<template>
  <form class="msg-input" @submit.prevent="submit">
    <input v-model="text" placeholder="Type a message..." />
    <button type="submit">Send</button>
  </form>
</template>

<script>
/**
 * MessageInput
 * - Input box for composing a message.
 * - Emits 'send' with the message text.
 */
export default {
  name: "MessageInput",
  data() {
    return { text: "" };
  },
  methods: {
    submit() {
      if (!this.text.trim()) return;
      this.$emit("send", this.text.trim());
      this.text = "";
    },
  },
};
</script>

<style scoped>
.msg-input {
  display: flex;
  gap: 8px;
}
.msg-input input {
  flex: 1;
}
</style>
