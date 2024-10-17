import { create } from "zustand";
import { createJSONStorage, persist } from "zustand/middleware";
import { TransactionHistory } from "@/models/history";
import { config } from "@/config";
import { OnChainMessageStatus } from "@consensys/linea-sdk";
import { isEmptyObject } from "@/utils/utils";

export type HistoryState = {
  isLoading: boolean;
  transactions: Record<string, TransactionHistory[]>;
};

export type HistoryActions = {
  setIsLoading: (isLoading: boolean) => void;
  setTransactions: (key: string, transactions: TransactionHistory[]) => void;
  getTransactionsByKey: (key: string) => TransactionHistory[];
  clearStorage: (key: string) => void;
  getMinEventBlockNumber: (key: string, fromChainId: number) => bigint;
};

export type HistoryStore = HistoryState & HistoryActions;

export const defaultInitState: HistoryState = {
  transactions: {},
  isLoading: false,
};

export const useHistoryStore = create<HistoryStore>()(
  persist(
    (set, get) => ({
      ...defaultInitState,
      setIsLoading: (isLoading) => set({ isLoading }),
      setTransactions: (key, transactions) =>
        set((state) => ({
          transactions: { ...state.transactions, [key]: transactions },
        })),
      clearStorage: (key) => {
        // eslint-disable-next-line @typescript-eslint/no-unused-vars
        const { [key]: _, ...newTransactions } = get().transactions;
        set({ transactions: newTransactions });
      },
      getTransactionsByKey: (key) => {
        const { transactions } = get();
        return transactions[key] ?? [];
      },
      getMinEventBlockNumber: (key, fromChainId) => {
        const { transactions } = get();

        if (isEmptyObject(transactions)) {
          return 0n;
        }
        let minBlockNumber = BigInt(Number.MAX_SAFE_INTEGER);

        const filteredTransactions = transactions[key].filter(
          (transaction) => transaction.fromChain.id === fromChainId,
        );

        filteredTransactions.forEach((transaction) => {
          if (transaction.message && transaction.message.status !== OnChainMessageStatus.CLAIMED) {
            if (transaction.event.blockNumber && transaction.event?.blockNumber < minBlockNumber) {
              minBlockNumber = transaction.event.blockNumber;
            }
          }
        });

        return minBlockNumber;
      },
    }),
    {
      name: "history-storage",
      version: parseInt(config.storage.minVersion),
      storage: createJSONStorage(() => localStorage, {
        replacer: (_, value) => {
          if (typeof value === "bigint") value = { __type: "bigint", value: value.toString() };
          if (value instanceof Map) value = { __type: "Map", value: Array.from(value.entries()) };
          return value;
        },
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        reviver: (_, value: any) => {
          if (value?.__type === "bigint") value = BigInt(value.value);
          if (value?.__type === "Map") value = new Map(value.value);
          return value;
        },
      }),
      migrate() {
        return defaultInitState;
      },
    },
  ),
);
