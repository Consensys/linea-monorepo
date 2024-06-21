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

package net.consensys.linea.extradata;

import java.util.function.Consumer;
import java.util.function.Function;

import lombok.extern.slf4j.Slf4j;
import net.consensys.linea.config.LineaProfitabilityConfiguration;
import org.apache.commons.lang3.mutable.MutableLong;
import org.apache.tuweni.bytes.Bytes;
import org.apache.tuweni.units.bigints.UInt32;
import org.hyperledger.besu.datatypes.Wei;
import org.hyperledger.besu.datatypes.rpc.JsonRpcResponseType;
import org.hyperledger.besu.plugin.services.RpcEndpointService;

@Slf4j
public class LineaExtraDataHandler {
  private final RpcEndpointService rpcEndpointService;
  private final ExtraDataConsumer[] extraDataConsumers;

  public LineaExtraDataHandler(
      final RpcEndpointService rpcEndpointService,
      final LineaProfitabilityConfiguration profitabilityConf) {
    this.rpcEndpointService = rpcEndpointService;
    this.extraDataConsumers = new ExtraDataConsumer[] {new Version1Consumer(profitabilityConf)};
  }

  public void handle(final Bytes rawExtraData) throws LineaExtraDataException {

    if (!Bytes.EMPTY.equals(rawExtraData)) {
      for (final ExtraDataConsumer extraDataConsumer : extraDataConsumers) {
        if (extraDataConsumer.canConsume(rawExtraData)) {
          // strip first byte since it is the version already used to select the actual consumer
          final var extraData = rawExtraData.slice(1);
          extraDataConsumer.accept(extraData);
          return;
        }
      }
      throw new LineaExtraDataException(
          LineaExtraDataException.ErrorType.INVALID_ARGUMENT,
          "Unsupported extra data field " + rawExtraData.toHexString());
    }
  }

  private interface ExtraDataConsumer extends Consumer<Bytes> {
    boolean canConsume(Bytes extraData);

    static Long toLong(final Bytes fieldBytes) {
      return UInt32.fromBytes(fieldBytes).toLong();
    }
  }

  @SuppressWarnings("rawtypes")
  private class Version1Consumer implements ExtraDataConsumer {
    private static final int WEI_IN_KWEI = 1_000;
    private final LineaProfitabilityConfiguration profitabilityConf;
    private final FieldConsumer[] fieldsSequence;
    private final MutableLong currFixedCostKWei = new MutableLong();
    private final MutableLong currVariableCostKWei = new MutableLong();

    public Version1Consumer(final LineaProfitabilityConfiguration profitabilityConf) {
      this.profitabilityConf = profitabilityConf;

      final FieldConsumer fixedGasCostField =
          new FieldConsumer<>(
              "fixedGasCost", 4, ExtraDataConsumer::toLong, currFixedCostKWei::setValue);
      final FieldConsumer variableGasCostField =
          new FieldConsumer<>(
              "variableGasCost", 4, ExtraDataConsumer::toLong, currVariableCostKWei::setValue);
      final FieldConsumer minGasPriceField =
          new FieldConsumer<>("minGasPrice", 4, ExtraDataConsumer::toLong, this::updateMinGasPrice);

      this.fieldsSequence =
          new FieldConsumer[] {fixedGasCostField, variableGasCostField, minGasPriceField};
    }

    public boolean canConsume(final Bytes rawExtraData) {
      return rawExtraData.get(0) == (byte) 1;
    }

    public synchronized void accept(final Bytes extraData) {
      log.info("Parsing extra data version 1: {}", extraData.toHexString());
      int startIndex = 0;
      for (final FieldConsumer fieldConsumer : fieldsSequence) {
        fieldConsumer.accept(extraData.slice(startIndex, fieldConsumer.length));
        startIndex += fieldConsumer.length;
      }

      profitabilityConf.updateFixedAndVariableCost(
          currFixedCostKWei.longValue() * WEI_IN_KWEI,
          currVariableCostKWei.longValue() * WEI_IN_KWEI);
    }

    void updateMinGasPrice(final Long minGasPriceKWei) {
      final var minGasPriceWei = Wei.of(minGasPriceKWei).multiply(WEI_IN_KWEI);
      final var resp =
          rpcEndpointService.call(
              "miner_setMinGasPrice", new Object[] {minGasPriceWei.toShortHexString()});
      if (!resp.getType().equals(JsonRpcResponseType.SUCCESS)) {
        throw new LineaExtraDataException(
            LineaExtraDataException.ErrorType.FAILED_CALLING_SET_MIN_GAS_PRICE,
            "Internal setMinGasPrice method failed: " + resp);
      }
    }
  }

  private record FieldConsumer<T>(
      String name, int length, Function<Bytes, T> converter, Consumer<T> consumer)
      implements Consumer<Bytes> {

    @Override
    public void accept(final Bytes fieldBytes) {
      final var converted = converter.apply(fieldBytes);
      log.debug("Field {}={} (raw bytes: {})", name, converted, fieldBytes.toHexString());
      consumer.accept(converted);
    }
  }
}
