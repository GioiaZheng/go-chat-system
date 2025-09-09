// English: Axios instance with baseURL + auth interceptor
import axios from "axios";

const instance = axios.create({
  baseURL: import.meta.env.VITE_API_BASE || "http://localhost:3000/api/v1",
  timeout: 15000,
});

instance.interceptors.request.use((config) => {
  const token = localStorage.getItem("token");
  if (token) config.headers.Authorization = `Bearer ${token}`;
  return config;
});

instance.interceptors.response.use(
  (r) => r,
  (err) => {
    if (err?.response?.status === 401) {
      localStorage.removeItem("token");
      // optional redirect can be done in views based on error handling
    }
    return Promise.reject(err);
  }
);

export default instance;
