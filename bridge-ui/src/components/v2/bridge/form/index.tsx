import Tooltip from "@/components/v2/ui/tooltip";
import styles from "./bridge-form.module.scss";
import ArrowRightIcon from "@/assets/icons/arrow-right.svg";
import Image from "next/image";
import ConnectButton from "@/components/v2/connect-button";
import FaqHelp from "@/components/v2/bridge/faq-help";
import TokenList from "@/components/v2/bridge/token-list";
import { useAccount } from "wagmi";
import { Balance } from "@/components/v2/bridge/balance";
import { Amount } from "@/components/v2/bridge/amount";

export default function BridgeForm() {
  const { isConnected } = useAccount();

  return (
    <>
      <form>
        <div className={styles["form-wrapper"]}>
          <div className={styles["exchange"]}>
            <div className={styles["from"]}>
              <Image src="/images/logo/ethereum-rounded.svg" width="40" height="40" alt="eth" />
              <div className={styles["info"]}>
                <div className={styles["info-name"]}>From</div>
                <div className={styles["info-value"]}>Ethereum</div>
              </div>
            </div>
            <div className={styles["icon-container"]}>
              <Tooltip text="Switch chains" position="top">
                <div className={styles["icon-wrapper"]}>
                  <ArrowRightIcon className={styles["arrow-icon"]} />
                </div>
              </Tooltip>
            </div>
            <div className={styles["to"]}>
              <Image src="/images/logo/linea-rounded.svg" width="40" height="40" alt="linea" />
              <div className={styles["info"]}>
                <div className={styles["info-name"]}>To</div>
                <div className={styles["info-value"]}>Linea</div>
              </div>
            </div>
          </div>
          <div className={styles["amount-wrapper"]}>
            <Amount />
            <div className={styles["right"]}>
              <TokenList />
              {isConnected && <Balance />}
            </div>
          </div>
          {/* <div className={styles["get-on-linea"]}>
            <div>
              <div>Get on Linea</div>
              <div>Native Bridge</div>
              <div>Setting</div>
            </div>
            <div>Value</div>
            <div>Estimate</div>
            <div>
              Can&apos;t wait? <button>Speed up with Across</button>
            </div>
          </div> */}
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
