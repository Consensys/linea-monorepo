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
export type NativeYieldInvariant =
  | "A_valid_yield_reporting"
  | "B_user_principal_protection"
  | "C_pause_deposits_on_deficit_or_liability_or_ossification"
  | "Other";

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
