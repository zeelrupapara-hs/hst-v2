import api from "../api";

// the terminal sends status 1 to send and 3 to keep a draft; ours takes the draft flag itself.
// `to` is the broker mailbox login picked from the dropdown, MT5-style.
const toMail = ({ to, subject, body, status, replyTo, attachmentIds }) => ({
  mailbox_login: to || 0,
  subject,
  body,
  draft: status === 3,
  reply_to: replyTo || "",
  attachment_ids: attachmentIds || [],
});

export const createMail = (data) => api.post("/mails", toMail(data));
export const updateMail = (id, data) => api.put(`/mails/${id}`, toMail(data));
export const deleteMail = (id) => api.delete(`/mails/${id}`);
export const getMail = (id) => api.get(`/mails/${id}`);

// the To dropdown: managers with a mailbox name
export const getMailboxes = () => api.get("/mailboxes");

export const uploadMailAttachments = (files) => {
  const form = new FormData();
  files.forEach((f) => form.append("files", f));
  return api.post("/mails/attachments", form, {
    headers: { "Content-Type": "multipart/form-data" },
  });
};

export const downloadMailAttachment = async (id, name) => {
  const res = await api.get(`/mails/attachments/${id}`, { responseType: "blob" });
  const url = URL.createObjectURL(res.data);
  const a = document.createElement("a");
  a.href = url;
  a.download = name;
  a.click();
  URL.revokeObjectURL(url);
};
