import axios from "./axios"

const instance = axios.create({
  baseURL: __API_URL__,   // 由上面的 vite.config.js 注入
  timeout: 10000
});

export default instance;
