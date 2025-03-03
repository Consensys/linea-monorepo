import ConnectButton from "@/components/v2/connect-button";
import FaqHelp from "@/components/v2/bridge/faq-help";
import TokenList from "@/components/v2/bridge/token-list";
import { useAccount, useWatchBlockNumber } from "wagmi";
import { Amount } from "@/components/v2/bridge/amount";
import SwapChain from "@/components/v2/bridge/swap-chain";
import FromChain from "@/components/v2/bridge/from-chain";
import ToChain from "@/components/v2/bridge/to-chain";
import Claiming from "@/components/v2/bridge/claiming";
import styles from "./bridge-form.module.scss";
import { Submit } from "@/components/v2/bridge/submit";
import TransactionPaperIcon from "@/assets/icons/transaction-paper.svg";
import Setting from "@/components/v2/setting";
import { useEffect, useState } from "react";
import { DestinationAddress } from "../destination-address";
import Button from "../../ui/button";
import { useNativeBridgeNavigationStore } from "@/stores/nativeBridgeNavigationStore";
import { useTokenBalance } from "@/hooks/useTokenBalance";
import { useChainStore } from "@/stores/chainStore";
import { ChainLayer } from "@/types";
import { useFormStore } from "@/stores/formStoreProvider";

export default function BridgeForm() {
  const [isDestinationAddressOpen, setIsDestinationAddressOpen] = useState(false);
  const setIsTransactionHistoryOpen = useNativeBridgeNavigationStore.useSetIsTransactionHistoryOpen();
  const setIsBridgeOpen = useNativeBridgeNavigationStore.useSetIsBridgeOpen();

  const { isConnected, address } = useAccount();
  const fromChain = useChainStore.useFromChain();
  const token = useFormStore((state) => state.token);
  const setRecipient = useFormStore((state) => state.setRecipient);
  const setBalance = useFormStore((state) => state.setBalance);
  const setClaim = useFormStore((state) => state.setClaim);
  const { balance, refetch } = useTokenBalance(token);

  useWatchBlockNumber({
    onBlockNumber() {
      refetch();
    },
    poll: true,
    pollingInterval: 20_000,
  });

  useEffect(() => {
    setBalance(balance);
    if (address) {
      setRecipient(address);
    }

    if (fromChain.layer === ChainLayer.L2) {
      setClaim("manual");
    }
  }, [balance, address, fromChain.layer, setBalance, setRecipient, setClaim]);

  return (
    <>
      <div className={styles["form-wrapper"]}>
        <div className={styles.headline}>
          <div className={styles["action"]}>
            <Button
              variant="link"
              onClick={() => {
                setIsBridgeOpen(false);
                setIsTransactionHistoryOpen(true);
              }}
            >
              <TransactionPaperIcon className={styles["transaction-icon"]} />
            </Button>
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
        {isDestinationAddressOpen && (
          <div className={styles["destination-address-wrapper"]}>
            <DestinationAddress />
          </div>
        )}
        <div className={styles["connect-btn-wrapper"]}>
          {isConnected ? (
            <Submit setIsDestinationAddressOpen={() => setIsDestinationAddressOpen((prev) => !prev)} />
          ) : (
            <ConnectButton fullWidth text={"Connect wallet"} />
          )}
        </div>
        <FaqHelp isMobile />
      </div>
      <FaqHelp />
    </>
  );
}
