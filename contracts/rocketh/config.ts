import type { UserConfig } from "rocketh/types";

export const config = {
  accounts: {
    deployer: {
      default: 0,
    },
  },
  data: {},
  scripts: "deploy",
  deployments: "deployments",
} as const satisfies UserConfig;
