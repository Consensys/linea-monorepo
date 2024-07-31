"use client";

import { useEffect, useState } from "react";
import classNames from "classnames";
import { isAddress } from "viem";
import { useAccount } from "wagmi";
import { useFormContext } from "react-hook-form";
import { MdKeyboardArrowDown } from "react-icons/md";

export default function Recipient() {
  const [isChecked, setIsChecked] = useState(false);

  // Form
  const { register, formState, setValue, setError, clearErrors, watch } = useFormContext();
  const { errors } = formState;
  const watchRecipient = watch("recipient", false);

  // Hooks
  const { isConnected } = useAccount();

  useEffect(() => {
    if (watchRecipient && !isAddress(watchRecipient)) {
      setError("recipient", {
        type: "custom",
        message: "Invalid address",
      });
    } else {
      clearErrors("recipient");
    }
  }, [watchRecipient, setError, clearErrors]);

  const toggleCheckbox = () => {
    setIsChecked(!isChecked);
    clearErrors("recipient");
    setValue("recipient", "");
  };

  return (
    <div
      className={classNames("rounded-none collapse", {
        "text-neutral-600": !isConnected,
      })}
    >
      <input type="checkbox" className="min-h-0" onChange={toggleCheckbox} />
      <div className="collapse-title flex min-h-0 flex-row justify-end space-x-1 p-0 text-sm">
        <div>Optional: Add recipient</div>{" "}
        <MdKeyboardArrowDown
          className={classNames("text-xl", {
            "rotate-180": isChecked,
          })}
        />
      </div>
      <div
        className={classNames("collapse-content p-1 !pb-1", {
          "mt-3 h-18": isChecked,
        })}
      >
        <div className="form-control w-full">
          <div className="flex flex-row">
            <input
              type="text"
              {...register("recipient", {
                validate: (value) => !value || isAddress(value) || "Invalid address",
              })}
              maxLength={42}
              disabled={!isConnected}
              placeholder="0x..."
              className="input input-bordered input-info w-full [appearance:textfield] [&::-webkit-inner-spin-button]:appearance-none [&::-webkit-outer-spin-button]:appearance-none"
            />
          </div>

          {errors.recipient && <div className="pt-2 text-error">{errors.recipient.message?.toString()}</div>}
        </div>
      </div>
    </div>
  );
}
