import { execSync } from "child_process";

export function getGitBranch(): string {
  return execSync("git rev-parse --abbrev-ref HEAD").toString().trim();
}

export function getGitCommitHash(): string {
  return execSync(`git rev-parse HEAD`).toString().trim();
}

export function getGitTagsAtCommitHash(): string[] {
  // Get the tags pointing specifically to this commit hash! This is vital in making sure that commit has a valid tag.
  const tags = execSync(`git tag --points-at ${getGitCommitHash()}`).toString().trim();
  return tags.length > 0 ? tags.split("\n") : [];
}

export function validateDeployBranchAndTags(networkName: string) {
  // Tag pattern - e.g. contract-audit-diligience-2022-08-28
  const tagPattern = /^contract-audit-\S+-\d{4}-\d{2}-\d{2}$/;
  const branchPattern = /^audit\/|main\//;

  const networksRequiringAuditedCode: string[] = ["mainnet", "linea_mainnet", "goerli", "linea_goerli"];

  console.log("Validating if the network to deploy to requires an audited version.");

  if (networksRequiringAuditedCode.includes(networkName)) {
    const branch = getGitBranch();
    const tags = getGitTagsAtCommitHash();

    if (!branchPattern.test(branch)) {
      throw `You must deploy from an audit branch, you are on ${branch}`;
    }

    if (!tags.some((value) => tagPattern.test(value))) {
      throw "Tags for this branch are missing. Format 'contract-audit-FIRM-DATE'";
    }
  }
}
