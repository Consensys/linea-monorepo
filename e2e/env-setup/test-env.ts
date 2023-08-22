import process from "node:process";
import { StartedDockerComposeEnvironment, DockerComposeEnvironment, Wait } from "testcontainers";

export class TestEnvironment {
  public dockerEnvironment: StartedDockerComposeEnvironment;

  public async startEnv(): Promise<void> {
    console.log("\nStarting tests environment...");
    this.bindProcess();
    const composeFilePath = "../docker";
    const composeFile = "compose.yml";

    this.dockerEnvironment = await new DockerComposeEnvironment(composeFilePath, composeFile)
      .withProfiles("l1", "l2")
      .withWaitStrategy(
        "l1-validator",
        Wait.forAll([Wait.forLogMessage("HTTP server started"), Wait.forLogMessage("WebSocket enabled")]),
      )
      .withWaitStrategy(
        "sequencer",
        Wait.forAll([Wait.forLogMessage("HTTP server started"), Wait.forLogMessage("WebSocket enabled")]),
      )
      .withWaitStrategy(
        "traces-node",
        Wait.forAll([Wait.forLogMessage("HTTP server started"), Wait.forLogMessage("WebSocket enabled")]),
      )
      .withWaitStrategy(
        "postgres",
        Wait.forHealthCheck()
      )
      .withWaitStrategy(
        "coordinator",
        Wait.forAll([Wait.forLogMessage("CoordinatorApp - Waiting for block number")])
      )
      .up();
    console.log("Tests environment started.");
  }

  public async stopEnv(): Promise<void> {
    console.log("Stopping tests environment...");
    await Promise.all([this.dockerEnvironment.down()]);
    console.log("Tests environment stopped.");
  }

  public async restartCoordinator(localSetup: Boolean): Promise<void> {
    console.log("Restarting coordinator...");
    if (localSetup) {
      await Promise.all([this.dockerEnvironment.getContainer("coordinator").restart()]);
    } else {
      //TODO restart k8 coordinator
    }
    console.log("Coordinator restarted.");
  }

  private async stopProcess() {
    await this.stopEnv();
    process.exit(1);
  }

  public bindProcess(): void {
    process.on("SIGINT", async () => {
      this.stopProcess()
    });
    process.on("SIGTERM", async () => {
      this.stopProcess()
    });
    process.on("uncaughtException", async () => {
      this.stopProcess()
    });
    process.on("unhandledRejection", async () => {
      this.stopProcess()
    });
  }
}
