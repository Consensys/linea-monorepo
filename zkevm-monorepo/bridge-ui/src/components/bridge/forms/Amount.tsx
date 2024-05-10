'use client';

import { ChangeEvent, useCallback, useContext, useEffect } from 'react';
import { useFormContext } from 'react-hook-form';
import Image from 'next/image';
import { useAccount } from 'wagmi';
import { formatEther, parseUnits } from 'viem';

import { ChainContext } from '@/contexts/chain.context';
import { useBridge, useIsConnected, useMessageService } from '@/hooks';
import { TokenType } from '@/config/config';

const MAX_AMOUNT_CHAR = 24;
const FEES_MARGIN_PERCENT = 20;
const AMOUNT_REGEX = '^[0-9]*[.,]?[0-9]*$';

interface Props {
  tokensModalRef: React.RefObject<HTMLDialogElement>;
}

export default function Amount({ tokensModalRef }: Props) {
  // Context
  const context = useContext(ChainContext);
  const { token } = context;

  // Form
  const { setValue, getValues, formState, setError, clearErrors, trigger, watch } = useFormContext();
  const { errors } = formState;
  const watchBalance = watch('balance', false);
  const amount = getValues('amount');
  const gasFees = getValues('gasFees') || BigInt(0);
  const minFees = getValues('minFees') || BigInt(0);

  // Wagmi
  const { address } = useAccount();

  // Hooks
  const isConnected = useIsConnected();
  const { estimateGasBridge } = useBridge();
  const { minimumFee } = useMessageService();

  const compareAmountBalance = useCallback(
    (_amount: string) => {
      if (!token) {
        return;
      }
      const amountToCompare =
        token.type === TokenType.ETH
          ? parseUnits(_amount, token.decimals) + gasFees + parseUnits(minFees, 18)
          : parseUnits(_amount, token.decimals);
      const balanceToCompare = parseUnits(watchBalance, token.decimals);
      if (amountToCompare > balanceToCompare) {
        setError('amount', {
          type: 'custom',
          message: 'Not enough funds (Incl fees)',
        });
      } else {
        clearErrors('amount');
      }
    },
    [token, gasFees, minFees, clearErrors, setError, watchBalance],
  );

  /**
   * Set Max Amount
   */
  const setMaxHandler = async () => {
    if (!token || !watchBalance) return;

    let maxAmount;
    if (token.type === TokenType.ETH) {
      const bridgeGasFee = await estimateGasBridge(watchBalance, minimumFee);
      if (!bridgeGasFee) return;

      // Add margin to gas fees for prevent error when gas fees change
      const gasFeesMargin = (bridgeGasFee * BigInt(100 + FEES_MARGIN_PERCENT)) / BigInt(100);

      maxAmount = formatEther(parseUnits(watchBalance, token.decimals) - gasFeesMargin - parseUnits(minFees, 18));
    } else {
      maxAmount = watchBalance;
    }

    setValue('amount', maxAmount);
    compareAmountBalance(maxAmount);
  };

  /**
   * Dynamic check amount
   * @param e
   * @returns
   */
  const checkAmountHandler = (e: ChangeEvent<HTMLInputElement>) => {
    // Replace minus
    const amount = e.target.value.replace(/,/g, '.');

    if (!token) {
      return;
    }

    if (new RegExp(AMOUNT_REGEX).test(amount) || amount === '') {
      // Limit max char
      if (amount.length > MAX_AMOUNT_CHAR) {
        setValue('amount', amount.substring(0, MAX_AMOUNT_CHAR));
      } else {
        setValue('amount', amount);
      }
    }

    compareAmountBalance(amount);
  };

  // Detect changes
  useEffect(() => {
    if (amount) {
      trigger(['amount']);
      compareAmountBalance(amount);
    }
  }, [amount, trigger, compareAmountBalance]);

  // Detect when changing account
  useEffect(() => {
    setValue('amount', '');
    clearErrors('amount');
  }, [address, setValue, clearErrors]);

  return (
    <div className="form-control">
      <div className="flex flex-row">
        <div className="dropdown">
          {token && (
            <button
              id={`token-select-btn`}
              type="button"
              className="flex flex-row pl-0 pr-0 mr-2 btn btn-neutral w-28"
              disabled={!isConnected}
              onClick={() => tokensModalRef?.current?.showModal()}
            >
              <Image src={token.image} alt={token.name} width={20} height={20} className="rounded-full" />
              {token.symbol}
            </button>
          )}
        </div>

        <input
          id="amount-input"
          type="text"
          pattern={AMOUNT_REGEX}
          autoCorrect="off"
          autoComplete="off"
          spellCheck="false"
          inputMode="decimal"
          value={amount}
          onChange={checkAmountHandler}
          disabled={!isConnected}
          placeholder="Enter amount"
          className="w-full max-w-xs input input-bordered input-info [appearance:textfield] [&::-webkit-outer-spin-button]:appearance-none [&::-webkit-inner-spin-button]:appearance-none"
        />
        {token?.type !== TokenType.ETH && (
          <button
            id="max-amount-btn"
            className="mt-3 rounded-full btn btn-primary btn-xs -ml-14"
            type="button"
            disabled={!isConnected}
            onClick={setMaxHandler}
          >
            Max
          </button>
        )}
      </div>
      {errors.amount && <div className="pt-2 text-right text-error">{errors.amount.message?.toString()}</div>}
    </div>
  );
}
