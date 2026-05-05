import { rmSync } from "node:fs";
import { dirname, resolve } from "node:path";

export default async function globalTeardown() {
  rmSync(resolve(dirname(__dirname), "oclif.manifest.json"), { force: true });
}
