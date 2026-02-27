import { createSelectorHooks, ZustandHookSelectors } from "auto-zustand-selectors-hook";
import { create } from "zustand";

export type NativeBridgeNavigationState = {
  isBridgeOpen: boolean;
  isTransactionHistoryOpen: boolean;
  hideNoFeesPill: boolean;
};

export type NativeBridgeNavigationActions = {
  setIsBridgeOpen: (opened: boolean) => void;
  setIsTransactionHistoryOpen: (opened: boolean) => void;
  setHideNoFeesPill: (hidden: boolean) => void;
};

export type NativeBridgeNavigationStore = NativeBridgeNavigationState & NativeBridgeNavigationActions;

export const defaultInitState: NativeBridgeNavigationState = {
  isBridgeOpen: true,
  isTransactionHistoryOpen: false,
  hideNoFeesPill: false,
};

const useNativeBridgeNavigationStoreBase = create<NativeBridgeNavigationStore>()((set) => ({
  ...defaultInitState,
  setIsBridgeOpen: (opened) => set({ isBridgeOpen: opened }),
  setIsTransactionHistoryOpen: (opened) => set({ isTransactionHistoryOpen: opened }),
  setHideNoFeesPill: (hidden) => set({ hideNoFeesPill: hidden }),
}));

export const useNativeBridgeNavigationStore = createSelectorHooks(
  useNativeBridgeNavigationStoreBase,
) as unknown as typeof useNativeBridgeNavigationStoreBase & ZustandHookSelectors<NativeBridgeNavigationStore>;
