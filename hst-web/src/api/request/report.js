import api from "../api";

export const createReport = (data) => api.post("/reports", data);
export const deleteReport = (id) => api.delete(`/reports/${id}`);
export const downloadReport = (id, headers) => api.get(`/public/reports/${id}`, headers);
