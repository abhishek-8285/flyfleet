import { useEffect, useState } from 'react';

/**
 * Landing page data types
 */
export interface SuccessStory {
  id: string;
  title: string;
  value: string;
  quote: string;
  company: string;
  fleet: string;
  icon: string;
  color: string;
  badge: string;
  badgeColor: string;
}

export interface Feature {
  id: string;
  title: string;
  desc: string;
  icon: string;
  color: string;
}

export interface Integration {
  id: string;
  name: string;
  icon: string;
  color: string;
}

export interface SecurityItem {
  id: string;
  title: string;
  desc: string;
  icon: string;
  color: string;
}

export interface HardwareAddon {
  id: string;
  name: string;
  price: string;
  icon: string;
  color: string;
}

export interface FAQ {
  id: string;
  question: string;
  answer: string;
}

export interface TrustedBrand {
  id: string;
  name: string;
  icon: string;
}

export interface TimelineItem {
  id: string;
  quarter: string;
  year: number;
  title: string;
  desc: string;
  icon: string;
  color: string;
}

export interface PricingTier {
  id: string;
  name: string;
  price: number;
  period: string;
  features: string[];
  cta: string;
  popular?: boolean;
  icon: string;
}

export interface LandingMetadata {
  sales_info?: {
    phone?: string;
  };
  user_metadata?: {
    full_name?: string;
  };
  [key: string]: unknown;
}

export interface SystemMetadata {
  location?: string;
  [key: string]: unknown;
}

export interface MarketingMetadata {
  conversion_rate?: string;
  current_temp?: string;
  driver_name?: string;
  [key: string]: unknown;
}

/**
 * Data returned by the useLandingData hook
 */
export interface LandingData {
  successStories: SuccessStory[];
  features: Feature[];
  integrations: Integration[];
  securityItems: SecurityItem[];
  hardwareAddons: HardwareAddon[];
  faqs: FAQ[];
  trustedBy: TrustedBrand[];
  timeline: TimelineItem[];
  pricing: PricingTier[];
  landingMetadata: LandingMetadata | null;
  systemMeta: SystemMetadata | null;
  marketingMeta: MarketingMetadata | null;
}

/**
 * State for the useLandingData hook
 */
interface UseLandingDataState {
  data: LandingData;
  isLoading: boolean;
  error: Error | null;
}

const INITIAL_STATE: UseLandingDataState = {
  data: {
    successStories: [],
    features: [],
    integrations: [],
    securityItems: [],
    hardwareAddons: [],
    faqs: [],
    trustedBy: [],
    timeline: [],
    pricing: [],
    landingMetadata: null,
    systemMeta: null,
    marketingMeta: null,
  },
  isLoading: true,
  error: null,
};

/**
 * Custom hook for fetching landing page data
 * Handles data fetching, loading states, and error handling
 */
export function useLandingData() {
  const [state, setState] = useState<UseLandingDataState>(INITIAL_STATE);

  useEffect(() => {
    let isMounted = true;

    const fetchData = async () => {
      try {
        // Fetch landing data
        const dataResponse = await fetch('/api/v1/data');
        if (!dataResponse.ok) {
          throw new Error(`Failed to fetch landing data: ${dataResponse.status}`);
        }
        const data = await dataResponse.json();

        // Fetch UI config
        const configResponse = await fetch('/api/v1/ui/global_config');
        if (!configResponse.ok) {
          throw new Error(`Failed to fetch UI config: ${configResponse.status}`);
        }
        const config = await configResponse.json();

        if (isMounted) {
          setState({
            data: {
              successStories: data.successStories || [],
              features: data.features || [],
              integrations: data.integrations || [],
              securityItems: data.securityItems || [],
              hardwareAddons: data.hardwareAddons || [],
              faqs: data.faqs || [],
              trustedBy: data.trustedBy || [],
              timeline: data.timeline || [],
              pricing: data.pricing || [],
              landingMetadata: config.landing_metadata || config,
              systemMeta: config.system_metadata,
              marketingMeta: config.marketing_metadata,
            },
            isLoading: false,
            error: null,
          });
        }
      } catch (error) {
        if (isMounted) {
          setState({
            ...INITIAL_STATE,
            error: error instanceof Error ? error : new Error('Unknown error'),
          });
        }
      }
    };

    fetchData();

    return () => {
      isMounted = false;
    };
  }, []);

  return state;
}

/**
 * Custom hook for theme toggle functionality
 */
export function useTheme(initialDark = false) {
  const [isDark, setIsDark] = useState(initialDark);

  useEffect(() => {
    if (isDark) {
      document.documentElement.classList.add('dark');
    } else {
      document.documentElement.classList.remove('dark');
    }
  }, [isDark]);

  const toggle = () => setIsDark((prev) => !prev);

  return { isDark, toggle };
}

/**
 * Custom hook for pricing toggle (monthly/yearly)
 */
export function usePricingToggle(initialYearly = false) {
  const [isYearly, setIsYearly] = useState(initialYearly);

  const toggle = () => setIsYearly((prev) => !prev);

  return { isYearly, toggle };
}
