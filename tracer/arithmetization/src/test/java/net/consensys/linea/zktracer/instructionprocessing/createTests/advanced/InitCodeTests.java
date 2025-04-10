/*
 * Copyright Consensys Software Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except in compliance with
 * the License. You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 *
 * SPDX-License-Identifier: Apache-2.0
 */
package net.consensys.linea.zktracer.instructionprocessing.createTests.advanced;

import static com.google.common.base.Preconditions.checkArgument;
import static net.consensys.linea.zktracer.instructionprocessing.utilities.MonoOpCodeSmcs.keyPair;
import static net.consensys.linea.zktracer.instructionprocessing.utilities.MonoOpCodeSmcs.userAccount;
import static org.junit.jupiter.api.Assertions.assertEquals;

import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

import net.consensys.linea.UnitTestWatcher;
import net.consensys.linea.testing.*;
import net.consensys.linea.testing.ToyTransaction.ToyTransactionBuilder;
import net.consensys.linea.testing.generated.ContractC;
import net.consensys.linea.testing.generated.CustomCreate2;
import net.consensys.linea.zktracer.instructionprocessing.utilities.SmartContractTestValidator;
import org.apache.tuweni.bytes.Bytes;
import org.hyperledger.besu.crypto.Hash;
import org.hyperledger.besu.datatypes.Address;
import org.hyperledger.besu.datatypes.Wei;
import org.hyperledger.besu.ethereum.core.Transaction;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.web3j.abi.EventEncoder;

@ExtendWith(UnitTestWatcher.class)
public class InitCodeTests {

  // This suite aims at testing advanced CREATE2 scenarii using smart contracts
  // ** CustomCreate2 **
  // CustomCreate2 is the smart contract used to pilot the deployment of a complex initCodeC
  // CustomCreate2 stores an initCodeC and a salt used for subsequent deployments
  // CustomCreate2 has different create2 methods to have deployment scenarii within the same
  // transaction
  // CustomCreate2 can CALL/STATICCALL itself and contractC
  // ** initCodeC **
  // initCodeC can be piloted by CustomCreate2 via the value passed in the transaction
  // value 1 Wei enables a storage modification in ContractC
  // value 2 Wei calls CustomCreate2 back to trigger an immediate redeployment of ContractC
  // value 3 Wei triggers a self-destruct of ContractC
  // value 4 Wei triggers a revert on demand
  // ** ContractC **
  // ContractC can modify storage, revert on demand, self-destruct on demand, and call back
  // CustomCreate2 to trigger a redeployment of ContractC

  static final Wei defaultBalance = Wei.of(4500L);

  // Deployment code from testing/src/main/solidity/ContractC.sol
  // Generated for the London EVM in testing/build/resources/main/solidity/ContractC.json
  // Value =
  // 6080604052600034905060003390506001820361006c578060008084815260200190815260200160002060006101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff160217905550610196565b60028203610161578073ffffffffffffffffffffffffffffffffffffffff16630161e6d76040518163ffffffff1660e01b8152600401600060405180830381600087803b1580156100bc57600080fd5b505af19250505080156100cd575060015b61015c576100d96101cd565b806308c379a0036100fa57506100ed61026a565b806100f857506100fc565b005b505b3d8060008114610128576040519150601f19603f3d011682016040523d82523d6000602084013e61012d565b606091505b507f95a42732c1514e697b8c2992b6ee317a9bc3c4568e341cdbd58ff87fc4c467cd60405160405180910390a1005b610195565b6003820361017c5761017761019d60201b60201c565b610194565b60048203610193576101926101bb60201b60201c565b5b5b5b5b50506102fa565b60003090508073ffffffffffffffffffffffffffffffffffffffff16ff5b600080fd5b60008160e01c9050919050565b600060033d11156101ec5760046000803e6101e96000516101c0565b90505b90565b6000604051905090565b6000601f19601f8301169050919050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052604160045260246000fd5b610242826101f9565b810181811067ffffffffffffffff821117156102615761026061020a565b5b80604052505050565b600060443d106102f75761027c6101ef565b60043d036004823e80513d602482011167ffffffffffffffff821117156102a45750506102f7565b808201805167ffffffffffffffff8111156102c257505050506102f7565b80602083010160043d0385018111156102df5750505050506102f7565b6102ee82602001850186610239565b82955050505050505b90565b610379806103096000396000f3fe608060405234801561001057600080fd5b50600436106100575760003560e01c80632199ecd71461005c578063545b079a146100665780635e666e4a146100825780636c0c412e146100b2578063983d1ce2146100ce575b600080fd5b6100646100d8565b005b610080600480360381019061007b9190610249565b6100f6565b005b61009c600480360381019061009791906102ac565b610159565b6040516100a991906102e8565b60405180910390f35b6100cc60048036038101906100c79190610303565b61018c565b005b6100d66101e1565b005b60003090508073ffffffffffffffffffffffffffffffffffffffff16ff5b8073ffffffffffffffffffffffffffffffffffffffff16630161e6d76040518163ffffffff1660e01b8152600401600060405180830381600087803b15801561013e57600080fd5b505af1158015610152573d6000803e3d6000fd5b5050505050565b60006020528060005260406000206000915054906101000a900473ffffffffffffffffffffffffffffffffffffffff1681565b8060008084815260200190815260200160002060006101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff1602179055505050565b600080fd5b600080fd5b600073ffffffffffffffffffffffffffffffffffffffff82169050919050565b6000610216826101eb565b9050919050565b6102268161020b565b811461023157600080fd5b50565b6000813590506102438161021d565b92915050565b60006020828403121561025f5761025e6101e6565b5b600061026d84828501610234565b91505092915050565b6000819050919050565b61028981610276565b811461029457600080fd5b50565b6000813590506102a681610280565b92915050565b6000602082840312156102c2576102c16101e6565b5b60006102d084828501610297565b91505092915050565b6102e28161020b565b82525050565b60006020820190506102fd60008301846102d9565b92915050565b6000806040838503121561031a576103196101e6565b5b600061032885828601610297565b925050602061033985828601610234565b915050925092905056fea26469706673582212203989edc877fe659634a18fccfea5c11d5bb1f6f616f6cacad1172e8af1c1e83064736f6c634300081a0033
  static final String initCodeC =
      SmartContractUtils.getSolidityContractCompiledByteCode(ContractC.class).toString();
  static final String salt = "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef";
  static final String buildCustomCreate2Address = "789101";

  ToyAccount customCreate2Account =
      ToyAccount.builder()
          .address(Address.fromHexString("0x" + buildCustomCreate2Address))
          .balance(defaultBalance)
          .nonce(1)
          .code(SmartContractUtils.getSolidityContractRuntimeByteCode(CustomCreate2.class))
          .build();

  static final Long gasLimit = 5000000L;

  @Test
  void deployContractCWithCreate2() {

    // Payloads preparation
    Bytes storeInitCodeC = CustomCreate2Payload.storeInitCodeC(initCodeC);
    Bytes storeSalt = CustomCreate2Payload.storeSalt(salt);
    Bytes create2WithInitCodeC = CustomCreate2Payload.create2WithInitCodeC();
    Bytes callContractCStoreInMapPayload =
        CustomCreate2Payload.callContractC(
            ContractCPayload.storeInMap(1, "0x0000000000000000000000000000000000001234"), false);
    Bytes callContractCSelfDestructPayload =
        CustomCreate2Payload.callContractC(ContractCPayload.selfDestructOnDemand(), false);
    Bytes create2WithCallBackAfterCreate2 = CustomCreate2Payload.create2WithCallBackAfterCreate2();
    Bytes create2CallCAndRevert = CustomCreate2Payload.create2CallCAndRevert();
    Bytes create2WithStaticCall =
        CustomCreate2Payload.callMyself(CustomCreate2Payload.create2WithInitCodeC(), true);
    Bytes create2FourTimes = CustomCreate2Payload.create2FourTimes();

    // Compute expected address for ContractC with Create2
    // address = keccak256(0xff + sender_address + salt + keccak256(initialisation_code))[12:]
    String senderAddNoOx = customCreate2Account.getAddress().toHexString().substring(2);
    String saltNoOx = salt.substring(2);
    String hashInitCodeCNo0x =
        Hash.keccak256(Bytes.fromHexString(initCodeC)).toHexString().substring(2);
    Bytes expectedContractCAddress =
        Hash.keccak256(Bytes.fromHexString("0xff" + senderAddNoOx + saltNoOx + hashInitCodeCNo0x))
            .slice(12);
    // padding left to fit the 32 bytes log data format
    Bytes expectedContractCAddressLogData =
        Bytes.fromHexString(
            "0x000000000000000000000000" + expectedContractCAddress.toString().substring(2));

    // Logs for transaction validator
    String contractCreatedEvent = EventEncoder.encode(CustomCreate2.CONTRACTCREATED_EVENT);
    String staticCallMyselfFailEvent =
        EventEncoder.encode(CustomCreate2.STATICCALLMYSELFFAIL_EVENT);
    String calledCreate2WithInitCodeCEvent =
        EventEncoder.encode(CustomCreate2.CALLEDCREATE2WITHINITCODEC_EVENT);
    Map<String, List<Integer>> logsTopicMap = new HashMap<>();
    Map<String, List<Bytes>> logsDataMap = new HashMap<>();

    // List all logs expected from each topic
    logsTopicMap.put(contractCreatedEvent, List.of(0, 0, 1, 0, 0, 0, 0, 0, 1));
    logsTopicMap.put(staticCallMyselfFailEvent, List.of(0, 0, 0, 0, 0, 0, 0, 1, 0));
    logsTopicMap.put(calledCreate2WithInitCodeCEvent, List.of(0, 0, 1, 0, 0, 0, 0, 0, 0));
    // List data expected for each topic
    logsDataMap.put(
        contractCreatedEvent,
        List.of(
            Bytes.EMPTY,
            Bytes.EMPTY,
            expectedContractCAddressLogData,
            Bytes.EMPTY,
            Bytes.EMPTY,
            Bytes.EMPTY,
            Bytes.EMPTY,
            Bytes.EMPTY,
            expectedContractCAddressLogData));
    // List status expected per transaction
    // 0 is FAILED
    // 1 is SUCCESSFUL
    List<Integer> txStatuses = List.of(1, 1, 1, 1, 1, 0, 0, 1, 1);

    // Instantiate validator
    TransactionProcessingResultValidator create2Validator =
        new SmartContractTestValidator(txStatuses, logsTopicMap, logsDataMap);

    List<Transaction> transactions =
        getTransactions(
            customCreate2Account,
            userAccount,
            // The list below contains `Bytes` payloads
            // Every item in that list will define a unique transaction having it as its payload.
            List.of(
                // CustomCreate2 initialization
                storeInitCodeC,
                storeSalt,
                // SCENARIO 1 - Deploy ContractC through CustomCreate2, modify storage,
                // self-destruct ContractC
                // ContractC is deployed
                // ContractC Storage is modified
                // ContractC is self-destructed
                // TXSTATUS : Successfull
                // LOGS: 1 CalledCreate2WithInitCodeC + 1 ContractCreated
                // Note : 1 CREATE2 opcode called
                create2WithInitCodeC,
                callContractCStoreInMapPayload,
                callContractCSelfDestructPayload,
                // SCENARIO 2 - Deploy ContractC and attempt redeployment after in same transaction
                // Transaction reverts, nothing is deployed
                // TXSTATUS : Failed
                // LOGS: no logs
                // Note : 2 CREATE2 opcode called
                create2WithCallBackAfterCreate2,
                // SCENARIO 3 - Deploy ContractC and the deployment attempts redeployment
                // The ContractC deployment is done with value 2 - this value pilots the initcode so
                // immediate redeployment is attempted
                // While deploying ContractC adds STOP opcode after immediate redeployment attempt
                // has failed
                // ContractC is deployed with empty bytecode
                // Call ContractC to modify storage
                // Revert on demand
                // Transaction reverts
                // TXSTATUS : Failed
                // LOGS: no logs
                // Note : 2 CREATE2 opcode called
                create2CallCAndRevert,
                // SCENARIO 4 - Attempt ContractC deployment with a staticCall
                // TXSTATUS : Successfull
                // LOGS: 1 StaticCallMyselfFail
                // Note 1 : no CalledCreate2WithInitCodeCEvent as attempt fails prior
                // Note : 1 CREATE2 opcode called
                create2WithStaticCall,
                // SCENARIO 5 - Four ContractC deployment attempts : (1) with max value, (2)
                // acceptable value, (3) max value and
                // (4) acceptable value
                // (1) is aborted
                // (2) results in deployment of ContractC
                // (3) is aborted
                // (4) fails as it's a collision with attempt (2)
                // TXSTATUS : Successfull
                // LOGS: 1 ContractCreated
                // Note : 4 CREATE2 opcode called
                create2FourTimes),
            // Values to pilot initCode : as many as there are transactions
            List.of(0L, 0L, 0L, 0L, 0L, 0L, 2L, 0L, 0L));

    final ToyExecutionEnvironmentV2 toyExecutionEnvironmentV2 =
        ToyExecutionEnvironmentV2.builder()
            .accounts(List.of(userAccount, customCreate2Account))
            .transactions(transactions)
            .transactionProcessingResultValidator(create2Validator)
            .build();
    toyExecutionEnvironmentV2.run();

    // Final check on the deployment number of ContractC
    // At start, deploymentNumber = 0
    // transaction 3 - create2WithInitCodeC : deploymentNumber ++
    // - Deploys contract C with non empty code
    // - deploymentNumber = 1
    // transaction 5 - callContractCSelfDestructPayload : deploymentNumber ++
    // - Self-destruct successful increments deployment number
    // - deploymentNumber = 2
    // transaction 6 - create2WithCallBackAfterCreate2 : deploymentNumber ++
    // - First create2 is successful so increments the deployment number, second create2 makes the
    // whole transaction revert
    // - deploymentNumber = 3
    // transaction 7 - create2CallCAndRevert : deploymentNumber ++
    // - Create2 deploys contractC with empty bytecode (thus the deployment number increment) and
    // reverts
    // - deploymentNumber = 4
    // transaction 9 - create2FourTimes : deploymentNumber ++
    // - Only one create2 is successful so increments the deployment number by 1
    // - deploymentNumber = 5
    int deploymentNumber =
        toyExecutionEnvironmentV2
            .getHub()
            .transients()
            .conflation()
            .deploymentInfo()
            .deploymentNumber(Address.fromHexString(expectedContractCAddress.toString()));
    int expectedDeploymentNumber = 5;
    assertEquals(expectedDeploymentNumber, deploymentNumber);
  }

  /// /////////////////////////////////////////////////////////////////////////////////////////////
  /// Common helpers
  /// Create transactions with payloads/values for the same user and the same to account
  List<Transaction> getTransactions(
      ToyAccount to, ToyAccount userAccount, List<Bytes> payloads, List<Long> values) {

    checkArgument(payloads.size() == values.size());
    final List<ToyTransactionBuilder> builders = new ArrayList<>();

    for (int i = 0; i < payloads.size(); i++) {
      ToyTransactionBuilder builder =
          ToyTransaction.builder()
              .to(to)
              .payload(payloads.get(i))
              .keyPair(keyPair)
              .gasLimit(gasLimit)
              .value(Wei.of(values.get(i)));
      builders.add(builder);
    }
    return ToyMultiTransaction.builder().build(builders, userAccount);
  }
}
