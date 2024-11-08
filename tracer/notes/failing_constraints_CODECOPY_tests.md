testCreateContractFromInitCodeThatDeploysItself
===============================================

hub.copy-instruction---CODECOPY---debug-consistency-constraints
hub.create-instruction---createe-account-row-first-appearance
hub.create-instruction---createe-balance-operation
hub.create-instruction---createe-code-operation
hub.create-instruction---createe-nonce-operation
hub.create-instruction---creator-balance-update
hub.create-instruction---final-context-row-for-deployment-successes-that-wont-revert
hub.create-instruction---setting-GAS_NEXT
hub.create-instruction---setting-the-MMU-instruction
hub.return-instruction---justifying-the-ICPX
hub.return-instruction---setting-MMU-data-first-call
hub.tx-finalization---success-setting-sender-account-row

testCreateContractFromInitCodeSimple
====================================

hub.copy-instruction---CODECOPY---debug-consistency-constraints
hub.create-instruction---createe-account-row-first-appearance
hub.create-instruction---createe-balance-operation
hub.create-instruction---createe-code-operation
hub.create-instruction---createe-nonce-operation
hub.create-instruction---creator-balance-update
hub.create-instruction---final-context-row-for-deployment-successes-that-wont-revert
hub.create-instruction---setting-GAS_NEXT
hub.create-instruction---setting-the-MMU-instruction
hub.tx-finalization---success-setting-sender-account-row

testDeploymentTransactionDeploysOwnInitCodeThroughCodeCopy
==========================================================

hub.return-instruction---first-account-row-for-nonempty-deployments
hub.return-instruction---justifying-the-ICPX
hub.return-instruction---setting-NSR
hub.return-instruction---setting-peeking-flags
hub.return-instruction---setting-the-callers-new-return-data-nonempty-deployments
hub.return-instruction---setting-the-second-MMU-instruction
hub.tx-finalization---success-setting-sender-account-row

testCreateContractFromInitCodeWithMload
=======================================

hub.copy-instruction---CODECOPY---debug-consistency-constraints
hub.create-instruction---createe-account-row-first-appearance
hub.create-instruction---createe-balance-operation
hub.create-instruction---createe-code-operation
hub.create-instruction---createe-nonce-operation
hub.create-instruction---creator-balance-update
hub.create-instruction---final-context-row-for-deployment-successes-that-wont-revert
hub.create-instruction---setting-GAS_NEXT
hub.create-instruction---setting-the-MMU-instruction
hub.tx-finalization---success-setting-sender-account-row


all failing constraints:
========================

hub.copy-instruction---CODECOPY---debug-consistency-constraints
hub.create-instruction---createe-account-row-first-appearance
hub.create-instruction---createe-balance-operation
hub.create-instruction---createe-code-operation
hub.create-instruction---createe-nonce-operation
hub.create-instruction---creator-balance-update
hub.create-instruction---final-context-row-for-deployment-successes-that-wont-revert
hub.create-instruction---setting-GAS_NEXT
hub.create-instruction---setting-the-MMU-instruction
hub.return-instruction---first-account-row-for-nonempty-deployments
hub.return-instruction---justifying-the-ICPX
hub.return-instruction---setting-NSR
hub.return-instruction---setting-peeking-flags
hub.return-instruction---setting-the-callers-new-return-data-nonempty-deployments
hub.return-instruction---setting-the-second-MMU-instruction
hub.tx-finalization---success-setting-sender-account-row

RESOLVED:
=========

hub.return-instruction---setting-MMU-data-first-call
