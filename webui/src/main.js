import { createApp } from 'vue'
import App from './App.vue'
import router from './router'
import axios from './services/axios.js';

import ErrorMsg from './components/ErrorMsg.vue'
import LoadingSpinner from './components/LoadingSpinner.vue'

// Load Bootstrap assets for layout and icons.
import 'bootstrap/dist/css/bootstrap.min.css'
import 'bootstrap/dist/js/bootstrap.bundle.min.js'
import 'bootstrap-icons/font/bootstrap-icons.css'

// Load application-specific stylesheets.
import './assets/dashboard.css'
import './assets/main.css'

const app = createApp(App)
app.config.globalProperties.$axios = axios;
// Register shared UI components for use throughout the app.
app.component('ErrorMsg', ErrorMsg)
app.component('LoadingSpinner', LoadingSpinner)

app.use(router)

// Log uncaught component errors to aid debugging in development.
app.config.errorHandler = (err, vm, info) => {
  console.error('Global error:', err)
  console.error('Error info:', info)
}

// Mount the root Vue instance.
app.mount('#app')
