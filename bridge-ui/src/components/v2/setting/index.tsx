"use client";
import styles from "./setting.module.scss";
import clsx from "clsx";
import SettingIcon from "@/assets/icons/setting.svg";
import ToggleSwitch from "@/components/v2/ui/toggle-switch";
import { useEffect, useRef, useState } from "react";
import CurrencyDropdown from "@/components/v2/bridge/currency-dropdown";
import { useConfigStore } from "@/stores/configStore";
import { useChainStore } from "@/stores/chainStore";
import { useChains } from "@/hooks/useChains";
import { ChainLayer } from "@/types";
import { useFormStore } from "@/stores/formStoreProvider";

export default function Setting() {
  const dropdownRef = useRef<HTMLDivElement | null>(null);
  const [isDropdownVisible, setDropdownVisible] = useState<boolean>(false);
  const setShowTestnet = useConfigStore.useSetShowTestnet();
  const showTestnet = useConfigStore.useShowTestnet();
  const chains = useChains();
  const setFromChain = useChainStore.useSetFromChain();
  const setToChain = useChainStore.useSetToChain();
  const resetForm = useFormStore((state) => state.resetForm);

  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (dropdownRef.current && !dropdownRef.current.contains(event.target as Node)) {
        setDropdownVisible(false);
      }
    };

    document.addEventListener("mousedown", handleClickOutside);
    return () => {
      document.removeEventListener("mousedown", handleClickOutside);
    };
  }, []);

  const toggleDropdown = () => {
    setDropdownVisible((prev) => !prev);
  };

  useEffect(() => {
    if (!showTestnet) {
      setFromChain(chains.find((c) => !c.testnet && c.layer === ChainLayer.L1));
      setToChain(chains.find((c) => !c.testnet && c.layer === ChainLayer.L2));
    } else {
      setFromChain(chains.find((c) => c.testnet && c.layer === ChainLayer.L1));
      setToChain(chains.find((c) => c.testnet && c.layer === ChainLayer.L2));
    }
    resetForm();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [showTestnet]);

  return (
    <div className={styles["wrapper"]}>
      <div className={styles["dropdown-wrapper"]} ref={dropdownRef}>
        <div
          className={clsx(styles["menu-button"], {
            [styles["visible"]]: isDropdownVisible,
          })}
          onClick={toggleDropdown}
        >
          <SettingIcon />
        </div>
        <div
          className={clsx(styles["menu-dropdown"], {
            [styles["visible"]]: isDropdownVisible,
          })}
        >
          <ul className={styles["dropdown-list"]}>
            <li className={clsx(styles["dropdown-item"])}>
              <span>Currency</span>
              <CurrencyDropdown />
            </li>
            <li className={clsx(styles["dropdown-item"])}>
              <span>Show Test Networks</span>
              <ToggleSwitch
                checked={showTestnet}
                onChange={(checked) => {
                  setShowTestnet(checked);
                }}
              />
            </li>
          </ul>
        </div>
      </div>
    </div>
  );
}
