import axios from "axios";
import {
  clearLocalStorage,
  getLocalItem,
  errorToast,
  setLocalItem,
} from "../utils/utils";
import { getRefreshToken } from "./request/auth";

let isRefreshing = false;
let failedQueue = [];

const processQueue = (error, token = null) => {
  failedQueue.forEach((prom) => {
    if (error) prom.reject(error);
    else prom.resolve(token);
  });

  failedQueue = [];
};

const logout = () => {
  clearLocalStorage();
  window.location.href = "/login";
};

const api = axios.create({
  baseURL: `${import.meta.env.VITE_BASE_URL}/api/trader/v1`,
});

api.interceptors.request.use((config) => {
  const token = getLocalItem("token");
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

api.interceptors.response.use(
  (response) => response,
  async (error) => {
    const originalRequest = error.config;

    if (error?.response?.status === 401 && !originalRequest._retry) {
      originalRequest._retry = true;

      const refreshToken = getLocalItem("refreshToken");
      if (!refreshToken) {
        logout();
        return Promise.reject(error);
      }

      if (isRefreshing) {
        return new Promise((resolve, reject) => {
          failedQueue.push({ resolve, reject });
        })
          .then((token) => {
            originalRequest.headers.Authorization = `Bearer ${token}`;
            return api(originalRequest);
          })
          .catch((error) => {
            return Promise.reject(error);
          });
      }

      isRefreshing = true;

      try {
        const { data } = await getRefreshToken({ refresh_token: refreshToken });
        const { access_token, refresh_token } = data?.data;

        setLocalItem("token", access_token);
        setLocalItem("refreshToken", refresh_token);

        originalRequest.headers.Authorization = `Bearer ${access_token}`;
        processQueue(null, access_token);

        return api(originalRequest);
      } catch (error) {
        processQueue(error, null);
        logout();
        return Promise.reject(error);
      } finally {
        isRefreshing = false;
      }
    }

    const message =
      error?.response?.data?.message ||
      error?.message ||
      "Something went wrong. Please try again later.";
    errorToast(message);
    return Promise.reject(error);
  }
);

export default api;
