import { sendGTMEvent } from "@next/third-parties/google";
import { useCallback } from "react";

/**
 * Custom hook for handling Google Tag Manager events
 * @returns {Object} GTM event handlers
 */
const useGTM = () => {
  /**
   * Handle click events with GTM tracking
   * @param {Object} data - The event data to send to GTM
   * @param {Function} callback - Optional callback to execute after sending event
   */
  const trackEvent = useCallback((data: any) => {
    try {
      sendGTMEvent({
        page_title: document.title,
        page_path_before: window.location.pathname,
        ...data,
      });
    } catch (error) {
      console.error("GTM Event Error:", error);
    }
  }, []);

  /**
   * Track page views with GTM
   * @param {Object} pageData - The page data to send to GTM
   */
  const trackPageView = useCallback((pageData: any) => {
    try {
      sendGTMEvent({
        event: "page_view",
        page_path: window.location.pathname,
        page_title: document.title,
        ...pageData,
      });
    } catch (error) {
      console.error("GTM Page View Error:", error);
    }
  }, []);

  return {
    trackEvent,
    trackPageView,
  };
};

export default useGTM;
