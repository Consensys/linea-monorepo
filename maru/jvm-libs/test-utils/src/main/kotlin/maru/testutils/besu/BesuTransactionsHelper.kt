/*
   Copyright 2025 Consensys Software Inc.

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

      http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
 */
package maru.testutils.besu

import org.hyperledger.besu.tests.acceptance.dsl.account.Account
import org.hyperledger.besu.tests.acceptance.dsl.account.Accounts
import org.hyperledger.besu.tests.acceptance.dsl.blockchain.Amount
import org.hyperledger.besu.tests.acceptance.dsl.condition.eth.EthConditions
import org.hyperledger.besu.tests.acceptance.dsl.transaction.account.AccountTransactions
import org.hyperledger.besu.tests.acceptance.dsl.transaction.account.TransferTransaction
import org.hyperledger.besu.tests.acceptance.dsl.transaction.eth.EthTransactions

class BesuTransactionsHelper {
  private val ethTransactions = EthTransactions()
  private val accounts = Accounts(ethTransactions)
  private val accountTransactions = AccountTransactions(accounts)
  val ethConditions = EthConditions(ethTransactions)
  private val whaleAccount =
    Account.fromPrivateKey(
      ethTransactions,
      "Whale",
      "0x3a4ff6d22d7502ef2452368165422861c01a0f72f851793b372b87888dc3c453",
    )

  fun createAccount(accountName: String): Account = accounts.createAccount(accountName)

  fun createTransfer(
    recipient: Account,
    amount: Amount,
  ): TransferTransaction = accountTransactions.createTransfer(whaleAccount, recipient, amount)

  fun createTransfers(numberOfTransactions: UInt): TransferTransaction {
    val recipient = accounts.createAccount("recipient")

    return accountTransactions.createTransfer(whaleAccount, recipient, numberOfTransactions.toInt())
  }
}
