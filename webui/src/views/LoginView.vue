<template>
  <div class="login-container">
    <div class="green-card">
      <div class="leaf-decoration leaf-1"><i class="fas fa-leaf"></i></div>
      <div class="leaf-decoration leaf-2"><i class="fas fa-leaf"></i></div>
      
      <h2 class="login-title">WASAText • Login</h2>

      <form class="login-form" @submit.prevent="onSubmit">
        <!-- Username input -->
        <div class="input-group">
          <label class="input-label">Username</label>
          <input
            v-model="name"
            placeholder="Enter your username"
            required
            autofocus
            class="green-input"
          />
        </div>

        <!-- Submit button -->
        <button type="submit" :disabled="loading || !name" class="green-btn">
          <span v-if="loading" class="btn-loading"></span>
          {{ loading ? "Signing in…" : "Sign in" }}
        </button>

        <!-- Error message -->
        <ErrorMsg v-if="error" class="error-message">{{ error }}</ErrorMsg>
      </form>

      <!-- Hint text -->
      <div class="hint-text">
        <i class="fas fa-lightbulb"></i> 
        Tip: Any name will create/login to your account.
      </div>
    </div>
  </div>
</template>

<script>
import { ref } from "vue";
import { useRouter } from "vue-router";
import { doLogin } from "../services/api";
import ErrorMsg from "../components/ErrorMsg.vue";

export default {
  name: "LoginView",
  components: { ErrorMsg },
  setup() {
    const router = useRouter();
    const name = ref("");
    const loading = ref(false);
    const error = ref("");

    async function onSubmit() {
      loading.value = true;
      error.value = "";
      try {
        const res = await doLogin(name.value);
        const { user, token } = (res.data && res.data[0]) || {};

        localStorage.setItem("token", token || user?.id);
        localStorage.setItem("me", JSON.stringify(user));

        router.replace("/conversations");
      } catch (e) {
        error.value = e?.message || "Login failed";
      } finally {
        loading.value = false;
      }
    }

    return { name, loading, error, onSubmit };
  },
};
</script>

<style scoped>
.login-container {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 100vh;
  background: linear-gradient(135deg, #f5f9f1 0%, #e8f4ea 100%);
  padding: 20px;
}

.green-card {
  background: white;
  border-radius: 16px;
  box-shadow: 0 10px 30px rgba(56, 161, 105, 0.15);
  padding: 40px;
  width: 100%;
  max-width: 440px;
  position: relative;
  overflow: hidden;
  transition: transform 0.3s ease, box-shadow 0.3s ease;
}

.green-card:hover {
  transform: translateY(-5px);
  box-shadow: 0 15px 35px rgba(56, 161, 105, 0.2);
}

.green-card::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 6px;
  background: linear-gradient(90deg, #47c77a 0%, #2ecc71 50%, #27ae60 100%);
}

.leaf-decoration {
  position: absolute;
  font-size: 120px;
  color: rgba(46, 204, 113, 0.08);
  z-index: 0;
}

.leaf-1 {
  top: -40px;
  right: -30px;
  transform: rotate(30deg);
}

.leaf-2 {
  bottom: -50px;
  left: -40px;
  transform: rotate(-20deg);
}

.login-title {
  color: #2c3e50;
  text-align: center;
  margin-bottom: 30px;
  font-weight: 700;
  font-size: 28px;
  position: relative;
  z-index: 1;
}

.login-title::after {
  content: '';
  display: block;
  width: 60px;
  height: 4px;
  background: linear-gradient(90deg, #47c77a 0%, #27ae60 100%);
  margin: 10px auto 0;
  border-radius: 2px;
}

.login-form {
  display: grid;
  gap: 20px;
  position: relative;
  z-index: 1;
}

.input-group {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.input-label {
  color: #34495e;
  font-weight: 600;
  font-size: 14px;
}

.green-input {
  width: 100%;
  padding: 14px 16px;
  border: 2px solid #e0e0e0;
  border-radius: 10px;
  font-size: 16px;
  transition: all 0.3s ease;
  outline: none;
}

.green-input:focus {
  border-color: #2ecc71;
  box-shadow: 0 0 0 3px rgba(46, 204, 113, 0.2);
}

.green-input::placeholder {
  color: #a0a0a0;
}

.green-btn {
  background: linear-gradient(90deg, #47c77a 0%, #27ae60 100%);
  color: white;
  border: none;
  padding: 14px 20px;
  border-radius: 10px;
  font-size: 16px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.3s ease;
  width: 100%;
  display: flex;
  justify-content: center;
  align-items: center;
  position: relative;
  overflow: hidden;
}

.green-btn:hover:not(:disabled) {
  box-shadow: 0 5px 15px rgba(39, 174, 96, 0.4);
  transform: translateY(-2px);
}

.green-btn:active:not(:disabled) {
  transform: translateY(0);
}

.green-btn:disabled {
  background: linear-gradient(90deg, #9ee6bf 0%, #86dbab 100%);
  cursor: not-allowed;
  transform: none;
  box-shadow: none;
}

.btn-loading {
  width: 18px;
  height: 18px;
  border: 2px solid rgba(255, 255, 255, 0.3);
  border-radius: 50%;
  border-top-color: white;
  animation: spin 1s linear infinite;
  margin-right: 8px;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}

.error-message {
  margin-top: 10px;
}

.hint-text {
  text-align: center;
  color: #7f8c8d;
  font-size: 14px;
  margin-top: 24px;
  line-height: 1.5;
  position: relative;
  z-index: 1;
}

/* Responsive design */
@media (max-width: 480px) {
  .green-card {
    padding: 30px 20px;
  }
  
  .login-title {
    font-size: 24px;
  }
  
  .leaf-decoration {
    font-size: 100px;
  }
}
</style>