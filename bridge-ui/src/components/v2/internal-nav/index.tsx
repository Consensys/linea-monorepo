"use client";
import styles from "./internal-nav.module.scss";
import clsx from "clsx";
import { usePathname } from "next/navigation";
import ConfigIcon from "@/assets/icons/config.svg";
import NavItem from "./item";
import ToggleSwitch from "@/components/v2/ui/toggle-switch";
import { useEffect, useRef, useState } from "react";

const NavData = [
  {
    title: "Bridge",
    href: "/",
  },
  {
    title: "Transactions",
    href: "/transactions",
  },
];

export default function InternalNav() {
  const dropdownRef = useRef<HTMLDivElement | null>(null);
  const [isDropdownVisible, setDropdownVisible] = useState<boolean>(false);

  const pathnane = usePathname();

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
      <div className={styles["list-nav"]}>
        {NavData.map((item, index) => (
          <NavItem key={index} {...item} active={pathnane === item.href} />
        ))}
      </div>
      <div className={styles["nav-wrapper"]} ref={dropdownRef}>
        <div
          className={clsx(styles["menu-button"], {
            [styles["visible"]]: isDropdownVisible,
          })}
          onClick={toggleDropdown}
        >
          <ConfigIcon />
        </div>
        <div
          className={clsx(styles["menu-dropdown"], {
            [styles["visible"]]: isDropdownVisible,
          })}
        >
          <ul className={styles["dropdown-list"]}>
            <li className={clsx(styles["dropdown-item"])}>
              <span>Currency</span>
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
