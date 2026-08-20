import api from "../api";

export const getChartLayouts = () =>
  api.get("/charts").then((r) => r.data?.data ?? []);

export const saveChartLayout = (chartId, content) =>
  api.put(`/charts/${chartId}`, { content });

export const getChartSettings = () =>
  api.get("/charts/settings").then((r) => r.data?.data ?? {});

export const saveChartSettings = (content) =>
  api.put("/charts/settings", { content });
