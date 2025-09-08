<script setup>
import { RouterLink, RouterView } from 'vue-router'
</script>

<template>
  <!-- Top navbar -->
  <header class="navbar navbar-dark sticky-top bg-dark flex-md-nowrap p-0 shadow">
    <a class="navbar-brand col-md-3 col-lg-2 me-0 px-3 fs-6" href="#/">WASAText</a>
    <button class="navbar-toggler position-absolute d-md-none collapsed" type="button"
      data-bs-toggle="collapse" data-bs-target="#sidebarMenu"
      aria-controls="sidebarMenu" aria-expanded="false" aria-label="Toggle navigation">
      <span class="navbar-toggler-icon"></span>
    </button>
  </header>

  <!-- Layout with sidebar -->
  <div class="container-fluid">
    <div class="row">
      <nav id="sidebarMenu" class="col-md-3 col-lg-2 d-md-block bg-light sidebar collapse">
        <div class="position-sticky pt-3 sidebar-sticky">
          <h6 class="sidebar-heading px-3 mt-4 mb-1 text-muted text-uppercase">
            <span>Main</span>
          </h6>
          <ul class="nav flex-column">
            <li class="nav-item">
              <RouterLink to="/conversations" class="nav-link">💬 Conversations</RouterLink>
            </li>
            <li class="nav-item">
              <RouterLink to="/groups" class="nav-link">👥 Groups</RouterLink>
            </li>
            <li class="nav-item">
              <RouterLink to="/me" class="nav-link">🙍 Profile</RouterLink>
            </li>
          </ul>

          <h6 class="sidebar-heading px-3 mt-4 mb-1 text-muted text-uppercase">
            <span>Account</span>
          </h6>
          <ul class="nav flex-column">
            <li class="nav-item">
              <RouterLink to="/login" class="nav-link">🚪 Logout</RouterLink>
            </li>
          </ul>
        </div>
      </nav>

      <main class="col-md-9 ms-sm-auto col-lg-10 px-md-4">
        <!-- Current route -->
        <RouterView />
      </main>
    </div>
  </div>
</template>
