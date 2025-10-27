import { SingletonMetricsService } from "@consensys/linea-shared-utils";
import { LineaNativeYieldAutomationServiceMetrics } from "../../core/metrics/LineaNativeYieldAutomationServiceMetrics.js";

export class NativeYieldAutomationMetricsService extends SingletonMetricsService<LineaNativeYieldAutomationServiceMetrics> {
  constructor(defaultLabels: Record<string, string> = { app: "native-yield-automation-service" }) {
    super(defaultLabels);
  }
}
