"use client";

import Link from "next/link";
import { format, fromUnixTime } from "date-fns";
import { motion, Variants } from "framer-motion";
import { MdOutlineArrowRightAlt } from "react-icons/md";
import { Address, formatUnits, zeroAddress } from "viem";

import ChainLogo from "@/components/widgets/ChainLogo";
import HistoryClaim from "./HistoryClaim";
import { formatAddress, safeGetAddress } from "@/utils/format";
import { TransactionHistory } from "@/models/history";

interface Props {
  transaction: TransactionHistory;
  variants: Variants;
}
export default function HistoryItem({ transaction, variants }: Props) {
  let destinationAddress: Address | null = null;
  const tokenAddress = safeGetAddress(transaction.tokenAddress);
  const l1Address = safeGetAddress(transaction.token.L1);
  const l2Address = safeGetAddress(transaction.token.L2);

  if (tokenAddress && tokenAddress === l1Address) {
    destinationAddress = transaction.token.L2;
  } else if (tokenAddress && tokenAddress === l2Address) {
    destinationAddress = transaction.token.L1;
  }

  const getL1Token = () => {
    if (transaction.tokenAddress && transaction.tokenAddress !== zeroAddress) {
      return (
        <Link
          href={`${transaction.fromChain.blockExplorers?.default.url}/address/${transaction.tokenAddress}`}
          target={"_blank"}
          className="link text-xs font-bold"
          passHref
        >
          {transaction.token.symbol}
        </Link>
      );
    } else {
      return <span className="font-bold">{transaction.token.symbol}</span>;
    }
  };

  const getL2Token = () => {
    if (destinationAddress && destinationAddress !== zeroAddress) {
      return (
        <Link
          href={`${transaction.toChain.blockExplorers?.default.url}/address/${destinationAddress}`}
          target={"_blank"}
          className="link text-xs font-bold"
          passHref
        >
          {transaction.token.symbol}
        </Link>
      );
    } else {
      return <span className="font-bold">{transaction.token.symbol}</span>;
    }
  };

  return (
    <motion.ul className="space-y-2" variants={variants} initial="hidden" animate="show">
      <li className="flex flex-row justify-between">
        <div className="flex flex-row items-center space-x-1">
          <ChainLogo chainId={transaction.fromChain.id} />
          <span className="flex flex-row">{transaction.fromChain.name}</span>
          <div>
            <MdOutlineArrowRightAlt className="mx-1" />
          </div>
          <ChainLogo chainId={transaction.toChain.id} />
          <span className="flex flex-row">{transaction.toChain.name}</span>
        </div>
        <div className="text-sm">
          {transaction && format(fromUnixTime(parseInt(transaction.timestamp.toString())), "LLL dd yyyy, HH:mm")}
        </div>
      </li>
      <li className="flex flex-row items-center justify-between text-xs">
        <div className="space-x-1">
          <span>{formatUnits(transaction.amount, transaction.token.decimals)}</span>
          <span>
            {transaction.fromChain.name} {getL1Token()}
          </span>
          <span>to</span>
          <span>
            {transaction.toChain.name} {getL2Token()}
          </span>
          <span>To</span>
          <span>
            <Link
              href={`${transaction.toChain.blockExplorers?.default.url}/address/${transaction.recipient}`}
              target={"_blank"}
              className="link"
              passHref
            >
              {formatAddress(transaction.recipient)}
            </Link>
          </span>
        </div>
      </li>
      <li className="space-x-1 text-xs">
        <span>{transaction.fromChain.name} transaction:</span>
        <Link
          href={`${transaction.fromChain.blockExplorers?.default.url}/tx/${transaction.transactionHash}`}
          target={"_blank"}
          className="link"
          passHref
        >
          {formatAddress(transaction.transactionHash)}
        </Link>
      </li>
      {transaction.messages?.map((message) => {
        return <HistoryClaim key={message.messageHash} message={message} transaction={transaction} />;
      })}
      <li className="divider"></li>
    </motion.ul>
  );
}
