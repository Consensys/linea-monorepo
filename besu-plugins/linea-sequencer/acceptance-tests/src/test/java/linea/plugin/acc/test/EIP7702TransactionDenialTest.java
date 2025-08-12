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
import static org.junit.jupiter.api.Assertions.assertThrows;

import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.databind.node.ArrayNode;
import com.fasterxml.jackson.databind.node.ObjectNode;
import java.util.Collections;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.Optional;
import okhttp3.Response;
import org.apache.tuweni.bytes.Bytes;
import org.apache.tuweni.bytes.Bytes32;
import org.hyperledger.besu.datatypes.Address;
import org.hyperledger.besu.datatypes.BlobGas;
import org.hyperledger.besu.datatypes.Hash;
import org.hyperledger.besu.datatypes.RequestType;
import org.hyperledger.besu.ethereum.api.jsonrpc.internal.parameters.EnginePayloadParameter;
import org.hyperledger.besu.ethereum.core.BlockHeader;
import org.hyperledger.besu.ethereum.core.Difficulty;
import org.hyperledger.besu.ethereum.core.Request;
import org.hyperledger.besu.ethereum.mainnet.BodyValidation;
import org.hyperledger.besu.ethereum.mainnet.MainnetBlockHeaderFunctions;
import org.hyperledger.besu.tests.acceptance.dsl.account.Accounts;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.web3j.crypto.Credentials;
import org.web3j.protocol.Web3j;

/**
 * Tests that verify the LineaTransactionValidationPlugin correctly rejects BLOB transactions from
 * being executed
 */
public class EIP7702TransactionDenialTest extends LineaPluginTestBasePrague {
  private Web3j web3j;
  private Credentials credentials;
  private String recipient;

  @Override
  @BeforeEach
  public void setup() throws Exception {
    super.setup();
    web3j = minerNode.nodeRequests().eth();
    credentials = Credentials.create(Accounts.GENESIS_ACCOUNT_ONE_PRIVATE_KEY);
    recipient = accounts.getSecondaryBenefactor().getAddress();
  }

  @Test
  public void eip7702TransactionIsRejectedFromTransactionPool() throws Exception {
    // Act - Send a blob transaction to transaction pool and expect it to be rejected
    // We use 'minerNode.execute' here which throw us a RuntimeException directly
    RuntimeException exception =
        assertThrows(
            RuntimeException.class,
            () -> {
              sendRawEIP7702Transaction(web3j, credentials, recipient);
            });
    // No need to build new block.

    // Assert
    assertThat(exception.getMessage()).contains("Plugin has marked the transaction as invalid");
  }

  // Ideally the block import test would be conducted with two nodes as follows:
  // 1. Start an additional minimal node with Prague config
  // 2. Ensure additional node is peered to minerNode
  // 3. Send EIP7702 tx to additional node
  // 4. Construct block on additional node
  // 5. Send 'debug_getBadBlocks' RPC request to minerNode, confirm that block is rejected from
  // import
  //
  // However we are currently unable to run more than one node per test, due to the CLI options
  // being
  // singleton options and this implemented in dependency repository - linea tracer.
  // Thus simulate the block import as below:
  // 1. Create a premade block containing a EIP7702 tx
  // 2. Import the premade block using 'engine_newPayloadV4' Engine API call

  @Test
  // @Disabled
  public void EIP7702TransactionIsRejectedFromNodeImport() throws Exception {
    // Arrange
    EngineNewPayloadRequest blockWithEIP7702TxRequest = getBlockWithEIP7702TxRequest(mapper);

    // Act
    Response response =
        this.importPremadeBlock(
            blockWithEIP7702TxRequest.executionPayload(),
            blockWithEIP7702TxRequest.expectedBlobVersionedHashes(),
            blockWithEIP7702TxRequest.parentBeaconBlockRoot(),
            blockWithEIP7702TxRequest.executionRequests());

    // Assert
    JsonNode result = mapper.readTree(response.body().string()).get("result");
    String status = result.get("status").asText();
    String validationError = result.get("validationError").asText();
    assertThat(status).isEqualTo("INVALID");
    assertThat(validationError)
        .contains("LineaTransactionValidatorPlugin - DELEGATE_CODE_TX_NOT_ALLOWED");
  }

  private record EngineNewPayloadRequest(
      ObjectNode executionPayload,
      ArrayNode expectedBlobVersionedHashes,
      String parentBeaconBlockRoot,
      ArrayNode executionRequests) {}

  private EngineNewPayloadRequest getBlockWithEIP7702TxRequest(ObjectMapper mapper)
      throws Exception {
    // Obtained following values by running `blobTransactionsIsRejectedFromTransactionPool` test
    // without the LineaTransactionSelectorPlugin and LineaTransactionValidatorPlugin plugins.
    Map<String, String> blockWithBlockTxParams = new HashMap<>();
    blockWithBlockTxParams.put(
        "STATE_ROOT", "0x217cc246352b4a22254ab139cc4a5a37e1dbe75b63fcf12161674773f5043bbe");
    blockWithBlockTxParams.put(
        "LOGS_BLOOM",
        "0x00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000");
    blockWithBlockTxParams.put(
        "RECEIPTS_ROOT", "0x036c7d20420edbce24b5062148bce80563b48fd4532fb1c068b2c96a53117019");
    blockWithBlockTxParams.put("EXTRA_DATA", "0x626573752032352e372e302d6c696e656134");
    blockWithBlockTxParams.put(
        "DELEGATE_CALL_TX",
        "0x04f8cd8205398084f461090084f46109008389544094fe3b557e8fb62b89f4916b721be55ceb828dbd738080c0f85ef85c82053994627306090abab3a6e1400e9345bc60c78a8bef570101a0972498bc9ef3b18ec9f16f3dd59e7b622cc07fce1459d7485f424658e4013aa6a038dfeeaaa952cf3eb3fb81ccac2b57a69f57b457d1350353ac80d11ddc5dfeb180a0713d685d1b0fd47e7e7e75d9d8aaf2fa0d4a8811aec37fa54cb0ca4deb632dcaa01a66a49c9bbd92f11a0e67fb096a6e862a5da8bf05af32f6390ee431011fba81");
    blockWithBlockTxParams.put(
        "EXECUTION_REQUEST",
        "0x01a4664c40aacebd82a2db79f0ea36c06bc6a19adbb10a4a15bf67b328c9b101d09e5c6ee6672978fdad9ef0d9e2ceffaee99223555d8601f0cb3bcc4ce1af9864779a416e0000000000000000");
    blockWithBlockTxParams.put(
        "TRANSACTIONS_ROOT", "0xb84030d9aae336c44f284a9710bc8f6771a38d0bccdeb1d837f871bacd1d07c9");
    blockWithBlockTxParams.put(
        "WITHDRAWALS_ROOT", "0x56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421");
    blockWithBlockTxParams.put("GAS_LIMIT", "0x1ca35ef");
    blockWithBlockTxParams.put("GAS_USED", "0x8fc0");
    blockWithBlockTxParams.put("TIMESTAMP", "0x5");
    blockWithBlockTxParams.put("BASE_FEE_PER_GAS", "0x7");
    blockWithBlockTxParams.put("EXCESS_BLOB_GAS", "0x0");
    blockWithBlockTxParams.put("BLOB_GAS_USED", "0x0");
    blockWithBlockTxParams.put("BLOCK_NUMBER", "0x1");
    blockWithBlockTxParams.put("FEE_RECIPIENT", Address.ZERO.toHexString());
    blockWithBlockTxParams.put("PREV_RANDAO", Hash.ZERO.toHexString());
    blockWithBlockTxParams.put("PARENT_BEACON_BLOCK_ROOT", Hash.ZERO.toHexString());
    blockWithBlockTxParams = Collections.unmodifiableMap(blockWithBlockTxParams);
    // Seems that the genesis block hash change with each run, despite a constant genesis file
    String genesisBlockHash = getLatestBlockHash();

    ObjectNode executionPayload =
        createExecutionPayload(mapper, genesisBlockHash, blockWithBlockTxParams);
    ArrayNode executionRequests = createExecutionRequests(mapper, blockWithBlockTxParams);
    // Compute block hash and update payload
    BlockHeader blockHeader = computeBlockHeader(executionPayload, mapper, blockWithBlockTxParams);
    updateExecutionPayloadWithBlockHash(executionPayload, blockHeader);
    return new EngineNewPayloadRequest(
        executionPayload,
        mapper.createArrayNode(),
        blockWithBlockTxParams.get("PARENT_BEACON_BLOCK_ROOT"),
        executionRequests);
  }

  private String getLatestBlockHash() throws Exception {
    return web3j
        .ethGetBlockByNumber(org.web3j.protocol.core.DefaultBlockParameterName.LATEST, false)
        .send()
        .getBlock()
        .getHash();
  }

  private ObjectNode createExecutionPayload(
      ObjectMapper mapper, String genesisBlockHash, Map<String, String> blockParams) {
    ObjectNode payload =
        mapper
            .createObjectNode()
            .put("parentHash", genesisBlockHash)
            .put("feeRecipient", blockParams.get("FEE_RECIPIENT"))
            .put("stateRoot", blockParams.get("STATE_ROOT"))
            .put("logsBloom", blockParams.get("LOGS_BLOOM"))
            .put("prevRandao", blockParams.get("PREV_RANDAO"))
            .put("gasLimit", blockParams.get("GAS_LIMIT"))
            .put("gasUsed", blockParams.get("GAS_USED"))
            .put("timestamp", blockParams.get("TIMESTAMP"))
            .put("extraData", blockParams.get("EXTRA_DATA"))
            .put("baseFeePerGas", blockParams.get("BASE_FEE_PER_GAS"))
            .put("excessBlobGas", blockParams.get("EXCESS_BLOB_GAS"))
            .put("blobGasUsed", blockParams.get("BLOB_GAS_USED"))
            .put("receiptsRoot", blockParams.get("RECEIPTS_ROOT"))
            .put("blockNumber", blockParams.get("BLOCK_NUMBER"));

    // Add transactions
    ArrayNode transactions = mapper.createArrayNode();
    transactions.add(blockParams.get("DELEGATE_CALL_TX"));
    payload.set("transactions", transactions);
    // Add withdrawals (empty list)
    ArrayNode withdrawals = mapper.createArrayNode();
    payload.set("withdrawals", withdrawals);

    return payload;
  }

  private ArrayNode createExecutionRequests(ObjectMapper mapper, Map<String, String> blockParams) {
    ArrayNode requests = mapper.createArrayNode();
    requests.add(blockParams.get("EXECUTION_REQUEST"));
    return requests;
  }

  private BlockHeader computeBlockHeader(
      ObjectNode executionPayload, ObjectMapper mapper, Map<String, String> blockParams)
      throws Exception {
    EnginePayloadParameter blockParam =
        mapper.readValue(executionPayload.toString(), EnginePayloadParameter.class);

    Hash transactionsRoot = Hash.fromHexString(blockParams.get("TRANSACTIONS_ROOT"));
    Hash withdrawalsRoot = Hash.fromHexString(blockParams.get("WITHDRAWALS_ROOT"));

    // Take code from AbstractEngineNewPayload in Besu codebase
    Bytes executionRequestBytes = Bytes.fromHexString(blockParams.get("EXECUTION_REQUEST"));
    Bytes executionRequestBytesData = executionRequestBytes.slice(1);
    Request executionRequest =
        new Request(RequestType.of(executionRequestBytes.get(0)), executionRequestBytesData);
    Optional<List<Request>> maybeRequests = Optional.of(List.of(executionRequest));

    return new BlockHeader(
        blockParam.getParentHash(),
        Hash.EMPTY_LIST_HASH, // OMMERS_HASH_CONSTANT
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
        0, // Nonce
        withdrawalsRoot,
        blockParam.getBlobGasUsed(),
        BlobGas.fromHexString(blockParam.getExcessBlobGas()),
        Bytes32.fromHexString(blockParams.get("PARENT_BEACON_BLOCK_ROOT")),
        maybeRequests.map(BodyValidation::requestsHash).orElse(null),
        new MainnetBlockHeaderFunctions());
  }

  private void updateExecutionPayloadWithBlockHash(
      ObjectNode executionPayload, BlockHeader blockHeader) {
    executionPayload.put("blockHash", blockHeader.getBlockHash().toHexString());
  }
}
