// English: bootstrap Vue app, provide $axios, register global components
import { createApp } from "vue";
import App from "./App.vue";
import router from "./router/index.js";
import axios from "./services/axios.js";

import "./assets/main.css";
import "./assets/dashboard.css";


import LoadingSpinner from "./components/LoadingSpinner.vue";
import ErrorMsg from "./components/ErrorMsg.vue";

const app = createApp(App);

// global provide for Options API components
app.config.globalProperties.$axios = axios;
// provide/inject for Composition API
app.provide("axios", axios);

app.component("LoadingSpinner", LoadingSpinner);
app.component("ErrorMsg", ErrorMsg);

app.use(router).mount("#app");
