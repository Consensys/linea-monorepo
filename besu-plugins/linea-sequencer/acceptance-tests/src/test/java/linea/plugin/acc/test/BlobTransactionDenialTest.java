/*
 * Copyright Consensys Software Inc.
 *
 * This file is dual-licensed under either the MIT license or Apache License 2.0.
 * See the LICENSE-MIT and LICENSE-APACHE files in the repository root for details.
 *
 * SPDX-License-Identifier: MIT OR Apache-2.0
 */
package linea.plugin.acc.test;

import static org.assertj.core.api.Assertions.assertThat;
import static java.util.stream.Collectors.toList;
import java.util.List;
import java.util.Optional;

// import org.hyperledger.besu.ethereum.rlp.RLP;
// import org.hyperledger.besu.ethereum.rlp.RLPInput;
// import org.hyperledger.besu.ethereum.rlp.RLPOutput;

import org.apache.tuweni.bytes.Bytes;
import org.apache.tuweni.bytes.Bytes32;


import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.databind.node.ArrayNode;
import com.fasterxml.jackson.databind.node.ObjectNode;

import org.hyperledger.besu.datatypes.BlobGas;
import org.hyperledger.besu.datatypes.Hash;
import org.hyperledger.besu.datatypes.RequestType;

import org.hyperledger.besu.ethereum.api.jsonrpc.internal.parameters.EnginePayloadParameter;
import org.hyperledger.besu.ethereum.api.jsonrpc.internal.parameters.WithdrawalParameter;
import org.hyperledger.besu.ethereum.core.BlockHeader;
import org.hyperledger.besu.ethereum.core.BlockHeaderFunctions;
import org.hyperledger.besu.ethereum.core.Difficulty;
import org.hyperledger.besu.ethereum.core.Request;
import org.hyperledger.besu.ethereum.core.Transaction;
import org.hyperledger.besu.ethereum.core.Withdrawal;
import org.hyperledger.besu.ethereum.core.encoding.EncodingContext;
import org.hyperledger.besu.ethereum.core.encoding.TransactionDecoder;

import org.hyperledger.besu.ethereum.mainnet.BodyValidation;
import org.hyperledger.besu.ethereum.mainnet.MainnetBlockHeaderFunctions;

// import org.hyperledger.besu.evm.log.LogsBloomFilter;
import org.hyperledger.besu.tests.acceptance.dsl.account.Accounts;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Disabled;
import org.junit.jupiter.api.Test;
import org.web3j.crypto.Credentials;
import org.web3j.protocol.Web3j;
import org.web3j.protocol.core.methods.response.EthSendTransaction;
import org.web3j.protocol.core.methods.response.EthBlock.Block;

/**
 * Tests that verify the LineaTransactionValidationPlugin correctly rejects BLOB transactions from
 * being executed
 */
public class BlobTransactionDenialTest extends LineaPluginTestBasePrague {
  private Web3j web3j;
  private Credentials credentials;
  private String recipient;

  @Override
  protected String getGenesisFileTemplatePath() {
    // We cannot use clique-prague-zero-blobs because `config.blobSchedule.prague.max = 0` will
    // block all blob txs
    return "/clique/clique-prague-one-blob.json.tpl";
  }

  @Override
  @BeforeEach
  public void setup() throws Exception {
    super.setup();
    web3j = minerNode.nodeRequests().eth();
    credentials = Credentials.create(Accounts.GENESIS_ACCOUNT_ONE_PRIVATE_KEY);
    recipient = accounts.getSecondaryBenefactor().getAddress();
  }

  @Test
  @Disabled("Disabled for dev workflow")
  public void blobTransactionsIsRejectedFromTransactionPool() throws Exception {

    // Act - Send a blob transaction to transaction pool
    EthSendTransaction response = sendRawBlobTransaction(web3j, credentials, recipient);
    this.buildNewBlock();

    Block newBlock = web3j
            .ethGetBlockByNumber(org.web3j.protocol.core.DefaultBlockParameterName.LATEST, false)
            .send()
            .getBlock();
    System.out.println("newBlock transactionsRoot: " + newBlock.getTransactionsRoot());
    System.out.println("newBlock withdrawalsRoot: " + newBlock.getWithdrawalsRoot());
    System.out.println("newBlock blobGasUsed: " + newBlock.getBlobGasUsed());
    System.out.println("newBlock excessBlobGas: " + newBlock.getExcessBlobGasRaw());
    
    // Assert
    assertThat(response.hasError()).isTrue();
    assertThat(response.getError().getMessage())
        .contains("Plugin has marked the transaction as invalid");
  }

  // Ideally the block import test would be conducted with two nodes as follows:
  // 1. Start an additional minimal node with Prague config
  // 2. Ensure additional node is peered to minerNode
  // 3. Send blob tx to additional node
  // 4. Construct block on additional node
  // 5. Send 'debug_getBadBlocks' RPC request to minerNode, confirm that block is rejected from
  // import
  //
  // However we are unable to run more than one node per test, due to the CLI options being
  // singleton options and this implemented in dependency repository - linea tracer.
  // Thus we are limited to 'simulating' the block import as below:
  // 1. Create a premade block containing a blob tx
  // 2. Import the premade block using 'engine_newPayloadV4' Engine API call

  @Test
  // @Disabled("Disabled for dev workflow")
  public void blobTransactionsIsRejectedFromNodeImport() throws Exception {

    // Arrange
    String genesisBlockHash =
        web3j
            .ethGetBlockByNumber(org.web3j.protocol.core.DefaultBlockParameterName.LATEST, false)
            .send()
            .getBlock()
            .getHash();
    Hash transactionsRoot = Hash.fromHexString("0x7a430a1c9da1f6e25ff8e6e96217c359784f3438dc1d983b4695355d66437f8f");
    Hash withdrawalsRoot = Hash.fromHexString("0x56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421");

    final ObjectMapper mapper = new ObjectMapper();
    // We obtained the below values by running the test `blobTransactionsIsRejectedFromTransactionPool` with the following plugins removed - LineaTransactionPoolValidatorPlugin, LineaTransactionSelectorPlugin, LineaTransactionValidatorPlugin
    ObjectNode executionPayloadWithoutBlockHash =
        mapper
            .createObjectNode()
            .put("parentHash", genesisBlockHash)
            .put("feeRecipient", "0x0000000000000000000000000000000000000000")
            .put("stateRoot", "0x2c1457760c057cf42f2d509648d725ec1f557b9d8729a5361e517952f91d050e")
            .put(
                "logsBloom",
                "0x00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000")
            .put("prevRandao", "0x0000000000000000000000000000000000000000000000000000000000000000")
            .put("gasLimit", "0x1ca35ef")
            .put("gasUsed", "0x5208")
            .put("timestamp", "0x5")
            .put("extraData", "0x626573752032352e362e302d6c696e656131")
            .put("baseFeePerGas", "0x7")
            .put("excessBlobGas", "0x0")
            .put("blobGasUsed", "0x20000");
            final ArrayNode transactions = mapper.createArrayNode();
            transactions.add(
                "0x03f8908205398084f461090084f46109008389544094627306090abab3a6e1400e9345bc60c78a8bef578080c001e1a0018ef96865998238a5e1783b6cafbc1253235d636f15d318f1fb50ef6a5b8f6a80a0576a95756f32ab705a22b591ab464d5affc8c1c7fcd14d777bac24d83bc44821a01f93b26f4f9989c3fe764f4a58d264bcd71b9deab72d6852f5dcdf19d55494f1");
            executionPayloadWithoutBlockHash.set("transactions", transactions);
            final ArrayNode withdrawals = mapper.createArrayNode();
            executionPayloadWithoutBlockHash.set("withdrawals", withdrawals);
            executionPayloadWithoutBlockHash.put("receiptsRoot", "0xeaa8c40899a61ae59615cf9985f5e2194f8fd2b57d273be63bde6733e89b12ab");
            executionPayloadWithoutBlockHash.put("blockHash", "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef");
            executionPayloadWithoutBlockHash.put("blockNumber", "0x1");
            
    final ArrayNode expectedBlobVersionedHashes = mapper.createArrayNode();
    expectedBlobVersionedHashes.add("0x018ef96865998238a5e1783b6cafbc1253235d636f15d318f1fb50ef6a5b8f6a");
    final String parentBeaconBlockRoot =
        "0x0000000000000000000000000000000000000000000000000000000000000000";
    
    String executionRequestString = "0x01a4664c40aacebd82a2db79f0ea36c06bc6a19adbb10a4a15bf67b328c9b101d09e5c6ee6672978fdad9ef0d9e2ceffaee99223555d8601f0cb3bcc4ce1af9864779a416e0000000000000000";
        final ArrayNode executionRequests = mapper.createArrayNode();
    executionRequests.add(executionRequestString);
    final Bytes executionRequestBytes = Bytes.fromHexString(executionRequestString);
    final Bytes executionRequestBytesData = executionRequestBytes.slice(1);
    final Request executionRequest = new Request(RequestType.of(executionRequestBytes.get(0)), executionRequestBytesData);
    Optional<List<Request>> maybeRequests = Optional.of(List.of(executionRequest));
        
    // We must compute the blockhash manually, as the genesis block hash changes each run
    String executionPayloadWithoutBlockHashStringified = executionPayloadWithoutBlockHash.toString();
    EnginePayloadParameter blockParam = mapper.readValue(executionPayloadWithoutBlockHashStringified, EnginePayloadParameter.class);

    // From AbstractEngineNewPayload.syncResponse method in Besu
    final BlockHeader blockHeader = new BlockHeader(
      blockParam.getParentHash(),
      Hash.ZERO,
      blockParam.getFeeRecipient(),
      blockParam.getStateRoot(),
      transactionsRoot,
      blockParam.getReceiptsRoot(),
      blockParam.getLogsBloom(),
      Difficulty.ZERO,
      blockParam.getBlockNumber(),
      blockParam.getGasLimit(),
      blockParam.getGasUsed(),
      blockParam.getTimestamp(),
      Bytes.fromHexString(blockParam.getExtraData()),
      blockParam.getBaseFeePerGas(),
      blockParam.getPrevRandao(),
      0,
      withdrawalsRoot,
      blockParam.getBlobGasUsed(),
      BlobGas.fromHexString(blockParam.getExcessBlobGas()),
      Bytes32.fromHexString(parentBeaconBlockRoot),
      maybeRequests.map(BodyValidation::requestsHash).orElse(null),
      new MainnetBlockHeaderFunctions()
    );
    executionPayloadWithoutBlockHash.set("transactions", transactions);

    // Create new executionPayload with computed block hash
    ObjectNode executionPayload = mapper.createObjectNode();
    executionPayloadWithoutBlockHash.fields().forEachRemaining(entry -> {
        executionPayload.set(entry.getKey(), entry.getValue());
    });
    executionPayload.put("blockHash", blockHeader.getBlockHash().toHexString());
    executionPayload.put("blockNumber", "0x1");

    // Act
    // this.importPremadeBlock(executionPayload, expectedBlobVersionedHashes, parentBeaconBlockRoot, executionRequests);

    EthSendTransaction response = sendRawBlobTransaction(web3j, credentials, recipient);
    this.buildNewBlock();

    Block nextBlock =
    web3j
        .ethGetBlockByNumber(org.web3j.protocol.core.DefaultBlockParameterName.LATEST, false)
        .send()
        .getBlock();
    
    // Print all properties of the Block object
    System.out.println("Block Properties:");
    System.out.println("Parent Hash: " + nextBlock.getParentHash());

    System.out.println("Miner: " + nextBlock.getMiner());
    System.out.println("StateRoot: " + nextBlock.getStateRoot());
    System.out.println("TransactionsRoot: " + nextBlock.getTransactionsRoot());
    System.out.println("ReceiptsRoot: " + nextBlock.getReceiptsRoot());
    System.out.println("LogsBloom: " + nextBlock.getLogsBloom());
    System.out.println("Difficulty: " + nextBlock.getDifficulty());
    System.out.println("Number: " + nextBlock.getNumber());
    System.out.println("GasLimit: " + nextBlock.getGasLimit());
    System.out.println("GasUsed: " + nextBlock.getGasUsed());
    System.out.println("Timestamp: " + nextBlock.getTimestamp());
    System.out.println("ExtraData: " + nextBlock.getExtraData());
    System.out.println("BaseFeePerGas: " + nextBlock.getBaseFeePerGas());
    System.out.println("MixHash: " + nextBlock.getMixHash());
    System.out.println("Nonce: " + nextBlock.getNonce());
    System.out.println("WithdrawalsRoot: " + nextBlock.getWithdrawalsRoot());
    System.out.println("BlobGasUsed: " + nextBlock.getBlobGasUsed());
    System.out.println("ExcessBlobGas: " + nextBlock.getExcessBlobGas());
    System.out.println("ParentBeaconBlockRoot: " + nextBlock.getParentBeaconBlockRoot());

    System.out.println("-------");
    System.out.println("computed blockHeader: " + blockHeader.toString());
    // System.out.println("Actual nextBlock: " + nextBlock.toString());

    assertThat(false).isTrue();

    // Assert
    // TODO: Verify the response indicates rejection of blob transactions
  }
}
