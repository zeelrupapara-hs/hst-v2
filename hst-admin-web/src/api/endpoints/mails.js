import { request } from "@/api/client.js";

/**
 * Send a manager mail over the internal mailbox and/or email.
 * @param {{to?: string, logins?: number[], group_mask?: string, subject: string, body: string,
 *   internal?: boolean, email?: boolean, mail_server_id?: number}} body
 * @returns {Promise<{ok: boolean, data?: {sent: number, emails_queued: number, skipped_no_email: number}, message?: string}>}
 */
export const sendMail = (body) => request("/api/v1/mails", { method: "POST", body });

/**
 * Count who a To expression reaches, for the badge beside the field.
 * @param {{to?: string, logins?: number[], group_mask?: string}} body
 * @returns {Promise<{ok: boolean, data?: {count: number, with_email: number}, message?: string}>}
 */
export const previewMail = (body) => request("/api/v1/mails/preview", { method: "POST", body });

/**
 * List the caller's own compose templates.
 * @returns {Promise<{ok: boolean, data?: Array<{template_id: number, name: string, subject: string, body: string}>}>}
 */
export const fetchMailTemplates = () => request("/api/v1/mail-templates");

/**
 * Save a compose template; saving under a taken name overwrites.
 * @param {{name: string, subject: string, body: string}} body
 * @returns {Promise<{ok: boolean, data?: object, message?: string}>}
 */
export const saveMailTemplate = (body) => request("/api/v1/mail-templates", { method: "POST", body });

/**
 * Delete one of the caller's compose templates.
 * @param {number} id
 * @returns {Promise<{ok: boolean, message?: string}>}
 */
export const deleteMailTemplate = (id) => request(`/api/v1/mail-templates/${id}`, { method: "DELETE" });
