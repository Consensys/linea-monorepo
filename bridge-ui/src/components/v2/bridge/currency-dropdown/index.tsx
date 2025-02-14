import React, { useState, useEffect, useRef } from "react";
import CaretDownIcon from "@/assets/icons/caret-down.svg";
import styles from "./currency-dropdown.module.scss";

type CurrencyOption = {
  value: string;
  label: string;
  flag: string;
};

type Props = {
  disabled?: boolean;
};

export default function CurrencyDropdown({ disabled }: Props) {
  const options: CurrencyOption[] = [
    { value: "usd", label: "USD", flag: "🇺🇸" },
    { value: "usd2", label: "USD2", flag: "🇺🇸" },
    { value: "usd3", label: "USD3", flag: "🇺🇸" },
  ];

  const [selectedCurrency, setSelectedCurrency] = useState<CurrencyOption>(options[0]);
  const [isOpen, setIsOpen] = useState(false);
  const dropdownRef = useRef<HTMLDivElement>(null);
  const buttonRef = useRef<HTMLButtonElement>(null);

  const handleSelect = (option: CurrencyOption) => {
    setSelectedCurrency(option);
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
      <button ref={buttonRef} type="button" className={styles.button} onClick={toggleDropdown} disabled={disabled}>
        <div className={styles["selected-label"]}>
          <span className={styles.flag}>{selectedCurrency.flag}</span>
          <span>{selectedCurrency.label}</span>
        </div>
        <CaretDownIcon className={styles.caret} />
      </button>
      {isOpen && (
        <div ref={dropdownRef} className={styles.dropdown}>
          {options.map((option) => (
            <div key={option.value} onClick={() => handleSelect(option)} className={styles.option}>
              <span className={styles.flag}>{option.flag}</span>
              {option.label}
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
