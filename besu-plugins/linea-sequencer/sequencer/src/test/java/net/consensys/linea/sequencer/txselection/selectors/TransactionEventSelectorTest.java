package net.consensys.linea.sequencer.txselection.selectors;

import java.util.Collections;
import java.util.List;
import java.util.Set;
import java.util.concurrent.atomic.AtomicReference;
import net.consensys.linea.bundles.TransactionBundle;
import org.hyperledger.besu.datatypes.PendingTransaction;
import org.hyperledger.besu.evm.log.Log;
import org.hyperledger.besu.plugin.data.TransactionProcessingResult;
import org.hyperledger.besu.plugin.data.TransactionSelectionResult;
import org.hyperledger.besu.plugin.services.txselection.TransactionEvaluationContext;
import org.junit.jupiter.api.AfterEach;
import org.junit.jupiter.api.Assertions;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.mockito.Mockito;

public class TransactionEventSelectorTest {
  private final AtomicReference<Set<TransactionEventSelectionDescription>> deniedEvents =
      new AtomicReference<>(Collections.emptySet());
  private final AtomicReference<Set<TransactionEventSelectionDescription>> deniedBundleEvents =
      new AtomicReference<>(Collections.emptySet());

  private TransactionEventSelector selector;

  @BeforeEach
  public void beforeTest() {
    selector = new TransactionEventSelector(deniedEvents, deniedBundleEvents);
  }

  @AfterEach
  public void afterTest() {
    deniedEvents.set(Collections.emptySet());
    deniedBundleEvents.set(Collections.emptySet());
  }

  @Test
  public void testEvaluateTransactionPostProcessingForSingleTransactionWithEmptyDenyList() {
    TransactionEvaluationContext evaluationContext =
        Mockito.mock(TransactionEvaluationContext.class);
    TransactionProcessingResult processingResult = Mockito.mock(TransactionProcessingResult.class);

    mockTransactionType(false, evaluationContext);

    TransactionSelectionResult actualResult =
        selector.evaluateTransactionPostProcessing(evaluationContext, processingResult);

    Mockito.verify(evaluationContext).getPendingTransaction();
    Mockito.verifyNoMoreInteractions(evaluationContext);
    Mockito.verifyNoInteractions(processingResult);

    Assertions.assertEquals(TransactionSelectionResult.SELECTED, actualResult);
  }

  @Test
  public void testEvaluateTransactionPostProcessingForSingleTransactionWithDenyListButSelected() {
    TransactionEventSelectionDescription transactionEventSelectionDescription =
        Mockito.mock(TransactionEventSelectionDescription.class);
    deniedEvents.set(Set.of(transactionEventSelectionDescription));
    TransactionEvaluationContext evaluationContext =
        Mockito.mock(TransactionEvaluationContext.class);
    TransactionProcessingResult processingResult = Mockito.mock(TransactionProcessingResult.class);

    mockTransactionType(false, evaluationContext);
    Log log = Mockito.mock(Log.class);
    Mockito.when(processingResult.getLogs()).thenReturn(List.of(log));
    Mockito.when(transactionEventSelectionDescription.matches(log)).thenReturn(false);

    TransactionSelectionResult actualResult =
        selector.evaluateTransactionPostProcessing(evaluationContext, processingResult);

    Mockito.verify(evaluationContext).getPendingTransaction();
    Mockito.verify(processingResult).getLogs();
    Mockito.verify(transactionEventSelectionDescription).matches(log);
    Mockito.verifyNoMoreInteractions(
        evaluationContext, processingResult, transactionEventSelectionDescription);

    Assertions.assertEquals(TransactionSelectionResult.SELECTED, actualResult);
  }

  @Test
  public void testEvaluateTransactionPostProcessingForSingleTransactionWithDenyListButInvalid() {
    TransactionEventSelectionDescription transactionEventSelectionDescription =
        Mockito.mock(TransactionEventSelectionDescription.class);
    deniedEvents.set(Set.of(transactionEventSelectionDescription));
    TransactionEvaluationContext evaluationContext =
        Mockito.mock(TransactionEvaluationContext.class);
    TransactionProcessingResult processingResult = Mockito.mock(TransactionProcessingResult.class);

    mockTransactionType(false, evaluationContext);
    Log log = Mockito.mock(Log.class);
    Mockito.when(processingResult.getLogs()).thenReturn(List.of(log));
    Mockito.when(transactionEventSelectionDescription.matches(log)).thenReturn(true);

    TransactionSelectionResult actualResult =
        selector.evaluateTransactionPostProcessing(evaluationContext, processingResult);

    Mockito.verify(evaluationContext).getPendingTransaction();
    Mockito.verify(processingResult).getLogs();
    Mockito.verify(transactionEventSelectionDescription).matches(log);
    Mockito.verifyNoMoreInteractions(
        evaluationContext, processingResult, transactionEventSelectionDescription);

    Assertions.assertEquals(
        TransactionSelectionResult.invalid("Transaction event logs match deny list"), actualResult);
  }

  @Test
  public void testEvaluateTransactionPostProcessingForBundleTransactionWithEmptyDenyList() {
    TransactionEvaluationContext evaluationContext =
        Mockito.mock(TransactionEvaluationContext.class);
    TransactionProcessingResult processingResult = Mockito.mock(TransactionProcessingResult.class);

    mockTransactionType(true, evaluationContext);

    TransactionSelectionResult actualResult =
        selector.evaluateTransactionPostProcessing(evaluationContext, processingResult);

    Mockito.verify(evaluationContext).getPendingTransaction();
    Mockito.verifyNoMoreInteractions(evaluationContext);
    Mockito.verifyNoInteractions(processingResult);

    Assertions.assertEquals(TransactionSelectionResult.SELECTED, actualResult);
  }

  @Test
  public void testEvaluateTransactionPostProcessingForBundleTransactionWithDenyListButSelected() {
    TransactionEventSelectionDescription transactionEventSelectionDescription =
        Mockito.mock(TransactionEventSelectionDescription.class);
    deniedBundleEvents.set(Set.of(transactionEventSelectionDescription));
    TransactionEvaluationContext evaluationContext =
        Mockito.mock(TransactionEvaluationContext.class);
    TransactionProcessingResult processingResult = Mockito.mock(TransactionProcessingResult.class);

    mockTransactionType(true, evaluationContext);
    Log log = Mockito.mock(Log.class);
    Mockito.when(processingResult.getLogs()).thenReturn(List.of(log));
    Mockito.when(transactionEventSelectionDescription.matches(log)).thenReturn(false);

    TransactionSelectionResult actualResult =
        selector.evaluateTransactionPostProcessing(evaluationContext, processingResult);

    Mockito.verify(evaluationContext).getPendingTransaction();
    Mockito.verify(processingResult).getLogs();
    Mockito.verify(transactionEventSelectionDescription).matches(log);
    Mockito.verifyNoMoreInteractions(
        evaluationContext, processingResult, transactionEventSelectionDescription);

    Assertions.assertEquals(TransactionSelectionResult.SELECTED, actualResult);
  }

  @Test
  public void testEvaluateTransactionPostProcessingForBundleTransactionWithDenyListButInvalid() {
    TransactionEventSelectionDescription transactionEventSelectionDescription =
        Mockito.mock(TransactionEventSelectionDescription.class);
    deniedBundleEvents.set(Set.of(transactionEventSelectionDescription));
    TransactionEvaluationContext evaluationContext =
        Mockito.mock(TransactionEvaluationContext.class);
    TransactionProcessingResult processingResult = Mockito.mock(TransactionProcessingResult.class);

    mockTransactionType(true, evaluationContext);
    Log log = Mockito.mock(Log.class);
    Mockito.when(processingResult.getLogs()).thenReturn(List.of(log));
    Mockito.when(transactionEventSelectionDescription.matches(log)).thenReturn(true);

    TransactionSelectionResult actualResult =
        selector.evaluateTransactionPostProcessing(evaluationContext, processingResult);

    Mockito.verify(evaluationContext).getPendingTransaction();
    Mockito.verify(processingResult).getLogs();
    Mockito.verify(transactionEventSelectionDescription).matches(log);
    Mockito.verifyNoMoreInteractions(
        evaluationContext, processingResult, transactionEventSelectionDescription);

    Assertions.assertEquals(
        TransactionSelectionResult.invalid("Transaction event logs match deny list"), actualResult);
  }

  private void mockTransactionType(
      final boolean isBundle, final TransactionEvaluationContext evaluationContext) {
    if (isBundle) {
      Mockito.when(evaluationContext.getPendingTransaction())
          .thenReturn(Mockito.mock(TransactionBundle.PendingBundleTx.class));
    } else {
      Mockito.when(evaluationContext.getPendingTransaction())
          .thenReturn(Mockito.mock(PendingTransaction.class));
    }
  }
}
