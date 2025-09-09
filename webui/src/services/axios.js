// src/services/axios.js
import axios from 'axios';

// 创建 axios 实例
const instance = axios.create({
  baseURL: 'http://localhost:3000', // 后端 API 地址
  timeout: 10000, // 请求超时时间
});

// 请求拦截器 - 添加认证 token
instance.interceptors.request.use(
  (config) => {
    // 从 localStorage 获取 token
    const token = localStorage.getItem('token');
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);

// 响应拦截器 - 处理错误
instance.interceptors.response.use(
  (response) => {
    return response;
  },
  (error) => {
    if (error.response?.status === 401) {
      // token 过期或无效，跳转到登录页
      localStorage.removeItem('token');
      localStorage.removeItem('me');
      window.location.href = '/login';
    }
    return Promise.reject(error);
  }
);

export default instance;