import { useEffect, useState } from "react";

export interface UIStrings {
  [key: string]: any;
}

export const useUIStrings = (section: string) => {
  const [uiStrings, setUiStrings] = useState<UIStrings | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const fetchStrings = async () => {
      try {
        const res = await fetch("/api/v1/ui/global_config");
        if (!res.ok) throw new Error("Failed to fetch UI strings");
        const data = await res.json();

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
