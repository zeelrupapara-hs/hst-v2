import { create } from "zustand";
import {
  getAllDraftMails,
  getAllInboxMails,
  getAllOutboxMails,
  getAllTrashMails,
} from "../api/request/activity";

const useMailStore = create((set, get) => ({
  inboxMails: [],
  outboxMails: [],
  draftMails: [],
  trashMails: [],
  loading: false,
  error: null,

  fetchMails: async () => {
    const { fetchInbox, fetchOutbox, fetchDraft, fetchTrash } = get();

    try {
      set({ loading: true });
      await Promise.all([
        fetchInbox(),
        fetchOutbox(),
        fetchDraft(),
        fetchTrash(),
      ]);
    } catch (error) {
      set({ error: error });
    } finally {
      set({ loading: false });
    }
  },

  fetchInbox: async () => {
    set({ loading: true });
    try {
      const { data } = await getAllInboxMails();
      set({ inboxMails: data?.data });
    } catch (error) {
      set({ error: error?.response?.data?.message || error?.message });
    } finally {
      set({ loading: false });
    }
  },

  fetchOutbox: async () => {
    set({ loading: true });
    try {
      const { data } = await getAllOutboxMails();
      set({ outboxMails: data?.data });
    } catch (error) {
      set({ error: error?.response?.data?.message || error?.message });
    } finally {
      set({ loading: false });
    }
  },

  fetchDraft: async () => {
    set({ loading: true });
    try {
      const { data } = await getAllDraftMails();
      set({ draftMails: data?.data });
    } catch (error) {
      set({ error: error?.response?.data?.message || error?.message });
    } finally {
      set({ loading: false });
    }
  },

  fetchTrash: async () => {
    set({ loading: true });
    try {
      const { data } = await getAllTrashMails();
      set({ trashMails: data?.data });
    } catch (error) {
      set({ error: error?.response?.data?.message || error?.message });
    } finally {
      set({ loading: false });
    }
  },

  addInboxMail: (mail) => {
    set((state) => ({ inboxMails: [...state.inboxMails, mail] }));
  },

  addOutboxMail: (mail) => {
    set((state) => ({
      outboxMails: [...state.outboxMails, mail],
      draftMails: state.draftMails.filter(
        (draft) => draft?.common_id !== mail?.common_id
      ),
    }));
  },

  addDraftMail: (mail) => {
    set((state) => ({ draftMails: [...state.draftMails, mail] }));
  },

  trashMail: (id) => {
    const { inboxMails, outboxMails } = get();

    const allMails = [...inboxMails, ...outboxMails];
    const mail = allMails.find((mail) => mail?.id === id);

    set((state) => ({
      inboxMails: state.inboxMails.filter((mail) => mail?.id !== id),
      outboxMails: state.outboxMails.filter((mail) => mail?.id !== id),
      trashMails: [...state.trashMails, mail],
    }));
  },

  deleteMail: (id) => {
    set((state) => ({
      draftMails: state.draftMails.filter((mail) => mail?.id !== id),
      trashMails: state.trashMails.filter((mail) => mail?.id !== id),
    }));
  },
}));

export default useMailStore;
