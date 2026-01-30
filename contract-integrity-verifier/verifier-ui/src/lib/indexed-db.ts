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

/** Session expiry time in milliseconds (24 hours) */
const SESSION_EXPIRY_MS = 24 * 60 * 60 * 1000;

/** Maximum number of sessions to keep */
const MAX_SESSIONS = 10;

// ============================================================================
// Error Types
// ============================================================================

export class IndexedDBError extends Error {
  constructor(
    message: string,
    public readonly code: "QUOTA_EXCEEDED" | "NOT_AVAILABLE" | "OPERATION_FAILED",
  ) {
    super(message);
    this.name = "IndexedDBError";
  }
}

/**
 * Checks if an error is a quota exceeded error.
 */
function isQuotaExceededError(error: unknown): boolean {
  if (error instanceof DOMException) {
    // Different browsers use different error names/codes
    const firefoxQuotaError = "NS_ERROR_DOM_QUOTA_REACHED";
    return error.name === "QuotaExceededError" || error.code === 22 || error.name === firefoxQuotaError;
  }
  return false;
}

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
 * Handles quota exceeded errors by cleaning up old sessions.
 */
export async function saveSession(session: StoredSession): Promise<void> {
  const db = await getDatabase();

  return new Promise((resolve, reject) => {
    const transaction = db.transaction(SESSIONS_STORE, "readwrite");
    const store = transaction.objectStore(SESSIONS_STORE);
    const request = store.put(session);

    request.onerror = async () => {
      if (isQuotaExceededError(request.error)) {
        // Try to free up space by cleaning old sessions
        try {
          await cleanupOldSessions();
          // Retry the save
          const retryDb = await getDatabase();
          const retryTx = retryDb.transaction(SESSIONS_STORE, "readwrite");
          const retryStore = retryTx.objectStore(SESSIONS_STORE);
          const retryRequest = retryStore.put(session);

          retryRequest.onerror = () => {
            reject(
              new IndexedDBError(
                "Storage quota exceeded. Please clear some browser data or use smaller files.",
                "QUOTA_EXCEEDED",
              ),
            );
          };
          retryRequest.onsuccess = () => resolve();
        } catch {
          reject(new IndexedDBError("Storage quota exceeded and cleanup failed.", "QUOTA_EXCEEDED"));
        }
      } else {
        reject(new IndexedDBError(`Failed to save session: ${request.error?.message}`, "OPERATION_FAILED"));
      }
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
 * Uses crypto.randomUUID() if available, falls back to crypto.getRandomValues().
 * Never uses Math.random() as it is not cryptographically secure.
 */
export function generateUUID(): string {
  // Prefer native randomUUID if available
  if (typeof crypto !== "undefined" && crypto.randomUUID) {
    return crypto.randomUUID();
  }

  // Fallback using crypto.getRandomValues() for older browsers
  // This is cryptographically secure, unlike Math.random()
  if (typeof crypto !== "undefined" && crypto.getRandomValues) {
    const bytes = new Uint8Array(16);
    crypto.getRandomValues(bytes);

    // Set version (4) and variant (RFC4122) bits
    bytes[6] = (bytes[6] & 0x0f) | 0x40; // Version 4
    bytes[8] = (bytes[8] & 0x3f) | 0x80; // Variant RFC4122

    // Convert to hex string with dashes
    const hex = Array.from(bytes)
      .map((b) => b.toString(16).padStart(2, "0"))
      .join("");

    return `${hex.slice(0, 8)}-${hex.slice(8, 12)}-${hex.slice(12, 16)}-${hex.slice(16, 20)}-${hex.slice(20)}`;
  }

  // Last resort: throw error if no secure random available
  // This should never happen in modern browsers
  throw new Error("No cryptographically secure random number generator available");
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

// ============================================================================
// Session Cleanup
// ============================================================================

/**
 * Checks if a session has expired.
 */
function isSessionExpired(session: StoredSession): boolean {
  const createdAt = new Date(session.createdAt).getTime();
  return Date.now() - createdAt > SESSION_EXPIRY_MS;
}

/**
 * Cleans up expired sessions and enforces maximum session count.
 * Called automatically when quota is exceeded.
 */
export async function cleanupOldSessions(): Promise<number> {
  const sessions = await listSessions();

  // Sort by creation date (oldest first)
  sessions.sort((a, b) => new Date(a.createdAt).getTime() - new Date(b.createdAt).getTime());

  const toDelete: string[] = [];

  // Mark expired sessions for deletion
  for (const session of sessions) {
    if (isSessionExpired(session)) {
      toDelete.push(session.id);
    }
  }

  // If we still have too many sessions, delete oldest ones
  const remainingSessions = sessions.filter((s) => !toDelete.includes(s.id));
  if (remainingSessions.length > MAX_SESSIONS) {
    const excess = remainingSessions.length - MAX_SESSIONS;
    for (let i = 0; i < excess; i++) {
      toDelete.push(remainingSessions[i].id);
    }
  }

  // Delete marked sessions
  for (const sessionId of toDelete) {
    try {
      await deleteSession(sessionId);
    } catch {
      // Ignore individual deletion errors
    }
  }

  return toDelete.length;
}

/**
 * Gets estimated storage usage for IndexedDB.
 * Returns null if not supported by browser.
 */
export async function getStorageEstimate(): Promise<{ usage: number; quota: number } | null> {
  if (typeof navigator !== "undefined" && navigator.storage?.estimate) {
    try {
      const estimate = await navigator.storage.estimate();
      return {
        usage: estimate.usage ?? 0,
        quota: estimate.quota ?? 0,
      };
    } catch {
      return null;
    }
  }
  return null;
}
