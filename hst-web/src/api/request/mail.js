import api from "../api";

// the terminal sends status 1 to send and 3 to keep a draft; ours takes the draft flag itself,
// and picks the support recipient on the server, so the addressee is not the caller's to choose
const toMail = ({ subject, body, status }) => ({
  subject,
  body,
  draft: status === 3,
});

export const createMail = (data) => api.post("/mails", toMail(data));
export const updateMail = (id, data) => api.put(`/mails/${id}`, toMail(data));
export const deleteMail = (id) => api.delete(`/mails/${id}`);
