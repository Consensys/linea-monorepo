import React, { useEffect, useRef, useState } from "react";
import Link from "next/link";
import classNames from "classnames";
import { useConfigStore } from "@/stores/configStore";

export default function TermsModal() {
  const termsModalRef = useRef<HTMLDivElement>(null);
  const { agreeToTerms, setAgreeToTerms } = useConfigStore((state) => ({
    agreeToTerms: state.agreeToTerms,
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
    if (window && isFirstTime()) {
      setTimeout(() => {
        setOpen(true);
        setVideoEnabled(true);
      }, 1000);
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  return (
    <div
      ref={termsModalRef}
      id="terms_modal"
      className={classNames(
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
      <button
        id="agree-terms-btn"
        onClick={handleAgreeToTerms}
        type="button"
        className="btn btn-primary btn-sm mt-3 w-full rounded-full font-medium uppercase"
      >
        Got It
      </button>
    </div>
  );
}
