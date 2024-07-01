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

package net.consensys.linea.zktracer.module.hub.precompiles;

import lombok.Builder;
import lombok.experimental.Accessors;
import net.consensys.linea.zktracer.module.hub.Hub;
import net.consensys.linea.zktracer.module.limits.precompiles.Blake2fRounds;
import net.consensys.linea.zktracer.module.limits.precompiles.EcAddEffectiveCall;
import net.consensys.linea.zktracer.module.limits.precompiles.EcMulEffectiveCall;
import net.consensys.linea.zktracer.module.limits.precompiles.EcPairingEffectiveCall;
import net.consensys.linea.zktracer.module.limits.precompiles.EcRecoverEffectiveCall;
import net.consensys.linea.zktracer.module.limits.precompiles.ModexpEffectiveCall;
import net.consensys.linea.zktracer.module.limits.precompiles.RipemdBlocks;
import net.consensys.linea.zktracer.module.limits.precompiles.Sha256Blocks;
import net.consensys.linea.zktracer.types.MemorySpan;
import net.consensys.linea.zktracer.types.Precompile;
import org.hyperledger.besu.datatypes.Address;
import org.hyperledger.besu.evm.internal.Words;

@Accessors(fluent = true)
@Builder
public record PrecompileInvocation(
    /* The precompile being called */
    Precompile precompile,
    /*
     * If applicable, some data related to the precompile required later. Only used for Blake2f for
     * now.
     */
    PrecompileMetadata metadata,
    /* The input data for the precompile */
    MemorySpan callDataSource,
    /* Where the caller wants the precompile return data to be stored */
    MemorySpan requestedReturnDataTarget,
    boolean hubFailure,
    boolean ramFailure,
    /* The price of the *CALL itself */
    long opCodeGas,
    /* The intrinsic cost of the precompile */
    long precompilePrice,
    /* The available gas just before the *CALL opcode execution */
    long gasAtCall,
    /* If applicable, the gas given to a precompile */
    long gasAllowance,
    /* The amount of gas to be given back to the caller */
    long returnGas,
    /* The HubStamp at the time of the call of the precompile*/
    int hubStamp) {

  public boolean success() {
    return !this.hubFailure && !this.ramFailure;
  }

  public boolean hubSuccess() {
    return !this.hubFailure;
  }

  public boolean ramSuccess() {
    return !this.ramFailure;
  }

  public static PrecompileInvocation of(final Hub hub, Precompile p) {
    final boolean hubFailure =
        switch (p) {
          case EC_RECOVER -> !EcRecoverEffectiveCall.hasEnoughGas(hub);
          case SHA2_256 -> !Sha256Blocks.hasEnoughGas(hub);
          case RIPEMD_160 -> !RipemdBlocks.hasEnoughGas(hub);
          case IDENTITY -> switch (hub.opCode()) {
            case CALL, STATICCALL, DELEGATECALL, CALLCODE -> {
              final Address target = Words.toAddress(hub.messageFrame().getStackItem(1));
              if (target.equals(Address.ID)) {
                final long dataByteLength = hub.transients().op().callDataSegment().length();
                final long wordCount = (dataByteLength + 31) / 32;
                final long gasNeeded = 15 + 3 * wordCount;

                yield hub.transients().op().gasAllowanceForCall() < gasNeeded;
              } else {
                throw new IllegalStateException();
              }
            }
            default -> throw new IllegalStateException();
          };
          case MODEXP -> false;
          case EC_ADD -> hub.transients().op().gasAllowanceForCall() < 150;
          case EC_MUL -> hub.transients().op().gasAllowanceForCall() < 6000;
          case EC_PAIRING -> EcPairingEffectiveCall.isHubFailure(hub);
          case BLAKE2F -> Blake2fRounds.isHubFailure(hub);
        };

    final boolean ramFailure =
        !hubFailure
            && switch (p) {
              case EC_RECOVER, IDENTITY, RIPEMD_160, SHA2_256 -> false;
              case MODEXP -> ModexpEffectiveCall.gasCost(hub)
                  > hub.transients().op().gasAllowanceForCall();
              case EC_ADD -> EcAddEffectiveCall.isRamFailure(hub);
              case EC_MUL -> EcMulEffectiveCall.isRamFailure(hub);
              case EC_PAIRING -> EcPairingEffectiveCall.isRamFailure(hub);
              case BLAKE2F -> Blake2fRounds.isRamFailure(hub);
            };

    final long opCodeGas = Hub.GAS_PROJECTOR.of(hub.messageFrame(), hub.opCode()).total();

    final long precompilePrice =
        hubFailure || ramFailure
            ? hub.transients().op().gasAllowanceForCall()
            : switch (p) {
              case EC_RECOVER -> EcRecoverEffectiveCall.gasCost();
              case SHA2_256 -> Sha256Blocks.gasCost(hub);
              case RIPEMD_160 -> RipemdBlocks.gasCost(hub);
              case IDENTITY -> switch (hub.opCode()) {
                case CALL, STATICCALL, DELEGATECALL, CALLCODE -> {
                  final Address target = Words.toAddress(hub.messageFrame().getStackItem(1));
                  if (target.equals(Address.ID)) {
                    final long dataByteLength = hub.transients().op().callDataSegment().length();
                    final long wordCount = (dataByteLength + 31) / 32;
                    yield 15 + 3 * wordCount;
                  } else {
                    throw new IllegalStateException();
                  }
                }
                default -> throw new IllegalStateException();
              };
              case MODEXP -> ModexpEffectiveCall.gasCost(hub);
              case EC_ADD -> EcAddEffectiveCall.gasCost();
              case EC_MUL -> EcMulEffectiveCall.gasCost();
              case EC_PAIRING -> EcPairingEffectiveCall.gasCost(hub);
              case BLAKE2F -> Blake2fRounds.gasCost(hub);
            };

    final long returnGas =
        hubFailure || ramFailure
            ? 0
            : hub.transients().op().gasAllowanceForCall() - precompilePrice;

    PrecompileMetadata metadata =
        switch (p) {
          case EC_RECOVER -> EcRecoverMetadata.of(hub);
          case SHA2_256 -> null;
          case RIPEMD_160 -> null;
          case IDENTITY -> null;
          case MODEXP -> ModExpMetadata.of(hub);
          case EC_ADD -> null;
          case EC_MUL -> null;
          case EC_PAIRING -> null;
          case BLAKE2F -> Blake2fRounds.metadata(hub);
        };

    return PrecompileInvocation.builder()
        .precompile(p)
        .metadata(metadata)
        .callDataSource(hub.transients().op().callDataSegment())
        .requestedReturnDataTarget(hub.transients().op().returnDataRequestedSegment())
        .hubFailure(hubFailure)
        .ramFailure(ramFailure)
        .opCodeGas(opCodeGas)
        .precompilePrice(precompilePrice)
        .gasAtCall(hub.messageFrame().getRemainingGas())
        .gasAllowance(hub.transients().op().gasAllowanceForCall())
        .returnGas(returnGas)
        .hubStamp(hub.stamp())
        .build();
  }
}
