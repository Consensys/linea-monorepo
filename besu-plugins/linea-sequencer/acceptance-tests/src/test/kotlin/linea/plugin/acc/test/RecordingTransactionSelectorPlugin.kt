/*
 * Copyright Consensys Software Inc.
 *
 * This file is dual-licensed under either the MIT license or Apache License 2.0.
 * See the LICENSE-MIT and LICENSE-APACHE files in the repository root for details.
 *
 * SPDX-License-Identifier: MIT OR Apache-2.0
 */
package linea.plugin.acc.test

import org.hyperledger.besu.datatypes.Hash
import org.hyperledger.besu.plugin.BesuPlugin
import org.hyperledger.besu.plugin.ServiceManager
import org.hyperledger.besu.plugin.data.TransactionProcessingResult
import org.hyperledger.besu.plugin.data.TransactionSelectionResult
import org.hyperledger.besu.plugin.services.TransactionSelectionService
import org.hyperledger.besu.plugin.services.txselection.PluginTransactionSelector
import org.hyperledger.besu.plugin.services.txselection.PluginTransactionSelectorFactory
import org.hyperledger.besu.plugin.services.txselection.SelectorsStateManager
import org.hyperledger.besu.plugin.services.txselection.TransactionEvaluationContext
import java.util.concurrent.ConcurrentHashMap

/**
 * Acceptance-test-only plugin that observes transaction selection outcomes without affecting them.
 *
 * Registers a [PluginTransactionSelector] whose sole purpose is to record which transactions were
 * not selected and why. The recorded rejections can be queried via the companion object so that
 * acceptance tests can assert the exact rejection reason instead of relying on indirect signals
 * such as receipt absence.
 *
 * Besu's [TransactionSelectionService] maintains a list of factories (each registered by a
 * different plugin) and composes all resulting selectors via an internal
 * [AggregatedPluginTransactionSelector]. Because [RecordingTransactionSelector] always returns
 * [TransactionSelectionResult.SELECTED] from its evaluation methods, it does not influence which
 * transactions are chosen; it only observes the final outcome through
 * [PluginTransactionSelector.onTransactionNotSelected].
 */
class RecordingTransactionSelectorPlugin : BesuPlugin {

  private lateinit var transactionSelectionService: TransactionSelectionService

  companion object {
    private val rejections: ConcurrentHashMap<Hash, TransactionSelectionResult> = ConcurrentHashMap()

    /** Clears all recorded rejections. Call this in a @BeforeEach to isolate tests. */
    fun reset() {
      rejections.clear()
    }

    /**
     * Returns the rejection reason for the given transaction hash, or null if the transaction was
     * selected (or not yet evaluated).
     *
     * @param txHash transaction hash
     */
    fun getRejectionReason(txHash: Hash): TransactionSelectionResult? = rejections[txHash]

    /**
     * Returns the rejection reason for the given transaction hash, or null if the transaction was
     * selected (or not yet evaluated).
     *
     * @param txHashHex transaction hash in the standard `0x…` hex form returned by web3j
     */
    fun getRejectionReason(txHashHex: String): TransactionSelectionResult? =
      getRejectionReason(Hash.fromHexString(txHashHex))
  }

  override fun register(serviceManager: ServiceManager) {
    transactionSelectionService =
      serviceManager
        .getService(TransactionSelectionService::class.java)
        .orElseThrow { RuntimeException("TransactionSelectionService not found in ServiceManager") }
  }

  override fun start() {
    transactionSelectionService.registerPluginTransactionSelectorFactory(
      object : PluginTransactionSelectorFactory {
        override fun create(stateManager: SelectorsStateManager): PluginTransactionSelector =
          RecordingTransactionSelector()
      },
    )
  }

  override fun stop() {
    rejections.clear()
  }

  private class RecordingTransactionSelector : PluginTransactionSelector {

    override fun evaluateTransactionPreProcessing(
      evaluationContext: TransactionEvaluationContext,
    ): TransactionSelectionResult = TransactionSelectionResult.SELECTED

    override fun evaluateTransactionPostProcessing(
      evaluationContext: TransactionEvaluationContext,
      processingResult: TransactionProcessingResult,
    ): TransactionSelectionResult = TransactionSelectionResult.SELECTED

    override fun onTransactionNotSelected(
      evaluationContext: TransactionEvaluationContext,
      transactionSelectionResult: TransactionSelectionResult,
    ) {
      val txHash = evaluationContext.pendingTransaction.transaction.hash
      rejections[txHash] = transactionSelectionResult
    }
  }
}
