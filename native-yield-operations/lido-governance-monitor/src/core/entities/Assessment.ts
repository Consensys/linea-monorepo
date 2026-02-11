export type ImpactType = "economic" | "technical" | "operational" | "governance-process";
export type RiskLevel = "low" | "medium" | "high" | "critical";
export type RecommendedAction = "no-action" | "monitor" | "comment" | "escalate";
export type ProposalType = "discourse" | "snapshot" | "onchain_vote";
export type Urgency = "none" | "routine" | "urgent" | "critical";
export type AffectedComponent =
  | "StakingVault"
  | "VaultHub"
  | "LazyOracle"
  | "OperatorGrid"
  | "PredepositGuarantee"
  | "Dashboard"
  | "Other";
export const NativeYieldInvariant = {
  VALID_YIELD_REPORTING: "Valid yield reporting",
  USER_PRINCIPAL_PROTECTION: "User principal protection",
  PAUSE_DEPOSITS: "Pause deposits only on deficit, or liability or ossification",
  OTHER: "Other",
} as const;
export type NativeYieldInvariant = (typeof NativeYieldInvariant)[keyof typeof NativeYieldInvariant];

export interface Assessment {
  riskScore: number;
  riskLevel: RiskLevel;
  confidence: number;
  proposalType: ProposalType;
  impactTypes: ImpactType[];
  affectedComponents: AffectedComponent[];
  whatChanged: string;
  nativeYieldInvariantsAtRisk: NativeYieldInvariant[];
  whyItMattersForLineaNativeYield: string;
  recommendedAction: RecommendedAction;
  urgency: Urgency;
  supportingQuotes: string[];
  keyUnknowns: string[];
}
