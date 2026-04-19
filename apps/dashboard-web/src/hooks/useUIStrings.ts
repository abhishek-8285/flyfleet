import { useEffect, useState } from "react";
import apiClient from "../services/apiClient";

export interface UIStrings {
  [key: string]: any;
}

export const useUIStrings = (section: string) => {
  const [uiStrings, setUiStrings] = useState<UIStrings | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const fetchStrings = async () => {
      try {
        const data = await apiClient.get("ui/global_config") as any;

        // Check both direct ui_strings and nested system_metadata
        const strings =
          data.ui_strings?.[section] ||
          data.system_metadata?.ui_strings?.[section];
        setUiStrings(strings || null);
      } catch (err) {
        console.error(`Error fetching UI strings for section: ${section}`, err);
      } finally {
        setLoading(false);
      }
    };

    fetchStrings();
  }, [section]);

  return { uiStrings, loading };
};
