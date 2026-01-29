/**
 * IndexedDB Storage Layer
 *
 * Provides a typed, promise-based wrapper around IndexedDB for
 * storing verification sessions and files in the browser.
 */

import type { StoredSession } from "@/services/types";

// ============================================================================
// Constants
// ============================================================================

const DB_NAME = "linea-verifier-ui";
const DB_VERSION = 1;
const SESSIONS_STORE = "sessions";

// ============================================================================
// Database Connection
// ============================================================================

let dbInstance: IDBDatabase | null = null;
let dbPromise: Promise<IDBDatabase> | null = null;

/**
 * Opens the IndexedDB database, creating object stores if needed.
 */
function openDatabase(): Promise<IDBDatabase> {
  return new Promise((resolve, reject) => {
    const request = indexedDB.open(DB_NAME, DB_VERSION);

    request.onerror = () => {
      reject(new Error(`Failed to open IndexedDB: ${request.error?.message}`));
    };

    request.onsuccess = () => {
      resolve(request.result);
    };

    request.onupgradeneeded = (event) => {
      const db = (event.target as IDBOpenDBRequest).result;

      // Create sessions store if it doesn't exist
      if (!db.objectStoreNames.contains(SESSIONS_STORE)) {
        const store = db.createObjectStore(SESSIONS_STORE, { keyPath: "id" });
        store.createIndex("createdAt", "createdAt", { unique: false });
      }
    };
  });
}

/**
 * Gets the database instance, opening it if necessary.
 */
export async function getDatabase(): Promise<IDBDatabase> {
  if (dbInstance) {
    return dbInstance;
  }

  if (!dbPromise) {
    dbPromise = openDatabase().then((db) => {
      dbInstance = db;

      // Handle connection close
      db.onclose = () => {
        dbInstance = null;
        dbPromise = null;
      };

      return db;
    });
  }

  return dbPromise;
}

/**
 * Closes the database connection.
 */
export function closeDatabase(): void {
  if (dbInstance) {
    dbInstance.close();
    dbInstance = null;
    dbPromise = null;
  }
}

// ============================================================================
// Session Operations
// ============================================================================

/**
 * Saves a session to IndexedDB.
 */
export async function saveSession(session: StoredSession): Promise<void> {
  const db = await getDatabase();

  return new Promise((resolve, reject) => {
    const transaction = db.transaction(SESSIONS_STORE, "readwrite");
    const store = transaction.objectStore(SESSIONS_STORE);
    const request = store.put(session);

    request.onerror = () => {
      reject(new Error(`Failed to save session: ${request.error?.message}`));
    };

    request.onsuccess = () => {
      resolve();
    };
  });
}

/**
 * Retrieves a session from IndexedDB.
 */
export async function getSession(sessionId: string): Promise<StoredSession | null> {
  const db = await getDatabase();

  return new Promise((resolve, reject) => {
    const transaction = db.transaction(SESSIONS_STORE, "readonly");
    const store = transaction.objectStore(SESSIONS_STORE);
    const request = store.get(sessionId);

    request.onerror = () => {
      reject(new Error(`Failed to get session: ${request.error?.message}`));
    };

    request.onsuccess = () => {
      resolve(request.result ?? null);
    };
  });
}

/**
 * Deletes a session from IndexedDB.
 */
export async function deleteSession(sessionId: string): Promise<void> {
  const db = await getDatabase();

  return new Promise((resolve, reject) => {
    const transaction = db.transaction(SESSIONS_STORE, "readwrite");
    const store = transaction.objectStore(SESSIONS_STORE);
    const request = store.delete(sessionId);

    request.onerror = () => {
      reject(new Error(`Failed to delete session: ${request.error?.message}`));
    };

    request.onsuccess = () => {
      resolve();
    };
  });
}

/**
 * Lists all sessions in IndexedDB.
 */
export async function listSessions(): Promise<StoredSession[]> {
  const db = await getDatabase();

  return new Promise((resolve, reject) => {
    const transaction = db.transaction(SESSIONS_STORE, "readonly");
    const store = transaction.objectStore(SESSIONS_STORE);
    const request = store.getAll();

    request.onerror = () => {
      reject(new Error(`Failed to list sessions: ${request.error?.message}`));
    };

    request.onsuccess = () => {
      resolve(request.result);
    };
  });
}

/**
 * Clears all sessions from IndexedDB.
 */
export async function clearAllSessions(): Promise<void> {
  const db = await getDatabase();

  return new Promise((resolve, reject) => {
    const transaction = db.transaction(SESSIONS_STORE, "readwrite");
    const store = transaction.objectStore(SESSIONS_STORE);
    const request = store.clear();

    request.onerror = () => {
      reject(new Error(`Failed to clear sessions: ${request.error?.message}`));
    };

    request.onsuccess = () => {
      resolve();
    };
  });
}

// ============================================================================
// Utility Functions
// ============================================================================

/**
 * Generates a UUID v4 for session IDs.
 * Uses crypto.randomUUID() if available, falls back to a simple implementation.
 */
export function generateUUID(): string {
  if (typeof crypto !== "undefined" && crypto.randomUUID) {
    return crypto.randomUUID();
  }

  // Fallback for older browsers
  return "xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx".replace(/[xy]/g, (c) => {
    const r = (Math.random() * 16) | 0;
    const v = c === "x" ? r : (r & 0x3) | 0x8;
    return v.toString(16);
  });
}

/**
 * Checks if IndexedDB is available in the current environment.
 */
export function isIndexedDBAvailable(): boolean {
  try {
    return typeof indexedDB !== "undefined" && indexedDB !== null;
  } catch {
    return false;
  }
}
