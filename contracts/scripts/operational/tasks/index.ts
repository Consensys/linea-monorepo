import { task } from "hardhat/config";
import { ArgumentType } from "hardhat/types/arguments";
import type { HardhatPlugin } from "hardhat/types/plugins";

export const operationalTasksPlugin: HardhatPlugin = {
  id: "linea-operational-tasks",
  tasks: [
    task("getCurrentFinalizedBlockNumber", "Gets the finalized block number")
      .addOption({
        name: "contractType",
        description: "The contract type (e.g., LineaRollup)",
        type: ArgumentType.STRING,
        defaultValue: "",
      })
      .addOption({
        name: "proxyAddress",
        description: "The proxy address",
        type: ArgumentType.STRING,
        defaultValue: "",
      })
      .setAction(() => import("./actions/getCurrentFinalizedBlockNumber.js"))
      .build(),

    task("grantContractRoles", "Grants roles to specific accounts")
      .addOption({
        name: "adminAddress",
        description: "The admin address to grant roles to",
        type: ArgumentType.STRING,
        defaultValue: "",
      })
      .addOption({
        name: "proxyAddress",
        description: "The proxy address",
        type: ArgumentType.STRING,
        defaultValue: "",
      })
      .addOption({
        name: "contractType",
        description: "The contract type",
        type: ArgumentType.STRING,
        defaultValue: "",
      })
      .addOption({
        name: "contractRoles",
        description: "Comma-separated role hashes",
        type: ArgumentType.STRING,
        defaultValue: "",
      })
      .setAction(() => import("./actions/grantContractRoles.js"))
      .build(),

    task("renounceContractRoles", "Renounces roles from specific accounts")
      .addOption({
        name: "adminAddress",
        description: "The admin address to renounce roles from",
        type: ArgumentType.STRING,
        defaultValue: "",
      })
      .addOption({
        name: "proxyAddress",
        description: "The proxy address",
        type: ArgumentType.STRING,
        defaultValue: "",
      })
      .addOption({
        name: "contractType",
        description: "The contract type",
        type: ArgumentType.STRING,
        defaultValue: "",
      })
      .addOption({
        name: "contractRoles",
        description: "Comma-separated role hashes",
        type: ArgumentType.STRING,
        defaultValue: "",
      })
      .setAction(() => import("./actions/renounceContractRoles.js"))
      .build(),

    task("setRateLimit", "Sets the rate limit on a Message Service contract")
      .addOption({
        name: "messageServiceAddress",
        description: "The message service address",
        type: ArgumentType.STRING,
        defaultValue: "",
      })
      .addOption({
        name: "messageServiceType",
        description: "The message service type",
        type: ArgumentType.STRING,
        defaultValue: "",
      })
      .addOption({
        name: "withdrawLimit",
        description: "The withdraw limit in wei",
        type: ArgumentType.STRING,
        defaultValue: "",
      })
      .setAction(() => import("./actions/setRateLimit.js"))
      .build(),

    task("setVerifierAddress", "Sets the verifier address on LineaRollup")
      .addOption({
        name: "proxyAddress",
        description: "The proxy address",
        type: ArgumentType.STRING,
        defaultValue: "",
      })
      .addOption({
        name: "contractType",
        description: "The contract type",
        type: ArgumentType.STRING,
        defaultValue: "",
      })
      .addOption({
        name: "newVerifierAddress",
        description: "The new verifier address",
        type: ArgumentType.STRING,
        defaultValue: "",
      })
      .addOption({
        name: "proofType",
        description: "The proof type",
        type: ArgumentType.STRING,
        defaultValue: "",
      })
      .setAction(() => import("./actions/setVerifierAddress.js"))
      .build(),

    task("setMessageServiceOnTokenBridge", "Sets the message service on TokenBridge")
      .addOption({
        name: "tokenBridgeAddress",
        description: "The token bridge address",
        type: ArgumentType.STRING,
        defaultValue: "",
      })
      .addOption({
        name: "newMessageServiceAddress",
        description: "The new message service address",
        type: ArgumentType.STRING,
        defaultValue: "",
      })
      .setAction(() => import("./actions/setMessageServiceOnTokenBridge.js"))
      .build(),
  ],
};

export default operationalTasksPlugin;
