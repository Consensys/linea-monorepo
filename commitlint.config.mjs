// commitlint.config.mjs
export default {
  extends: ["@commitlint/config-conventional"],
  rules: {
    "scope-enum": [2, "always", [
      "coordinator", "prover", "postman", "tx-exclusion-api", "linea-besu",
      "prover-ray", "contracts", "sdk-core", "sdk-ethers", "sdk-viem",
      "tracer", "sequencer", "state-recovery",
      "jvm-libs", "blob-libs",
      "e2e", "ci", "docker", "deps", "misc"
    ]],
    "scope-empty": [2, "never"]
  }
};
