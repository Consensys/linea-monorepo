import { fetchWithTimeout, ILogger, IRetryService } from "@consensys/linea-shared-utils";

import { IDiscourseClient } from "../core/clients/IDiscourseClient.js";
import { RawDiscourseProposal, RawDiscourseProposalList } from "../core/entities/RawDiscourseProposal.js";

export class DiscourseClient implements IDiscourseClient {
  private readonly baseUrl: string;

  constructor(
    private readonly logger: ILogger,
    private readonly retryService: IRetryService,
    private readonly proposalsUrl: string,
    private readonly httpTimeoutMs: number,
  ) {
    // Derive base URL from proposals URL for fetching individual proposals
    this.baseUrl = new URL(proposalsUrl).origin;
  }

  async fetchLatestProposals(): Promise<RawDiscourseProposalList | undefined> {
    const url = this.proposalsUrl;
    try {
      const response = await this.retryService.retry(() => fetchWithTimeout(url, {}, this.httpTimeoutMs));
      if (!response.ok) {
        this.logger.error("Failed to fetch latest proposals", {
          status: response.status,
          statusText: response.statusText,
        });
        return undefined;
      }
      const data = await response.json();
      this.logger.debug("Fetched latest proposals", { count: data.topic_list?.topics?.length });
      return data as RawDiscourseProposalList;
    } catch (error) {
      this.logger.error("Error fetching latest proposals", { error });
      return undefined;
    }
  }

  getBaseUrl(): string {
    return this.baseUrl;
  }

  async fetchProposalDetails(proposalId: number): Promise<RawDiscourseProposal | undefined> {
    const url = `${this.baseUrl}/t/${proposalId}.json`;
    try {
      const response = await this.retryService.retry(() => fetchWithTimeout(url, {}, this.httpTimeoutMs));
      if (!response.ok) {
        this.logger.error("Failed to fetch proposal details", {
          proposalId,
          status: response.status,
        });
        return undefined;
      }
      const data = await response.json();
      this.logger.debug("Fetched proposal details", { proposalId, title: data.title });
      return data as RawDiscourseProposal;
    } catch (error) {
      this.logger.error("Error fetching proposal details", { proposalId, error });
      return undefined;
    }
  }
}
