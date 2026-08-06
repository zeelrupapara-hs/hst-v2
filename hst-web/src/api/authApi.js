import axios from "axios";
import { errorToast } from "../utils/utils";

const authApi = axios.create({
  baseURL: `${import.meta.env.VITE_BASE_URL}/auth/trader/v1`,
});

authApi.interceptors.response.use(
  (response) => response,
  (error) => {
    const message =
      error?.response?.data?.message ||
      error?.message ||
      "Something went wrong. Please try again later.";
    errorToast(message);
    return Promise.reject(error);
  }
);

export default authApi;
