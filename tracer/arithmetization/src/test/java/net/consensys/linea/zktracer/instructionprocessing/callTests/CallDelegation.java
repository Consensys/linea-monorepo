package net.consensys.linea.zktracer.instructionprocessing.callTests;

import net.consensys.linea.reporting.TracerTestBase;
import net.consensys.linea.testing.ToyAccount;
import net.consensys.linea.testing.ToyExecutionEnvironmentV2;
import net.consensys.linea.testing.ToyTransaction;
import net.consensys.linea.testing.TransactionProcessingResultValidator;
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

public class CallDelegation extends TracerTestBase {

  // smc byte code
  // designed to test environment variables after call to delegated account
  // ADDRESS
  // SELFBALANCE
  // CODESIZE
  // CALLER
  // CALLVALUE
  // GAS

  final String smcBytecode = "0x30473833345a";

  final KeyPair keyPair = new SECP256K1().generateKeyPair();
  final Address senderAddress = Address.extract(Hash.hash(keyPair.getPublicKey().getEncodedBytes()));
  final ToyAccount senderAccount =
    ToyAccount.builder().balance(Wei.of(1_550_000_000_000L)).nonce(23).address(senderAddress).build();

  final ToyAccount smcAccount =
    ToyAccount.builder()
      .address(Address.fromHexString("aaaaa")) // identity caller
      .nonce(99)
      .balance(Wei.of(1_000_000L))
      .code(
        Bytes.concatenate(Bytes.fromHexString(smcBytecode)))
      .build();

  @Test
  void EOADelegatedToSmc(TestInfo testInfo) {

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

  @Test
  void EOADelegatedEOADelegatedToSmc(TestInfo testInfo) {

    String delegationCode = "0xef0100" + smcAccount.getAddress().toHexString().substring(2);

    ToyAccount eoaDelegatedToSmc =
      ToyAccount.builder()
        .address(Address.fromHexString("ca11ee2")) // identity caller
        .nonce(99)
        .balance(Wei.of(1_000_000L))
        .code(
          Bytes.concatenate(Bytes.fromHexString(delegationCode)))
        .build();

    String firstDelegationCode = "0xef0100" + eoaDelegatedToSmc.getAddress().toHexString().substring(2);

    ToyAccount eoaDelegatedToEoa =
      ToyAccount.builder()
        .address(Address.fromHexString("ca11ee1")) // identity caller
        .nonce(99)
        .balance(Wei.of(1_000_000L))
        .code(
          Bytes.concatenate(Bytes.fromHexString(firstDelegationCode)))
        .build();

    Transaction tx = ToyTransaction.builder()
      .sender(senderAccount)
      .keyPair(keyPair)
      .to(eoaDelegatedToEoa).build();

    ToyExecutionEnvironmentV2.builder(chainConfig, testInfo)
      .accounts(List.of(senderAccount, eoaDelegatedToSmc, eoaDelegatedToEoa, smcAccount))
      .transaction(tx)
      .transactionProcessingResultValidator(TransactionProcessingResultValidator.EMPTY_VALIDATOR)
      .build()
      .run();
  }

  @Test
  void EOADelegatedToItself(TestInfo testInfo) {

    String delegationCode = "0xef0100" + Address.fromHexString("ca11ee1").toHexString().substring(2);

    ToyAccount eoa =
      ToyAccount.builder()
        .address(Address.fromHexString("ca11ee1")) // identity caller
        .nonce(99)
        .balance(Wei.of(1_000_000L))
        .code(
          Bytes.concatenate(Bytes.fromHexString(delegationCode)))
        .build();

    Transaction tx = ToyTransaction.builder()
      .sender(senderAccount)
      .keyPair(keyPair)
      .to(eoa).build();

    ToyExecutionEnvironmentV2.builder(chainConfig, testInfo)
      .accounts(List.of(senderAccount, eoa, smcAccount))
      .transaction(tx)
      .transactionProcessingResultValidator(TransactionProcessingResultValidator.EMPTY_VALIDATOR)
      .build()
      .run();
  }

  @Test
  void targetIsPrc(TestInfo testInfo) {

    String delegationCode = "0xef0100" + Address.P256_VERIFY.toHexString().substring(2);

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
      .accounts(List.of(senderAccount, eoaDelegatedToSmc))
      .transaction(tx)
      .transactionProcessingResultValidator(TransactionProcessingResultValidator.EMPTY_VALIDATOR)
      .build()
      .run();
  }

}
