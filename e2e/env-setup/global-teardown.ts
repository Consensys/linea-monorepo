import { globalTestEnvironment } from "./global-test-env";

export default async (): Promise<void> => {
  await globalTestEnvironment.stopEnv();
};
