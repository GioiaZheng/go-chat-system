<template>
  <main class="login-bg min-vh-100 d-flex align-items-center">
    <div class="container">
      <div class="row justify-content-center">
        <div class="col-11 col-sm-9 col-md-7 col-lg-5">
          <div class="card border-0 shadow-lg rounded-4 overflow-hidden">
            <div class="card-body p-4 p-sm-5">
              <!-- 品牌 + 标题 -->
              <div class="text-center mb-4">
                <div class="brand mb-1">
                  WASA <span class="text-gradient">Chat</span>
                </div>
                <h1 class="h5 text-secondary fw-semibold m-0">Sign in</h1>
              </div>

              <ErrorMsg v-if="err" :text="err" class="mb-3" />

              <form @submit.prevent="login" novalidate>
                <label class="form-label fw-medium">Name</label>
                <input
                  v-model.trim="name"
                  type="text"
                  class="form-control form-control-lg"
                  placeholder="alice"
                  autocomplete="username"
                  required
                  minlength="3"
                  maxlength="16"
                  autofocus
                />

                <button
                  class="btn btn-gradient btn-lg w-100 mt-3"
                  :disabled="busy || (name?.length ?? 0) < 3"
                >
                  <span
                    v-if="busy"
                    class="spinner-border spinner-border-sm me-2"
                    role="status"
                    aria-hidden="true"
                  ></span>
                  {{ busy ? 'Signing in…' : 'Login' }}
                </button>
              </form>

              <p class="text-center text-muted small mt-3 mb-0">
                Press <kbd>Enter</kbd> to login
              </p>
            </div>
          </div>
        </div>
      </div>
    </div>
  </main>
</template>

<style scoped>
.login-bg{
  background:
    radial-gradient(1200px 1200px at 10% 10%, #f3f8ff 0, transparent 60%),
    radial-gradient(1000px 1000px at 90% 20%, #f7f7ff 0, transparent 60%),
    linear-gradient(180deg, #f8fafc, #f4f6f9);
}

.brand{
  font-size: 1.65rem;
  font-weight: 800;
  letter-spacing: .5px;
}

.text-gradient{
  background: linear-gradient(90deg, #22c55e, #3b82f6);
  -webkit-background-clip: text;
  background-clip: text;
  color: transparent;
}

.btn-gradient{
  background-image: linear-gradient(135deg, #22c55e 0%, #16a34a 45%, #3b82f6 120%);
  color: #fff;
  border: 0;
  box-shadow: 0 .5rem 1.25rem rgba(16,185,129,.25);
}
.btn-gradient:disabled{ opacity:.6 }
.card{ backdrop-filter: blur(6px); }
</style>
