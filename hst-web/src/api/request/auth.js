import api from "../api";
import authApi from "../authApi";

// this is a web terminal, and the platform has a connection type that says so; sending the
// plain client type made every journal line claim it came from a desktop terminal
const CONNECTION_TYPE_WEB = 11;

// the form collects a name in two fields and a numeric trade_type; the server takes one name
// and a named type, and issues the passwords itself
const registerBody = ({ first_name, last_name, trade_type, address, ...rest }) => ({
  ...rest,
  name: [first_name, last_name].filter(Boolean).join(" "),
  type: trade_type === 1 ? "demo" : "real",
  city: address,
});

export const register = (data) => authApi.post("/register", registerBody(data));
export const login = (headers) =>
  authApi.post("/login", { connection_type: CONNECTION_TYPE_WEB }, headers);
export const getRefreshToken = (data) => authApi.post("/refresh", data);

export const forgetPasswordSendOtp = (data) => authApi.post("/forgot-password", data);
export const forgetPasswordVerifyOtp = (data) => authApi.post("/verify-code", data);
export const forgotPassword = (data) => authApi.post("/reset-password", data);

export const resetPassword = (data) => api.post("/auth/change-password", data);
export const updateProfile = (id, data) => api.put(`/users/${id}`, data);
export const getProfile = () => api.get("/profile");
