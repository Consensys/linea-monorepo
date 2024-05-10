'use client';

import { useEffect, useState } from 'react';
import classNames from 'classnames';
import { isAddress } from 'viem';
import { useFormContext } from 'react-hook-form';
import { MdKeyboardArrowDown } from 'react-icons/md';

import { useIsConnected } from '@/hooks';

export default function Recipient() {
  const [isChecked, setIsChecked] = useState(false);

  // Form
  const { register, formState, setValue, setError, clearErrors, watch } = useFormContext();
  const { errors } = formState;
  const watchRecipient = watch('recipient', false);

  // Hooks
  const isConnected = useIsConnected();

  useEffect(() => {
    if (watchRecipient && !isAddress(watchRecipient)) {
      setError('recipient', {
        type: 'custom',
        message: 'Invalid address',
      });
    } else {
      clearErrors('recipient');
    }
  }, [watchRecipient, setError, clearErrors]);

  const toggleCheckbox = () => {
    setIsChecked(!isChecked);
    clearErrors('recipient');
    setValue('recipient', '');
  };

  return (
    <div
      className={classNames('rounded-none collapse', {
        'text-neutral-600': !isConnected,
      })}
    >
      <input type="checkbox" className="min-h-0" onChange={toggleCheckbox} />
      <div className="flex flex-row justify-end min-h-0 p-0 space-x-1 text-sm collapse-title">
        <div>Optional: Add recipient</div>{' '}
        <MdKeyboardArrowDown
          className={classNames('text-xl', {
            'rotate-180': isChecked,
          })}
        />
      </div>
      <div
        className={classNames('collapse-content p-1 !pb-1', {
          'mt-3 h-18': isChecked,
        })}
      >
        <div className="w-full form-control">
          <div className="flex flex-row">
            <input
              type="text"
              {...register('recipient', {
                validate: (value) => !value || isAddress(value) || 'Invalid address',
              })}
              maxLength={42}
              disabled={!isConnected}
              placeholder="0x..."
              className="w-full input input-bordered input-info [appearance:textfield] [&::-webkit-outer-spin-button]:appearance-none [&::-webkit-inner-spin-button]:appearance-none"
            />
          </div>

          {errors.recipient && <div className="pt-2 text-error">{errors.recipient.message?.toString()}</div>}
        </div>
      </div>
    </div>
  );
}
