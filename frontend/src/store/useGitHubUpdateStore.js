import { create } from "zustand";

export const useGitHubUpdateStore = create((set, get) => ({
  updateInfo: null,
  loading: false,
  syncing: false,
  modalOpen: false,
  syncOutput: null,
  error: null,

  setModalOpen: (modalOpen) => set({ modalOpen }),

  fetchStatus: async () => {
    try {
      const res = await fetch("/api/version");
      if (!res.ok) return null;
      const data = await res.json();
      set({ updateInfo: data, error: null });
      return data;
    } catch (err) {
      set({ error: err.message });
      return null;
    }
  },

  checkNow: async () => {
    set({ loading: true, error: null });
    try {
      const res = await fetch("/api/version/check", { method: "POST" });
      const data = await res.json();
      const updated = data.status || data;
      set({ updateInfo: updated, loading: false });
      return updated;
    } catch (err) {
      set({ loading: false, error: err.message });
      return null;
    }
  },

  syncNow: async () => {
    set({ syncing: true, error: null, syncOutput: null });
    try {
      const res = await fetch("/api/version/sync", { method: "POST" });
      const data = await res.json();
      if (!res.ok || data.success === false) {
        throw new Error(data.error || "Gagal melakukan sync dari GitHub");
      }
      set({
        syncing: false,
        syncOutput: data.output || data.message || "Berhasil di-sync ke commit terbaru",
        updateInfo: data.status || { ...(get().updateInfo || {}), hasUpdate: false }
      });
      // Re-fetch status to refresh state
      setTimeout(() => get().fetchStatus(), 1500);
      return data;
    } catch (err) {
      set({ syncing: false, error: err.message });
      return null;
    }
  }
}));
