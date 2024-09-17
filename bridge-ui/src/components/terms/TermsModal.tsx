import React, { useEffect, useRef, useState } from "react";
import Link from "next/link";
import { useConfigStore } from "@/stores/configStore";
import { cn } from "@/utils/cn";
import Button from "../bridge/Button";

export default function TermsModal() {
  const termsModalRef = useRef<HTMLDivElement>(null);

  const { agreeToTerms, rehydrated, setAgreeToTerms } = useConfigStore((state) => ({
    agreeToTerms: state.agreeToTerms,
    rehydrated: state.rehydrated,
    setAgreeToTerms: state.setAgreeToTerms,
  }));

  const [open, setOpen] = useState(false);
  const [videoEnabled, setVideoEnabled] = useState(false);

  const isFirstTime = () => {
    return !agreeToTerms;
  };

  const handleAgreeToTerms = () => {
    setAgreeToTerms(true);
    setOpen(false);
    setVideoEnabled(false);
  };

  useEffect(() => {
    if (rehydrated && window && isFirstTime()) {
      setTimeout(() => {
        setOpen(true);
        setVideoEnabled(true);
      }, 1000);
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [rehydrated]);

  if (!rehydrated) {
    return null;
  }

  return (
    <div
      ref={termsModalRef}
      id="terms_modal"
      className={cn(
        "p-4 fixed right-2 left-2 md:left-auto md:right-5 md:max-w-[20rem] bg-white rounded text-black z-50 transition-all duration-500",
        !open ? "invisible -bottom-full" : "visible bottom-2 md:bottom-16",
      )}
    >
      {videoEnabled && (
        <>
          <div className="text-lg font-medium">Tutorial</div>
          <div className="pb-2">
            <iframe
              className="aspect-video w-full"
              src="https://www.youtube.com/embed/V4DflPkxqE8"
              title="YouTube video player"
              frameBorder="0"
              allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture; web-share"
              allowFullScreen
            ></iframe>
          </div>
        </>
      )}
      <div className="text-lg font-medium">Terms of Use</div>
      <div className="mb-2 text-xs leading-relaxed">
        Linea Mainnet is in Alpha — click{" "}
        <Link href="https://docs.linea.build/risk-disclosures" target="_blank" className="mr-1 font-bold">
          here
        </Link>
        to learn more about the risks and security measures. I agree and accept that my use of the services is subject
        to the Linea Terms of Use, which contains an arbitration provision and class action waiver{" "}
        <Link href="https://linea.build/terms-of-service" target="_blank" className="ml-1 font-bold">
          (Terms of Service | Linea )
        </Link>{" "}
      </div>
      <Button
        id="agree-terms-btn"
        onClick={handleAgreeToTerms}
        type="button"
        variant="primary"
        size="sm"
        className="mt-3 w-full font-medium"
      >
        Got It
      </Button>
    </div>
  );
}
