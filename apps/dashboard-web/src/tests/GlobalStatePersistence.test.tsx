/**
 * Multi-Route Global State Persistence Tests
 *
 * Per the FlyFleet global rule, any global feature change must be verified across:
 *   - At least 2 distinct routes
 *   - 1 navigation event
 */
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

// Mock the API client so no real HTTP requests are made in tests
jest.mock("../services/apiClient", () => ({
  __esModule: true,
  default: {
    get: jest.fn().mockResolvedValue({
      user: {
        id: "test-user-id",
        name: "Test User",
        email: "test@flyfleet.io",
        theme: "light",
        language: "en",
        sidebar_collapsed: false,
        permissions: [],
      },
      permissions: [],
      orgs: [],
    }),
    put: jest.fn().mockResolvedValue({}),
  },
}));

const PreferencesReadout: React.FC = () => {
  const { user } = useAuth();
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
        <button data-testid="go-b" onClick={() => setRoute("b")}>
          Go B
        </button>
        <button data-testid="go-a" onClick={() => setRoute("a")}>
          Go A
        </button>
        {route === "a" ? <RouteA /> : <RouteB />}
      </AuthProvider>
    </MemoryRouter>
  );
};

const renderAndWait = async () => {
  // Pre-seed a session token so AuthProvider hydrates user from mock
  localStorage.setItem("access_token", "test-token");
  const utils = render(<NavigatingApp />);
  // Wait for the async initAuth to resolve
  await waitFor(() => screen.getByTestId("theme"));
  return utils;
};

describe("Multi-Route Global State Persistence", () => {
  beforeEach(() => {
    localStorage.clear();
    document.documentElement.classList.remove("dark");
  });

  afterEach(() => {
    jest.clearAllMocks();
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
    expect(screen.getByTestId("theme").textContent).toBe("dark");
    expect(document.documentElement.classList.contains("dark")).toBe(true);

    // Navigate to Route B
    await act(async () => {
      fireEvent.click(screen.getByTestId("go-b"));
    });

    await waitFor(() => {
      expect(screen.getByTestId("route-b")).toBeInTheDocument();
      // Theme must still be dark — context is shared, not re-created
      expect(screen.getByTestId("theme").textContent).toBe("dark");
    });
  });

  it("Language change persists across Route A → Route B → Route A", async () => {
    await renderAndWait();

    await act(async () => {
      fireEvent.click(screen.getByTestId("btn-lang"));
    });
    expect(screen.getByTestId("language").textContent).toBe("fr");

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
    expect(screen.getByTestId("sidebar").textContent).toBe("true");

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
    expect(localStorage.getItem("theme")).toBe("dark");

    await act(async () => {
      fireEvent.click(screen.getByTestId("btn-lang"));
    });
    expect(localStorage.getItem("flyfleet_language")).toBe("fr");

    await act(async () => {
      fireEvent.click(screen.getByTestId("btn-sidebar"));
    });
    expect(localStorage.getItem("flyfleet_sidebar_collapsed")).toBe("true");
  });
});
