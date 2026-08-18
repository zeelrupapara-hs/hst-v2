import { downloadFile, request, requestForm } from "@/api/client.js";

/**
 * Send a manager mail over the internal mailbox and/or email.
 * @param {{to?: string, logins?: number[], group_mask?: string, subject: string, body: string,
 *   internal?: boolean, email?: boolean, mail_server_id?: number}} body
 * @returns {Promise<{ok: boolean, data?: {sent: number, emails_queued: number, skipped_no_email: number}, message?: string}>}
 */
export const sendMail = (body) => request("/api/v1/mails", { method: "POST", body });

/**
 * List the caller's own mailbox: type inbox|outbox|draft|bin, or one thread by thread_id.
 * @param {{type?: string, thread_id?: string}} [q]
 * @returns {Promise<{ok: boolean, data?: Array<object>, message?: string}>}
 */
export const fetchMails = (q = {}) => {
  const p = new URLSearchParams(Object.entries(q).filter(([, v]) => v)).toString();
  return request(`/api/v1/mails${p ? `?${p}` : ""}`);
};

/**
 * Read one message (marks it read when the caller is the recipient), attachments included.
 * @param {string} trackingId
 */
export const fetchMail = (trackingId) => request(`/api/v1/mails/${trackingId}`);

/**
 * Bin a message, or purge it when it is already in the bin.
 * @param {string} trackingId
 */
export const deleteMail = (trackingId) => request(`/api/v1/mails/${trackingId}`, { method: "DELETE" });

/**
 * Stage attachment files for a send; returns their ids to pass in attachment_ids.
 * @param {File[]} files
 * @returns {Promise<{ok: boolean, data?: Array<{attachment_id: number, name: string, size: number}>, message?: string}>}
 */
export const uploadMailAttachments = (files) => {
  const form = new FormData();
  files.forEach((f) => form.append("files", f));
  return requestForm("/api/v1/mails/attachments", form);
};

/** Download one attachment through the authorized API. */
export const downloadMailAttachment = (id, name) =>
  downloadFile(`/api/v1/mails/attachments/${id}`, name);

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
