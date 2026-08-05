import { request } from "@/api/client.js";

/** The whole navigator: nodes, counts and the can{} rights map, already rights-filtered. */
export const fetchNavigation = () => request("/api/v1/navigation");
