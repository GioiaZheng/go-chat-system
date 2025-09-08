<template>
  <div class="page">
    <header class="bar">
      <h2 class="title">My Profile</h2>
    </header>

    <section class="body" v-if="loading">
      <LoadingSpinner text="Loading profile..." />
    </section>

    <section class="body" v-else>
      <!-- Basic info -->
      <div class="row">
        <label>ID</label>
        <div>#{{ me?.id }}</div>
      </div>

      <div class="row">
        <label>Username</label>
        <form class="inline" @submit.prevent="saveName">
          <input v-model="username" placeholder="Enter new username" />
          <button :disabled="savingName">{{ savingName ? "Saving…" : "Save" }}</button>
        </form>
      </div>

      <div class="row">
        <label>Photo</label>
        <div class="photo">
          <img v-if="me?.avatarUrl" :src="me.avatarUrl" class="avatar" alt="me" />
          <div class="btns">
            <!-- Preset quick set -->
            <button @click="setPreset('avatar6')">Use preset: avatar6</button>
            <!-- Upload -->
            <input type="file" accept="image/*" @change="uploadPhoto" />
          </div>
        </div>
      </div>

      <!-- Feedback -->
      <ErrorMsg v-if="error">{{ error }}</ErrorMsg>
      <div v-if="ok" class="ok">Saved ✔</div>
    </section>
  </div>
</template>

<script>
/**
 * ProfileView
 * - English comments for TA/teacher clarity.
 * - Shows current user info and allows:
 *   (1) setMyUserName(username)
 *   (2) setMyPhoto (preset or file upload)
 * - Uses APIs: me, setMyUserName, setMyPhotoPreset, setMyPhotoUpload.
 */

import { ref, onMounted } from "vue";
import {
  me as apiMe,
  setMyUserName,
  setMyPhotoPreset,
  setMyPhotoUpload,
} from "../services/api";
import LoadingSpinner from "../components/LoadingSpinner.vue";
import ErrorMsg from "../components/ErrorMsg.vue";

export default {
  name: "ProfileView",
  components: { LoadingSpinner, ErrorMsg },
  setup() {
    const loading = ref(true);
    const me = ref(null);
    const username = ref("");
    const savingName = ref(false);
    const error = ref("");
    const ok = ref(false);

    onMounted(fetchMe);

    async function fetchMe() {
      loading.value = true;
      error.value = "";
      ok.value = false;
      try {
        const r = await apiMe();
        me.value = r?.data?.user || r?.data || null;
        username.value = me.value?.username || me.value?.name || "";
      } catch (e) {
        error.value = e?.message || "Failed to load profile";
      } finally {
        loading.value = false;
      }
    }

    async function saveName() {
      if (!username.value.trim()) return;
      savingName.value = true;
      error.value = "";
      ok.value = false;
      try {
        await setMyUserName(username.value.trim());
        await fetchMe();
        ok.value = true;
        // persist local cache
        localStorage.setItem("me", JSON.stringify(me.value));
      } catch (e) {
        error.value = e?.message || "Failed to save username";
      } finally {
        savingName.value = false;
      }
    }

    async function setPreset(preset) {
      error.value = "";
      ok.value = false;
      try {
        await setMyPhotoPreset(preset);
        await fetchMe();
        ok.value = true;
        localStorage.setItem("me", JSON.stringify(me.value));
      } catch (e) {
        error.value = e?.message || "Failed to set preset photo";
      }
    }

    async function uploadPhoto(ev) {
      const f = ev?.target?.files?.[0];
      if (!f) return;
      error.value = "";
      ok.value = false;
      try {
        await setMyPhotoUpload(f);
        await fetchMe();
        ok.value = true;
        localStorage.setItem("me", JSON.stringify(me.value));
      } catch (e) {
        error.value = e?.message || "Failed to upload photo";
      } finally {
        ev.target.value = ""; // reset file input
      }
    }

    return {
      loading,
      me,
      username,
      savingName,
      error,
      ok,
      saveName,
      setPreset,
      uploadPhoto,
    };
  },
};
</script>

<style scoped>
.page { display: grid; grid-template-rows: auto 1fr; height: 100vh; }
.bar { display: flex; align-items: center; gap: 12px; padding: 12px; border-bottom: 1px solid #eee; }
.title { margin: 0; font-size: 18px; font-weight: 700; }
.body { padding: 12px 14px; display: grid; gap: 14px; max-width: 720px; }
.row { display: grid; grid-template-columns: 140px 1fr; gap: 10px; align-items: center; }
.inline { display: flex; gap: 8px; }
.photo { display: flex; align-items: center; gap: 12px; }
.avatar { width: 64px; height: 64px; border-radius: 14px; object-fit: cover; }
.btns { display: flex; gap: 8px; align-items: center; }
.ok { color: #2a7a2a; }
</style>
