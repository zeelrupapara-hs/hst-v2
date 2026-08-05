import { request } from "@/api/client.js";

export const fetchAccountPositions = (login) => request(`/api/v1/positions/accounts/${login}`);
export const fetchAccountOrders = (login) => request(`/api/v1/orders/accounts/${login}`);
