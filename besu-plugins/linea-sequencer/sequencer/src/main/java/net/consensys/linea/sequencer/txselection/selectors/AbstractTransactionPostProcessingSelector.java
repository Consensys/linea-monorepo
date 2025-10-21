package net.consensys.linea.sequencer.txselection.selectors;

import org.hyperledger.besu.plugin.data.TransactionSelectionResult;
import org.hyperledger.besu.plugin.services.txselection.PluginTransactionSelector;
import org.hyperledger.besu.plugin.services.txselection.TransactionEvaluationContext;

public abstract class AbstractTransactionPostProcessingSelector
    implements PluginTransactionSelector {

  @Override
  public TransactionSelectionResult evaluateTransactionPreProcessing(
      final TransactionEvaluationContext evaluationContext) {
    return null;
  }
}
