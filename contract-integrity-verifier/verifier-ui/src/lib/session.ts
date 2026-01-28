import { mkdir, readFile, writeFile, rm, readdir, stat } from "fs/promises";
import { join, resolve } from "path";
import { randomUUID } from "crypto";
import { SESSIONS_DIR, SESSION_EXPIRY_MS } from "./constants";
import type { Session } from "@/types";

// ============================================================================
// Path Safety
// ============================================================================

/**
 * Ensures the resolved path is within the allowed base directory.
 * Prevents path traversal attacks.
 */
function ensurePathWithinDir(basePath: string, targetPath: string): string {
  const resolvedBase = resolve(basePath);
  const resolvedTarget = resolve(basePath, targetPath);

  if (!resolvedTarget.startsWith(resolvedBase)) {
    throw new Error("Path traversal detected");
  }

  return resolvedTarget;
}

// ============================================================================
// Session File Operations
// ============================================================================

function getSessionDir(sessionId: string): string {
  return join(SESSIONS_DIR, sessionId);
}

function getSessionMetaPath(sessionId: string): string {
  return join(getSessionDir(sessionId), "session.json");
}

// ============================================================================
// Session Management
// ============================================================================

export async function createSession(): Promise<Session> {
  // Cleanup expired sessions periodically (non-blocking)
  cleanupExpiredSessions().catch(() => {
    // Ignore cleanup errors
  });

  const sessionId = randomUUID();
  const now = new Date();
  const expiresAt = new Date(now.getTime() + SESSION_EXPIRY_MS);

  const session: Session = {
    id: sessionId,
    createdAt: now.toISOString(),
    expiresAt: expiresAt.toISOString(),
    config: null,
    fileMap: {},
    envVarValues: {},
  };

  const sessionDir = getSessionDir(sessionId);
  await mkdir(sessionDir, { recursive: true });
  await mkdir(join(sessionDir, "schemas"), { recursive: true });
  await mkdir(join(sessionDir, "artifacts"), { recursive: true });

  await writeFile(getSessionMetaPath(sessionId), JSON.stringify(session, null, 2));

  return session;
}

export async function getSession(sessionId: string): Promise<Session | null> {
  try {
    const metaPath = getSessionMetaPath(sessionId);
    const content = await readFile(metaPath, "utf-8");
    const session = JSON.parse(content) as Session;

    // Check expiry
    if (new Date(session.expiresAt) < new Date()) {
      await deleteSession(sessionId);
      return null;
    }

    return session;
  } catch {
    return null;
  }
}

export async function updateSession(
  sessionId: string,
  updates: Partial<Pick<Session, "config" | "fileMap" | "envVarValues">>,
): Promise<Session | null> {
  const session = await getSession(sessionId);
  if (!session) return null;

  const updated: Session = {
    ...session,
    ...updates,
  };

  await writeFile(getSessionMetaPath(sessionId), JSON.stringify(updated, null, 2));
  return updated;
}

export async function deleteSession(sessionId: string): Promise<void> {
  try {
    await rm(getSessionDir(sessionId), { recursive: true, force: true });
  } catch {
    // Ignore errors
  }
}

// ============================================================================
// File Storage
// ============================================================================

export async function storeConfigFile(sessionId: string, filename: string, content: string): Promise<string> {
  const sessionDir = getSessionDir(sessionId);
  // Ensure path is within session directory (prevents path traversal)
  const configPath = ensurePathWithinDir(sessionDir, filename);

  await writeFile(configPath, content);
  return configPath;
}

export async function storeFile(
  sessionId: string,
  type: "schema" | "artifact",
  originalPath: string,
  content: string | Buffer,
): Promise<string> {
  const sessionDir = getSessionDir(sessionId);
  const filename = originalPath.split("/").pop() || "file.json";
  const targetDir = join(sessionDir, type === "schema" ? "schemas" : "artifacts");
  // Ensure path is within target directory (prevents path traversal)
  const targetPath = ensurePathWithinDir(targetDir, filename);

  await writeFile(targetPath, content);
  return targetPath;
}

export async function getConfigContent(sessionId: string): Promise<string | null> {
  const session = await getSession(sessionId);
  if (!session?.config) return null;

  const configPath = join(getSessionDir(sessionId), session.config.filename);

  try {
    return await readFile(configPath, "utf-8");
  } catch {
    return null;
  }
}

// ============================================================================
// Session Cleanup
// ============================================================================

export async function cleanupExpiredSessions(): Promise<number> {
  let cleaned = 0;

  try {
    await mkdir(SESSIONS_DIR, { recursive: true });
    const entries = await readdir(SESSIONS_DIR);

    for (const entry of entries) {
      const sessionDir = join(SESSIONS_DIR, entry);
      const stats = await stat(sessionDir);

      if (!stats.isDirectory()) continue;

      const session = await getSession(entry);
      if (!session) {
        // Session expired or invalid
        await rm(sessionDir, { recursive: true, force: true });
        cleaned++;
      }
    }
  } catch {
    // Ignore errors during cleanup
  }

  return cleaned;
}
