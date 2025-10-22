package net.consensys.linea.sequencer.txselection.selectors;

import java.util.Set;
import java.util.concurrent.atomic.AtomicReference;
import lombok.RequiredArgsConstructor;
import net.consensys.linea.bundles.TransactionBundle;
import org.hyperledger.besu.evm.log.Log;
import org.hyperledger.besu.plugin.data.TransactionProcessingResult;
import org.hyperledger.besu.plugin.data.TransactionSelectionResult;
import org.hyperledger.besu.plugin.services.txselection.TransactionEvaluationContext;

@RequiredArgsConstructor
public class TransactionEventPostProcessingSelector
    extends AbstractTransactionPostProcessingSelector {
  private final AtomicReference<Set<TransactionEventSelectionDescription>> deniedEvents;
  private final AtomicReference<Set<TransactionEventSelectionDescription>> deniedBundleEvents;

  @Override
  public TransactionSelectionResult evaluateTransactionPostProcessing(
      final TransactionEvaluationContext evaluationContext,
      final TransactionProcessingResult processingResult) {
    final boolean isBundle =
        evaluationContext.getPendingTransaction() instanceof TransactionBundle.PendingBundleTx;
    Set<TransactionEventSelectionDescription> deniedEventsForTransaction =
        isBundle ? deniedBundleEvents.get() : deniedEvents.get();
    for (TransactionEventSelectionDescription deniedEvent : deniedEventsForTransaction) {
      for (Log log : processingResult.getLogs()) {
        if (deniedEvent.matches(log)) {
          return TransactionSelectionResult.invalid("Transaction event logs match deny list");
        }
      }
    }
    return TransactionSelectionResult.SELECTED;
  }
}
