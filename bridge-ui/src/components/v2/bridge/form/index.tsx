import ConnectButton from "@/components/v2/connect-button";
import FaqHelp from "@/components/v2/bridge/faq-help";
import TokenList from "@/components/v2/bridge/token-list";
import { useAccount } from "wagmi";
import { Amount } from "@/components/v2/bridge/amount";
import SwapChain from "@/components/v2/bridge/swap-chain";
import FromChain from "@/components/v2/bridge/from-chain";
import ToChain from "@/components/v2/bridge/to-chain";
import Claiming from "@/components/v2/bridge/claiming";
import styles from "./bridge-form.module.scss";
import { Submit } from "@/components/v2/bridge/submit";
import TransactionPaperIcon from "@/assets/icons/transaction-paper.svg";
import Link from "next/link";
import Setting from "@/components/v2/setting";

export default function BridgeForm() {
  const { isConnected } = useAccount();

  return (
    <>
      <form>
        <div className={styles["form-wrapper"]}>
          <div className={styles.headline}>
            <h1 className={styles.title}>Bridge</h1>
            <div className={styles["action"]}>
              <Link href="/transactions">
                <TransactionPaperIcon className={styles["transaction-icon"]} />
              </Link>
              <Setting />
            </div>
          </div>
          <div className={styles["exchange"]}>
            <FromChain />
            <div className={styles["swap-chain-container"]}>
              <SwapChain />
            </div>
            <ToChain />
          </div>
          <div className={styles["amount-wrapper"]}>
            <Amount />
            <div className={styles["right"]}>
              <TokenList />
            </div>
          </div>
          <Claiming />
          <div className={styles["connect-btn-wrapper"]}>
            {isConnected ? <Submit /> : <ConnectButton fullWidth text={"Connect wallet"} />}
          </div>
          <FaqHelp isMobile />
        </div>
      </form>
      <FaqHelp />
    </>
  );
}
