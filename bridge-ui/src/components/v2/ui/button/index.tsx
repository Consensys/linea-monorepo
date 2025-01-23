import { clsx } from "clsx";
import React, { forwardRef, ButtonHTMLAttributes, ReactNode } from "react";

import styles from "./button.module.scss";

interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: "primary" | "link";
  children: ReactNode;
  disabled?: boolean;
  fullWidth?: boolean;
}

const Button = forwardRef<HTMLButtonElement, ButtonProps>(
  ({ variant = "primary", onClick, children, disabled, className, fullWidth, type = "button", ...rest }, ref) => {
    const buttonClassNames = clsx(styles["button"], className, styles[variant], {
      [styles["full-width"]]: fullWidth,
    });

    return (
      <button ref={ref} onClick={onClick} type={type} className={buttonClassNames} disabled={disabled} {...rest}>
        {children}
      </button>
    );
  },
);

Button.displayName = "Button";

export default Button;
