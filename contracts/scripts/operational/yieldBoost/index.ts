import { task } from "hardhat/config";
import { ArgumentType } from "hardhat/types/arguments";
import type { HardhatPlugin } from "hardhat/types/plugins";

export const yieldBoostTasksPlugin: HardhatPlugin = {
  id: "linea-yieldboost-tasks",
  tasks: [
    task(
      "addLidoStVaultYieldProvider",
      "Generates parameters for adding and configuring a new LidoStVaultYieldProvider",
    )
      .addOption({
        name: "yieldManager",
        description: "The yield manager address",
        type: ArgumentType.STRING,
        defaultValue: "",
      })
      .addOption({
        name: "yieldProvider",
        description: "The yield provider address",
        type: ArgumentType.STRING,
        defaultValue: "",
      })
      .addOption({
        name: "nodeOperator",
        description: "The node operator address",
        type: ArgumentType.STRING,
        defaultValue: "",
      })
      .addOption({
        name: "securityCouncil",
        description: "The security council address",
        type: ArgumentType.STRING,
        defaultValue: "",
      })
      .addOption({
        name: "nodeOperatorFee",
        description: "The node operator fee",
        type: ArgumentType.STRING,
        defaultValue: "0",
      })
      .addOption({
        name: "confirmExpiry",
        description: "The confirm expiry",
        type: ArgumentType.STRING,
        defaultValue: "0",
      })
      .setAction(() => import("./actions/addLidoStVaultYieldProvider.js"))
      .build(),

    task("prepareInitiateOssification", "Generates parameters for initiating ossification")
      .addOption({
        name: "yieldManager",
        description: "The yield manager address",
        type: ArgumentType.STRING,
        defaultValue: "",
      })
      .addOption({
        name: "yieldProvider",
        description: "The yield provider address",
        type: ArgumentType.STRING,
        defaultValue: "",
      })
      .setAction(() => import("./actions/prepareInitiateOssification.js"))
      .build(),

    task("addAndClaimMessage", "Adds and claims a message for testing")
      .addOption({
        name: "lineaRollup",
        description: "The Linea rollup address",
        type: ArgumentType.STRING,
        defaultValue: "",
      })
      .addOption({
        name: "yieldManager",
        description: "The yield manager address",
        type: ArgumentType.STRING,
        defaultValue: "",
      })
      .addOption({
        name: "amount",
        description: "The amount",
        type: ArgumentType.STRING,
        defaultValue: "",
      })
      .setAction(() => import("./actions/addAndClaimMessage.js"))
      .build(),

    task("addAndClaimMessageForLST", "Adds and claims a message for LST testing")
      .addOption({
        name: "lineaRollup",
        description: "The Linea rollup address",
        type: ArgumentType.STRING,
        defaultValue: "",
      })
      .addOption({
        name: "yieldManager",
        description: "The yield manager address",
        type: ArgumentType.STRING,
        defaultValue: "",
      })
      .addOption({
        name: "amount",
        description: "The amount",
        type: ArgumentType.STRING,
        defaultValue: "",
      })
      .setAction(() => import("./actions/addAndClaimMessageForLST.js"))
      .build(),

    task("unstakePermissionless", "Unstakes permissionlessly for testing")
      .addOption({
        name: "lineaRollup",
        description: "The Linea rollup address",
        type: ArgumentType.STRING,
        defaultValue: "",
      })
      .addOption({
        name: "yieldManager",
        description: "The yield manager address",
        type: ArgumentType.STRING,
        defaultValue: "",
      })
      .addOption({
        name: "yieldProvider",
        description: "The yield provider address",
        type: ArgumentType.STRING,
        defaultValue: "",
      })
      .setAction(() => import("./actions/unstakePermissionless.js"))
      .build(),
  ],
};

export default yieldBoostTasksPlugin;
