import { isLoggedIn } from "../lib/api.js";

/** Whether the app is showing demo/mock data (not signed in). */
export function useDemo() {
  return !isLoggedIn();
}
