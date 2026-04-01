import { Octokit } from "@octokit/rest";

const DEFAULT_DAYS = 30;
const OWNER = "Consensys";
const REPO = "linea-monorepo";
const WORKFLOW_FILE = "main.yml";
const E2E_JOB_NAME = "run-e2e-tests / run-e2e-tests";

function parseArgs(): { days: number } {
  const args = process.argv.slice(2);
  const daysIndex = args.indexOf("--days");
  const days = daysIndex !== -1 && args[daysIndex + 1] ? parseInt(args[daysIndex + 1], 10) : DEFAULT_DAYS;
  if (isNaN(days) || days < 1 || days > 365) {
    console.error("Invalid --days value. Must be between 1 and 365.");
    process.exit(1);
  }
  return { days };
}

async function main(): Promise<void> {
  const { days } = parseArgs();
  const token = process.env.GITHUB_TOKEN;
  if (!token) {
    console.error("GITHUB_TOKEN environment variable is required.");
    process.exit(1);
  }

  const octokit = new Octokit({ auth: token });

  console.log(`Generating E2E runtime report for the last ${days} days...`);

  // Phase 1: Fetch
  // Phase 2: Parse
  // Phase 3: Render

  console.log("Done.");
}

main().catch((err) => {
  console.error("Fatal error:", err);
  process.exit(1);
});
