import { ILogger } from "@consensys/linea-shared-utils";
import { IDiscourseClient } from "../core/clients/IDiscourseClient.js";
import { RawDiscourseProposal, RawDiscourseProposalList } from "../core/entities/RawDiscourseProposal.js";

export class DiscourseClient implements IDiscourseClient {
  constructor(
    private readonly logger: ILogger,
    private readonly baseUrl: string,
    private readonly categoryId: number
  ) {}

  async fetchLatestProposals(): Promise<RawDiscourseProposalList | undefined> {
    const url = `${this.baseUrl}/c/proposals/${this.categoryId}/l/latest.json`;
    try {
      const response = await fetch(url);
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

  async fetchProposalDetails(proposalId: number): Promise<RawDiscourseProposal | undefined> {
    const url = `${this.baseUrl}/t/${proposalId}.json`;
    try {
      const response = await fetch(url);
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
