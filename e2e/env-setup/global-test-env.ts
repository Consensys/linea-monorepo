import { TestEnvironment } from "./test-env";

export class GlobalTestEnvironment {
  public readonly testingEnv = new TestEnvironment();

  public async startEnv(): Promise<void> {
    await this.testingEnv.startEnv();
  }

  public async stopEnv(): Promise<void> {
    await this.testingEnv.stopEnv();
  }

  public async restartCoordinator(): Promise<void> {
    await this.testingEnv.restartCoordinator(global.useLocalSetup);
  }
}

export const globalTestEnvironment = new GlobalTestEnvironment();
