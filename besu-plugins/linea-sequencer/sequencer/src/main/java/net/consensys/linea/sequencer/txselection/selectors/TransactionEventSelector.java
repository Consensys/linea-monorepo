package net.consensys.linea.sequencer.txselection.selectors;

import java.util.Map;
import java.util.Set;
import java.util.concurrent.atomic.AtomicReference;
import lombok.RequiredArgsConstructor;
import net.consensys.linea.bundles.TransactionBundle;
import org.hyperledger.besu.datatypes.Address;
import org.hyperledger.besu.evm.log.Log;
import org.hyperledger.besu.plugin.data.TransactionProcessingResult;
import org.hyperledger.besu.plugin.data.TransactionSelectionResult;
import org.hyperledger.besu.plugin.services.txselection.PluginTransactionSelector;
import org.hyperledger.besu.plugin.services.txselection.TransactionEvaluationContext;

@RequiredArgsConstructor
public class TransactionEventSelector implements PluginTransactionSelector {
  private final AtomicReference<Map<Address, Set<TransactionEventFilter>>> deniedEvents;
  private final AtomicReference<Map<Address, Set<TransactionEventFilter>>> deniedBundleEvents;

  @Override
  public TransactionSelectionResult evaluateTransactionPreProcessing(
      final TransactionEvaluationContext evaluationContext) {
    return TransactionSelectionResult.SELECTED;
  }

  @Override
  public TransactionSelectionResult evaluateTransactionPostProcessing(
      final TransactionEvaluationContext evaluationContext,
      final TransactionProcessingResult processingResult) {
    final boolean isBundle =
        evaluationContext.getPendingTransaction() instanceof TransactionBundle.PendingBundleTx;
    final Map<Address, Set<TransactionEventFilter>> deniedEventsByAddress =
        isBundle ? deniedBundleEvents.get() : deniedEvents.get();
    for (Log log : processingResult.getLogs()) {
      Set<TransactionEventFilter> deniedEventsForTransaction =
          deniedEventsByAddress.get(log.getLogger());
      if (deniedEventsForTransaction != null) {
        for (TransactionEventFilter deniedEvent : deniedEventsForTransaction) {
          if (deniedEvent.matches(log)) {
            return TransactionSelectionResult.invalid(
                String.format(
                    "Transaction %s is blocked due to contract address and event logs appearing on SDN or other legally prohibited list",
                    evaluationContext
                        .getPendingTransaction()
                        .getTransaction()
                        .getHash()
                        .toShortHexString()));
          }
        }
      }
    }
    return TransactionSelectionResult.SELECTED;
  }
}
