"use client";

import React, { createContext, ReactNode, useCallback, useRef, useState } from "react";
import Modal from "@/components/layouts/Modal";

interface ModalContextType {
  ref: React.MutableRefObject<HTMLDialogElement | null>;
  handleShow: (content: ReactNode, options?: ModalOptions) => void;
  handleClose: () => void;
  modalContent: ReactNode;
  options: ModalOptions;
}

export const ModalContext = createContext<ModalContextType>({} as ModalContextType);

interface ModalProviderProps {
  children: ReactNode;
}

interface ModalOptions {
  width?: string;
  showCloseButton?: boolean;
  onClose?: () => void;
}

export const ModalProvider: React.FC<ModalProviderProps> = ({ children }) => {
  const ref = useRef<HTMLDialogElement>(null);
  const [modalContent, setModalContent] = useState<ReactNode>(null);
  const [options, setOptions] = useState<ModalOptions>({});

  const handleShow = useCallback((content: ReactNode, options?: ModalOptions) => {
    options && setOptions(options);
    setModalContent(content);
    ref.current?.showModal();
  }, []);

  const handleClose = useCallback(() => {
    ref.current?.close();
    setModalContent(null);
  }, []);

  return (
    <ModalContext.Provider value={{ ref, handleShow, handleClose, modalContent, options }}>
      {children}
      <Modal />
    </ModalContext.Provider>
  );
};
