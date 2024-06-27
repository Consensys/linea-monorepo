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

package net.consensys.linea.zktracer.module.ecdata;

import static net.consensys.linea.zktracer.module.ecdata.EcDataOperation.SECP256K1N;
import static org.junit.jupiter.api.Assertions.assertEquals;

import java.math.BigInteger;
import java.util.ArrayList;
import java.util.List;
import java.util.stream.Stream;

import net.consensys.linea.zktracer.opcode.OpCode;
import net.consensys.linea.zktracer.testing.BytecodeCompiler;
import net.consensys.linea.zktracer.testing.BytecodeRunner;
import net.consensys.linea.zktracer.testing.EvmExtension;
import net.consensys.linea.zktracer.types.EWord;
import org.apache.tuweni.bytes.Bytes;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.junit.jupiter.params.ParameterizedTest;
import org.junit.jupiter.params.provider.Arguments;
import org.junit.jupiter.params.provider.MethodSource;

@ExtendWith(EvmExtension.class)
public class EcDataTest {
  @Test
  void testEcData() {
    BytecodeRunner.of(
            Bytes.fromHexString(
                "608060405234801561001057600080fd5b5061004a6001601b6001620f00007ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffe610b9760201b60201c565b61005757610056610e05565b5b61006f60016019600060016000610b9760201b60201c565b61007c5761007b610e05565b5b6100936001601e6001806000610b9760201b60201c565b6100a05761009f610e05565b5b6100d76001601b60017fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff6000610b9760201b60201c565b6100e4576100e3610e05565b5b610152600160027f30644e72e131a029b85045b68181585d97816a916871ca8d3c208c16d87cfd476101169190610e6d565b600160027f30644e72e131a029b85045b68181585d97816a916871ca8d3c208c16d87cfd476101459190610e6d565b6003610bd260201b60201c565b61015f5761015e610e05565b5b610176600080600160026000610bd260201b60201c565b61018357610182610e05565b5b6101c5600060018060027f30644e72e131a029b85045b68181585d97816a916871ca8d3c208c16d87cfd476101b89190610e6d565b6000610bd260201b60201c565b156101d3576101d2610e05565b5b6102146000806001807f30644e72e131a029b85045b68181585d97816a916871ca8d3c208c16d87cfd476102079190610ea1565b6000610bd260201b60201c565b1561022257610221610e05565b5b610283600160027f30644e72e131a029b85045b68181585d97816a916871ca8d3c208c16d87cfd476102549190610e6d565b610f007fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff610c0c60201b60201c565b6102905761028f610e05565b5b6102d0600160027f30644e72e131a029b85045b68181585d97816a916871ca8d3c208c16d87cfd476102c29190610e6d565b600080610c0c60201b60201c565b6102dd576102dc610e05565b5b6102f260008060036000610c0c60201b60201c565b6102ff576102fe610e05565b5b6103156000600460036000610c0c60201b60201c565b1561032357610322610e05565b5b61033860006004600080610c0c60201b60201c565b1561034657610345610e05565b5b60006040518060c001604052806001815260200160027f30644e72e131a029b85045b68181585d97816a916871ca8d3c208c16d87cfd476103879190610e6d565b81526020016000815260200160008152602001600081526020016000815250905060006040518060c001604052806001815260200160027f30644e72e131a029b85045b68181585d97816a916871ca8d3c208c16d87cfd476103e99190610e6d565b81526020017f1800deef121f1e76426a00665e5c4479674322d4f75edadd46debd5cd992f6ed81526020017f198e9393920d483a7260bfb731fb5d25f1aa493335a9e71297e485b7aef312c281526020017f12c85ea5db8c6deb4aab71808dcb408fe3d1e7690c43d37b4ce6cc0166fa7daa81526020017f090689d0585ff075ec9e99ad690c3395bc4b313370b38ef355acdadcd122975b815250905060006040518060c0016040528060008152602001600081526020017f1800deef121f1e76426a00665e5c4479674322d4f75edadd46debd5cd992f6ed81526020017f198e9393920d483a7260bfb731fb5d25f1aa493335a9e71297e485b7aef312c281526020017f12c85ea5db8c6deb4aab71808dcb408fe3d1e7690c43d37b4ce6cc0166fa7daa81526020017f090689d0585ff075ec9e99ad690c3395bc4b313370b38ef355acdadcd122975b815250905060006040518060c001604052806000815260200160008152602001600181526020017f198e9393920d483a7260bfb731fb5d25f1aa493335a9e71297e485b7aef312c281526020017f12c85ea5db8c6deb4aab71808dcb408fe3d1e7690c43d37b4ce6cc0166fa7daa81526020017f090689d0585ff075ec9e99ad690c3395bc4b313370b38ef355acdadcd122975b815250905060006040518060c0016040528060008152602001600c81526020017f1800deef121f1e76426a00665e5c4479674322d4f75edadd46debd5cd992f6ed81526020017f198e9393920d483a7260bfb731fb5d25f1aa493335a9e71297e485b7aef312c281526020017f12c85ea5db8c6deb4aab71808dcb408fe3d1e7690c43d37b4ce6cc0166fa7daa81526020017f090689d0585ff075ec9e99ad690c3395bc4b313370b38ef355acdadcd122975b815250905060008067ffffffffffffffff81111561069c5761069b610ed5565b5b6040519080825280602002602001820160405280156106d557816020015b6106c2610dcf565b8152602001906001900390816106ba5790505b5090506106e9816000610c4060201b60201c565b6106f6576106f5610e05565b5b600167ffffffffffffffff81111561071157610710610ed5565b5b60405190808252806020026020018201604052801561074a57816020015b610737610dcf565b81526020019060019003908161072f5790505b509050858160008151811061076257610761610f04565b5b602002602001018190525061077e816000610c4060201b60201c565b61078b5761078a610e05565b5b85816000815181106107a05761079f610f04565b5b6020026020010181905250600267ffffffffffffffff8111156107c6576107c5610ed5565b5b6040519080825280602002602001820160405280156107ff57816020015b6107ec610dcf565b8152602001906001900390816107e45790505b509050858160008151811061081757610816610f04565b5b6020026020010181905250848160018151811061083757610836610f04565b5b6020026020010181905250610853816000610c4060201b60201c565b6108605761085f610e05565b5b61087181600a610c4060201b60201c565b1561087f5761087e610e05565b5b828160018151811061089457610893610f04565b5b60200260200101819052506108b0816000610c4060201b60201c565b156108be576108bd610e05565b5b600367ffffffffffffffff8111156108d9576108d8610ed5565b5b60405190808252806020026020018201604052801561091257816020015b6108ff610dcf565b8152602001906001900390816108f75790505b509050858160008151811061092a57610929610f04565b5b6020026020010181905250848160018151811061094a57610949610f04565b5b6020026020010181905250838160028151811061096a57610969610f04565b5b6020026020010181905250610986816000610c4060201b60201c565b61099357610992610e05565b5b6109a4816001610c4060201b60201c565b156109b2576109b1610e05565b5b6109e2817fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff610c4060201b60201c565b156109f0576109ef610e05565b5b8181600281518110610a0557610a04610f04565b5b6020026020010181905250610a21816000610c4060201b60201c565b15610a2f57610a2e610e05565b5b600a67ffffffffffffffff811115610a4a57610a49610ed5565b5b604051908082528060200260200182016040528015610a8357816020015b610a70610dcf565b815260200190600190039081610a685790505b50905060005b8151811015610b36576000600382610aa19190610f62565b03610aca5786828281518110610aba57610ab9610f04565b5b6020026020010181905250610b23565b6001600382610ad99190610f62565b03610b025785828281518110610af257610af1610f04565b5b6020026020010181905250610b22565b84828281518110610b1657610b15610f04565b5b60200260200101819052505b5b8080610b2e90610f93565b915050610a89565b50610b48816000610c4060201b60201c565b610b5557610b54610e05565b5b60006040518060400160405280600e81526020017f7a6b2d65766d206973206c6966650000000000000000000000000000000000008152509050805160208201f35b600060405186815285602082015284604082015283606082015260008084608001836001610bb8fa9150608081016040525095945050505050565b6000604051868152856020820152846040820152836060820152600080846080018360066096fa9150608081016040525095945050505050565b600060405185815284602082015283604082015260008084606001836007611770fa91506060810160405250949350505050565b60008061afc884516184d0610c559190610fdb565b610c5f9190610ea1565b90506000604051905060005b8551811015610d91576000868281518110610c8957610c88610f04565b5b60200260200101516000015190506000878381518110610cac57610cab610f04565b5b60200260200101516020015190506000888481518110610ccf57610cce610f04565b5b60200260200101516040015190506000898581518110610cf257610cf1610f04565b5b602002602001015160600151905060008a8681518110610d1557610d14610f04565b5b602002602001015160800151905060008b8781518110610d3857610d37610f04565b5b602002602001015160a0015190508660c00286818a015285602082018a015283604082018a015284606082018a015281608082018a01528260a082018a0152505050505050508080610d8990610f93565b915050610c6b565b50600060c08651610da29190610fdb565b905060008183610db29190610ea1565b905060008087840185600888fa9450806040525050505092915050565b6040518060c001604052806000815260200160008152602001600081526020016000815260200160008152602001600081525090565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052600160045260246000fd5b6000819050919050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052601160045260246000fd5b6000610e7882610e34565b9150610e8383610e34565b9250828203905081811115610e9b57610e9a610e3e565b5b92915050565b6000610eac82610e34565b9150610eb783610e34565b9250828201905080821115610ecf57610ece610e3e565b5b92915050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052604160045260246000fd5b7f4e487b7100000000000000000000000000000000000000000000000000000000600052603260045260246000fd5b7f4e487b7100000000000000000000000000000000000000000000000000000000600052601260045260246000fd5b6000610f6d82610e34565b9150610f7883610e34565b925082610f8857610f87610f33565b5b828206905092915050565b6000610f9e82610e34565b91507fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff8203610fd057610fcf610e3e565b5b600182019050919050565b6000610fe682610e34565b9150610ff183610e34565b9250828202610fff81610e34565b9150828204841483151761101657611015610e3e565b5b509291505056fe"))
        .run();
  }

  @Test
  void testEcRecoverWithEmptyExt() {
    BytecodeRunner.of(
            Bytes.fromHexString(
                "6080604052348015600f57600080fd5b5060476001601b6001620f00007ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffe609360201b60201c565b605157605060ce565b5b60006040518060400160405280600e81526020017f7a6b2d65766d206973206c6966650000000000000000000000000000000000008152509050805160208201f35b600060405186815285602082015284604082015283606082015260008084608001836001610bb8fa9150608081016040525095945050505050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052600160045260246000fdfe"))
        .run();
  }

  private static Stream<Arguments> ecRecoverSource() {
    EWord h =
        EWord.ofHexString("0x456e9aea5e197a1f1af7a3e85a3212fa4049a3ba34c2289b4c860fc0b0c64ef3");
    List<EWord> v =
        List.of(
            EWord.of(28),
            EWord.ZERO,
            EWord.of(BigInteger.ONE, BigInteger.valueOf(27)),
            EWord.of(BigInteger.ONE, BigInteger.valueOf(28)));
    List<EWord> r =
        List.of(
            EWord.ofHexString("0x9242685bf161793cc25603c231bc2f568eb630ea16aa137d2664ac8038825608"),
            EWord.ZERO,
            SECP256K1N,
            SECP256K1N.add(EWord.of(1)));
    List<EWord> s =
        List.of(
            EWord.ofHexString("0x4f8ae3bd7535248d0bd448298cc2e2071e56992d0774dc340c368ae950852ada"),
            EWord.ZERO,
            SECP256K1N,
            SECP256K1N.add(EWord.of(1)));

    List<Arguments> arguments = new ArrayList<>();

    // Test cases where ICP = successBit = 1 (first one) or ICP = successBit = 0 (all the others)
    for (int i = 0; i < v.size(); i++) {
      for (int j = 0; j < r.size(); j++) {
        for (int k = 0; k < s.size(); k++) {
          arguments.add(
              Arguments.of(
                  i + j + k == 0 ? "[ICP = 1, successBit = 1]" : "[ICP = 0, successBit = 0]",
                  h,
                  v.get(i),
                  r.get(j),
                  s.get(k),
                  i + j + k == 0,
                  i + j + k == 0));
        }
      }
    }

    // Test cases where ICP = successBit = 1
    arguments.add(
        argumentsFromStrings(
            "[ICP = 1, successBit = 1]",
            "0x279d94621558f755796898fc4bd36b6d407cae77537865afe523b79c74cc680b",
            "0x1b",
            "0xc2ff96feed8749a5ad1c0714f950b5ac939d8acedbedcbc2949614ab8af06312",
            "0x1feecd50adc6273fdd5d11c6da18c8cfe14e2787f5a90af7c7c1328e7d0a2c42",
            true,
            true));

    arguments.add(
        argumentsFromStrings(
            "[ICP = 1, successBit = 1]",
            "0x4be146e06cc1b37342b6b7b1fa8542ae58a62103b8af0f7d58f8a1ffffcf7914",
            "0x1b",
            "0xa7b0f504b652b3a621921c78c587fdf80a3ab590e22c304b0b0930e90c4e081d",
            "0x5428459ef7e6bd079fbbb7c6fd95cc6c7fe68c93ed4ae75cee36810e79e8a0e5",
            true,
            true));

    arguments.add(
        argumentsFromStrings(
            "[ICP = 1, successBit = 1]",
            "0xca3e75570aea0e3dd8e7a9d38c2efa866f5ee2b18bf527a0f4e3248b7c7cf376",
            "0x1c",
            "0xf1136900c2cd16eacc676f2c7b70f3dfec13fd16a426aab4eda5d8047c30a9e9",
            "0x4dad8f009ebe31bdc38133bc5fa60e9dca59d0366bd90e2ef12b465982c696aa",
            true,
            true));

    arguments.add(
        argumentsFromStrings(
            "[ICP = 1, successBit = 1]",
            "0x9a3fa82837622a34408888b40af937f21f4e6d051f04326d3d7717848c434448",
            "0x1b",
            "0x52a734f01d14d161795ba3b38ce329eba468e109b4f2e330af671649ffef4e0e",
            "0xe3e2a22b830edf61554ab6c18c7efb9e37e1953c913784db0ef74e1e07c227d3",
            true,
            true));

    // Test cases where ICP = 1 but successBit = 0
    // Failing reason QNR
    arguments.add(
        argumentsFromStrings(
            "[ICP = 1, successBit = 0 due to QNR]",
            "0x94f66d57fb0a3854a44d94956447e01f8b3f09845860f18856e792c821359162",
            "0x1c",
            "0x000000000000000000000000000000014551231950b75fc4402da1722fc9baed",
            "0x44c819d6b971e456562fefc2408536bdfd9567ee1c6c7dd2a7076625953a1859",
            true,
            false));

    // Failing reason INFINITY
    arguments.add(
        argumentsFromStrings(
            "[ICP = 1, successBit = 0 due to INFINITY]",
            "0xd33cfae367f4f7413985ff82dc7db3ffbf7a027fb5dad7097b4a15cc85ab6580",
            "0x1c",
            "0xa12b54d413c4ffaaecd59468de6a7d414d2fa7f2ba700d8e0753ca226410c806",
            "0xe9956ef412dceeda0016fe0edfc4746452a8f4d02f21e28cfa6019ee1a8976e8",
            true,
            false));

    // Failing reason INFINITY
    arguments.add(
        argumentsFromStrings(
            "[ICP = 1, successBit = 0 due to INFINITY]",
            "0x6ec17edf5cecd83ed50c08adfeba8146f69769231f4b7903eba38c2e7e98e173",
            "0x1b",
            "0xaeb8ffe3655e07edd6bde0ab79edd92d4e7a155385c3d8c8ca117bfd13633516",
            "0x4da31701c798fe3078ee9de6e4d892242e235dc078df76b15a9ad82137c6250e",
            true,
            false));

    return arguments.stream();
  }

  private static Arguments argumentsFromStrings(
      String description,
      String h,
      String v,
      String r,
      String s,
      boolean expectedInternalChecksPassed,
      boolean expectedSuccessBit) {
    return Arguments.of(description, h, v, r, s, expectedInternalChecksPassed, expectedSuccessBit);
  }

  @ParameterizedTest
  @MethodSource("ecRecoverSource")
  void testEcRecover(
      String description,
      EWord h,
      EWord v,
      EWord r,
      EWord s,
      boolean expectedInternalChecksPassed,
      boolean expectedSuccessBit) {
    BytecodeCompiler program =
        BytecodeCompiler.newProgram()
            // First place the parameters in memory
            .push(h)
            .push(0)
            .op(OpCode.MSTORE)
            .push(v) // v
            .push(0x20)
            .op(OpCode.MSTORE)
            .push(r) // r
            .push(0x40)
            .op(OpCode.MSTORE)
            .push(s) // s
            .push(0x60)
            .op(OpCode.MSTORE)
            // Do the call
            .push(32) // retSize
            .push(0x80) // retOffset
            .push(0x80) // argSize
            .push(0) // argOffset
            .push(1) // address
            .push(Bytes.fromHexStringLenient("0xFFFFFFFF")) // gas
            .op(OpCode.STATICCALL);

    BytecodeRunner bytecodeRunner = BytecodeRunner.of(program.compile());
    bytecodeRunner.run();

    // Retrieve recoveredAddress, internalChecksPassed, successBit
    // Assert internalChecksPassed and successBit are what expected
    EcDataOperation ecDataOperation =
        bytecodeRunner.getHub().ecData().getOperations().stream().toList().get(0);
    EWord recoveredAddress =
        EWord.of(
            ecDataOperation.limb().get(8).toUnsignedBigInteger(),
            ecDataOperation.limb().get(9).toUnsignedBigInteger());
    boolean internalChecksPassed = ecDataOperation.internalChecksPassed();
    boolean successBit = ecDataOperation.successBit();

    assertEquals(internalChecksPassed, expectedInternalChecksPassed);
    assertEquals(successBit, expectedSuccessBit);

    System.out.println("recoveredAddress: " + recoveredAddress);
    System.out.println("internalChecksPassed: " + internalChecksPassed);
    System.out.println("successBit: " + successBit);
  }

  // TODO: continue testing the other precompiles in a more structured way
  @Test
  void testEcAddGeneric() {
    // TODO: The same inputs in failingMmuModexp return 0x, debug it!
    BytecodeCompiler program =
        BytecodeCompiler.newProgram()
            // First place the parameters in memory
            .push("070375d4eec4f22aa3ad39cb40ccd73d2dbab6de316e75f81dc2948a996795d5") // pX
            .push(0)
            .op(OpCode.MSTORE)
            .push("041b98f07f44aa55ce8bd97e32cacf55f1e42229d540d5e7a767d1138a5da656") // pY
            .push(0x20)
            .op(OpCode.MSTORE)
            .push("185f6f5cf93c8afa0461a948c2da7c403b6f8477c488155dfa8d2da1c62517b8") // qX
            .push(0x40)
            .op(OpCode.MSTORE)
            .push("13d83d7a51eb18fdb51225873c87d44f883e770ce2ca56c305d02d6cb99ca5b8") // qY
            .push(0x60)
            .op(OpCode.MSTORE)
            // Do the call
            .push(0x40) // retSize
            .push(0x80) // retOffset
            .push(0x80) // argSize
            .push(0) // argOffset
            .push(6) // address
            .push(Bytes.fromHexStringLenient("0xFFFFFFFF")) // gas
            .op(OpCode.STATICCALL);

    BytecodeRunner bytecodeRunner = BytecodeRunner.of(program.compile());
    bytecodeRunner.run();
  }

  @Test
  void testEcAddWithPointAtInfinityAsResult() {
    BytecodeCompiler program =
        BytecodeCompiler.newProgram()
            // First place the parameters in memory
            .push(1) // pX
            .push(0)
            .op(OpCode.MSTORE)
            .push(2) // pY
            .push(0x20)
            .op(OpCode.MSTORE)
            .push(1) // qX
            .push(0x40)
            .op(OpCode.MSTORE)
            .push("30644e72e131a029b85045b68181585d97816a916871ca8d3c208c16d87cfd45") // qY
            .push(0x60)
            .op(OpCode.MSTORE)
            // Do the call
            .push(0x40) // retSize
            .push(0x80) // retOffset
            .push(0x80) // argSize
            .push(0) // argOffset
            .push(6) // address
            .push(Bytes.fromHexStringLenient("0xFFFFFFFF")) // gas
            .op(OpCode.STATICCALL);

    BytecodeRunner bytecodeRunner = BytecodeRunner.of(program.compile());
    bytecodeRunner.run();
  }

  @Test
  void testEcPairingWithTrivialPairing() {
    BytecodeCompiler program =
        BytecodeCompiler.newProgram()
            .push(0x20) // retSize
            .push(0) // retOffset
            .push(192) // argSize
            .push(0) // argOffset
            .push(8) // address
            .push(Bytes.fromHexStringLenient("0xFFFFFFFF")) // gas
            .op(OpCode.STATICCALL);

    BytecodeRunner bytecodeRunner = BytecodeRunner.of(program.compile());
    bytecodeRunner.run();
  }
}
