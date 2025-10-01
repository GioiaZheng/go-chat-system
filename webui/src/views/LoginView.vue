<template>
  <div class="container d-flex align-items-center justify-content-center min-vh-100">
    <div class="card shadow-sm" style="max-width: 420px; width: 100%;">
      <div class="card-body">
        <h1 class="h4 mb-3 text-center">Sign in</h1>

        <ErrorMsg v-if="err" :text="err" class="mb-3" />

        <form @submit.prevent="login" novalidate>
          <div class="mb-3">
            <label class="form-label">Name</label>
            <input
              v-model.trim="name"
              type="text"
              class="form-control"
              placeholder="alice"
              autocomplete="username"
              required
              minlength="3"
              maxlength="16"
            />
          </div>

          <button class="btn btn-dark w-100" :disabled="busy || !name">
            <span
              v-if="busy"
              class="spinner-border spinner-border-sm me-2"
              role="status"
              aria-hidden="true"
            ></span>
            {{ busy ? 'Signing in…' : 'Login' }}
          </button>
        </form>
      </div>
    </div>
  </div>
</template>
