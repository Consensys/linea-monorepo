package net.consensys.linea.zktracer.instructionprocessing.callTests;

import net.consensys.linea.reporting.TracerTestBase;
import net.consensys.linea.testing.ToyAccount;
import net.consensys.linea.testing.ToyExecutionEnvironmentV2;
import net.consensys.linea.testing.ToyTransaction;
import net.consensys.linea.testing.TransactionProcessingResultValidator;
import net.consensys.linea.zktracer.types.Bytecode;
import org.apache.tuweni.bytes.Bytes;
import org.hyperledger.besu.crypto.KeyPair;
import org.hyperledger.besu.crypto.SECP256K1;
import org.hyperledger.besu.datatypes.Address;
import org.hyperledger.besu.datatypes.Hash;
import org.hyperledger.besu.datatypes.Wei;
import org.hyperledger.besu.ethereum.core.Transaction;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.TestInfo;

import java.util.List;

public class CallToDelegated extends TracerTestBase {
  @Test
  void targetIsSmc(TestInfo testInfo) {

    KeyPair keyPair = new SECP256K1().generateKeyPair();
    Address senderAddress = Address.extract(Hash.hash(keyPair.getPublicKey().getEncodedBytes()));
    ToyAccount senderAccount =
      ToyAccount.builder().balance(Wei.of(1_550_000_000_000L)).nonce(23).address(senderAddress).build();

    // smc byte code
    // designed to test environment variables after call to delegated account
    // ADDRESS
    // SELFBALANCE
    // CODESIZE
    // CALLER
    // CALLVALUE
    // GAS

    String smcBytecode = "0x30473833345a";
    ToyAccount smcAccount =
      ToyAccount.builder()
        .address(Address.fromHexString("aaaaa")) // identity caller
        .nonce(99)
        .balance(Wei.of(1_000_000L))
        .code(
          Bytes.concatenate(Bytes.fromHexString(smcBytecode)))
        .build();


    String delegationCode = "0xef0100" + smcAccount.getAddress().toHexString().substring(2);

    ToyAccount eoaDelegatedToSmc =
      ToyAccount.builder()
        .address(Address.fromHexString("ca11ee")) // identity caller
        .nonce(99)
        .balance(Wei.of(1_000_000L))
        .code(
          Bytes.concatenate(Bytes.fromHexString(delegationCode)))
        .build();

    Transaction tx = ToyTransaction.builder()
        .sender(senderAccount)
        .keyPair(keyPair)
        .to(eoaDelegatedToSmc).build();

    ToyExecutionEnvironmentV2.builder(chainConfig, testInfo)
      .accounts(List.of(senderAccount, eoaDelegatedToSmc, smcAccount))
      .transaction(tx)
      .transactionProcessingResultValidator(TransactionProcessingResultValidator.EMPTY_VALIDATOR)
      .build()
      .run();
  }


}
