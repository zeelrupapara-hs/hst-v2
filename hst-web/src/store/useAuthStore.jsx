import { create } from "zustand";
import { persist } from "zustand/middleware";
import { removeLocalItem } from "../utils/utils";

const useAuthStore = create(
  persist(
    (set) => ({
      isLogin: false,
      user: null,

      setIsLogin: (isLogin) => set({ isLogin }),
      setUser: (user) => set({ user }),
      updateUser: (user) => set({ user }),

      logout: () => {
        removeLocalItem("token");
        removeLocalItem("refreshToken");
        set({ isLogin: false, user: null });
      },
    }),
    { name: "auth-store" } // localStorage key
  )
);

export default useAuthStore;
