import { ethers } from "ethers";
import { describe, expect, it } from "@jest/globals";
import { config } from "./config/tests-config";
import { waitForEvents, etherToWei } from "./common/utils";

const l1AccountManager = config.getL1AccountManager();
const l2AccountManager = config.getL2AccountManager();
const tokenTotalSuppy = ethers.parseEther("100000");
const bridgeAmount = ethers.parseEther("100");

describe("Bridge ERC20 Tokens L1 -> L2 and L2 -> L1", () => {
  it.concurrent("Bridge a token from L1 to L2", async () => {
    const [l1Account, l2Account] = await Promise.all([
      l1AccountManager.whaleAccount(4),
      l2AccountManager.whaleAccount(4),
    ]);

    const [lineaRollup, l2MessageService, l1TokenBridge, l2TokenBridge, l1Token] = await Promise.all([
      config.getLineaRollupContract(),
      config.getL2MessageServiceContract(),
      config.getL1TokenBridgeContract(),
      config.getL2TokenBridgeContract(),
      config.getL1TokenContract(),
    ]);

    console.log("Approving tokens to L1 TokenBridge");
    const approveTx = await l1Token.connect(l1Account).approve(l1TokenBridge.getAddress(), ethers.parseEther("100"));
    await approveTx.wait();

    const l1TokenBridgeAddress = await l1TokenBridge.getAddress();
    const l1TokenAddress = await l1Token.getAddress();

    const allowanceL1Account = await l1Token.allowance(l1Account.address, l1TokenBridgeAddress);
    console.log("Current allowance of L1 account to L1 TokenBridge is :", allowanceL1Account.toString());

    console.log("Calling the bridgeToken function on the L1 TokenBridge contract");

    const bridgeTokenTx = await l1TokenBridge
      .connect(l1Account)
      .bridgeToken(l1TokenAddress, bridgeAmount, l2Account.address, {
        value: etherToWei("0.01"),
        gasPrice: ethers.parseUnits("300", "gwei"),
      });

    let receipt = await bridgeTokenTx.wait();
    while (!receipt) {
      console.log("Waiting for transaction to be mined...");
      receipt = await bridgeTokenTx.wait();
      console.log("receipt", receipt);
    }

    const l1TokenBalance = await l1Token.balanceOf(l1Account.address);
    console.log("Token balance of L1 account :", l1TokenBalance.toString());

    expect(l1TokenBalance).toEqual(tokenTotalSuppy - ethers.parseEther("100"));

    console.log("Waiting for MessageSent event on L1.");

    const [messageSentEvent] = await waitForEvents(
      lineaRollup,
      lineaRollup.filters.MessageSent(),
      500,
      bridgeTokenTx.blockNumber!,
    );
    const messageEventArgs = messageSentEvent.args;
    const messageNumber = messageEventArgs._nonce;
    const messageHash = messageEventArgs._messageHash;

    console.log(`Message sent on L1 : ${JSON.stringify(messageSentEvent)}`);

    console.log("Waiting for anchoring...");

    const [rollingHashUpdatedEvent] = await waitForEvents(
      l2MessageService,
      l2MessageService.filters.RollingHashUpdated(),
      1_000,
      0,
      "latest",
      async (events) => events.filter((event) => event.args.messageNumber >= messageNumber),
    );
    expect(rollingHashUpdatedEvent).not.toBeNull();

    const anchoredStatus = await l2MessageService.inboxL1L2MessageStatus(messageHash);

    expect(anchoredStatus).toBeGreaterThan(0);

    console.log(`Message anchored : ${JSON.stringify(rollingHashUpdatedEvent)}`);

    console.log("Waiting for MessageClaimed event on L2...");

    const [claimedEvent] = await waitForEvents(l2MessageService, l2MessageService.filters.MessageClaimed(messageHash));
    expect(claimedEvent).not.toBeNull();

    const [newTokenDeployed] = await waitForEvents(l2TokenBridge, l2TokenBridge.filters.NewTokenDeployed());
    expect(newTokenDeployed).not.toBeNull();

    console.log(`Message claimed on L2 : ${JSON.stringify(claimedEvent)}.`);

    const l2Token = config.getL2BridgedTokenContract(newTokenDeployed.args.bridgedToken);

    console.log("Verify the token balance on L2");

    const l2TokenBalance = await l2Token.balanceOf(l2Account.address);
    console.log("Token balance of L2 account :", l2TokenBalance.toString());

    expect(l2TokenBalance).toEqual(bridgeAmount);
  });

  it.concurrent("Bridge a token from L2 to L1", async () => {
    const [l1Account, l2Account] = await Promise.all([
      l1AccountManager.whaleAccount(4),
      l2AccountManager.whaleAccount(4),
    ]);

    const lineaRollup = config.getLineaRollupContract(l1Account);
    const l2MessageService = config.getL2MessageServiceContract(l2Account);
    const l2TokenBridge = config.getL2TokenBridgeContract(l2Account);
    const l1Token = config.getL1TokenContract(l1Account);

    const [newTokenDeployed] = await waitForEvents(l2TokenBridge, l2TokenBridge.filters.NewTokenDeployed());
    expect(newTokenDeployed).not.toBeNull();

    const l2Token = config.getL2BridgedTokenContract(newTokenDeployed.args.bridgedToken);

    console.log("Approving tokens to L2 TokenBridge");

    const l2Provider = config.getL2Provider();
    const { maxPriorityFeePerGas: l2MaxPriorityFeePerGas, maxFeePerGas: l2MaxFeePerGas } =
      await l2Provider.getFeeData();

    const approveTx = await l2Token.connect(l2Account).approve(l2TokenBridge.getAddress(), ethers.parseEther("100"), {
      maxPriorityFeePerGas: l2MaxPriorityFeePerGas,
      maxFeePerGas: l2MaxFeePerGas,
    });
    await approveTx.wait();

    const allowanceL2Account = await l2Token.allowance(l2Account.address, l2TokenBridge.getAddress());
    console.log("Current allowance of L2 account to L2 TokenBridge is :", allowanceL2Account.toString());
    console.log("Current balance of  L2 account is :", await l2Token.balanceOf(l2Account));

    console.log("Calling the bridgeToken function on the L1 TokenBridge contract");

    const bridgeTokenTx = await l2TokenBridge
      .connect(l2Account)
      .bridgeToken(await l2Token.getAddress(), bridgeAmount, l1Account.address, {
        value: etherToWei("0.01"),
        maxPriorityFeePerGas: l2MaxPriorityFeePerGas,
        maxFeePerGas: l2MaxFeePerGas,
      });
    let receipt = await bridgeTokenTx.wait();
    while (!receipt) {
      console.log("Waiting for transaction to be mined...");
      receipt = await bridgeTokenTx.wait();
      console.log("receipt", receipt);
    }

    console.log("Waiting for MessageSent event on L2...");

    const [messageSentEvent] = await waitForEvents(
      l2MessageService,
      l2MessageService.filters.MessageSent(),
      500,
      receipt!.blockNumber,
    );

    console.log(`L2 message sent : ${JSON.stringify(messageSentEvent)}`);

    console.log("Waiting for L1 MessageClaimed event.");

    const [claimedEvent] = await waitForEvents(
      lineaRollup,
      lineaRollup.filters.MessageClaimed(messageSentEvent.args._messageHash),
    );
    expect(claimedEvent).not.toBeNull();
    console.log(`Message claimed on L1 : ${JSON.stringify(claimedEvent)}`);

    console.log("Verify the token balance on L1");

    const l1TokenBalance = await l1Token.balanceOf(l1Account.address);
    console.log("Token balance of L1 account :", l1TokenBalance.toString());

    expect(l1TokenBalance).toEqual(tokenTotalSuppy);
  });
});
