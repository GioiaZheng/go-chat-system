<template>
  <div class="search">
    <!-- Search input -->
    <input
      :disabled="busy"
      v-model="q"
      placeholder="Search users by username..."
      @input="emitQuery"
    />

    <!-- Results dropdown -->
    <div v-if="q && results?.length" class="dropdown">
      <button
        v-for="u in results"
        :key="u.id"
        class="row"
        @click="pick(u)"
      >
        <span class="name">{{ u.username || u.name || u.id }}</span>
        <span class="id">#{{ u.id }}</span>
      </button>
    </div>

    <!-- Clear icon -->
    <button v-if="q" class="clear" @click="clear">✕</button>
  </div>
</template>

<script>
/**
 * UserSearchBox
 * - Emits 'query' on input (parent performs API call to searchUsers).
 * - Emits 'pick' when a user is selected.
 * - Emits 'clear' to reset results in parent.
 */
export default {
  name: "UserSearchBox",
  props: {
    busy: { type: Boolean, default: false },
    results: { type: Array, default: () => [] },
  },
  emits: ["query", "pick", "clear"],
  data() {
    return { q: "" };
  },
  methods: {
    emitQuery() {
      this.$emit("query", this.q.trim());
    },
    pick(u) {
      this.$emit("pick", u);
      this.q = "";
      this.$emit("clear");
    },
    clear() {
      this.q = "";
      this.$emit("clear");
    },
  },
};
</script>

<style scoped>
.search {
  position: relative;
  display: flex;
  align-items: center;
  gap: 8px;
}
.search input {
  width: 280px;
}
.clear {
  border: 0;
  background: transparent;
  cursor: pointer;
}
.dropdown {
  position: absolute;
  top: 36px;
  left: 0;
  width: 100%;
  max-height: 240px;
  overflow: auto;
  border: 1px solid #eee;
  background: #fff;
  z-index: 3;
}
.row {
  display: flex;
  justify-content: space-between;
  width: 100%;
  border: 0;
  background: #fff;
  padding: 8px 10px;
  text-align: left;
  cursor: pointer;
  border-bottom: 1px solid #f7f7f7;
}
.row:hover {
  background: #fafafa;
}
.name { overflow: hidden; text-overflow: ellipsis; }
.id   { font-size: 12px; opacity: 0.65; }
</style>
