import { Address, serialize, deserialize } from 'wagmi';
import log from 'loglevel';

export enum StorageKeys {
  TRANSACTIONS = 'transactions',
  HASHES = 'hashes',
  USER_TOKENS = 'user_tokens',
  STORAGE_VERSION = 'storage_version',
}

export const serializeWithBigInt = (object: unknown) => {
  return JSON.parse(JSON.stringify(object, (key, value) => (typeof value === 'bigint' ? value.toString() : value)));
};

export const generateKey = (baseKey: string, address: Address, networkType: string) => {
  return `${address}_${networkType}_${baseKey}`;
};

export const loadFromLocalStorage = (key: string, defaultValue: any) => {
  try {
    const serializedState = localStorage.getItem(key);
    if (serializedState === null) {
      return defaultValue;
    }
    return deserialize(serializedState);
  } catch (e) {
    log.warn(`Error getting data from localStorage for key “${key}”:`, e);
    return defaultValue;
  }
};

export const saveToLocalStorage = (key: string, value: any) => {
  try {
    const serializedState = serialize(value);
    localStorage.setItem(key, serializedState);
  } catch (e) {
    log.warn(`Error setting localStorage for key “${key}”:`, e);
  }
};