import { config } from "dotenv";

config();

type BuildConfig = {
  nativeLibReleaseTag: string;
};

export function getBuildConfig(): BuildConfig {
  const nativeLibReleaseTag = process.env.NATIVE_LIBS_RELEASE_TAG;

  if (!nativeLibReleaseTag) {
    throw new Error("NATIVE_LIBS_RELEASE_TAG environment variable is not set");
  }

  return {
    nativeLibReleaseTag,
  };
}
