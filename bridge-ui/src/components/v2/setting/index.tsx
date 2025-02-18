"use client";
import styles from "./setting.module.scss";
import clsx from "clsx";
import SettingIcon from "@/assets/icons/setting.svg";
import ToggleSwitch from "@/components/v2/ui/toggle-switch";
import { useEffect, useRef, useState } from "react";
import CurrencyDropdown from "@/components/v2/bridge/currency-dropdown";

export default function Setting() {
  const dropdownRef = useRef<HTMLDivElement | null>(null);
  const [isDropdownVisible, setDropdownVisible] = useState<boolean>(false);

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
              <CurrencyDropdown disabled />
            </li>
            <li className={clsx(styles["dropdown-item"])}>
              <span>Show Test Networks</span>
              <ToggleSwitch />
            </li>
          </ul>
        </div>
      </div>
    </div>
  );
}
