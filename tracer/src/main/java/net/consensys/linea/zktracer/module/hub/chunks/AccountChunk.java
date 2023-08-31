/*
 * Copyright ConsenSys AG.
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

package net.consensys.linea.zktracer.module.hub.chunks;

import java.math.BigInteger;

import net.consensys.linea.zktracer.EWord;
import net.consensys.linea.zktracer.module.hub.Hub;
import net.consensys.linea.zktracer.module.hub.Trace;
import org.hyperledger.besu.datatypes.Address;
import org.hyperledger.besu.datatypes.Hash;
import org.hyperledger.besu.evm.account.Account;
import org.hyperledger.besu.evm.frame.MessageFrame;

public record AccountChunk(
    MessageFrame frame,
    Address who,
    int deploymentNumber,
    boolean deploymentStatus,
    long oldNonce,
    long newNonce,
    long oldBalance,
    long newBalance,
    Hash oldCodeHash,
    Hash newCodeHash,
    long oldCodeSize,
    long newCodeSize,
    boolean oldWarm,
    boolean newWarm,
    int oldDeploymentNumber,
    int newDeploymentNumber,
    boolean oldDeploymentStatus,
    boolean newDeploymentStatus,
    boolean debit,
    long cost,
    boolean createAddress,
    int deploymentNumberInfnty,
    boolean existsInfinity)
    implements TraceChunk {
  @Override
  public Trace.TraceBuilder trace(Trace.TraceBuilder trace) {
    final Account account = frame.getWorldUpdater().getAccount(who);
    final EWord eWho = EWord.of(who);
    final EWord eCodeHash = EWord.of(oldCodeHash);
    final EWord eCodeHashNew = EWord.of(newCodeHash);

    return trace
        .peekAtAccount(true)
        .pAccountAddressHi(eWho.hiBigInt())
        .pAccountAddressLo(eWho.loBigInt())
        .pAccountIsPrecompile(Hub.isPrecompile(who))
        .pAccountNonce(BigInteger.valueOf(oldNonce))
        .pAccountNonceNew(BigInteger.valueOf(newNonce))
        .pAccountBalance(BigInteger.valueOf(oldBalance))
        .pAccountBalanceNew(BigInteger.valueOf(newBalance))
        .pAccountCodeSize(BigInteger.valueOf(oldCodeSize))
        .pAccountCodeSizeNew(BigInteger.valueOf(newCodeSize))
        .pAccountCodeHashHi(eCodeHash.hiBigInt())
        .pAccountCodeHashLo(eCodeHash.loBigInt())
        .pAccountCodeHashHiNew(eCodeHashNew.hiBigInt())
        .pAccountCodeHashLoNew(eCodeHashNew.loBigInt())
        .pAccountHasCode(oldCodeHash != Hash.EMPTY)
        .pAccountHasCodeNew(newCodeHash != Hash.EMPTY)
        .pAccountExists(oldNonce > 0 || oldCodeHash != Hash.EMPTY || oldBalance > 0)
        .pAccountExistsNew(newNonce > 0 || newCodeHash != Hash.EMPTY || newBalance > 0)
        .pAccountWarm(oldWarm)
        .pAccountWarmNew(newWarm)
        .pAccountDeploymentNumber(BigInteger.valueOf(oldDeploymentNumber))
        .pAccountDeploymentNumberNew(BigInteger.valueOf(newDeploymentNumber))
        .pAccountDeploymentStatus(oldDeploymentStatus ? BigInteger.ONE : BigInteger.ZERO)
        .pAccountDeploymentStatusNew(newDeploymentStatus ? BigInteger.ONE : BigInteger.ZERO)
        //      .pAccountDebit(debit)
        //      .pAccountCost(cost)
        .pAccountSufficientBalance(!debit || cost <= oldBalance)
        //      .pAccountCreateAddress(createAddress)
        .pAccountDeploymentNumberInfty(BigInteger.valueOf(deploymentNumberInfnty))
    //    .pAccountExistsInfty(existsInfinity)
    ;
  }
}
