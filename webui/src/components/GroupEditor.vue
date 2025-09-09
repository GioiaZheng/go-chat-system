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
              <button @click="setPreset('group6')">Use preset: group6</button>
              <input type="file" accept="image/*" @change="uploadPhoto" />
            </div>
          </div>
        </div>

        <!-- Members -->
        <div class="row">
          <label>Members ({{ (group.members||[]).length }})</label>
          <div class="members">
            <div v-for="m in group.members || []" :key="m.id" class="member">
              <span class="mn">{{ m.username || m.name || m.id }}</span>
              <span class="mid">#{{ m.id }}</span>
            </div>
          </div>
        </div>

        <!-- Add members via search -->
        <div class="row">
          <label>Add members</label>
          <UserSearchBox
            :busy="searching"
            :results="results"
            @query="onSearch"
            @pick="onPick"
            @clear="clearResults"
          />
          <button class="add" :disabled="adding || !toAdd.length" @click="addMembers">
            {{ adding ? "Adding…" : "Add selected" }} ({{ toAdd.length }})
          </button>
          <div v-if="toAdd.length" class="chips">
            <span v-for="u in toAdd" :key="u.id" class="chip">
              {{ u.username || u.name || u.id }}
            </span>
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

<script>
/**
 * GroupEditor
 * - English comments for TA/teacher clarity.
 * - Edits a single group: rename, set photo (preset/upload), add members, leave group.
 * - Uses APIs: getGroup, setGroupName, setGroupPhotoPreset/setGroupPhotoUpload,
 *             addToGroup, leaveGroup.
 */

import { ref, onMounted } from "vue";
import {
  getGroup,
  setGroupName,
  setGroupPhotoPreset,
  setGroupPhotoUpload,
  addToGroup,
  leaveGroup,
  searchUsers,
} from "../services/api";
import LoadingSpinner from "./LoadingSpinner.vue";
import UserSearchBox from "./UserSearchBox.vue";

export default {
  name: "GroupEditor",
  components: { LoadingSpinner, UserSearchBox },
  props: {
    groupId: { type: [String, Number], required: true },
  },
  setup(props, { emit }) {
    const loading = ref(true);
    const group = ref({});
    const name = ref("");

    const savingName = ref(false);
    const searching = ref(false);
    const results = ref([]);
    const toAdd = ref([]);
    const adding = ref(false);
    const leaving = ref(false);

    onMounted(fetchGroup);

    async function fetchGroup() {
      loading.value = true;
      try {
        const r = await getGroup(props.groupId);
        group.value = r?.data?.group || r?.data || {};
        name.value = group.value.name || "";
      } catch (e) {
        console.error("getGroup failed:", e);
      } finally {
        loading.value = false;
      }
    }

    async function saveName() {
      if (!name.value.trim()) return;
      savingName.value = true;
      try {
        await setGroupName(props.groupId, name.value.trim());
        await fetchGroup();
      } catch (e) {
        console.error("setGroupName failed:", e);
      } finally {
        savingName.value = false;
      }
    }

    async function setPreset(preset) {
      try {
        await setGroupPhotoPreset(props.groupId, preset);
        await fetchGroup();
      } catch (e) {
        console.error("setGroupPhotoPreset failed:", e);
      }
    }

    async function uploadPhoto(ev) {
      const f = ev?.target?.files?.[0];
      if (!f) return;
      try {
        await setGroupPhotoUpload(props.groupId, f);
        await fetchGroup();
      } catch (e) {
        console.error("setGroupPhotoUpload failed:", e);
      } finally {
        ev.target.value = "";
      }
    }

    // --- Search & add members ---
    let timer = null;
    function onSearch(q) {
      clearTimeout(timer);
      if (!q) {
        results.value = [];
        return;
      }
      timer = setTimeout(async () => {
        searching.value = true;
        try {
          const r = await searchUsers(q);
          results.value = r?.data?.users || r?.data || [];
        } catch (e) {
          console.error("searchUsers failed:", e);
          results.value = [];
        } finally {
          searching.value = false;
        }
      }, 250);
    }

    function clearResults() {
      results.value = [];
    }

    function onPick(u) {
      // prevent duplicates
      if (!toAdd.value.find((x) => String(x.id) === String(u.id))) {
        toAdd.value.push(u);
      }
    }

    async function addMembers() {
      if (!toAdd.value.length) return;
      adding.value = true;
      try {
        const ids = toAdd.value.map((u) => u.id);
        await addToGroup(props.groupId, ids);
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
        await leaveGroup(props.groupId);
        emit("close"); // parent will refresh list
      } catch (e) {
        console.error("leaveGroup failed:", e);
      } finally {
        leaving.value = false;
      }
    }

    return {
      loading,
      group,
      name,
      savingName,
      setPreset,
      uploadPhoto,
      searching,
      results,
      toAdd,
      onSearch,
      onPick,
      clearResults,
      adding,
      addMembers,
      leaving,
      leave,
    };
  },
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
  background: #fff; border-radius: 12px;
  box-shadow: 0 10px 28px rgba(0,0,0,0.15);
  display: grid; grid-template-rows: auto 1fr;
}
.head {
  display: flex; justify-content: space-between; align-items: center;
  padding: 12px 14px; border-bottom: 1px solid #eee;
}
.close { border: 0; background: transparent; cursor: pointer; }
.body { padding: 12px 14px; overflow: auto; display: grid; gap: 14px; }
.row { display: grid; grid-template-columns: 140px 1fr; gap: 10px; align-items: center; }
.inline { display: flex; gap: 8px; }
.photo-actions { display: flex; align-items: center; gap: 12px; }
.avatar { width: 64px; height: 64px; border-radius: 14px; object-fit: cover; }
.members { display: flex; flex-wrap: wrap; gap: 8px; }
.member { background: #f7f7f7; padding: 4px 8px; border-radius: 8px; }
.mn { font-weight: 600; margin-right: 6px; }
.mid { font-size: 12px; opacity: .6; }
.add { margin-top: 8px; }
.chips { display: flex; gap: 6px; flex-wrap: wrap; }
.chip { background: #eef6ff; padding: 2px 6px; border-radius: 6px; font-size: 12px; }
.danger { border-top: 1px dashed #eee; padding-top: 12px; margin-top: 8px; }
.danger-btn { border: 0; background: #ffe9e9; color: #b10000; padding: 8px 10px; border-radius: 8px; cursor: pointer; }
</style>
