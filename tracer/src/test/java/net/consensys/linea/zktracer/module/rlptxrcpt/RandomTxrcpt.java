/// *
// * Copyright ConsenSys AG.
// *
// * Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file
// except in compliance with
// * the License. You may obtain a copy of the License at
// *
// * http://www.apache.org/licenses/LICENSE-2.0
// *
// * Unless required by applicable law or agreed to in writing, software distributed under the
// License is distributed on
// * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See
// the License for the
// * specific language governing permissions and limitations under the License.
// *
// * SPDX-License-Identifier: Apache-2.0
// */
//
// package net.consensys.linea.zktracer.module.rlptxrcpt;
//
// import static org.assertj.core.api.Assertions.assertThat;
//
// import java.util.List;
// import java.util.Optional;
//
// import net.consensys.linea.zktracer.ZkTracer;
// import net.consensys.linea.zktracer.corset.CorsetValidator;
// import org.apache.tuweni.bytes.Bytes;
// import org.hyperledger.besu.datatypes.Address;
// import org.hyperledger.besu.datatypes.TransactionType;
// import org.hyperledger.besu.ethereum.core.TransactionReceipt;
// import org.hyperledger.besu.evm.log.Log;
// import org.hyperledger.besu.evm.log.LogTopic;
// import org.junit.jupiter.api.Test;
//
// public class RandomTxrcpt {
//  @Test
//  public void testRandomTxrcpt() {
//    ZkTracer tracer = new ZkTracer();
//    TransactionReceipt randomTxrcpt =
//        new TransactionReceipt(
//            TransactionType.of(2),
//            0,
//            2123006,
//            List.of(
//                new Log(
//                    Address.wrap(Bytes.random(20)),
//                    Bytes.random(16),
//                    List.of(LogTopic.of(Bytes.random(32)), LogTopic.of(Bytes.random(32))))),
//            Optional.empty());
//    // TransactionReceipt randomTxrcpt = new
//    // TransactionReceipt(TransactionType.FRONTIER,0,21000,List.of(),Optional.empty());
//    tracer.rlpTxrcpt.traceTransaction(randomTxrcpt, TransactionType.FRONTIER);
//    tracer.traceEndConflation();
//    assertThat(CorsetValidator.isValid(tracer.getTrace().toJson())).isTrue();
//  }
// }
