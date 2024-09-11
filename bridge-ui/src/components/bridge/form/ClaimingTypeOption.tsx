import React from "react";
import { UseFormRegisterReturn } from "react-hook-form";
import { MdInfo } from "react-icons/md";
import { Tooltip } from "../../tooltip";
import { cn } from "@/utils/cn";

interface ClaimOptionProps {
  id: string;
  value?: string;
  label: string;
  tooltip: string;
  disabled: boolean;
  isConnected: boolean;
  onClick?: () => void;
  register?: UseFormRegisterReturn;
  isSelected: boolean;
}

const ClaimingTypeOption: React.FC<ClaimOptionProps> = ({
  id,
  value,
  label,
  tooltip,
  disabled,
  isConnected,
  onClick,
  register,
  isSelected,
}) => (
  <div>
    <input
      {...(register ? register : {})}
      id={id}
      type="radio"
      value={value}
      className="peer hidden"
      disabled={disabled}
      onClick={onClick}
      checked={isSelected}
      readOnly
    />
    <label
      htmlFor={id}
      className={cn("btn btn-outline normal-case font-normal w-full rounded-full", {
        "btn-disabled": disabled,
        "peer-checked:border-primary peer-checked:bg-cardBg peer-checked:text-white": isConnected,
      })}
    >
      {label}
      <Tooltip text={tooltip}>
        <MdInfo />
      </Tooltip>
    </label>
  </div>
);

export default ClaimingTypeOption;
