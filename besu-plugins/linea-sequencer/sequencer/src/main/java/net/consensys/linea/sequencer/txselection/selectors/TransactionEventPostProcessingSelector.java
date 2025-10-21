package net.consensys.linea.sequencer.txselection.selectors;

import java.util.Set;
import java.util.concurrent.atomic.AtomicReference;
import lombok.RequiredArgsConstructor;
import org.hyperledger.besu.evm.log.Log;
import org.hyperledger.besu.plugin.data.TransactionProcessingResult;
import org.hyperledger.besu.plugin.data.TransactionSelectionResult;
import org.hyperledger.besu.plugin.services.txselection.TransactionEvaluationContext;

@RequiredArgsConstructor
public class TransactionEventPostProcessingSelector
    extends AbstractTransactionPostProcessingSelector {
  private final AtomicReference<Set<TransactionEventSelectionDescription>> deniedEvents;

  @Override
  public TransactionSelectionResult evaluateTransactionPostProcessing(
      final TransactionEvaluationContext evaluationContext,
      final TransactionProcessingResult processingResult) {
    for (TransactionEventSelectionDescription deniedEvent : deniedEvents.get()) {
      for (Log log : processingResult.getLogs()) {
        if (deniedEvent.matches(log)) {
          return TransactionSelectionResult.invalid("Transaction event logs match deny list");
        }
      }
    }
    return TransactionSelectionResult.SELECTED;
  }
}
