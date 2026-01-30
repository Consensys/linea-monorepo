import { ILogger } from "@consensys/linea-shared-utils";
import { IProposalProcessor } from "../core/services/IProposalProcessor.js";
import { IAIClient } from "../core/clients/IAIClient.js";
import { IProposalRepository } from "../core/repositories/IProposalRepository.js";
import { ProposalState } from "../core/entities/ProposalState.js";
import { ProposalSource } from "../core/entities/ProposalSource.js";
import { Proposal } from "../core/entities/Proposal.js";
import { ProposalType } from "../core/entities/Assessment.js";

export class ProposalProcessor implements IProposalProcessor {
  private intervalId: NodeJS.Timeout | null = null;

  constructor(
    private readonly logger: ILogger,
    private readonly aiClient: IAIClient,
    private readonly proposalRepository: IProposalRepository,
    private readonly riskThreshold: number,
    private readonly promptVersion: string,
    private readonly domainContext: string,
    private readonly processingIntervalMs: number
  ) {}

  start(): void {
    this.logger.info("ProposalProcessor started", { intervalMs: this.processingIntervalMs });

    // Initial process
    void this.processOnce();

    // Schedule subsequent processing
    this.intervalId = setInterval(() => {
      void this.processOnce();
    }, this.processingIntervalMs);
  }

  stop(): void {
    if (this.intervalId) {
      clearInterval(this.intervalId);
      this.intervalId = null;
    }
    this.logger.info("ProposalProcessor stopped");
  }

  async processOnce(): Promise<void> {
    const newProposals = await this.proposalRepository.findByState(ProposalState.NEW);
    const failedProposals = await this.proposalRepository.findByState(ProposalState.ANALYSIS_FAILED);
    const proposals = [...newProposals, ...failedProposals];

    if (proposals.length === 0) {
      this.logger.debug("No proposals to process");
      return;
    }

    this.logger.debug("Processing proposals", { count: proposals.length });

    for (const proposal of proposals) {
      await this.processProposal(proposal);
    }
  }

  private async processProposal(proposal: Proposal): Promise<void> {
    try {
      // Increment attempt count first
      const updated = await this.proposalRepository.incrementAnalysisAttempt(proposal.id);

      // Perform AI analysis
      const assessment = await this.aiClient.analyzeProposal({
        proposalTitle: proposal.title,
        proposalText: proposal.text,
        proposalUrl: proposal.url,
        proposalType: this.mapSourceToProposalType(proposal.source),
        domainContext: this.domainContext,
      });

      if (!assessment) {
        // Analysis failed - will retry on next cycle
        this.logger.warn("AI analysis failed, will retry", {
          proposalId: proposal.id,
          attempt: updated.analysisAttemptCount,
        });
        return;
      }

      // Save analysis result
      await this.proposalRepository.saveAnalysis(
        proposal.id,
        assessment,
        assessment.riskScore,
        this.aiClient.getModelName(),
        this.riskThreshold,
        this.promptVersion
      );

      // Transition based on risk score
      if (assessment.riskScore >= this.riskThreshold) {
        await this.proposalRepository.updateState(proposal.id, ProposalState.PENDING_NOTIFY);
        this.logger.info("Proposal requires notification", {
          proposalId: proposal.id,
          riskScore: assessment.riskScore,
          threshold: this.riskThreshold,
        });
      } else {
        await this.proposalRepository.updateState(proposal.id, ProposalState.NOT_NOTIFIED);
        this.logger.info("Proposal below notification threshold", {
          proposalId: proposal.id,
          riskScore: assessment.riskScore,
          threshold: this.riskThreshold,
        });
      }
    } catch (error) {
      this.logger.error("Error processing proposal", { proposalId: proposal.id, error });
    }
  }

  private mapSourceToProposalType(source: ProposalSource): ProposalType {
    switch (source) {
      case ProposalSource.DISCOURSE:
        return "discourse";
      case ProposalSource.SNAPSHOT:
        return "snapshot";
      case ProposalSource.LDO_VOTING_CONTRACT:
      case ProposalSource.STETH_VOTING_CONTRACT:
        return "onchain_vote";
    }
  }
}
