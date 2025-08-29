// import { ethers, toBigInt, Transaction } from "ethers";
// import { describe, expect, it } from "@jest/globals";
// import { config } from "./config/tests-config";
// import { LineaEstimateGasClient, RollupGetZkEVMBlockNumberClient, etherToWei } from "./common/utils";
// import { error } from "console";

// const l2AccountManager = config.getL2AccountManager();

// describe("Layer 2 test suite", () => {
//   const lineaEstimateGasClient = new LineaEstimateGasClient(config.getL2BesuNodeEndpoint()!);
//   it.concurrent("Should successfully send an EIP1559 transaction", async () => {
//     const account = await l2AccountManager.generateAccount();
//     let nonce = await account.getNonce();

//     // logger.debug(`Fetched fee data. maxPriorityFeePerGas=${maxPriorityFeePerGas} maxFeePerGas=${maxFeePerGas}`);
//     for (let i = 0; i < 50000; i++) {
//       // logger.info(i);
//       const transaction = Transaction.from({
//         type: 2,
//         nonce,
//         to: "0x8D97689C9818892B700e27F316cc3E41e17fBeb9",
//         value: etherToWei("0.0005"),
//         chainId: config.getL2ChainId(),
//       });

//       nonce += 1;

//       const { maxPriorityFeePerGas, maxFeePerGas, gasLimit } = await lineaEstimateGasClient.lineaEstimateGas(
//         account.address,
//         "0x8D97689C9818892B700e27F316cc3E41e17fBeb9",
//         transaction.unsignedSerialized,
//       );
//       // logger.info("saljem");
//       account
//         .sendTransaction({
//           ...transaction.toJSON(),
//           gasLimit,
//           maxPriorityFeePerGas,
//           maxFeePerGas,
//         })
//         .then((tx) => logger.info(`EIP1559 transaction sent. transactionHash=${tx.hash}`))
//         .catch((error) => console.log(JSON.stringify(error)));

//       // logger.debug(`EIP1559 transaction sent. transactionHash=${tx.hash}`);
//       // const receipt = await tx.wait();
//       // logger.debug(`EIP1559 transaction receipt received. transactionHash=${tx.hash} status=${receipt?.status}`);
//       // expect(receipt).not.toBeNull();
//     }
//   });
// });
