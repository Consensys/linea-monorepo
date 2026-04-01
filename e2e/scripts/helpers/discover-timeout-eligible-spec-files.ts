import { readdirSync } from "node:fs";
import { join } from "node:path";

export const TEST_LOCAL_IGNORED_SPEC_FILES = new Set(["src/liveness.spec.ts", "src/linea-besu-fleet.spec.ts"]);

export function discoverTimeoutEligibleSpecFiles(projectDir: string = process.cwd()): string[] {
  const srcDir = join(projectDir, "src");

  return readdirSync(srcDir, { withFileTypes: true })
    .filter((entry) => entry.isFile() && entry.name.endsWith(".spec.ts"))
    .map((entry) => `src/${entry.name}`)
    .filter((specFile) => !TEST_LOCAL_IGNORED_SPEC_FILES.has(specFile))
    .sort();
}
