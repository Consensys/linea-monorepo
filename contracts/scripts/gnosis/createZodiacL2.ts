import { deployAndSetUpModule, KnownContracts } from "@gnosis.pm/zodiac";
import Safe, { CreateTransactionProps, EthersAdapter } from "@safe-global/protocol-kit";
import { ethers } from "ethers";
import { L1_L2_MESSAGE_SETTER_ROLE, OPERATOR_ROLE } from "../../test/common/constants";
import { requireEnv } from "../hardhat/utils";
import { get1559Fees } from "../utils";

const main = async () => {
  const RPC_URL = requireEnv("BLOCKCHAIN_NODE");
  const SAFE_OWNER1_PRIVATE_KEY = requireEnv("SAFE_OWNER1_PRIVATE_KEY");
  // const SAFE_OWNERS = requireEnv("SAFE_OWNERS");
  const safe4OutOf8Contract = requireEnv("SAFE_4_OUT_OF_8_CONTRACT");
  const safe5OutOf8Contract = requireEnv("SAFE_5_OUT_OF_8_CONTRACT");
  const timelockContract = requireEnv("TIMELOCK_CONTRACT");
  const proxyContract = requireEnv("PROXY_CONTRACT");

  const provider = new ethers.JsonRpcProvider(RPC_URL);
  const signer = new ethers.Wallet(SAFE_OWNER1_PRIVATE_KEY, provider);
  const deployedSafeAddress = requireEnv("SAFE_SECURITY_COUNCIL_ADDRESS");

  const ethAdapter = new EthersAdapter({ ethers, signerOrProvider: signer });

  const chainId = await ethAdapter.getChainId();
  console.log(`ChainId: ${chainId}`);
  const eip1559Fees = await get1559Fees(provider);

  // const safeVersion = '1.3.0'
  // const isL1SafeMasterCopy = false
  const gnosisSafe = await Safe.create({ ethAdapter: ethAdapter, safeAddress: deployedSafeAddress });

  // const safeAccountConfig = {
  //   threshold: 1, // Setting the Threshold to 1
  //   owners: SAFE_OWNERS?.split(","),
  // };

  // console.log("Deploying Security Council safe..");

  // const gnosisSafe = await safeFactory.deploySafe({ safeAccountConfig });
  const safeAddress = await gnosisSafe.getAddress();

  console.log(`Security Council Safe deployed at: ${safeAddress}`);

  // Deploy Zodiac
  const deployZodiac = await deployAndSetUpModule(
    KnownContracts.ROLES,
    {
      types: ["address", "address", "address"],
      values: [safeAddress, safeAddress, safeAddress],
    },
    provider,
    5, // Can change into 'chainId' once the network is supported in https://github.com/gnosis/zodiac/blob/master/sdk/contracts.ts
    (await provider.getTransactionCount(await signer.getAddress())).toString(),
  );

  console.log("deployZodiac.expectedModuleAddress :", deployZodiac.expectedModuleAddress);

  const tx = await signer.sendTransaction({ ...deployZodiac.transaction, ...(await get1559Fees(provider)) });
  await tx.wait();
  console.log("deployZodiac SendTransaction");

  console.log("checking to see if module is enabled");
  const isEnabled = await gnosisSafe.isModuleEnabled(deployZodiac.expectedModuleAddress);
  console.log("Is module enabled? :", isEnabled);

  const enableModuleTx = await gnosisSafe.createEnableModuleTx(deployZodiac.expectedModuleAddress);

  console.log("User connecting to Safe");

  console.log("Connect with ethAdapter & Safe address ");
  const userSafeConnection = await gnosisSafe.connect({ ethAdapter: ethAdapter, safeAddress });

  console.log("Getting enable module transaction hash");
  const enableModuleTxHash = await userSafeConnection.getTransactionHash(enableModuleTx);

  console.log("Signing...");
  const enableModuleTxResponse = await userSafeConnection.approveTransactionHash(enableModuleTxHash, {
    maxFeePerGas: eip1559Fees.maxFeePerGas?.toString(),
    maxPriorityFeePerGas: eip1559Fees.maxPriorityFeePerGas?.toString(),
  });

  await enableModuleTxResponse.transactionResponse?.wait();
  console.log("Signed enable module transaction");

  console.log("Executing enable module transaction");
  const executeTxResponse = await userSafeConnection.executeTransaction(enableModuleTx, {
    maxFeePerGas: eip1559Fees.maxFeePerGas?.toString(),
    maxPriorityFeePerGas: eip1559Fees.maxPriorityFeePerGas?.toString(),
  });

  await executeTxResponse.transactionResponse?.wait();

  console.log("Verifying module enabled");
  const isEnabled2 = await gnosisSafe.isModuleEnabled(deployZodiac.expectedModuleAddress);
  console.log("Is module enabled? :", isEnabled2);

  // Multisend
  console.log("Getting multisend address");
  const multisendAddress = gnosisSafe.getMultiSendAddress();

  console.log("Creating transaction data payload");
  const data = ethers.concat(["0x8b95eccd", ethers.AbiCoder.defaultAbiCoder().encode(["address"], [multisendAddress])]);

  const safeTransactionData: CreateTransactionProps = {
    transactions: [
      {
        to: deployZodiac.expectedModuleAddress,
        value: "0",
        data,
      },
    ],
    options: {
      safeTxGas: "100000", // required for setting the multiSend address
    },
  };

  console.log("Crafting setMultisend transaction");

  const setMultisendTx = await gnosisSafe.createTransaction(safeTransactionData);

  console.log("Getting setMultisend transaction hash");
  const setMultisendTxHash = await userSafeConnection.getTransactionHash(setMultisendTx);

  console.log("Signing setMultisend transaction hash");
  await userSafeConnection.signHash(setMultisendTxHash);

  console.log("Executing setMultisend transaction");
  const executeSendMultisendTxResponse = await userSafeConnection.executeTransaction(setMultisendTx, {
    maxFeePerGas: eip1559Fees.maxFeePerGas?.toString(),
    maxPriorityFeePerGas: eip1559Fees.maxPriorityFeePerGas?.toString(),
  });
  await executeSendMultisendTxResponse.transactionResponse?.wait();

  console.log("Zodiac Role Creation Complete!");

  // Create Role 1
  console.log("Creating Zodiac Role 1");

  // AssignRoles
  console.log("Creating assignRoles 1 transaction data payload");
  const assignRolesData = ethers.concat([
    "0xa6edf38f", // 0xa6edf38f = assignRoles in Roles
    ethers.AbiCoder.defaultAbiCoder().encode(["address", "uint16[]", "bool[]"], [safe4OutOf8Contract, [1], [true]]),
  ]);

  console.log("Crafting createAssignRoles 1 transaction");

  const safeTransactionData2: CreateTransactionProps = {
    transactions: [
      {
        to: deployZodiac.expectedModuleAddress,
        value: "0",
        data: assignRolesData,
      },
    ],
  };

  const createAssignRolesTx = await gnosisSafe.createTransaction(safeTransactionData2);
  console.log("createAssignRoles transaction created");

  console.log("Getting createAssignRoles 1 transaction hash");
  const getAssignRolesTxHash = await userSafeConnection.getTransactionHash(createAssignRolesTx);

  console.log("Signing setMultisend 1 transaction hash");
  await userSafeConnection.signHash(getAssignRolesTxHash);

  console.log("Executing createAssignRoles 1 transaction");
  const executecreateAssignRolesTxResponse = await userSafeConnection.executeTransaction(createAssignRolesTx, {
    maxFeePerGas: eip1559Fees.maxFeePerGas?.toString(),
    maxPriorityFeePerGas: eip1559Fees.maxPriorityFeePerGas?.toString(),
  });
  await executecreateAssignRolesTxResponse.transactionResponse?.wait();

  // ScopeTarget
  console.log("Creating ScopeTarget transaction data payload");
  const scopeTargetData = ethers.concat([
    "0x5e826695", // 0x5e826695 = scopeTarget function in Roles
    ethers.AbiCoder.defaultAbiCoder().encode(["uint16", "address"], [1, timelockContract]),
  ]);

  const safeTransactionData3: CreateTransactionProps = {
    transactions: [
      {
        to: deployZodiac.expectedModuleAddress,
        value: "0",
        data: scopeTargetData,
      },
    ],
  };
  const createScopeTargetTx = await gnosisSafe.createTransaction(safeTransactionData3);
  console.log("Created createScopeTarget transaction");

  console.log("Getting createScopeTarget transaction hash");
  const getScopeTargetTxHash = await userSafeConnection.getTransactionHash(createScopeTargetTx);

  console.log("Signing getScopeTarget transaction hash");
  await userSafeConnection.signHash(getScopeTargetTxHash);

  console.log("Executing createScopeTargetTx transaction");
  const executeCreateScopeTargetTxResponse = await userSafeConnection.executeTransaction(createScopeTargetTx, {
    maxFeePerGas: eip1559Fees.maxFeePerGas?.toString(),
    maxPriorityFeePerGas: eip1559Fees.maxPriorityFeePerGas?.toString(),
  });
  await executeCreateScopeTargetTxResponse.transactionResponse?.wait();

  // ScopeFunction - Schedule Function with 7 day delay

  console.log("Creating scopeFunction (schedule w/ 7days delay) transaction data payload");
  const scopeFunctionData = ethers.concat([
    "0x33a0480c", // 0x33a0480c = scopeFunction in Roles
    ethers.AbiCoder.defaultAbiCoder().encode(
      ["uint16", "address", "bytes4", "bool[]", "uint8[]", "uint8[]", "bytes[]", "uint8"],
      [
        1,
        timelockContract,
        "0x01d5062a", // 0x01d5062a = schedule function in Timelock
        [false, false, false, false, false, true],
        [0, 0, 0, 0, 0, 0],
        [0, 0, 0, 0, 0, 0],
        ["0x", "0x", "0x", "0x", "0x", ethers.AbiCoder.defaultAbiCoder().encode(["uint256"], [604800])],
        0,
      ],
    ),
  ]);

  console.log("Crafting scopeFunctionData (schedule w/ 7days delay) transaction");

  const safeTransactionData4: CreateTransactionProps = {
    transactions: [
      {
        to: deployZodiac.expectedModuleAddress,
        value: "0",
        data: scopeFunctionData,
      },
    ],
  };

  const createScopeFunctionTx = await gnosisSafe.createTransaction(safeTransactionData4);
  console.log("Created createScopeFunction (schedule w/ 7days delay) transaction");

  console.log("Getting createScopeFunction (schedule w/ 7days delay) transaction hash");
  const getCreateScopeFunctionTxHash = await userSafeConnection.getTransactionHash(createScopeFunctionTx);

  console.log("Signing getCreateScopeFunction (schedule w/ 7days delay) transaction hash");
  await userSafeConnection.signHash(getCreateScopeFunctionTxHash);

  console.log("Executing createScopeAllow (schedule w/ 7days delay) transaction");
  const executeCreateScopeFunctionTxResponse = await userSafeConnection.executeTransaction(createScopeFunctionTx, {
    maxFeePerGas: eip1559Fees.maxFeePerGas?.toString(),
    maxPriorityFeePerGas: eip1559Fees.maxPriorityFeePerGas?.toString(),
  });
  await executeCreateScopeFunctionTxResponse.transactionResponse?.wait();

  // ScopeAllow - Execute Function

  console.log("Creating scopeAllow (execute) transaction data payload");
  const scopeAllow10 = ethers.concat([
    "0x2fcf52d1", // 0x2fcf52d1 = scopeAllow function in Roles
    ethers.AbiCoder.defaultAbiCoder().encode(
      ["uint16", "address", "bytes4", "uint8"],
      [1, timelockContract, "0x134008d3", 0], // 0x134008d3 = execute function in Timelock
    ),
  ]);
  // console.log("scopeAllowData3 :", scopeAllowData3);

  console.log("Crafting createScopeAllow (execute) transaction");

  const safeTransactionData10: CreateTransactionProps = {
    transactions: [
      {
        to: deployZodiac.expectedModuleAddress,
        value: "0",
        data: scopeAllow10,
      },
    ],
  };

  const createScopeAllowTx10 = await gnosisSafe.createTransaction(safeTransactionData10);
  console.log("Created createScopeAllow (execute) transaction");

  console.log("Getting createScopeAllow (execute) transaction hash");
  const getScopeAllowTxTxHash10 = await userSafeConnection.getTransactionHash(createScopeAllowTx10);

  console.log("Signing getScopeAllow (execute) transaction hash");
  await userSafeConnection.signHash(getScopeAllowTxTxHash10);

  console.log("Executing createScopeAllow (execute) transaction");
  const executeCreateScopeAllowTxResponse10 = await userSafeConnection.executeTransaction(createScopeAllowTx10, {
    maxFeePerGas: eip1559Fees.maxFeePerGas?.toString(),
    maxPriorityFeePerGas: eip1559Fees.maxPriorityFeePerGas?.toString(),
  });
  await executeCreateScopeAllowTxResponse10.transactionResponse?.wait();

  // Create Role 2
  console.log("Creating Zodiac Role 2");

  // AssignRoles
  console.log("Creating assignRoles 2 transaction data payload");
  const assignRolesData21 = ethers.concat([
    "0xa6edf38f", // 0xa6edf38f = assignRoles in Roles
    ethers.AbiCoder.defaultAbiCoder().encode(["address", "uint16[]", "bool[]"], [safe5OutOf8Contract, [2], [true]]),
  ]);

  console.log("Crafting createAssignRoles 2 transaction");

  const safeTransactionData21: CreateTransactionProps = {
    transactions: [
      {
        to: deployZodiac.expectedModuleAddress,
        value: "0",
        data: assignRolesData21,
      },
    ],
  };

  const createAssignRolesTx21 = await gnosisSafe.createTransaction(safeTransactionData21);
  console.log("createAssignRoles transaction created");

  console.log("Getting createAssignRoles 2 transaction hash");
  const getAssignRolesTxHash21 = await userSafeConnection.getTransactionHash(createAssignRolesTx21);

  console.log("Signing setMultisend 2 transaction hash");
  await userSafeConnection.signHash(getAssignRolesTxHash21);

  console.log("Executing createAssignRoles 2 transaction");
  const executecreateAssignRolesTxResponse21 = await userSafeConnection.executeTransaction(createAssignRolesTx21, {
    maxFeePerGas: eip1559Fees.maxFeePerGas?.toString(),
    maxPriorityFeePerGas: eip1559Fees.maxPriorityFeePerGas?.toString(),
  });
  await executecreateAssignRolesTxResponse21.transactionResponse?.wait();

  // ScopeTarget
  console.log("Creating ScopeTarget transaction data payload");
  const scopeTargetData5 = ethers.concat([
    "0x5e826695", // 0x5e826695 = scopeTarget function in Roles
    ethers.AbiCoder.defaultAbiCoder().encode(["uint16", "address"], [2, timelockContract]),
  ]);

  const safeTransactionData5: CreateTransactionProps = {
    transactions: [
      {
        to: deployZodiac.expectedModuleAddress,
        value: "0",
        data: scopeTargetData5,
      },
    ],
  };

  const createScopeTargetTx5 = await gnosisSafe.createTransaction(safeTransactionData5);
  console.log("Created createScopeTarget transaction");

  console.log("Getting createScopeTarget transaction hash");
  const getScopeTargetTxHash5 = await userSafeConnection.getTransactionHash(createScopeTargetTx5);

  console.log("Signing getScopeTarget transaction hash");
  await userSafeConnection.signHash(getScopeTargetTxHash5);

  console.log("Executing createScopeTargetTx transaction");
  const executeCreateScopeTargetTxResponse5 = await userSafeConnection.executeTransaction(createScopeTargetTx5, {
    maxFeePerGas: eip1559Fees.maxFeePerGas?.toString(),
    maxPriorityFeePerGas: eip1559Fees.maxPriorityFeePerGas?.toString(),
  });
  await executeCreateScopeTargetTxResponse5.transactionResponse?.wait();

  // ScopeFunction - Schedule Function with 1 day delay

  console.log("Creating scopeFunction (schedule w/ 1 day delay) transaction data payload");
  const scopeFunctionData25 = ethers.concat([
    "0x33a0480c", // 0x33a0480c = ScopeFunction in Roles
    ethers.AbiCoder.defaultAbiCoder().encode(
      ["uint16", "address", "bytes4", "bool[]", "uint8[]", "uint8[]", "bytes[]", "uint8"],
      [
        2,
        timelockContract,
        "0x01d5062a", // 0x01d5062a = schedule function in Timelock
        [false, false, false, false, false, true],
        [0, 0, 0, 0, 0, 0],
        [0, 0, 0, 0, 0, 0],
        ["0x", "0x", "0x", "0x", "0x", ethers.AbiCoder.defaultAbiCoder().encode(["uint256"], [86400])],
        0,
      ],
    ),
  ]);

  console.log("Crafting scopeFunctionData (schedule w/ 1 day delay) transaction");

  const safeTransactionData25: CreateTransactionProps = {
    transactions: [
      {
        to: deployZodiac.expectedModuleAddress,
        value: "0",
        data: scopeFunctionData25,
      },
    ],
  };

  const createScopeFunctionTx25 = await gnosisSafe.createTransaction(safeTransactionData25);
  console.log("Created createScopeFunction (schedule w/ 1 day delay) transaction");

  console.log("Getting createScopeFunction (schedule w/ 1 day delay) transaction hash");
  const getCreateScopeFunctionTxHash25 = await userSafeConnection.getTransactionHash(createScopeFunctionTx25);

  console.log("Signing getCreateScopeFunction (schedule w/ 1 day delay) transaction hash");
  await userSafeConnection.signHash(getCreateScopeFunctionTxHash25);

  console.log("Executing createScopeAllow (schedule w/ 1 day delay) transaction");
  const executeCreateScopeFunctionTxResponse25 = await userSafeConnection.executeTransaction(createScopeFunctionTx25, {
    maxFeePerGas: eip1559Fees.maxFeePerGas?.toString(),
    maxPriorityFeePerGas: eip1559Fees.maxPriorityFeePerGas?.toString(),
  });
  await executeCreateScopeFunctionTxResponse25.transactionResponse?.wait();

  // ScopeAllow - Execute Function

  console.log("Creating scopeAllow (execute) transaction data payload");
  const scopeAllow30 = ethers.concat([
    "0x2fcf52d1", // 0x2fcf52d1 = scopeAllow function in Roles
    ethers.AbiCoder.defaultAbiCoder().encode(
      ["uint16", "address", "bytes4", "uint8"],
      [2, timelockContract, "0x134008d3", 0], // 0x134008d3 = execute function in Timelock
    ),
  ]);

  console.log("Crafting createScopeAllow (execute) transaction");

  const safeTransactionData30: CreateTransactionProps = {
    transactions: [
      {
        to: deployZodiac.expectedModuleAddress,
        value: "0",
        data: scopeAllow30,
      },
    ],
  };

  const createScopeAllowTx30 = await gnosisSafe.createTransaction(safeTransactionData30);
  console.log("Created createScopeAllow (execute) transaction");

  console.log("Getting createScopeAllow (execute) transaction hash");
  const getScopeAllowTxTxHash30 = await userSafeConnection.getTransactionHash(createScopeAllowTx30);

  console.log("Signing getScopeAllow (execute) transaction hash");
  await userSafeConnection.signHash(getScopeAllowTxTxHash30);

  console.log("Executing createScopeAllow (execute) transaction");
  const executeCreateScopeAllowTxResponse30 = await userSafeConnection.executeTransaction(createScopeAllowTx30, {
    maxFeePerGas: eip1559Fees.maxFeePerGas?.toString(),
    maxPriorityFeePerGas: eip1559Fees.maxPriorityFeePerGas?.toString(),
  });
  await executeCreateScopeAllowTxResponse30.transactionResponse?.wait();

  // Creating Role 3
  console.log("Creating Zodiac Role 3");

  // AssignRoles 3
  console.log("Creating assignRoles 3 transaction data payload");
  const assignRolesData35 = ethers.concat([
    "0xa6edf38f", // 0xa6edf38f = assignRoles function in Roles
    ethers.AbiCoder.defaultAbiCoder().encode(["address", "uint16[]", "bool[]"], [safe4OutOf8Contract, [3], [true]]),
  ]);

  console.log("Crafting createAssignRoles 3 transaction");

  const safeTransactionData35: CreateTransactionProps = {
    transactions: [
      {
        to: deployZodiac.expectedModuleAddress,
        value: "0",
        data: assignRolesData35,
      },
    ],
  };

  const createAssignRolesTx35 = await gnosisSafe.createTransaction(safeTransactionData35);
  console.log("Created createAssignRoles 3 transaction");

  console.log("Getting createAssignRoles 3 transaction hash");
  const getAssignRolesTxHash35 = await userSafeConnection.getTransactionHash(createAssignRolesTx35);

  console.log("Signing setMultisend transaction hash");
  await userSafeConnection.signHash(getAssignRolesTxHash35);

  console.log("Executing createAssignRoles 3 transaction");
  const executecreateAssignRolesTxResponse35 = await userSafeConnection.executeTransaction(createAssignRolesTx35, {
    maxFeePerGas: eip1559Fees.maxFeePerGas?.toString(),
    maxPriorityFeePerGas: eip1559Fees.maxPriorityFeePerGas?.toString(),
  });
  await executecreateAssignRolesTxResponse35.transactionResponse?.wait();

  // ScopeTarget 3

  console.log("Creating ScopeTarget 3 transaction data payload");
  const ScopeTargetData40 = ethers.concat([
    "0x5e826695", // 0x5e826695 = ScopeTarget function in Roles
    ethers.AbiCoder.defaultAbiCoder().encode(["uint16", "address"], [3, proxyContract]),
  ]);

  const safeTransactionData40: CreateTransactionProps = {
    transactions: [
      {
        to: deployZodiac.expectedModuleAddress,
        value: "0",
        data: ScopeTargetData40,
      },
    ],
  };

  const createAllowTargetTx40 = await gnosisSafe.createTransaction(safeTransactionData40);
  console.log("Created createAllowTarget 3 transaction");

  console.log("Getting createAllowTarget 3 transaction hash");
  const getAllowTargetTxHash40 = await userSafeConnection.getTransactionHash(createAllowTargetTx40);

  console.log("Signing getAllowTarget 3 transaction hash");
  await userSafeConnection.signHash(getAllowTargetTxHash40);

  console.log("Executing createAllowTarget 3 transaction");
  const executeCreateAllowTargetTxResponse40 = await userSafeConnection.executeTransaction(createAllowTargetTx40, {
    maxFeePerGas: eip1559Fees.maxFeePerGas?.toString(),
    maxPriorityFeePerGas: eip1559Fees.maxPriorityFeePerGas?.toString(),
  });
  await executeCreateAllowTargetTxResponse40.transactionResponse?.wait();

  // ScopeAllowFunction 3 - pauseByType
  console.log("Creating scopeAllow (pauseByTime) transaction data payload");
  const scopeAllowData45 = ethers.concat([
    "0x2fcf52d1",
    ethers.AbiCoder.defaultAbiCoder().encode(
      ["uint16", "address", "bytes4", "uint8"],
      [3, proxyContract, "0x8264bd82", 0],
    ), // 0x8264bd82 = pauseByType function in Proxy
  ]);

  console.log("Crafting createScopeAllow (pauseByTime) transaction");

  const safeTransactionData45: CreateTransactionProps = {
    transactions: [
      {
        to: deployZodiac.expectedModuleAddress,
        value: "0",
        data: scopeAllowData45,
      },
    ],
  };

  const createScopeAllowTx45 = await gnosisSafe.createTransaction(safeTransactionData45);
  console.log("Created createScopeAllow (pauseByTime) transaction");

  console.log("Getting createScopeAllow (pauseByTime) transaction hash");
  const getScopeAllowTxTxHash45 = await userSafeConnection.getTransactionHash(createScopeAllowTx45);

  console.log("Signing getScopeAllow (pauseByTime) transaction hash");
  await userSafeConnection.signHash(getScopeAllowTxTxHash45);

  console.log("Executing createScopeAllow (pauseByTime) transaction");
  const executeCreateScopeAllowTxResponse45 = await userSafeConnection.executeTransaction(createScopeAllowTx45, {
    maxFeePerGas: eip1559Fees.maxFeePerGas?.toString(),
    maxPriorityFeePerGas: eip1559Fees.maxPriorityFeePerGas?.toString(),
  });
  await executeCreateScopeAllowTxResponse45.transactionResponse?.wait();

  // ScopeFunction - grantRole Function with keccak256("OPERATOR_ROLE") condition

  console.log("Creating scopeFunction transaction data payload");
  const scopeFunctionData50 = ethers.concat([
    "0x33a0480c",
    ethers.AbiCoder.defaultAbiCoder().encode(
      ["uint16", "address", "bytes4", "bool[]", "uint8[]", "uint8[]", "bytes[]", "uint8"],
      [3, proxyContract, "0x2f2ff15d", [true, false], [0, 0], [0, 0], [OPERATOR_ROLE, "0x"], 0],
    ),
  ]);

  console.log("Crafting scopeFunctionData (grantRole) transaction");

  const safeTransactionData50: CreateTransactionProps = {
    transactions: [
      {
        to: deployZodiac.expectedModuleAddress,
        value: "0",
        data: scopeFunctionData50,
      },
    ],
  };

  const createScopeFunctionTx50 = await gnosisSafe.createTransaction(safeTransactionData50);
  console.log("Created createScopeFunction (grantRole) transaction");

  console.log("Getting createScopeFunction (grantRole) transaction hash");
  const getCreateScopeFunctionTxHash50 = await userSafeConnection.getTransactionHash(createScopeFunctionTx50);

  console.log("Signing getCreateScopeFunction (grantRole) transaction hash");
  await userSafeConnection.signHash(getCreateScopeFunctionTxHash50);

  console.log("Executing createScopeAllow (grantRole) transaction");
  const executeCreateScopeFunctionTxResponse50 = await userSafeConnection.executeTransaction(createScopeFunctionTx50, {
    maxFeePerGas: eip1559Fees.maxFeePerGas?.toString(),
    maxPriorityFeePerGas: eip1559Fees.maxPriorityFeePerGas?.toString(),
  });
  await executeCreateScopeFunctionTxResponse50.transactionResponse?.wait();

  // ScopeFunction - revokeRole Function with keccak256("OPERATOR_ROLE") condition

  console.log("Creating scopeFunction (revokeRole) transaction data payload");
  const scopeFunctionData55 = ethers.concat([
    "0x33a0480c",
    ethers.AbiCoder.defaultAbiCoder().encode(
      ["uint16", "address", "bytes4", "bool[]", "uint8[]", "uint8[]", "bytes[]", "uint8"],
      [3, proxyContract, "0xd547741f", [true, false], [0, 0], [0, 0], [OPERATOR_ROLE, "0x"], 0],
    ),
  ]);

  console.log("Crafting scopeFunctionData (revokeRole) transaction");

  const safeTransactionData55: CreateTransactionProps = {
    transactions: [
      {
        to: deployZodiac.expectedModuleAddress,
        value: "0",
        data: scopeFunctionData55,
      },
    ],
  };

  const createScopeFunctionTx55 = await gnosisSafe.createTransaction(safeTransactionData55);
  console.log("Created createScopeFunction (revokeRole) transaction");

  console.log("Getting createScopeFunction (revokeRole) transaction hash");
  const getCreateScopeFunctionTxHash55 = await userSafeConnection.getTransactionHash(createScopeFunctionTx55);

  console.log("Signing getCreateScopeFunction (revokeRole) transaction hash");
  await userSafeConnection.signHash(getCreateScopeFunctionTxHash55);

  console.log("Executing createScopeAllow (revokeRole) transaction");
  const executeCreateScopeFunctionTxResponse55 = await userSafeConnection.executeTransaction(createScopeFunctionTx55, {
    maxFeePerGas: eip1559Fees.maxFeePerGas?.toString(),
    maxPriorityFeePerGas: eip1559Fees.maxPriorityFeePerGas?.toString(),
  });
  await executeCreateScopeFunctionTxResponse55.transactionResponse?.wait();

  // Creating Role 4
  console.log("Creating Zodiac Role 4");

  // AssignRoles 4
  console.log("Creating assignRoles 4 transaction data payload");
  const assignRolesData60 = ethers.concat([
    "0xa6edf38f", // 0xa6edf38f = assignRoles function in Roles
    ethers.AbiCoder.defaultAbiCoder().encode(["address", "uint16[]", "bool[]"], [safe4OutOf8Contract, [4], [true]]),
  ]);

  console.log("Crafting createAssignRoles 4 transaction");

  const safeTransactionData60: CreateTransactionProps = {
    transactions: [
      {
        to: deployZodiac.expectedModuleAddress,
        value: "0",
        data: assignRolesData60,
      },
    ],
  };

  const createAssignRolesTx60 = await gnosisSafe.createTransaction(safeTransactionData60);
  console.log("Created createAssignRoles 4 transaction");

  console.log("Getting createAssignRoles 4 transaction hash");
  const getAssignRolesTxHash60 = await userSafeConnection.getTransactionHash(createAssignRolesTx60);

  console.log("Signing setMultisend transaction hash");
  await userSafeConnection.signHash(getAssignRolesTxHash60);

  console.log("Executing createAssignRoles 4 transaction");
  const executecreateAssignRolesTxResponse60 = await userSafeConnection.executeTransaction(createAssignRolesTx60, {
    maxFeePerGas: eip1559Fees.maxFeePerGas?.toString(),
    maxPriorityFeePerGas: eip1559Fees.maxPriorityFeePerGas?.toString(),
  });
  await executecreateAssignRolesTxResponse60.transactionResponse?.wait();

  // ScopeTarget 4

  console.log("Creating ScopeTarget 4 transaction data payload");
  const ScopeTargetData65 = ethers.concat([
    "0x5e826695", // 0x5e826695 = ScopeTarget function in Roles
    ethers.AbiCoder.defaultAbiCoder().encode(["uint16", "address"], [4, proxyContract]),
  ]);

  const safeTransactionData65: CreateTransactionProps = {
    transactions: [
      {
        to: deployZodiac.expectedModuleAddress,
        value: "0",
        data: ScopeTargetData65,
      },
    ],
  };

  const createAllowTargetTx65 = await gnosisSafe.createTransaction(safeTransactionData65);
  console.log("Created createAllowTarget 4 transaction");

  console.log("Getting createAllowTarget 4 transaction hash");
  const getAllowTargetTxHash65 = await userSafeConnection.getTransactionHash(createAllowTargetTx65);

  console.log("Signing getAllowTarget 4 transaction hash");
  await userSafeConnection.signHash(getAllowTargetTxHash65);

  console.log("Executing createAllowTarget 4 transaction");
  const executeCreateAllowTargetTxResponse65 = await userSafeConnection.executeTransaction(createAllowTargetTx65, {
    maxFeePerGas: eip1559Fees.maxFeePerGas?.toString(),
    maxPriorityFeePerGas: eip1559Fees.maxPriorityFeePerGas?.toString(),
  });
  await executeCreateAllowTargetTxResponse65.transactionResponse?.wait();

  // ScopeAllowFunction 4 - resetRateLimitAmount

  console.log("Creating scopeAllow (resetRateLimitAmount) transaction data payload");
  const scopeAllowData70 = ethers.concat([
    "0x2fcf52d1",
    ethers.AbiCoder.defaultAbiCoder().encode(
      ["uint16", "address", "bytes4", "uint8"],
      [4, proxyContract, "0x557eac73", 0],
    ), // 0x557eac73 = resetRateLimitAmount function in Proxy
  ]);

  console.log("Crafting createScopeAllow (resetRateLimitAmount) transaction");

  const safeTransactionData70: CreateTransactionProps = {
    transactions: [
      {
        to: deployZodiac.expectedModuleAddress,
        value: "0",
        data: scopeAllowData70,
      },
    ],
  };

  const createScopeAllowTx70 = await gnosisSafe.createTransaction(safeTransactionData70);
  console.log("Created createScopeAllow (resetRateLimitAmount) transaction");

  console.log("Getting createScopeAllow (resetRateLimitAmount) transaction hash");
  const getScopeAllowTxTxHash70 = await userSafeConnection.getTransactionHash(createScopeAllowTx70);

  console.log("Signing getScopeAllow (resetRateLimitAmount) transaction hash");
  await userSafeConnection.signHash(getScopeAllowTxTxHash70);

  console.log("Executing createScopeAllow (resetRateLimitAmount) transaction");
  const executeCreateScopeAllowTxResponse70 = await userSafeConnection.executeTransaction(createScopeAllowTx70, {
    maxFeePerGas: eip1559Fees.maxFeePerGas?.toString(),
    maxPriorityFeePerGas: eip1559Fees.maxPriorityFeePerGas?.toString(),
  });
  await executeCreateScopeAllowTxResponse70.transactionResponse?.wait();

  // ScopeAllowFunction 4 - setMinimumFee

  console.log("Creating scopeAllow (setMinimumFee) transaction data payload");
  const scopeAllowData75 = ethers.concat([
    "0x2fcf52d1",
    ethers.AbiCoder.defaultAbiCoder().encode(
      ["uint16", "address", "bytes4", "uint8"],
      [4, proxyContract, "0x182a7506", 0],
    ), // 0x182a7506 = setMinimumFee function in Proxy
  ]);

  console.log("Crafting createScopeAllow (setMinimumFee) transaction");

  const safeTransactionData75: CreateTransactionProps = {
    transactions: [
      {
        to: deployZodiac.expectedModuleAddress,
        value: "0",
        data: scopeAllowData75,
      },
    ],
  };

  const createScopeAllowTx75 = await gnosisSafe.createTransaction(safeTransactionData75);
  console.log("Created createScopeAllow (setMinimumFee) transaction");

  console.log("Getting createScopeAllow (setMinimumFee) transaction hash");
  const getScopeAllowTxTxHash75 = await userSafeConnection.getTransactionHash(createScopeAllowTx75);

  console.log("Signing getScopeAllow (setMinimumFee) transaction hash");
  await userSafeConnection.signHash(getScopeAllowTxTxHash75);

  console.log("Executing createScopeAllow (setMinimumFee) transaction");
  const executeCreateScopeAllowTxResponse75 = await userSafeConnection.executeTransaction(createScopeAllowTx75, {
    maxFeePerGas: eip1559Fees.maxFeePerGas?.toString(),
    maxPriorityFeePerGas: eip1559Fees.maxPriorityFeePerGas?.toString(),
  });
  await executeCreateScopeAllowTxResponse75.transactionResponse?.wait();

  // Creating Role 5
  console.log("Creating Zodiac Role 5");

  // AssignRoles 5
  console.log("Creating assignRoles 5 transaction data payload");
  const assignRolesData80 = ethers.concat([
    "0xa6edf38f", // 0xa6edf38f = assignRoles function in Roles
    ethers.AbiCoder.defaultAbiCoder().encode(["address", "uint16[]", "bool[]"], [safe4OutOf8Contract, [5], [true]]),
  ]);

  console.log("Crafting createAssignRoles 5 transaction");

  const safeTransactionData80: CreateTransactionProps = {
    transactions: [
      {
        to: deployZodiac.expectedModuleAddress,
        value: "0",
        data: assignRolesData80,
      },
    ],
  };

  const createAssignRolesTx80 = await gnosisSafe.createTransaction(safeTransactionData80);
  console.log("Created createAssignRoles 5 transaction");

  console.log("Getting createAssignRoles 5 transaction hash");
  const getAssignRolesTxHash80 = await userSafeConnection.getTransactionHash(createAssignRolesTx80);

  console.log("Signing setMultisend transaction hash");
  await userSafeConnection.signHash(getAssignRolesTxHash80);

  console.log("Executing createAssignRoles 5 transaction");
  const executecreateAssignRolesTxResponse80 = await userSafeConnection.executeTransaction(createAssignRolesTx80, {
    maxFeePerGas: eip1559Fees.maxFeePerGas?.toString(),
    maxPriorityFeePerGas: eip1559Fees.maxPriorityFeePerGas?.toString(),
  });
  await executecreateAssignRolesTxResponse80.transactionResponse?.wait();

  // ScopeTarget 5

  console.log("Creating ScopeTarget 5 transaction data payload");
  const ScopeTargetData85 = ethers.concat([
    "0x5e826695", // 0x5e826695 = ScopeTarget function in Roles
    ethers.AbiCoder.defaultAbiCoder().encode(["uint16", "address"], [5, proxyContract]),
  ]);

  const safeTransactionData85: CreateTransactionProps = {
    transactions: [
      {
        to: deployZodiac.expectedModuleAddress,
        value: "0",
        data: ScopeTargetData85,
      },
    ],
  };

  const createAllowTargetTx85 = await gnosisSafe.createTransaction(safeTransactionData85);
  console.log("Created createAllowTarget 5 transaction");

  console.log("Getting createAllowTarget 5 transaction hash");
  const getAllowTargetTxHash85 = await userSafeConnection.getTransactionHash(createAllowTargetTx85);

  console.log("Signing getAllowTarget 5 transaction hash");
  await userSafeConnection.signHash(getAllowTargetTxHash85);

  console.log("Executing createAllowTarget 5 transaction");
  const executeCreateAllowTargetTxResponse85 = await userSafeConnection.executeTransaction(createAllowTargetTx85, {
    maxFeePerGas: eip1559Fees.maxFeePerGas?.toString(),
    maxPriorityFeePerGas: eip1559Fees.maxPriorityFeePerGas?.toString(),
  });
  await executeCreateAllowTargetTxResponse85.transactionResponse?.wait();

  // ScopeFunction - grantRole Function with keccak256("L1_L2_MESSAGE_SETTER_ROLE") condition

  console.log("Creating scopeFunction transaction data payload");
  const scopeFunctionData90 = ethers.concat([
    "0x33a0480c",
    ethers.AbiCoder.defaultAbiCoder().encode(
      ["uint16", "address", "bytes4", "bool[]", "uint8[]", "uint8[]", "bytes[]", "uint8"],
      [5, proxyContract, "0x2f2ff15d", [true, false], [0, 0], [0, 0], [L1_L2_MESSAGE_SETTER_ROLE, "0x"], 0],
    ),
  ]);

  console.log("Crafting scopeFunctionData (grantRole) transaction");

  const safeTransactionData90: CreateTransactionProps = {
    transactions: [
      {
        to: deployZodiac.expectedModuleAddress,
        value: "0",
        data: scopeFunctionData90,
      },
    ],
  };

  const createScopeFunctionTx90 = await gnosisSafe.createTransaction(safeTransactionData90);
  console.log("Created createScopeFunction (grantRole) transaction");

  console.log("Getting createScopeFunction (grantRole) transaction hash");
  const getCreateScopeFunctionTxHash90 = await userSafeConnection.getTransactionHash(createScopeFunctionTx90);

  console.log("Signing getCreateScopeFunction (grantRole) transaction hash");
  await userSafeConnection.signHash(getCreateScopeFunctionTxHash90);

  console.log("Executing createScopeAllow (grantRole) transaction");
  const executeCreateScopeFunctionTxResponse90 = await userSafeConnection.executeTransaction(createScopeFunctionTx90, {
    maxFeePerGas: eip1559Fees.maxFeePerGas?.toString(),
    maxPriorityFeePerGas: eip1559Fees.maxPriorityFeePerGas?.toString(),
  });
  await executeCreateScopeFunctionTxResponse90.transactionResponse?.wait();

  // ScopeFunction - revokeRole Function with keccak256("L1_L2_MESSAGE_SETTER_ROLE") condition

  console.log("Creating scopeFunction (revokeRole) transaction data payload");
  const scopeFunctionData95 = ethers.concat([
    "0x33a0480c",
    ethers.AbiCoder.defaultAbiCoder().encode(
      ["uint16", "address", "bytes4", "bool[]", "uint8[]", "uint8[]", "bytes[]", "uint8"],
      [5, proxyContract, "0xd547741f", [true, false], [0, 0], [0, 0], [L1_L2_MESSAGE_SETTER_ROLE, "0x"], 0],
    ),
  ]);

  console.log("Crafting scopeFunctionData (revokeRole) transaction");

  const safeTransactionData95: CreateTransactionProps = {
    transactions: [
      {
        to: deployZodiac.expectedModuleAddress,
        value: "0",
        data: scopeFunctionData95,
      },
    ],
  };

  const createScopeFunctionTx95 = await gnosisSafe.createTransaction(safeTransactionData95);
  console.log("Created createScopeFunction (revokeRole) transaction");

  console.log("Getting createScopeFunction (revokeRole) transaction hash");
  const getCreateScopeFunctionTxHash95 = await userSafeConnection.getTransactionHash(createScopeFunctionTx95);

  console.log("Signing getCreateScopeFunction (revokeRole) transaction hash");
  await userSafeConnection.signHash(getCreateScopeFunctionTxHash95);

  console.log("Executing createScopeAllow (revokeRole) transaction");
  const executeCreateScopeFunctionTxResponse95 = await userSafeConnection.executeTransaction(createScopeFunctionTx95, {
    maxFeePerGas: eip1559Fees.maxFeePerGas?.toString(),
    maxPriorityFeePerGas: eip1559Fees.maxPriorityFeePerGas?.toString(),
  });
  await executeCreateScopeFunctionTxResponse95.transactionResponse?.wait();

  console.log("Zodiac Roles setup is complete!");
};

main()
  .then(() => process.exit(0))
  .catch((error) => {
    console.error(error);
    process.exit(1);
  });
