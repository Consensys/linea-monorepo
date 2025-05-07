"use client";

import { useState } from "react";
import FaqItem from "./faq-item";
import styles from "./faq.module.scss";

type FaqProps = {
  data: { title: string; content: string }[];
};

export default function Faq({ data }: FaqProps) {
  const [openIndex, setOpenIndex] = useState<number | null>(null);

  const handleToggle = (index: number) => {
    setOpenIndex(openIndex === index ? null : index);
  };

  return (
    <ul className={styles["list"]}>
      {data.map((faq, index) => (
        <FaqItem
          key={`faq-item-${index}`}
          data={faq}
          isOpen={openIndex === index}
          onToggle={() => handleToggle(index)}
        />
      ))}
    </ul>
  );
}
