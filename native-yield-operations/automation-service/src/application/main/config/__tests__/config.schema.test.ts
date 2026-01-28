import { configSchema } from "../config.schema.js";
import { getAddress } from "viem";

const createValidEnv = () => ({
  CHAIN_ID: "11155111",
  L1_RPC_URL: "https://rpc.linea.build",
  BEACON_CHAIN_RPC_URL: "https://beacon.linea.build",
  STAKING_GRAPHQL_URL: "https://staking.linea.build/graphql",
  IPFS_BASE_URL: "https://ipfs.linea.build",
  CONSENSYS_STAKING_OAUTH2_TOKEN_ENDPOINT: "https://auth.linea.build/token",
  CONSENSYS_STAKING_OAUTH2_CLIENT_ID: "client-id",
  CONSENSYS_STAKING_OAUTH2_CLIENT_SECRET: "client-secret",
  CONSENSYS_STAKING_OAUTH2_AUDIENCE: "audience",
  LINEA_ROLLUP_ADDRESS: "0x1111111111111111111111111111111111111111",
  LAZY_ORACLE_ADDRESS: "0x2222222222222222222222222222222222222222",
  VAULT_HUB_ADDRESS: "0x3333333333333333333333333333333333333333",
  YIELD_MANAGER_ADDRESS: "0x4444444444444444444444444444444444444444",
  LIDO_YIELD_PROVIDER_ADDRESS: "0x5555555555555555555555555555555555555555",
  STETH_ADDRESS: "0x6666666666666666666666666666666666666666",
  L2_YIELD_RECIPIENT: "0x7777777777777777777777777777777777777777",
  TRIGGER_EVENT_POLL_INTERVAL_MS: "1000",
  TRIGGER_MAX_INACTION_MS: "5000",
  CONTRACT_READ_RETRY_TIME_MS: "250",
  GAUGE_METRICS_POLL_INTERVAL_MS: "5000",
  REBALANCE_TOLERANCE_AMOUNT_WEI: "5000000000000000000",
  MAX_VALIDATOR_WITHDRAWAL_REQUESTS_PER_TRANSACTION: "16",
  MIN_WITHDRAWAL_THRESHOLD_ETH: "42",
  STAKING_REBALANCE_QUOTA_BPS: "1800",
  STAKING_REBALANCE_QUOTA_WINDOW_SIZE_IN_CYCLES: "24",
  WEB3SIGNER_URL: "https://web3signer.linea.build",
  WEB3SIGNER_PUBLIC_KEY: `0x${"a".repeat(128)}`,
  WEB3SIGNER_KEYSTORE_PATH: "/path/to/keystore",
  WEB3SIGNER_KEYSTORE_PASSPHRASE: "keystore-pass",
  WEB3SIGNER_TRUSTSTORE_PATH: "/path/to/truststore",
  WEB3SIGNER_TRUSTSTORE_PASSPHRASE: "truststore-pass",
  WEB3SIGNER_TLS_ENABLED: "true",
  API_PORT: "3000",
  SHOULD_SUBMIT_VAULT_REPORT: "true",
  SHOULD_REPORT_YIELD: "true",
  IS_UNPAUSE_STAKING_ENABLED: "true",
  MIN_NEGATIVE_YIELD_DIFF_TO_REPORT_YIELD_WEI: "1000000000000000000",
  CYCLES_PER_YIELD_REPORT: "12",
});

describe("configSchema", () => {
  it("parses and normalizes valid environment variables", () => {
    const env = createValidEnv();

    const parsed = configSchema.parse(env);

    expect(parsed.CHAIN_ID).toBe(11155111);
    expect(parsed.TRIGGER_EVENT_POLL_INTERVAL_MS).toBe(1000);
    expect(parsed.API_PORT).toBe(3000);
    expect(parsed.MIN_WITHDRAWAL_THRESHOLD_ETH).toBe(42n);
    expect(parsed.STAKING_REBALANCE_QUOTA_BPS).toBe(1800);
    expect(parsed.STAKING_REBALANCE_QUOTA_WINDOW_SIZE_IN_CYCLES).toBe(24);
    expect(parsed.MIN_NEGATIVE_YIELD_DIFF_TO_REPORT_YIELD_WEI).toBe(1000000000000000000n);
    expect(parsed.WEB3SIGNER_TLS_ENABLED).toBe(true);
    expect(parsed.SHOULD_SUBMIT_VAULT_REPORT).toBe(true);
    expect(parsed.SHOULD_REPORT_YIELD).toBe(true);
    expect(parsed.IS_UNPAUSE_STAKING_ENABLED).toBe(true);
    expect(parsed.LINEA_ROLLUP_ADDRESS).toBe(getAddress(env.LINEA_ROLLUP_ADDRESS));
    expect(parsed.LAZY_ORACLE_ADDRESS).toBe(getAddress(env.LAZY_ORACLE_ADDRESS));
    expect(parsed.VAULT_HUB_ADDRESS).toBe(getAddress(env.VAULT_HUB_ADDRESS));
    expect(parsed.STETH_ADDRESS).toBe(getAddress(env.STETH_ADDRESS));
    expect(parsed.L2_YIELD_RECIPIENT).toBe(getAddress(env.L2_YIELD_RECIPIENT));
  });

  it("rejects invalid Ethereum addresses", () => {
    const env = {
      ...createValidEnv(),
      LINEA_ROLLUP_ADDRESS: "0xinvalid",
    };

    const result = configSchema.safeParse(env);

    expect(result.success).toBe(false);
    if (!result.success) {
      expect(result.error.issues.some((issue) => issue.path.join(".") === "LINEA_ROLLUP_ADDRESS")).toBe(true);
    }
  });

  it("rejects invalid Web3Signer public key values", () => {
    const env = {
      ...createValidEnv(),
      WEB3SIGNER_PUBLIC_KEY: "0x1234",
    };

    const result = configSchema.safeParse(env);

    expect(result.success).toBe(false);
    if (!result.success) {
      expect(result.error.issues.some((issue) => issue.path.join(".") === "WEB3SIGNER_PUBLIC_KEY")).toBe(true);
    }
  });

  describe("MIN_NEGATIVE_YIELD_DIFF_TO_REPORT_YIELD_WEI", () => {
    it("parses string values to bigint", () => {
      const env = {
        ...createValidEnv(),
        MIN_NEGATIVE_YIELD_DIFF_TO_REPORT_YIELD_WEI: "1000000000000000000",
      };

      const parsed = configSchema.parse(env);

      expect(parsed.MIN_NEGATIVE_YIELD_DIFF_TO_REPORT_YIELD_WEI).toBe(1000000000000000000n);
    });

    it("parses number values to bigint", () => {
      const env = {
        ...createValidEnv(),
        MIN_NEGATIVE_YIELD_DIFF_TO_REPORT_YIELD_WEI: 2000000000000000000,
      };

      const parsed = configSchema.parse(env);

      expect(parsed.MIN_NEGATIVE_YIELD_DIFF_TO_REPORT_YIELD_WEI).toBe(2000000000000000000n);
    });

    it("parses bigint values", () => {
      const env = {
        ...createValidEnv(),
        MIN_NEGATIVE_YIELD_DIFF_TO_REPORT_YIELD_WEI: 3000000000000000000n,
      };

      const parsed = configSchema.parse(env);

      expect(parsed.MIN_NEGATIVE_YIELD_DIFF_TO_REPORT_YIELD_WEI).toBe(3000000000000000000n);
    });

    it("accepts zero value", () => {
      const env = {
        ...createValidEnv(),
        MIN_NEGATIVE_YIELD_DIFF_TO_REPORT_YIELD_WEI: "0",
      };

      const parsed = configSchema.parse(env);

      expect(parsed.MIN_NEGATIVE_YIELD_DIFF_TO_REPORT_YIELD_WEI).toBe(0n);
    });

    it("rejects negative values", () => {
      const env = {
        ...createValidEnv(),
        MIN_NEGATIVE_YIELD_DIFF_TO_REPORT_YIELD_WEI: "-1",
      };

      const result = configSchema.safeParse(env);

      expect(result.success).toBe(false);
      if (!result.success) {
        expect(
          result.error.issues.some((issue) => issue.path.join(".") === "MIN_NEGATIVE_YIELD_DIFF_TO_REPORT_YIELD_WEI"),
        ).toBe(true);
      }
    });
  });

  describe("STAKING_REBALANCE_QUOTA_BPS", () => {
    it("parses string values to number", () => {
      const env = {
        ...createValidEnv(),
        STAKING_REBALANCE_QUOTA_BPS: "1800",
      };

      const parsed = configSchema.parse(env);

      expect(parsed.STAKING_REBALANCE_QUOTA_BPS).toBe(1800);
    });

    it("parses number values", () => {
      const env = {
        ...createValidEnv(),
        STAKING_REBALANCE_QUOTA_BPS: 2000,
      };

      const parsed = configSchema.parse(env);

      expect(parsed.STAKING_REBALANCE_QUOTA_BPS).toBe(2000);
    });

    it("accepts zero value (disables quota mechanism)", () => {
      const env = {
        ...createValidEnv(),
        STAKING_REBALANCE_QUOTA_BPS: "0",
      };

      const parsed = configSchema.parse(env);

      expect(parsed.STAKING_REBALANCE_QUOTA_BPS).toBe(0);
    });

    it("rejects negative values", () => {
      const env = {
        ...createValidEnv(),
        STAKING_REBALANCE_QUOTA_BPS: "-1",
      };

      const result = configSchema.safeParse(env);

      expect(result.success).toBe(false);
      if (!result.success) {
        expect(result.error.issues.some((issue) => issue.path.join(".") === "STAKING_REBALANCE_QUOTA_BPS")).toBe(true);
      }
    });

    it("rejects non-integer values", () => {
      const env = {
        ...createValidEnv(),
        STAKING_REBALANCE_QUOTA_BPS: "1800.5",
      };

      const result = configSchema.safeParse(env);

      expect(result.success).toBe(false);
      if (!result.success) {
        expect(result.error.issues.some((issue) => issue.path.join(".") === "STAKING_REBALANCE_QUOTA_BPS")).toBe(true);
      }
    });
  });

  describe("STAKING_REBALANCE_QUOTA_WINDOW_SIZE_IN_CYCLES", () => {
    it("parses string values to number", () => {
      const env = {
        ...createValidEnv(),
        STAKING_REBALANCE_QUOTA_WINDOW_SIZE_IN_CYCLES: "24",
      };

      const parsed = configSchema.parse(env);

      expect(parsed.STAKING_REBALANCE_QUOTA_WINDOW_SIZE_IN_CYCLES).toBe(24);
    });

    it("parses number values", () => {
      const env = {
        ...createValidEnv(),
        STAKING_REBALANCE_QUOTA_WINDOW_SIZE_IN_CYCLES: 48,
      };

      const parsed = configSchema.parse(env);

      expect(parsed.STAKING_REBALANCE_QUOTA_WINDOW_SIZE_IN_CYCLES).toBe(48);
    });

    it("accepts zero value (disables quota mechanism)", () => {
      const env = {
        ...createValidEnv(),
        STAKING_REBALANCE_QUOTA_WINDOW_SIZE_IN_CYCLES: "0",
      };

      const parsed = configSchema.parse(env);

      expect(parsed.STAKING_REBALANCE_QUOTA_WINDOW_SIZE_IN_CYCLES).toBe(0);
    });

    it("rejects negative values", () => {
      const env = {
        ...createValidEnv(),
        STAKING_REBALANCE_QUOTA_WINDOW_SIZE_IN_CYCLES: "-1",
      };

      const result = configSchema.safeParse(env);

      expect(result.success).toBe(false);
      if (!result.success) {
        expect(
          result.error.issues.some((issue) => issue.path.join(".") === "STAKING_REBALANCE_QUOTA_WINDOW_SIZE_IN_CYCLES"),
        ).toBe(true);
      }
    });

    it("rejects non-integer values", () => {
      const env = {
        ...createValidEnv(),
        STAKING_REBALANCE_QUOTA_WINDOW_SIZE_IN_CYCLES: "24.5",
      };

      const result = configSchema.safeParse(env);

      expect(result.success).toBe(false);
      if (!result.success) {
        expect(
          result.error.issues.some((issue) => issue.path.join(".") === "STAKING_REBALANCE_QUOTA_WINDOW_SIZE_IN_CYCLES"),
        ).toBe(true);
      }
    });
  });

  describe("SHOULD_SUBMIT_VAULT_REPORT", () => {
    it("parses string 'false' as false", () => {
      const env = {
        ...createValidEnv(),
        SHOULD_SUBMIT_VAULT_REPORT: "false",
      };

      const parsed = configSchema.parse(env);

      expect(parsed.SHOULD_SUBMIT_VAULT_REPORT).toBe(false);
    });

    it("parses string 'true' as true", () => {
      const env = {
        ...createValidEnv(),
        SHOULD_SUBMIT_VAULT_REPORT: "true",
      };

      const parsed = configSchema.parse(env);

      expect(parsed.SHOULD_SUBMIT_VAULT_REPORT).toBe(true);
    });

    it("parses case-insensitive string 'FALSE' as false", () => {
      const env = {
        ...createValidEnv(),
        SHOULD_SUBMIT_VAULT_REPORT: "FALSE",
      };

      const parsed = configSchema.parse(env);

      expect(parsed.SHOULD_SUBMIT_VAULT_REPORT).toBe(false);
    });

    it("parses case-insensitive string 'True' as true", () => {
      const env = {
        ...createValidEnv(),
        SHOULD_SUBMIT_VAULT_REPORT: "True",
      };

      const parsed = configSchema.parse(env);

      expect(parsed.SHOULD_SUBMIT_VAULT_REPORT).toBe(true);
    });

    it("parses numeric string '0' as false", () => {
      const env = {
        ...createValidEnv(),
        SHOULD_SUBMIT_VAULT_REPORT: "0",
      };

      const parsed = configSchema.parse(env);

      expect(parsed.SHOULD_SUBMIT_VAULT_REPORT).toBe(false);
    });

    it("parses numeric string '1' as true", () => {
      const env = {
        ...createValidEnv(),
        SHOULD_SUBMIT_VAULT_REPORT: "1",
      };

      const parsed = configSchema.parse(env);

      expect(parsed.SHOULD_SUBMIT_VAULT_REPORT).toBe(true);
    });

    it("parses actual boolean false as false", () => {
      const env = {
        ...createValidEnv(),
        SHOULD_SUBMIT_VAULT_REPORT: false,
      };

      const parsed = configSchema.parse(env);

      expect(parsed.SHOULD_SUBMIT_VAULT_REPORT).toBe(false);
    });

    it("parses actual boolean true as true", () => {
      const env = {
        ...createValidEnv(),
        SHOULD_SUBMIT_VAULT_REPORT: true,
      };

      const parsed = configSchema.parse(env);

      expect(parsed.SHOULD_SUBMIT_VAULT_REPORT).toBe(true);
    });

    it("parses numeric 0 as false", () => {
      const env = {
        ...createValidEnv(),
        SHOULD_SUBMIT_VAULT_REPORT: 0,
      };

      const parsed = configSchema.parse(env);

      expect(parsed.SHOULD_SUBMIT_VAULT_REPORT).toBe(false);
    });

    it("parses numeric 1 as true", () => {
      const env = {
        ...createValidEnv(),
        SHOULD_SUBMIT_VAULT_REPORT: 1,
      };

      const parsed = configSchema.parse(env);

      expect(parsed.SHOULD_SUBMIT_VAULT_REPORT).toBe(true);
    });

    it("handles strings with whitespace", () => {
      const env = {
        ...createValidEnv(),
        SHOULD_SUBMIT_VAULT_REPORT: "  false  ",
      };

      const parsed = configSchema.parse(env);

      expect(parsed.SHOULD_SUBMIT_VAULT_REPORT).toBe(false);
    });

    it("rejects invalid string values", () => {
      const env = {
        ...createValidEnv(),
        SHOULD_SUBMIT_VAULT_REPORT: "invalid",
      };

      const result = configSchema.safeParse(env);

      expect(result.success).toBe(false);
      if (!result.success) {
        expect(result.error.issues.some((issue) => issue.path.join(".") === "SHOULD_SUBMIT_VAULT_REPORT")).toBe(true);
      }
    });

    it("rejects empty string", () => {
      const env = {
        ...createValidEnv(),
        SHOULD_SUBMIT_VAULT_REPORT: "",
      };

      const result = configSchema.safeParse(env);

      expect(result.success).toBe(false);
      if (!result.success) {
        expect(result.error.issues.some((issue) => issue.path.join(".") === "SHOULD_SUBMIT_VAULT_REPORT")).toBe(true);
      }
    });
  });

  describe("IS_UNPAUSE_STAKING_ENABLED", () => {
    it("parses string 'false' as false", () => {
      const env = {
        ...createValidEnv(),
        IS_UNPAUSE_STAKING_ENABLED: "false",
      };

      const parsed = configSchema.parse(env);

      expect(parsed.IS_UNPAUSE_STAKING_ENABLED).toBe(false);
    });

    it("parses string 'true' as true", () => {
      const env = {
        ...createValidEnv(),
        IS_UNPAUSE_STAKING_ENABLED: "true",
      };

      const parsed = configSchema.parse(env);

      expect(parsed.IS_UNPAUSE_STAKING_ENABLED).toBe(true);
    });

    it("parses case-insensitive string 'FALSE' as false", () => {
      const env = {
        ...createValidEnv(),
        IS_UNPAUSE_STAKING_ENABLED: "FALSE",
      };

      const parsed = configSchema.parse(env);

      expect(parsed.IS_UNPAUSE_STAKING_ENABLED).toBe(false);
    });

    it("parses case-insensitive string 'True' as true", () => {
      const env = {
        ...createValidEnv(),
        IS_UNPAUSE_STAKING_ENABLED: "True",
      };

      const parsed = configSchema.parse(env);

      expect(parsed.IS_UNPAUSE_STAKING_ENABLED).toBe(true);
    });

    it("parses numeric string '0' as false", () => {
      const env = {
        ...createValidEnv(),
        IS_UNPAUSE_STAKING_ENABLED: "0",
      };

      const parsed = configSchema.parse(env);

      expect(parsed.IS_UNPAUSE_STAKING_ENABLED).toBe(false);
    });

    it("parses numeric string '1' as true", () => {
      const env = {
        ...createValidEnv(),
        IS_UNPAUSE_STAKING_ENABLED: "1",
      };

      const parsed = configSchema.parse(env);

      expect(parsed.IS_UNPAUSE_STAKING_ENABLED).toBe(true);
    });

    it("parses actual boolean false as false", () => {
      const env = {
        ...createValidEnv(),
        IS_UNPAUSE_STAKING_ENABLED: false,
      };

      const parsed = configSchema.parse(env);

      expect(parsed.IS_UNPAUSE_STAKING_ENABLED).toBe(false);
    });

    it("parses actual boolean true as true", () => {
      const env = {
        ...createValidEnv(),
        IS_UNPAUSE_STAKING_ENABLED: true,
      };

      const parsed = configSchema.parse(env);

      expect(parsed.IS_UNPAUSE_STAKING_ENABLED).toBe(true);
    });

    it("parses numeric 0 as false", () => {
      const env = {
        ...createValidEnv(),
        IS_UNPAUSE_STAKING_ENABLED: 0,
      };

      const parsed = configSchema.parse(env);

      expect(parsed.IS_UNPAUSE_STAKING_ENABLED).toBe(false);
    });

    it("parses numeric 1 as true", () => {
      const env = {
        ...createValidEnv(),
        IS_UNPAUSE_STAKING_ENABLED: 1,
      };

      const parsed = configSchema.parse(env);

      expect(parsed.IS_UNPAUSE_STAKING_ENABLED).toBe(true);
    });

    it("handles strings with whitespace", () => {
      const env = {
        ...createValidEnv(),
        IS_UNPAUSE_STAKING_ENABLED: "  false  ",
      };

      const parsed = configSchema.parse(env);

      expect(parsed.IS_UNPAUSE_STAKING_ENABLED).toBe(false);
    });

    it("rejects invalid string values", () => {
      const env = {
        ...createValidEnv(),
        IS_UNPAUSE_STAKING_ENABLED: "invalid",
      };

      const result = configSchema.safeParse(env);

      expect(result.success).toBe(false);
      if (!result.success) {
        expect(result.error.issues.some((issue) => issue.path.join(".") === "IS_UNPAUSE_STAKING_ENABLED")).toBe(true);
      }
    });

    it("rejects empty string", () => {
      const env = {
        ...createValidEnv(),
        IS_UNPAUSE_STAKING_ENABLED: "",
      };

      const result = configSchema.safeParse(env);

      expect(result.success).toBe(false);
      if (!result.success) {
        expect(result.error.issues.some((issue) => issue.path.join(".") === "IS_UNPAUSE_STAKING_ENABLED")).toBe(true);
      }
    });
  });

  describe("LOG_LEVEL", () => {
    it("accepts valid log levels", () => {
      const validLevels = ["error", "warn", "info", "verbose", "debug", "silly"];

      for (const level of validLevels) {
        const env = {
          ...createValidEnv(),
          LOG_LEVEL: level,
        };

        const parsed = configSchema.parse(env);

        expect(parsed.LOG_LEVEL).toBe(level);
      }
    });

    it("accepts undefined LOG_LEVEL (optional)", () => {
      const env = createValidEnv();
      // LOG_LEVEL is not set

      const parsed = configSchema.parse(env);

      expect(parsed.LOG_LEVEL).toBeUndefined();
    });

    it("rejects invalid log levels", () => {
      const env = {
        ...createValidEnv(),
        LOG_LEVEL: "invalid",
      };

      const result = configSchema.safeParse(env);

      expect(result.success).toBe(false);
      if (!result.success) {
        expect(result.error.issues.some((issue) => issue.path.join(".") === "LOG_LEVEL")).toBe(true);
      }
    });
  });
});
