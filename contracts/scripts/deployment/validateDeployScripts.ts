#!/usr/bin/env ts-node
/**
 * Validates deploy script initializer signatures against compiled contract ABIs and
 * checks that registry-backed lookups used by deploy scripts resolve on known networks.
 *
 * Run after `pnpm -F contracts run build`.
 */
import { Interface, InterfaceAbi } from "ethers";
import fs from "fs";
import path from "path";

import {
  EMPTY_INITIALIZE_SIGNATURE,
  LINEA_ROLLUP_INITIALIZE_SIGNATURE,
  L2_MESSAGE_SERVICE_INITIALIZE_SIGNATURE,
  ROLLUP_REVENUE_VAULT_INITIALIZE_SIGNATURE,
  VALIDIUM_INITIALIZE_SIGNATURE,
  YIELD_MANAGER_INITIALIZE_SIGNATURE,
} from "../../common/constants/general";
import {
  getPopulatedAddresses,
  loadRegistry,
  lookupRegistryEntry,
  REGISTRY_NETWORKS,
} from "../../common/helpers/addressRegistry";

const BUILD_DIR = path.join(__dirname, "..", "..", "build");

type ValidationIssue = {
  category: "signature" | "registry";
  target: string;
  message: string;
};

type InitializerCheck = {
  label: string;
  artifactRelativePath: string;
  signature: string;
};

const INITIALIZER_CHECKS: InitializerCheck[] = [
  {
    label: "LineaRollup",
    artifactRelativePath: "src/rollup/LineaRollup.sol/LineaRollup.json",
    signature: LINEA_ROLLUP_INITIALIZE_SIGNATURE,
  },
  {
    label: "Validium",
    artifactRelativePath: "src/rollup/Validium.sol/Validium.json",
    signature: VALIDIUM_INITIALIZE_SIGNATURE,
  },
  {
    label: "L2MessageService",
    artifactRelativePath: "src/messaging/l2/L2MessageService.sol/L2MessageService.json",
    signature: L2_MESSAGE_SERVICE_INITIALIZE_SIGNATURE,
  },
  {
    label: "RollupRevenueVault",
    artifactRelativePath: "src/operational/RollupRevenueVault.sol/RollupRevenueVault.json",
    signature: ROLLUP_REVENUE_VAULT_INITIALIZE_SIGNATURE,
  },
  {
    label: "YieldManager",
    artifactRelativePath: "src/yield/YieldManager.sol/YieldManager.json",
    signature: YIELD_MANAGER_INITIALIZE_SIGNATURE,
  },
  {
    label: "RecoverFunds",
    artifactRelativePath: "src/recovery/RecoverFunds.sol/RecoverFunds.json",
    signature: "initialize(address,address)",
  },
  {
    label: "CustomBridgedToken",
    artifactRelativePath: "src/bridging/token/CustomBridgedToken.sol/CustomBridgedToken.json",
    signature: "initializeV2(string,string,uint8,address)",
  },
  {
    label: "UpgradeableWithdrawalQueuePredeploy",
    artifactRelativePath:
      "src/predeploy/UpgradeableWithdrawalQueuePredeploy.sol/UpgradeableWithdrawalQueuePredeploy.json",
    signature: EMPTY_INITIALIZE_SIGNATURE,
  },
  {
    label: "UpgradeableConsolidationQueuePredeploy",
    artifactRelativePath:
      "src/predeploy/UpgradeableConsolidationQueuePredeploy.sol/UpgradeableConsolidationQueuePredeploy.json",
    signature: EMPTY_INITIALIZE_SIGNATURE,
  },
  {
    label: "UpgradeableBeaconChainDepositPredeploy",
    artifactRelativePath:
      "src/predeploy/UpgradeableBeaconChainDepositPredeploy.sol/UpgradeableBeaconChainDepositPredeploy.json",
    signature: EMPTY_INITIALIZE_SIGNATURE,
  },
];

/** Registry lookups that must resolve via registry or env on the given network. */
const REGISTRY_LOOKUP_CHECKS: Array<{
  network: string;
  contractKey: string;
  envVarName: string;
  usedBy: string;
}> = [
  { network: "sepolia", contractKey: "L1_SECURITY_COUNCIL", envVarName: "L1_SECURITY_COUNCIL", usedBy: "LineaRollup" },
  {
    network: "sepolia",
    contractKey: "LINEA_ROLLUP_OPERATORS",
    envVarName: "LINEA_ROLLUP_OPERATORS",
    usedBy: "LineaRollup",
  },
  {
    network: "sepolia",
    contractKey: "YIELD_MANAGER_ADDRESS",
    envVarName: "YIELD_MANAGER_ADDRESS",
    usedBy: "LineaRollup",
  },
  {
    network: "linea_sepolia",
    contractKey: "L2_SECURITY_COUNCIL",
    envVarName: "L2_SECURITY_COUNCIL",
    usedBy: "L2MessageService",
  },
  {
    network: "linea_sepolia",
    contractKey: "L2_MESSAGE_SERVICE_L1L2_MESSAGE_SETTER",
    envVarName: "L2_MESSAGE_SERVICE_L1L2_MESSAGE_SETTER",
    usedBy: "L2MessageService",
  },
  {
    network: "linea_sepolia",
    contractKey: "TOKEN_BRIDGE_ADDRESS",
    envVarName: "TOKEN_BRIDGE_ADDRESS",
    usedBy: "RollupRevenueVault",
  },
  {
    network: "linea_sepolia",
    contractKey: "L2MessageService",
    envVarName: "L2_MESSAGE_SERVICE_ADDRESS",
    usedBy: "RollupRevenueVault",
  },
];

function loadArtifact(relativePath: string): { abi: InterfaceAbi } {
  const filePath = path.join(BUILD_DIR, relativePath);
  if (!fs.existsSync(filePath)) {
    throw new Error(`Missing artifact ${relativePath}. Run pnpm -F contracts run build first.`);
  }
  return JSON.parse(fs.readFileSync(filePath, "utf-8")) as { abi: InterfaceAbi };
}

function validateInitializerSignatures(): ValidationIssue[] {
  const issues: ValidationIssue[] = [];

  for (const check of INITIALIZER_CHECKS) {
    try {
      const artifact = loadArtifact(check.artifactRelativePath);
      const iface = new Interface(artifact.abi);
      const fn = iface.getFunction(check.signature);
      if (!fn) {
        issues.push({
          category: "signature",
          target: check.label,
          message: `No initializer "${check.signature}" in ${check.artifactRelativePath}`,
        });
      }
    } catch (error) {
      issues.push({
        category: "signature",
        target: check.label,
        message: error instanceof Error ? error.message : String(error),
      });
    }
  }

  return issues;
}

function validateRegistryLookups(): ValidationIssue[] {
  const issues: ValidationIssue[] = [];

  for (const check of REGISTRY_LOOKUP_CHECKS) {
    if (!REGISTRY_NETWORKS.has(check.network)) {
      issues.push({
        category: "registry",
        target: `${check.network}/${check.contractKey}`,
        message: `Unknown registry network "${check.network}"`,
      });
      continue;
    }

    const registry = loadRegistry(check.network);
    if (!registry) {
      issues.push({
        category: "registry",
        target: `${check.network}/${check.contractKey}`,
        message: `Registry file missing for network "${check.network}"`,
      });
      continue;
    }

    const entry = lookupRegistryEntry(registry, check.contractKey, check.envVarName);
    if (!entry) {
      issues.push({
        category: "registry",
        target: `${check.network}/${check.contractKey}`,
        message: `No registry entry for ${check.contractKey} or ${check.envVarName} (required by ${check.usedBy}; env var needed at deploy time)`,
      });
      continue;
    }

    if (getPopulatedAddresses(entry) === undefined) {
      issues.push({
        category: "registry",
        target: `${check.network}/${check.contractKey}`,
        message: `Registry entry is zero/unpopulated (required by ${check.usedBy}; env var needed at deploy time)`,
      });
    }
  }

  return issues;
}

function main(): void {
  const signatureIssues = validateInitializerSignatures();
  const registryIssues = validateRegistryLookups();
  const issues = [...signatureIssues, ...registryIssues];

  if (issues.length === 0) {
    console.log("Deploy script validation passed (signatures + registry lookups).");
    return;
  }

  console.error(`Deploy script validation failed with ${issues.length} issue(s):\n`);
  for (const issue of issues) {
    console.error(`- [${issue.category}] ${issue.target}: ${issue.message}`);
  }
  process.exitCode = 1;
}

main();
