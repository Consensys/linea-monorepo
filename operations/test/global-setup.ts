import { execFileSync } from "node:child_process";
import { dirname, resolve } from "node:path";

export default async function globalSetup() {
  const operationsRoot = resolve(dirname(__dirname));

  execFileSync("pnpm", ["run", "build"], {
    cwd: operationsRoot,
    stdio: "inherit",
  });

  execFileSync("pnpm", ["exec", "oclif", "manifest"], {
    cwd: operationsRoot,
    stdio: "inherit",
  });
}
