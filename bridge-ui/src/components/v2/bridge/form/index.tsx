import ConnectButton from "@/components/v2/connect-button";
import FaqHelp from "@/components/v2/bridge/faq-help";
import TokenList from "@/components/v2/bridge/token-list";
import { useAccount } from "wagmi";
import { Balance } from "@/components/v2/bridge/balance";
import { Amount } from "@/components/v2/bridge/amount";
import SwapChain from "@/components/v2/bridge/swap-chain";
import FromChain from "@/components/v2/bridge/from-chain";
import ToChain from "@/components/v2/bridge/to-chain";
import Claiming from "@/components/v2/bridge/claiming";
import styles from "./bridge-form.module.scss";

export default function BridgeForm() {
  const { isConnected } = useAccount();

  return (
    <>
      <form>
        <div className={styles["form-wrapper"]}>
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
              {isConnected && <Balance />}
            </div>
          </div>
          <Claiming />
          <div className={styles["connect-btn-wrapper"]}>
            <ConnectButton fullWidth text={"Connect"} />
          </div>
          <FaqHelp isMobile />
        </div>
      </form>
      <FaqHelp />
    </>
  );
}
