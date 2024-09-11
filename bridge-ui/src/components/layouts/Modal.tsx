"use client";

import classNames from "classnames";
import { useContext } from "react";
import { ModalContext } from "@/contexts/modal.context";
import { MdOutlineClose } from "react-icons/md";

type ModalProps = Record<string, never>;

const Modal: React.FC<ModalProps> = () => {
  const { ref, modalContent, options } = useContext(ModalContext);

  const width = options?.width ? options.width : "w-screen md:w-[600px]";

  return (
    <dialog ref={ref} className="modal">
      <div className={classNames("modal-box bg-cardBg border-card border-2", width)}>
        <form method="dialog">
          <button className="btn btn-circle btn-ghost btn-sm absolute right-2 top-2">
            <MdOutlineClose className="text-lg" />
          </button>
        </form>
        {modalContent}
      </div>
      <form method="dialog" className="modal-backdrop">
        <button>Close</button>
      </form>
    </dialog>
  );
};

export default Modal;
