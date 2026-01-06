import { task } from "hardhat/config";
import { getTaskCliOrEnvValue } from "../../common/helpers/environmentHelper";
import {
  fetchBeaconHeader,
  fetchBeaconState,
  fetchBeaconHeaderByParentRoot,
} from "@lidofinance/lsv-cli/dist/utils/fetchCL";
import { createBeaconHeaderProof, createStateProof } from "@lidofinance/lsv-cli/dist/utils/proof/proofs.js";
import { hexlify, AbiCoder } from "ethers";

/*
  *******************************************************************************************
  Performs YieldManager::unstakePermissionless

  Note is for older contract versions that have been deployed
  - YieldManager implementation at [0xBae1F082825E0C64F724c5568D495A4Aa43698a3](https://hoodi.etherscan.io/address/0xBae1F082825E0C64F724c5568D495A4Aa43698a3)
  - LidoStVaultYieldProvider implementation deployed by [0x87f9b74D6A3EbD66A8081b3781c24E0DFEd4C2F5](https://hoodi.etherscan.io/address/0x87f9b74d6a3ebd66a8081b3781c24e0dfed4c2f5#code)

  -------------------------------------------------------------------------------------------
  Example (Hoodi):
  -------------------------------------------------------------------------------------------
  CUSTOM_PRIVATE_KEY=0000000000000000000000000000000000000000000000000000000000000002 \
  CUSTOM_BLOCKCHAIN_URL=https://0xrpc.io/hoodi \
  npx hardhat unstakePermissionless \
    --yield-manager <address> \
    --yield-provider <address> \
    --validator-index <uint64> \
    --slot <uint64> \
    --unstake-amount-gwei <string> \
    --beacon-rpc-url <string> \
    --network custom
  -------------------------------------------------------------------------------------------

  Env var alternatives (used if CLI params omitted):
    YIELD_MANAGER
    YIELD_PROVIDER
    VALIDATOR_INDEX
    SLOT
    BEACON_RPC_URL
  *******************************************************************************************
*/
task("unstakePermissionless", "Performs YieldManager::unstakePermissionless")
  .addOptionalParam("yieldManager")
  .addOptionalParam("yieldProvider")
  .addOptionalParam("validatorIndex")
  .addOptionalParam("slot")
  .addOptionalParam("unstakeAmountGwei")
  .addOptionalParam("beaconRpcUrl")
  .setAction(async (taskArgs, hre) => {
    // Get signer
    const { ethers } = hre;
    const [signer] = await ethers.getSigners();

    // --- Resolve inputs from CLI or ENV ---
    const yieldManagerAddress = getTaskCliOrEnvValue(taskArgs, "yieldManager", "YIELD_MANAGER");
    const yieldProvider = getTaskCliOrEnvValue(taskArgs, "yieldProvider", "YIELD_PROVIDER");
    const validatorIndexRaw = getTaskCliOrEnvValue(taskArgs, "validatorIndex", "VALIDATOR_INDEX");
    const slotRaw = getTaskCliOrEnvValue(taskArgs, "slot", "SLOT");
    const unstakeAmountGweiRaw = getTaskCliOrEnvValue(taskArgs, "unstakeAmountGwei", "UNSTAKE_AMOUNT_GWEI");
    const beaconRpcUrl = getTaskCliOrEnvValue(taskArgs, "beaconRpcUrl", "BEACON_RPC_URL");

    // --- Basic required fields check (adjust as needed) ---
    const missing: string[] = [];
    if (!yieldManagerAddress) missing.push("yieldManager / YIELD_MANAGER");
    if (!yieldProvider) missing.push("yieldProvider / YIELD_PROVIDER");
    if (!validatorIndexRaw) missing.push("validatorIndex / VALIDATOR_INDEX");
    if (!slotRaw) missing.push("slot / SLOT");
    if (!unstakeAmountGweiRaw) missing.push("unstakeAmountGwei / UNSTAKE_AMOUNT_GWEI");
    if (!beaconRpcUrl) missing.push("beaconRpcUrl / BEACON_RPC_URL");
    if (missing.length) {
      throw new Error(`Missing required params/envs: ${missing.join(", ")}`);
    }

    // --- Parse numeric params ---
    const validatorIndex = parseInt(validatorIndexRaw!, 10);
    const slot = parseInt(slotRaw!, 10);
    const unstakeAmountGwei = unstakeAmountGweiRaw ? BigInt(unstakeAmountGweiRaw) : 0n;

    // --- Log params ---
    console.log("Params:");
    console.log("  yieldManager:", yieldManagerAddress);
    console.log("  yieldProvider:", yieldProvider);
    console.log("  validatorIndex:", validatorIndex.toString());
    console.log("  slot:", slot.toString());
    console.log("  unstakeAmountGwei:", unstakeAmountGwei.toString());
    console.log("  beaconRpcUrl:", beaconRpcUrl);

    // Get beacon state
    const beaconHeaderJson = await fetchBeaconHeader(slot, beaconRpcUrl);
    const beaconHeader = beaconHeaderJson.data.header.message;
    console.log("beaconHeader:", beaconHeader);
    const { stateBodyBytes, forkName } = await fetchBeaconState(slot, beaconRpcUrl);

    // Proofs
    const { proof: beaconHeaderProof, root: beaconHeaderRoot } = await createBeaconHeaderProof(beaconHeader);
    const {
      proof: validatorStateProof,
      validator,
      view: validatorStateView,
    } = await createStateProof(validatorIndex, stateBodyBytes, forkName);
    const proofConcat = [...validatorStateProof.witnesses, ...beaconHeaderProof.witnesses];
    const proofHex = proofConcat.map((w) => hexlify(w));

    // Fill out Validator Witness fields
    const headerByParentJson = await fetchBeaconHeaderByParentRoot(beaconHeaderRoot, beaconRpcUrl);
    const headerByParentSlot = headerByParentJson.data[0].header.message.slot;
    const headerByParentTimestamp = validatorStateView.genesisTime + headerByParentSlot * 12;

    // struct ValidatorContainerWitness {
    //   bytes32[] proof;
    //   uint256 validatorIndex;
    //   uint64 effectiveBalance;
    //   uint64 childBlockTimestamp;
    //   uint64 slot;
    //   uint64 proposerIndex;
    //   uint64 activationEpoch;
    //   uint64 activationEligibilityEpoch;
    // }

    const witness = {
      proof: proofHex,
      validatorIndex: BigInt(validatorIndex),
      effectiveBalance: BigInt(validator.effectiveBalance),
      childBlockTimestamp: BigInt(headerByParentTimestamp),
      slot: BigInt(beaconHeader.slot),
      proposerIndex: BigInt(beaconHeader.proposer_index),
      activationEpoch: BigInt(validator.activationEpoch),
      activationEligibilityEpoch: BigInt(validator.activationEligibilityEpoch),
    };
    console.log("witness:", witness);

    // Encode params
    const withdrawalParams = AbiCoder.defaultAbiCoder().encode(
      ["bytes", "uint64[]", "address"],
      [hexlify(validator.pubkey), [unstakeAmountGwei], signer.address],
    );

    const VALIDATOR_CONTAINER_WITNESS_TYPE =
      "tuple(bytes32[] proof, uint256 validatorIndex, uint64 effectiveBalance, uint64 childBlockTimestamp, uint64 slot, uint64 proposerIndex, uint64 activationEpoch, uint64 activationEligibilityEpoch)";

    const withdrawalParamsProof = AbiCoder.defaultAbiCoder().encode([VALIDATOR_CONTAINER_WITNESS_TYPE], [witness]);

    // Define minimal ABI for YieldManager
    const YIELD_MANAGER_ABI = [
      {
        inputs: [
          { internalType: "address", name: "_yieldProvider", type: "address" },
          { internalType: "bytes", name: "_withdrawalParams", type: "bytes" },
          { internalType: "bytes", name: "_withdrawalParamsProof", type: "bytes" },
        ],
        name: "unstakePermissionless",
        outputs: [{ internalType: "uint256", name: "maxUnstakeAmount", type: "uint256" }],
        stateMutability: "payable",
        type: "function",
      },
    ];

    // Create contract instance
    const yieldManagerContract = await ethers.getContractAt(YIELD_MANAGER_ABI, yieldManagerAddress!, signer);

    // Call function
    console.log("Calling unstakePermissionless...");
    try {
      const tx = await yieldManagerContract.unstakePermissionless(
        yieldProvider!,
        withdrawalParams,
        withdrawalParamsProof,
        { value: 1n },
      );
      console.log("Transaction sent, hash:", tx.hash);
      const receipt = await tx.wait();
      console.log("Transaction receipt:", receipt);
      console.log("Transaction successful!");
    } catch (error: unknown) {
      // Extract error data
      if (error && typeof error === "object") {
        const errorObj = error as Record<string, unknown>;
        if (errorObj.data) {
          console.error("Error data:", errorObj.data);
        } else if (errorObj.error && typeof errorObj.error === "object") {
          const nestedError = errorObj.error as Record<string, unknown>;
          if (nestedError.data) {
            console.error("Error data:", nestedError.data);
          }
        }
      }
      throw error;
    }
  });
