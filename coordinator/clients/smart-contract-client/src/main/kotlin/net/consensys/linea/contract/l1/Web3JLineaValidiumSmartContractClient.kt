package net.consensys.linea.contract.l1

import linea.contract.l1.Web3JLineaValidiumSmartContractClientReadOnly
import linea.domain.BlockParameter.Companion.toBlockParameter
import linea.domain.gas.GasPriceCaps
import linea.kotlin.toULong
import linea.web3j.transactionmanager.AsyncFriendlyTransactionManager
import net.consensys.linea.contract.Web3JContractAsyncHelper
import net.consensys.zkevm.coordinator.clients.smartcontract.BlockAndNonce
import net.consensys.zkevm.coordinator.clients.smartcontract.LineaValidiumSmartContractClient
import net.consensys.zkevm.domain.BlobRecord
import net.consensys.zkevm.domain.ProofToFinalize
import org.web3j.protocol.Web3j
import org.web3j.protocol.core.DefaultBlockParameter
import org.web3j.tx.Contract
import tech.pegasys.teku.infrastructure.async.SafeFuture
import java.math.BigInteger

class Web3JLineaValidiumSmartContractClient(
  web3j: Web3j,
  contractAddress: String,
  private val transactionManager: AsyncFriendlyTransactionManager,
  private val web3jContractHelper: Web3JContractAsyncHelper,
  private val web3jLineaClient: Contract,
) : Web3JLineaValidiumSmartContractClientReadOnly(web3j, contractAddress), LineaValidiumSmartContractClient {

  override fun currentNonce(): ULong {
    return transactionManager.currentNonce().toULong()
  }

  private fun resetNonce(blockNumber: BigInteger): SafeFuture<ULong> {
    return transactionManager
      .resetNonce(blockNumber.toBlockParameter())
      .thenApply { currentNonce() }
  }

  override fun updateNonceAndReferenceBlockToLastL1Block(): SafeFuture<BlockAndNonce> {
    return web3jContractHelper.getCurrentBlock()
      .thenCompose { blockNumber ->
        web3jLineaClient.setDefaultBlockParameter(DefaultBlockParameter.valueOf(blockNumber))
        resetNonce(blockNumber)
          .thenApply { BlockAndNonce(blockNumber.toULong(), currentNonce()) }
      }
  }

  override fun acceptShnarfData(
    blobs: List<BlobRecord>,
    gasPriceCaps: GasPriceCaps?,
  ): SafeFuture<String> {
    return getVersion()
      .thenCompose { version ->
        val function = Web3JLineaValidiumFunctionBuilders.buildAcceptShnarfDataFunction(version, blobs)
        web3jContractHelper.sendShnarfDataTransactionAndGetTxHash(
          function = function,
          gasPriceCaps = gasPriceCaps,
        )
      }
  }

  override fun acceptShnarfDataEthCall(
    blobs: List<BlobRecord>,
    gasPriceCaps: GasPriceCaps?,
  ): SafeFuture<String?> {
    return getVersion()
      .thenCompose { version ->
        val function = Web3JLineaValidiumFunctionBuilders.buildAcceptShnarfDataFunction(version, blobs)
        web3jContractHelper.executeEthCall(
          function = function,
          gasPriceCaps = gasPriceCaps,
        )
      }
  }

  override fun finalizeBlocks(
    aggregation: ProofToFinalize,
    aggregationLastBlob: BlobRecord,
    parentL1RollingHash: ByteArray,
    parentL1RollingHashMessageNumber: Long,
    gasPriceCaps: GasPriceCaps?,
  ): SafeFuture<String> {
    return getVersion()
      .thenCompose { version ->
        val function = Web3JLineaValidiumFunctionBuilders.buildFinalizeBlocksFunction(
          version = version,
          aggregationProof = aggregation,
          aggregationLastBlob = aggregationLastBlob,
          parentL1RollingHash = parentL1RollingHash,
          parentL1RollingHashMessageNumber = parentL1RollingHashMessageNumber,
        )
        web3jContractHelper
          .sendTransactionAsync(function, BigInteger.ZERO, gasPriceCaps)
          .thenApply { result -> result.transactionHash }
      }
  }

  override fun finalizeBlocksEthCall(
    aggregation: ProofToFinalize,
    aggregationLastBlob: BlobRecord,
    parentL1RollingHash: ByteArray,
    parentL1RollingHashMessageNumber: Long,
  ): SafeFuture<String?> {
    return getVersion()
      .thenCompose { version ->
        val function = Web3JLineaValidiumFunctionBuilders.buildFinalizeBlocksFunction(
          version,
          aggregation,
          aggregationLastBlob,
          parentL1RollingHash,
          parentL1RollingHashMessageNumber,
        )
        web3jContractHelper.executeEthCall(function)
      }
  }

  override fun finalizeBlocksAfterEthCall(
    aggregation: ProofToFinalize,
    aggregationLastBlob: BlobRecord,
    parentL1RollingHash: ByteArray,
    parentL1RollingHashMessageNumber: Long,
    gasPriceCaps: GasPriceCaps?,
  ): SafeFuture<String> {
    return getVersion()
      .thenCompose { version ->
        val function = Web3JLineaValidiumFunctionBuilders.buildFinalizeBlocksFunction(
          version,
          aggregation,
          aggregationLastBlob,
          parentL1RollingHash,
          parentL1RollingHashMessageNumber,
        )
        web3jContractHelper.sendTransactionAfterEthCallAsync(function, BigInteger.ZERO, gasPriceCaps)
          .thenApply { result -> result.transactionHash }
      }
  }
}
