<!-- Group editor modal: fetches a single group, allows renaming, photo updates (preset or upload),
     member additions, and leaving the group while keeping the view in sync after each mutation.
     Uses axios directly against the backend endpoints:
       GET    /groups/:id
       PUT    /groups/:id/name           { name }
       PUT    /groups/:id/photo          { preset }  OR  multipart/form-data with field "upload"
       POST   /groups/:id/members        { member_ids: [...] }
       POST   /groups/:id/leave
     Integrates with <UserSearchBox /> (emits "select" only); selections are staged locally before submission.
-->

<template>
  <div class="editor">
    <div class="panel">
      <header class="head">
        <strong>Group settings</strong>
        <button class="close" @click="$emit('close')">✕</button>
      </header>

      <section v-if="loading" class="body">
        <LoadingSpinner text="Loading group..." />
      </section>

      <section v-else class="body">
        <!-- Basic info -->
        <div class="row">
          <label>ID</label>
          <div>#{{ group.id }}</div>
        </div>

        <div class="row">
          <label>Name</label>
          <form class="inline" @submit.prevent="saveName">
            <input v-model="name" placeholder="Group name" />
            <button :disabled="savingName">{{ savingName ? "Saving…" : "Save" }}</button>
          </form>
        </div>

        <!-- Photo: preset quick set or upload -->
        <div class="row">
          <label>Photo</label>
          <div class="photo-actions">
            <img v-if="group.avatarUrl" :src="group.avatarUrl" class="avatar" alt="group" />
            <div class="buttons">
              <button type="button" @click="setPreset('group6')">Use preset: group6</button>
              <input type="file" accept="image/*" @change="uploadPhoto" />
            </div>
          </div>
        </div>

        <!-- Members -->
        <div class="row">
          <label>Members ({{ (group.members || []).length }})</label>
          <div class="members">
            <div v-for="m in group.members || []" :key="m.id" class="member">
              <span class="mn">{{ m.username || m.name || m.id }}</span>
              <span class="mid">#{{ m.id }}</span>
            </div>
          </div>
        </div>

        <!-- Add members via search (uses our UserSearchBox emitting 'select') -->
        <div class="row">
          <label>Add members</label>
          <div class="w-full">
            <UserSearchBox @select="onPick" />
            <button class="add" :disabled="adding || !toAdd.length" @click="addMembers">
              {{ adding ? "Adding…" : "Add selected" }} ({{ toAdd.length }})
            </button>
            <div v-if="toAdd.length" class="chips">
              <span v-for="u in toAdd" :key="u.id" class="chip">
                {{ u.username || u.name || u.id }}
              </span>
            </div>
          </div>
        </div>

        <!-- Danger zone -->
        <div class="row danger">
          <label>Leave group</label>
          <button class="danger-btn" :disabled="leaving" @click="leave">
            {{ leaving ? "Leaving…" : "Leave this group" }}
          </button>
        </div>
      </section>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from "vue";
import axios from "../services/axios";
import LoadingSpinner from "./LoadingSpinner.vue";
import UserSearchBox from "./UserSearchBox.vue";

defineProps({
  groupId: { type: [String, Number], required: true },
});

const emit = defineEmits(["close"]);

const loading = ref(true);
const group = ref({});
const name = ref("");

const savingName = ref(false);
const adding = ref(false);
const leaving = ref(false);

// Users queued for addition.
const toAdd = ref([]);

// API helpers (axios-only).

async function fetchGroup() {
  loading.value = true;
  try {
    const r = await axios.get(`/groups/${groupId.value ?? groupId}`);
    // Normalize API shapes so both {data:{group}} and {group} are accepted.
    group.value = r?.data?.group || r?.data?.data?.group || r?.data || {};
    name.value = group.value.name || "";
  } finally {
    loading.value = false;
  }
}

async function saveName() {
  if (!name.value.trim()) return;
  savingName.value = true;
  try {
    await axios.put(`/groups/${groupId.value ?? groupId}/name`, { name: name.value.trim() });
    await fetchGroup();
  } finally {
    savingName.value = false;
  }
}

async function setPreset(preset) {
  try {
    await axios.put(`/groups/${groupId.value ?? groupId}/photo`, { preset });
    await fetchGroup();
  } catch (e) {
    console.error("setGroupPhotoPreset failed:", e);
  }
}

async function uploadPhoto(ev) {
  const f = ev?.target?.files?.[0];
  if (!f) return;
  try {
    const form = new FormData();
    form.append("upload", f);
    await axios.put(`/groups/${groupId.value ?? groupId}/photo`, form, {
      headers: { "Content-Type": "multipart/form-data" },
    });
    await fetchGroup();
  } catch (e) {
    console.error("setGroupPhotoUpload failed:", e);
  } finally {
    if (ev?.target) ev.target.value = "";
  }
}

function onPick(u) {
  // Avoid adding the same user more than once.
  if (!toAdd.value.find((x) => String(x.id) === String(u.id))) {
    toAdd.value.push(u);
  }
}

async function addMembers() {
  if (!toAdd.value.length) return;
  adding.value = true;
  try {
    const member_ids = toAdd.value.map((u) => u.id);
    await axios.post(`/groups/${groupId.value ?? groupId}/members`, { member_ids });
    toAdd.value = [];
    await fetchGroup();
  } catch (e) {
    console.error("addToGroup failed:", e);
  } finally {
    adding.value = false;
  }
}

async function leave() {
  if (!confirm("Are you sure to leave this group?")) return;
  leaving.value = true;
  try {
    await axios.post(`/groups/${groupId.value ?? groupId}/leave`);
    emit("close"); // Allow the parent to refresh the list or navigate away.
  } catch (e) {
    console.error("leaveGroup failed:", e);
  } finally {
    leaving.value = false;
  }
}

onMounted(fetchGroup);
</script>

<script>
export default {
  name: "GroupEditor",
};
</script>

<style scoped>
.editor {
  position: fixed; inset: 0; background: rgba(0,0,0,0.3);
  display: grid; place-items: center;
}
.panel {
  width: min(680px, 92vw);
  max-height: 90vh;
  background: #fff; border-radius: 0;
  box-shadow: 0 10px 28px rgba(0,0,0,0.15);
  display: grid; grid-template-rows: auto 1fr;
}
.head {
  display: flex; justify-content: space-between; align-items: center;
  padding: 12px 14px; border-bottom: 1px solid #eee;
}
.close { border: 0; background: transparent; cursor: pointer; border-radius: var(--radius-control); }
.body { padding: 12px 14px; overflow: auto; display: grid; gap: 14px; }
.row { display: grid; grid-template-columns: 140px 1fr; gap: 10px; align-items: center; }
.inline { display: flex; gap: 8px; }
.photo-actions { display: flex; align-items: center; gap: 12px; }
.avatar { width: 64px; height: 64px; border-radius: 50%; object-fit: cover; }
.members { display: flex; flex-wrap: wrap; gap: 8px; }
.member { background: #f7f7f7; padding: 4px 8px; border-radius: 0; }
.mn { font-weight: 600; margin-right: 6px; }
.mid { font-size: 12px; opacity: .6; }
.add { margin-top: 8px; }
.chips { display: flex; gap: 6px; flex-wrap: wrap; margin-top: 6px; }
.chip { background: #eef6ff; padding: 2px 6px; border-radius: 0; font-size: 12px; }
.danger { border-top: 1px dashed #eee; padding-top: 12px; margin-top: 8px; }
.danger-btn { border: 0; background: #ffe9e9; color: #b10000; padding: 8px 10px; border-radius: var(--radius-control); cursor: pointer; }
</style>
