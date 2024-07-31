import { config } from "@/config";
import { create } from "zustand";
import { createJSONStorage, persist } from "zustand/middleware";

export type ConfigState = {
  agreeToTerms: boolean;
};

export type ConfigActions = {
  setAgreeToTerms: (agree: boolean) => void;
};

export type ConfigStore = ConfigState & ConfigActions;

export const defaultInitState: ConfigState = {
  agreeToTerms: false,
};

export const useConfigStore = create<ConfigStore>()(
  persist(
    (set) => ({
      ...defaultInitState,
      setAgreeToTerms: (agree) => set({ agreeToTerms: agree }),
    }),
    {
      name: "config-storage",
      version: parseInt(config.storage.minVersion),
      storage: createJSONStorage(() => localStorage),
      migrate: () => {
        return defaultInitState;
      },
    },
  ),
);
