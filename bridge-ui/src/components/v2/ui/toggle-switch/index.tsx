import { useState } from "react";
import styles from "./toggle-switch.module.scss";
import clsx from "clsx";

type Props = {
  checked?: boolean;
  onChange?: (checked: boolean) => void;
  disabled?: boolean;
};

export default function ToggleSwitch({ checked = false, onChange, disabled }: Props) {
  const [isChecked, setIsChecked] = useState<boolean>(checked);

  const handleChange = () => {
    if (!disabled) {
      const newChecked = !isChecked;
      setIsChecked(newChecked);

      if (onChange) {
        onChange(isChecked);
      }
    }
  };

  return (
    <label
      className={clsx(styles["switch"], {
        [styles["disabled"]]: disabled,
      })}
    >
      <input type="checkbox" checked={isChecked} onChange={handleChange} disabled={disabled} />
      <span className={styles["slider"]} />
    </label>
  );
}
