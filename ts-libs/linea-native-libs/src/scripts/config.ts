import { config } from "dotenv";

config();

type BuildConfig = {
  authToken: string;
  nativeLibReleaseTag: string;
};

export function getBuildConfig(): BuildConfig {
  const authToken = process.env.GITHUB_API_ACCESS_TOKEN;

  if (!authToken) {
    throw new Error("GITHUB_API_ACCESS_TOKEN environment variable is not set");
  }

  const nativeLibReleaseTag = process.env.NATIVE_LIBS_RELEASE_TAG;

  if (!nativeLibReleaseTag) {
    throw new Error("NATIVE_LIBS_RELEASE_TAG environment variable is not set");
  }

  return {
    authToken,
    nativeLibReleaseTag,
  };
}
