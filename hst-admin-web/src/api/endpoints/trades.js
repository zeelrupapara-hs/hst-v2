import { request } from "@/api/client.js";

export const fetchAccountPositions = (login) => request(`/api/v1/positions/accounts/${login}`);
export const fetchAccountOrders = (login) => request(`/api/v1/orders/accounts/${login}`);

export const fetchAllPositions = () => request("/api/v1/positions?limit=500");
export const fetchAllOrders = () => request("/api/v1/orders?limit=500");
export const fetchAllDeals = () => request("/api/v1/deals?limit=500");
export const fetchAccountDeals = (login) => request(`/api/v1/deals/accounts/${login}`);
