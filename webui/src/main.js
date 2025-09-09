// src/main.js
import { createApp } from 'vue'
import App from './App.vue'
import router from './router' // 确保这行存在
import axios from './services/axios.js'
import ErrorMsg from './components/ErrorMsg.vue'
import LoadingSpinner from './components/LoadingSpinner.vue'

import './assets/dashboard.css'
import './assets/main.css'

const app = createApp(App)
app.config.globalProperties.$axios = axios
app.component("ErrorMsg", ErrorMsg)
app.component("LoadingSpinner", LoadingSpinner)
app.use(router) // 确保这行存在
app.mount('#app') // 确保这行存在