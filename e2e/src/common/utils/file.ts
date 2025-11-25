import { readdirSync, readFileSync } from "fs";
import path from "path";

export function getFiles(directory: string, fileRegex: RegExp[]): string[] {
  const files = readdirSync(directory, { withFileTypes: true });
  const filteredFiles = files.filter((file) => fileRegex.map((regex) => regex.test(file.name)).includes(true));
  return filteredFiles.map((file) => readFileSync(path.join(directory, file.name), "utf-8"));
}
