/**
 * Multi-Route Global State Persistence Tests
 *
 * Per the FlyFleet global rule, any global feature change must be verified across:
 *   - At least 2 distinct routes
 *   - 1 navigation event
 */
/* eslint-disable testing-library/no-unnecessary-act, testing-library/no-wait-for-multiple-assertions */
import {
    act,
    fireEvent,
    render,
    screen,
    waitFor,
} from "@testing-library/react";
import React from "react";
import { MemoryRouter } from "react-router-dom";
import { AuthProvider, useAuth } from "../context/AuthProvider";

import { RBACProvider } from "../context/RBACProvider";
import { vi } from "vitest";

// Mock the API client so no real HTTP requests are made in tests
vi.mock("../services/apiClient", () => ({
  __esModule: true,
  default: {
    get: vi.fn(),
    post: vi.fn(),
    put: vi.fn(),
  },
}));

import apiClient from "../services/apiClient";

const PreferencesReadout: React.FC = () => {
  const { user, isLoading } = useAuth();
  if (isLoading) return <div data-testid="loading">Loading...</div>;
  return (
    <div>
      <span data-testid="theme">{user?.theme ?? "none"}</span>
      <span data-testid="language">{user?.language ?? "none"}</span>
      <span data-testid="sidebar">
        {String(user?.sidebar_collapsed ?? false)}
      </span>
    </div>
  );
};

const PreferencesControls: React.FC = () => {
  const { toggleTheme, setLanguage, toggleSidebar } = useAuth();
  return (
    <div>
      <button onClick={toggleTheme} data-testid="btn-theme">
        Toggle Theme
      </button>
      <button onClick={() => setLanguage("fr")} data-testid="btn-lang">
        Set French
      </button>
      <button onClick={toggleSidebar} data-testid="btn-sidebar">
        Toggle Sidebar
      </button>
    </div>
  );
};

const RouteA: React.FC = () => (
  <div data-testid="route-a">
    <PreferencesReadout />
    <PreferencesControls />
  </div>
);

const RouteB: React.FC = () => (
  <div data-testid="route-b">
    <PreferencesReadout />
  </div>
);

const NavigatingApp: React.FC = () => {
  const [route, setRoute] = React.useState<"a" | "b">("a");
  return (
    <MemoryRouter>
      <AuthProvider>
        <RBACProvider>
          <button data-testid="go-b" onClick={() => setRoute("b")}>
            Go B
          </button>
          <button data-testid="go-a" onClick={() => setRoute("a")}>
            Go A
          </button>
          {route === "a" ? <RouteA /> : <RouteB />}
        </RBACProvider>
      </AuthProvider>
    </MemoryRouter>
  );
};

const renderAndWait = async () => {
  // Clear any existing state
  window.localStorage.clear();
  // Pre-seed a session token so AuthProvider hydrates user from mock
  window.localStorage.setItem("access_token", "test-token");
  // Seed defaults to avoid race conditions with local storage fallbacks
  window.localStorage.setItem("theme", "light");
  window.localStorage.setItem("flyfleet_language", "en");
  window.localStorage.setItem("flyfleet_sidebar_collapsed", "false");

  const utils = render(<NavigatingApp />);

  // Wait for the async initAuth to resolve (isLoading goes false)
  await waitFor(
    () => {
      expect(screen.queryByTestId("loading")).not.toBeInTheDocument();
    },
    { timeout: 3000 },
  );

  return utils;
};

describe("Multi-Route Global State Persistence", () => {
  beforeEach(() => {
    (apiClient.get as any).mockImplementation(() =>
      Promise.resolve({
        user: {
          id: "test-user-id",
          name: "Test User",
          email: "test@flyfleet.io",
          theme: "light",
          language: "en",
          sidebar_collapsed: false,
        },
        permissions: [],
        orgs: [],
      }),
    );
    (apiClient.post as any).mockResolvedValue({
      token: "new-access-token",
      user: {
        id: "test-user-id",
        name: "Test User",
        email: "test@flyfleet.io",
        theme: "light",
        language: "en",
        sidebar_collapsed: false,
      },
    });
    (apiClient.put as any).mockResolvedValue({});

    window.localStorage.clear();
    window.localStorage.setItem("access_token", "test-token");
    document.documentElement.classList.remove("dark");
    vi.clearAllMocks();
  });

  it("Route A: theme starts as light from mock DB value", async () => {
    await renderAndWait();
    expect(screen.getByTestId("theme").textContent).toBe("light");
  });

  it("Theme toggle persists across Route A → Route B navigation", async () => {
    await renderAndWait();

    // Toggle to dark on Route A
    await act(async () => {
      fireEvent.click(screen.getByTestId("btn-theme"));
    });

    await waitFor(() => {
      expect(screen.getByTestId("theme").textContent).toBe("dark");
      expect(document.documentElement.classList.contains("dark")).toBe(true);
    });

    // Navigate to Route B
    await act(async () => {
      fireEvent.click(screen.getByTestId("go-b"));
    });

    await waitFor(() => {
      expect(screen.getByTestId("route-b")).toBeInTheDocument();
      expect(screen.getByTestId("theme").textContent).toBe("dark");
    });
  });

  it("Language change persists across Route A → Route B → Route A", async () => {
    await renderAndWait();

    await act(async () => {
      fireEvent.click(screen.getByTestId("btn-lang"));
    });

    await waitFor(() => {
      expect(screen.getByTestId("language").textContent).toBe("fr");
    });

    // Navigate to Route B
    await act(async () => {
      fireEvent.click(screen.getByTestId("go-b"));
    });
    await waitFor(() =>
      expect(screen.getByTestId("language").textContent).toBe("fr"),
    );

    // Navigate back to Route A
    await act(async () => {
      fireEvent.click(screen.getByTestId("go-a"));
    });
    await waitFor(() =>
      expect(screen.getByTestId("language").textContent).toBe("fr"),
    );
  });

  it("Sidebar collapsed state persists across navigation", async () => {
    await renderAndWait();

    await act(async () => {
      fireEvent.click(screen.getByTestId("btn-sidebar"));
    });

    await waitFor(() => {
      expect(screen.getByTestId("sidebar").textContent).toBe("true");
    });

    await act(async () => {
      fireEvent.click(screen.getByTestId("go-b"));
    });
    await waitFor(() =>
      expect(screen.getByTestId("sidebar").textContent).toBe("true"),
    );
  });

  it("localStorage is kept in sync as a fallback for all three preferences", async () => {
    await renderAndWait();

    await act(async () => {
      fireEvent.click(screen.getByTestId("btn-theme"));
    });
    await waitFor(() => {
      expect(window.localStorage.getItem("theme")).toBe("dark");
    });

    await act(async () => {
      fireEvent.click(screen.getByTestId("btn-lang"));
    });
    await waitFor(() => {
      expect(window.localStorage.getItem("flyfleet_language")).toBe("fr");
    });

    await act(async () => {
      fireEvent.click(screen.getByTestId("btn-sidebar"));
    });
    await waitFor(() => {
      expect(window.localStorage.getItem("flyfleet_sidebar_collapsed")).toBe("true");
    });
  });
});
