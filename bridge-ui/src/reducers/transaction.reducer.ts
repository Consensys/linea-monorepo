// transactionsReducer.ts

import { NetworkType } from '@/contexts/chain.context';
import { TransactionHistory } from '@/models/history';

// Action types
type ActionType =
  | {
      type: 'SET_TRANSACTIONS';
      networkType: NetworkType;
      transactions: TransactionHistory[];
    }
  | { type: 'CLEAR_TRANSACTIONS'; networkType: NetworkType }
  | {
      type: 'UPDATE_TRANSACTION_STATUS';
      networkType: NetworkType;
      transactionHash: string;
      status: {
        pending: boolean;
        success: boolean;
        error: boolean;
        checking: boolean;
      };
    };

type TransactionState = {
  [NetworkType.MAINNET]: TransactionHistory[];
  [NetworkType.SEPOLIA]: TransactionHistory[];
};

// Reducer
const transactionReducer = (state: TransactionState, action: ActionType): TransactionState => {
  switch (action.type) {
    case 'SET_TRANSACTIONS':
      return {
        ...state,
        [action.networkType]: action.transactions,
      };
    case 'CLEAR_TRANSACTIONS':
      return {
        ...state,
        [action.networkType]: [],
      };
    case 'UPDATE_TRANSACTION_STATUS':
      if (action.networkType !== NetworkType.MAINNET && action.networkType !== NetworkType.SEPOLIA) {
        return {
          ...state,
          [action.networkType]: [],
        };
      }

      return {
        ...state,
        [action.networkType]: state[action.networkType].map((tx) =>
          tx.transactionHash === action.transactionHash ? { ...tx, ...action.status } : tx,
        ),
      };
    default:
      return state;
  }
};

// Transactions type from state
export const getTransactionsByNetwork = (state: TransactionState, network: NetworkType): TransactionHistory[] => {
  return state[network as keyof TransactionState] || [];
};

export type { ActionType, TransactionState };
export default transactionReducer;
