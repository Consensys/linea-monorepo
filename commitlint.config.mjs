// commitlint.config.mjs
export default {
  extends: ["@commitlint/config-conventional"],
  rules: {
    "scope-enum": [2, "always", [
      "coordinator", "prover", "contracts", "sdk", "sdk-core", "sdk-viem",
      "bridge-ui", "postman", "tracer", "sequencer", "state-recovery",
      "tx-exclusion-api", "jvm-libs", "linea-besu", "blob-libs",
      "e2e", "ci", "docker", "deps", "misc"
    ]],
    "scope-empty": [2, "never"]
  }
};
