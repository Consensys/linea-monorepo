/* eslint-disable @typescript-eslint/no-explicit-any */
import fetch from "node-fetch";
import * as fs from "fs";
import * as path from "path";
import { Open } from "unzipper";
import { exec } from "child_process";
import { getBuildConfig } from "./config";

async function downloadAndParseJson(url: string, headers: Record<string, string> = {}): Promise<any> {
  const response = await fetch(url, {
    method: "GET",
    headers: {
      Accept: "application/json",
      ...headers,
    },
  });

  if (!response.ok) {
    throw new Error(`Failed to load JSON from ${url}. HTTP error code: ${response.status}`);
  }

  return await response.json();
}

async function getReleaseAssetUrl(authToken: string, nativeLibReleaseTag: string): Promise<string> {
  const urlStr = "https://api.github.com/repos/ConsenSys/linea-monorepo/releases";

  const json = await downloadAndParseJson(urlStr, { Authorization: `token ${authToken}` });
  const release = json.find((release: any) => release.tag_name === nativeLibReleaseTag);

  if (!release) {
    const releases = json.map((release: any) => release.tag_name);
    throw new Error(`Release ${nativeLibReleaseTag} not found! releases: ${releases}`);
  }

  if (release.assets.length === 0) {
    throw new Error(`Release ${nativeLibReleaseTag} has no assets!`);
  }

  const asset = release.assets.find((asset: any) => asset.name.includes(nativeLibReleaseTag));
  return `https://${authToken}:@api.github.com/repos/Consensys/linea-monorepo/releases/assets/${asset.id}`;
}

async function downloadFileUsingCurl(authToken: string, url: string, outputFilePath: string): Promise<string> {
  const outputDirectory = path.dirname(outputFilePath);

  // Ensure the output directory exists
  fs.mkdirSync(outputDirectory, { recursive: true });
  const command = `curl -L -H 'Accept:application/octet-stream' -u ${authToken}: -o ${outputFilePath} ${url}`;

  return new Promise((resolve, reject) => {
    exec(command, (error: any, _: any, stderr: any) => {
      if (error) {
        reject(new Error(`Failed to download file using curl: ${stderr}`));
      } else {
        resolve(outputFilePath);
      }
    });
  });
}

const architectureResourceDirMapping: Record<string, string> = {
  darwin_arm64: "darwin-arm64",
  darwin_x86_64: "darwin-x64",
  linux_arm64: "linux-arm64",
  linux_amd64: "linux-x64",
  linux_x86_64: "linux-x64",
};

async function downloadReleaseAsset(authToken: string, nativeLibReleaseTag: string): Promise<string> {
  const assetReleaseUrl = await getReleaseAssetUrl(authToken, nativeLibReleaseTag);
  const fileName = `${nativeLibReleaseTag}.zip`;
  const destPath = path.resolve("build", fileName);
  console.log(`Downloading ${fileName} from ${assetReleaseUrl} to ${destPath}`);
  return await downloadFileUsingCurl(authToken, assetReleaseUrl, destPath);
}

function getBinaryResourceFolder(libFile: string): string {
  const destResource = Object.entries(architectureResourceDirMapping).find(([key]) => libFile.includes(key));
  if (!destResource) {
    throw new Error(`No architecture found for ${libFile}`);
  }
  return destResource[1];
}

function getBinaryResourceFileName(libFile: string, libName: string): string {
  const versionPattern = /v\d+\.\d+\.\d+/;
  const match = libFile.match(versionPattern);
  const version = match ? match[0] : null;
  const extension = path.extname(libFile);
  return `${libName}_${version}${extension}`;
}

async function downloadReleaseAndExtractToResources(
  authToken: string,
  nativeLibReleaseTag: string,
  libName: string,
): Promise<void> {
  const outputFile = await downloadReleaseAsset(authToken, nativeLibReleaseTag);

  if (!fs.existsSync(outputFile)) {
    throw new Error(`Output file ${outputFile} does not exist`);
  }

  const extractPath = path.resolve("build", nativeLibReleaseTag);

  const zipFile = await Open.file(outputFile);
  await zipFile.extract({ path: extractPath, concurrency: 5 });

  console.log("Extraction complete");
  const files = fs.readdirSync(extractPath);

  if (files.length === 0) {
    throw new Error("No files found in the extracted zip file.");
  }

  for (const file of files) {
    if (file.includes(libName) && (file.endsWith(".so") || file.endsWith(".dylib"))) {
      const destResourceDir = getBinaryResourceFolder(file);
      const destResourceFileName = getBinaryResourceFileName(file, libName);
      const destPath = path.resolve("src", "compressor", "lib", destResourceDir);

      fs.mkdirSync(destPath, { recursive: true });
      fs.copyFileSync(path.join(extractPath, file), path.join(destPath, destResourceFileName));
      console.log(`Copying ${file} to ${path.join(destPath, destResourceFileName)}`);
    }
  }
}

async function fetchLib(authToken: string, nativeLibReleaseTag: string, libName: string): Promise<void> {
  await downloadReleaseAndExtractToResources(authToken, nativeLibReleaseTag, libName);
}

async function main() {
  const { authToken, nativeLibReleaseTag } = getBuildConfig();
  await fetchLib(authToken, nativeLibReleaseTag, "blob_compressor");
}

main()
  .then(() => {
    process.exit(0);
  })
  .catch((error) => {
    console.error(error);
    process.exit(1);
  });
