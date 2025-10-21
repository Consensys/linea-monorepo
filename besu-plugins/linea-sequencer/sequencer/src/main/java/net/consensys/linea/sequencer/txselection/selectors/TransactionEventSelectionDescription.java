package net.consensys.linea.sequencer.txselection.selectors;

import org.hyperledger.besu.datatypes.Address;
import org.hyperledger.besu.evm.log.Log;
import org.hyperledger.besu.evm.log.LogTopic;

public record TransactionEventSelectionDescription(
    Address contractAddress, LogTopic topic0, LogTopic topic1, LogTopic topic2, LogTopic topic3) {
  /**
   * Checks whether the supplied log matches this TransactionEventSelectionDescription
   *
   * @param log the Log to be checked
   * @return true if the supplied log matches this TransactionEventSelectionDescription, false
   *     otherwise
   */
  public boolean matches(final Log log) {
    return log.getLogger().equals(contractAddress)
        && (topic0 == null || log.getTopics().indexOf(topic0) == 0)
        && (topic1 == null || log.getTopics().indexOf(topic1) == 1)
        && (topic2 == null || log.getTopics().indexOf(topic2) == 2)
        && (topic3 == null || log.getTopics().indexOf(topic3) == 3);
  }
}
