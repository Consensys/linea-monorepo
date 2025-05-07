import { motion } from "motion/react";
import PlusIcon from "@/assets/icons/plus.svg";
import styles from "./faq-item.module.scss";

type FaqItemProps = {
  data: {
    title: string;
    content: string;
  };
  isOpen: boolean;
  onToggle: () => void;
};

export default function FaqItem({ data, isOpen, onToggle }: FaqItemProps) {
  const { title, content } = data;
  return (
    <li className={styles["item"]}>
      <motion.div
        className={styles["header"]}
        onClick={onToggle}
        role="button"
        tabIndex={0}
        aria-expanded={isOpen}
        onKeyDown={(e) => e.key === "Enter" && onToggle()}
        animate={{
          marginBottom: isOpen ? "1.6rem" : "0rem",
        }}
        transition={{
          marginBottom: {
            duration: 0.3,
            ease: "easeInOut",
          },
        }}
      >
        <h2 className={styles["item-title"]}>{title}</h2>
        <motion.div
          animate={{ rotate: isOpen ? 225 : 0 }}
          transition={{ duration: 0.3, ease: "easeInOut" }}
          className={styles["button"]}
        >
          <PlusIcon />
        </motion.div>
      </motion.div>

      <motion.div
        className={styles["content"]}
        initial={false}
        animate={{
          height: isOpen ? "auto" : 0,
          opacity: isOpen ? 1 : 0,
        }}
        transition={{
          duration: 0.3,
        }}
        style={{ overflow: "hidden" }}
      >
        <motion.div
          initial={{ opacity: 0, y: -10 }}
          animate={{ opacity: isOpen ? 1 : 0, y: isOpen ? 0 : -10 }}
          transition={{
            duration: 0.3,
            ease: "easeInOut",
            delay: isOpen ? 0.1 : 0,
          }}
          dangerouslySetInnerHTML={{ __html: content }}
        />
      </motion.div>
    </li>
  );
}
