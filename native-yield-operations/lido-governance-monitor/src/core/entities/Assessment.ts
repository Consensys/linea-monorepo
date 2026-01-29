export type ImpactType = "economic" | "technical" | "operational" | "governance-process";
export type RiskLevel = "low" | "medium" | "high";
export type RecommendedAction = "monitor" | "comment" | "escalate" | "no-action";

export interface Assessment {
  riskScore: number;
  impactType: ImpactType;
  riskLevel: RiskLevel;
  whatChanged: string;
  whyItMattersForLineaNativeYield: string;
  recommendedAction: RecommendedAction;
  supportingQuotes: string[];
}
