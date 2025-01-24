import React, { useState, useEffect, useRef } from "react";
import CaretDownIcon from "@/assets/icons/caret-down.svg";
import styles from "./bridge-mode-dropdown.module.scss";
import Image from "next/image";
import { BridgeModeOption } from "@/components/v2/bridge/claiming";

type Props = {
  selectedMode: BridgeModeOption;
  setSelectedMode: (mode: BridgeModeOption) => void;
  options: BridgeModeOption[];
};

export default function BridgeModeDropdown({ selectedMode, setSelectedMode, options }: Props) {
  const [isOpen, setIsOpen] = useState(false);
  const dropdownRef = useRef<HTMLDivElement>(null);
  const buttonRef = useRef<HTMLButtonElement>(null);

  const handleSelect = (option: BridgeModeOption) => {
    setSelectedMode(option);
    setIsOpen(false);
  };

  const toggleDropdown = () => {
    setIsOpen((prev) => !prev);
  };

  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (
        dropdownRef.current &&
        !dropdownRef.current.contains(event.target as Node) &&
        buttonRef.current &&
        !buttonRef.current.contains(event.target as Node)
      ) {
        setIsOpen(false);
      }
    };

    document.addEventListener("mousedown", handleClickOutside);
    return () => {
      document.removeEventListener("mousedown", handleClickOutside);
    };
  }, []);

  return (
    <div className={styles.container}>
      <button ref={buttonRef} type="button" className={styles.button} onClick={toggleDropdown}>
        <div className={styles["selected-label"]}>
          <Image src={selectedMode.image} width={16} height={16} alt={selectedMode.label} />
          <span>{selectedMode.label}</span>
        </div>
        <CaretDownIcon className={styles.caret} />
      </button>
      {isOpen && (
        <div ref={dropdownRef} className={styles.dropdown}>
          {options.map((option) => (
            <div key={option.value} onClick={() => handleSelect(option)} className={styles.option}>
              <Image src={option.image} width={16} height={16} alt={option.label} />
              {option.label}
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
