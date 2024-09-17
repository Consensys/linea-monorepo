import Image from "next/image";
import { MdKeyboardArrowDown } from "react-icons/md";
import { ModalContext } from "@/contexts/modal.context";
import { useChainStore } from "@/stores/chainStore";
import { useContext } from "react";
import { useAccount } from "wagmi";
import TokenModal from "./TokenModal";
import { useFormContext } from "react-hook-form";
import Button from "./Button";

export default function TokenList() {
  const token = useChainStore((state) => state.token);
  const { isConnected } = useAccount();

  const { handleShow } = useContext(ModalContext);
  const { setValue, clearErrors } = useFormContext();

  return (
    <div className="dropdown">
      {token && (
        <Button
          id="token-select-btn"
          type="button"
          variant="outline"
          className="px-2 py-1"
          disabled={!isConnected}
          onClick={() =>
            handleShow(<TokenModal setValue={setValue} clearErrors={clearErrors} />, {
              showCloseButton: false,
            })
          }
        >
          <Image src={token.image} alt={token.name} width={25} height={25} className="rounded-full" />
          {token.symbol}
          <MdKeyboardArrowDown className="text-white" size={20} />
        </Button>
      )}
    </div>
  );
}
