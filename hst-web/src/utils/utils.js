import dayjs from "dayjs";
import toast from "react-hot-toast";
import { ORDER_TYPES, POSITION_SIDE } from "./constants";

export const setLocalItem = (key, value) => {
  localStorage.setItem(key, JSON.stringify(value));
};

export const getLocalItem = (key) => {
  const item = localStorage.getItem(key);
  return item ? JSON.parse(item) : null;
};

export const removeLocalItem = (key) => {
  localStorage.removeItem(key);
};

export const clearLocalStorage = () => {
  localStorage.clear();
};

export const successToast = (message) => {
  toast.success(message);
};

export const errorToast = (message) => {
  toast.error(message);
};

export const qs = (params) => {
  if (!params || typeof params !== "object") return "";
  const queryString = new URLSearchParams(params).toString();
  return queryString ? `${queryString}` : "";
};

export const getInitials = (name) => {
  if (!name) return null;

  const nameParts = name.trim()?.split(" ");
  if (nameParts) {
    const initials = nameParts[0][0] + nameParts[nameParts.length - 1][0];
    return initials.toUpperCase();
  }
};

export const sortByNumber = (data, key, order = "desc") => {
  if (!Array.isArray(data) || !data.length) return [];

  return [...data].sort((a, b) => {
    const valueA = a[key] || 0;
    const valueB = b[key] || 0;

    return order === "asc" ? valueA - valueB : valueB - valueA;
  });
};

export const defaultRender = (value) => (value ?? "--");

export const formateProfit = (number) => {
  if (!number) return "$0";
  const formattedNumber = Number(Math.abs(number)).toFixed(2);
  if (number < 0) return `-$${formattedNumber}`;
  else return `$${formattedNumber}`;
};

export const formatDate = (date, format = "DD-MM-YY HH:mm:ss") => {
  if (!date) return null;

  // Convert nanoseconds to milliseconds
  const milliseconds = Math.floor(date / 1_000_000);

  const parsed = dayjs(milliseconds);
  return parsed.isValid() ? parsed.format(format) : null;
};

export const getSystemTheme = () => {
  return window.matchMedia("(prefers-color-scheme: dark)").matches
    ? "dark"
    : "light";
};

export const getDefaultTheme = () => {
  if (getLocalItem("theme")) return getLocalItem("theme");
  return getSystemTheme();
};

export const openLink = (url, target = "_blank") => {
  if (!url) return;
  window.open(url, target, "noopener,noreferrer");
};

export const blobToUtf8 = async (blob) => {
  const text = await new Response(blob).text();
  return text;
};

export const getOptions = (obj) => {
  return Object.entries(obj).map(([value, label]) => ({
    value: Number(value),
    label: label,
  }));
};

export const getOrderType = (type, side) => {
  if (type) return ORDER_TYPES[type];
  return POSITION_SIDE[side];
};
