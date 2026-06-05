import kzg from "c-kzg";
import path from "path";
import { fileURLToPath } from "url";

let initialized = false;
const currentDir = path.dirname(fileURLToPath(import.meta.url));

/**
 * Idempotent KZG trusted setup initialization.
 * Safe to call from multiple test files — only runs once per process.
 */
export function ensureKzgSetup(): void {
  if (initialized) return;
  const trustedSetupPath = path.resolve(currentDir, "../../_testData/trusted_setup.txt");
  kzg.loadTrustedSetup(0, trustedSetupPath);
  initialized = true;
}
