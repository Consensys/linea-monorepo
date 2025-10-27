import express, { Express, Request, Response } from "express";
import { IMetricsService } from "../core/services/IMetricsService";
import { ILogger } from "../logging/ILogger";
import { IApplication } from "../core/applications/IApplication";

// Use protected instead of private, to overrides
export class ExpressApiApplication implements IApplication {
  protected readonly app: Express;
  protected server?: ReturnType<Express["listen"]>;

  constructor(
    protected readonly port: number,
    protected readonly metricsService: IMetricsService,
    protected readonly logger: ILogger,
  ) {
    this.app = express();

    this.setupMiddleware();
    this.setupMetricsRoute();
  }

  protected setupMiddleware(): void {
    this.app.use(express.json());
  }

  protected setupMetricsRoute() {
    this.app.get("/metrics", async (_req: Request, res: Response) => {
      try {
        const registry = this.metricsService.getRegistry();
        res.set("Content-Type", registry.contentType);
        res.end(await registry.metrics());
      } catch (error) {
        this.logger.warn("Failed to collect metrics", { error });
        res.status(500).json({ error: "Failed to collect metrics" });
      }
    });
  }

  public async start(): Promise<void> {
    await this.onBeforeStart?.();
    this.server = this.app.listen(this.port);

    await new Promise<void>((resolve) => {
      this.server?.on("listening", () => {
        this.logger.info(`Listening on port ${this.port}`);
        resolve();
      });
    });
    await this.onAfterStart?.();
  }

  public async stop(): Promise<void> {
    await this.onBeforeStop?.();
    if (!this.server) return;

    await new Promise<void>((resolve, reject) => {
      this.server?.close((err) => {
        if (err) return reject(err);
        this.logger.info(`Closing API server on port ${this.port}`);
        this.server = undefined;
        resolve();
      });
    });
    await this.onAfterStop?.();
  }

  // Optional hooks to assign from outside
  onBeforeStart?: () => Promise<void> | void;
  onAfterStart?: () => Promise<void> | void;
  onBeforeStop?: () => Promise<void> | void;
  onAfterStop?: () => Promise<void> | void;
}
