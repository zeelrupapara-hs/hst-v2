import api from "../api";
import { alertFromFormula } from "../adapt";

export const createAlert = (data) => api.post("/alerts", alertFromFormula(data));

// ours replaces the whole alert, so the formula is expanded again rather than patched
export const updateAlert = (id, data) => api.put(`/alerts/${id}`, alertFromFormula(data));

export const deleteAlert = (id) => api.delete(`/alerts/${id}`);
