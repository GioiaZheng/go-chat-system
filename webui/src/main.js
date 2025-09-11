import { createApp } from 'vue';
import App from './App.vue';
import router from './router';
import axios from './services/axios.js';

import ErrorMsg from './components/ErrorMsg.vue';
import LoadingSpinner from './components/LoadingSpinner.vue';

// Bootstrap & icons
import 'bootstrap/dist/css/bootstrap.min.css';
import 'bootstrap/dist/js/bootstrap.bundle.min.js';
import 'bootstrap-icons/font/bootstrap-icons.css';

// Custom styles
import './assets/dashboard.css';
import './assets/main.css';

const app = createApp(App);

// legacy-friendly: $axios available everywhere via `this.$axios`
app.config.globalProperties.$axios = axios;

// register global components
app.component('ErrorMsg', ErrorMsg);
app.component('LoadingSpinner', LoadingSpinner);

// vue-router
app.use(router);

// global error handler (optional but useful)
app.config.errorHandler = (err, vm, info) => {
  console.error('Global error:', err);
  console.error('Error info:', info);
};

app.mount('#app');
