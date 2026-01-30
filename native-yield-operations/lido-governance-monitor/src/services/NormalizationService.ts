import { ILogger } from "@consensys/linea-shared-utils";

import { CreateProposalInput } from "../core/entities/Proposal.js";
import { ProposalSource } from "../core/entities/ProposalSource.js";
import { RawDiscourseProposal } from "../core/entities/RawDiscourseProposal.js";
import { INormalizationService } from "../core/services/INormalizationService.js";

export class NormalizationService implements INormalizationService {
  constructor(
    private readonly logger: ILogger,
    private readonly discourseBaseUrl: string,
  ) {}

  normalizeDiscourseProposal(proposal: RawDiscourseProposal): CreateProposalInput {
    const posts = proposal.post_stream.posts;
    const firstPost = posts[0];
    const text = posts.map((post) => this.stripHtml(post.cooked)).join("\n\n");

    const input: CreateProposalInput = {
      source: ProposalSource.DISCOURSE,
      sourceId: proposal.id.toString(),
      url: `${this.discourseBaseUrl}/t/${proposal.slug}/${proposal.id}`,
      title: proposal.title,
      author: firstPost?.username ?? null,
      sourceCreatedAt: new Date(proposal.created_at),
      text: text.trim(),
    };

    this.logger.debug("Normalized Discourse proposal", {
      sourceId: input.sourceId,
      title: input.title,
      textLength: input.text.length,
    });

    return input;
  }

  stripHtml(html: string): string {
    if (!html) return "";

    // Replace block elements with newlines
    let text = html.replace(/<\/?(p|div|br|li|h[1-6])[^>]*>/gi, "\n");

    // Remove all remaining HTML tags
    text = text.replace(/<[^>]+>/g, "");

    // Decode HTML entities
    text = text
      .replace(/&amp;/g, "&")
      .replace(/&lt;/g, "<")
      .replace(/&gt;/g, ">")
      .replace(/&quot;/g, '"')
      .replace(/&#39;/g, "'")
      .replace(/&nbsp;/g, " ");

    // Normalize whitespace
    text = text.replace(/\n{3,}/g, "\n\n").trim();

    return text;
  }
}
