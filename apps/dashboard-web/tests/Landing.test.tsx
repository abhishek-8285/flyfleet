import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { BrowserRouter } from 'react-router-dom';
import Landing from '../src/pages/Landing';

// Mock react-router-dom
vi.mock('react-router-dom', () => ({
  useNavigate: vi.fn(),
  BrowserRouter: vi.fn(({ children }) => children),
}));

// Mock hooks
vi.mock('../src/hooks/useUIStrings', () => ({
  useUIStrings: () => ({
    uiStrings: {
      landing: {
        security_text: 'Operations Grade Security',
        sales_btn: 'Talk to Sales',
        locale_en: 'US (EN)',
        success_stories: 'Success Stories',
        platform: 'Platform',
        apps: 'Apps',
        pricing: 'Pricing',
        security: 'Security',
        login_btn: 'Login',
        trial_btn: 'Start Free Trial',
        security_badge: 'Operations Grade Security & Reliability',
        hero_title: 'The Trust Standard for Commercial Fleets',
        hero_sub: 'Secure, scalable, and compliant fleet management.',
        demo_btn: 'Request Demo',
        tour_btn: 'Watch Product Tour',
        trusted_title: 'Trusted by Global Logistics Leaders',
        success_title: 'Commercial Success Stories',
        success_sub: 'Real results from companies scaling their operations.',
      },
      common: {
        search_placeholder: 'Search Fleet...',
      },
    },
  }),
}));

// Mock branding config
vi.mock('@flyfleet/config', () => ({
  default: {
    companyName: 'FlyFleet',
  },
}));

describe('Landing Page', () => {
  const mockNavigate = vi.fn();

  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(useNavigate).mockReturnValue(mockNavigate);
  });

  describe('TopBar Component', () => {
    it('renders security badge with icon', () => {
      render(<Landing />);
      
      expect(screen.getByText('Operations Grade Security')).toBeInTheDocument();
      expect(screen.getByTestId('security-icon')).toBeInTheDocument();
    });

    it('renders sales phone number', () => {
      render(<Landing />);
      
      expect(screen.getByText(/Talk to Sales/i)).toBeInTheDocument();
    });

    it('renders language selector', () => {
      render(<Landing />);
      
      expect(screen.getByText(/US \(EN\)/i)).toBeInTheDocument();
    });
  });

  describe('Header Component', () => {
    it('renders FlyFleet logo and name', () => {
      render(<Landing />);
      
      expect(screen.getByText('FlyFleet')).toBeInTheDocument();
    });

    it('renders navigation links', () => {
      render(<Landing />);
      
      expect(screen.getByText(/Success Stories/i)).toBeInTheDocument();
      expect(screen.getByText(/Platform/i)).toBeInTheDocument();
      expect(screen.getByText(/Apps/i)).toBeInTheDocument();
      expect(screen.getByText(/Pricing/i)).toBeInTheDocument();
      expect(screen.getByText(/Security/i)).toBeInTheDocument();
    });

    it('renders theme toggle button', () => {
      render(<Landing />);
      
      expect(screen.getByRole('button', { name: /theme/i })).toBeInTheDocument();
    });

    it('renders login button that navigates to /login', () => {
      render(<Landing />);
      
      const loginButton = screen.getByText(/Login/i);
      fireEvent.click(loginButton);
      
      expect(mockNavigate).toHaveBeenCalledWith('/login');
    });

    it('renders trial button', () => {
      render(<Landing />);
      
      expect(screen.getByText(/Start Free Trial/i)).toBeInTheDocument();
    });
  });

  describe('Hero Section', () => {
    it('renders security badge', () => {
      render(<Landing />);
      
      expect(screen.getByText(/Operations Grade Security & Reliability/i)).toBeInTheDocument();
    });

    it('renders hero title', () => {
      render(<Landing />);
      
      expect(screen.getByText(/The Trust Standard for Commercial Fleets/i)).toBeInTheDocument();
    });

    it('renders hero subtitle', () => {
      render(<Landing />);
      
      expect(screen.getByText(/Secure, scalable, and compliant fleet management/i)).toBeInTheDocument();
    });

    it('renders demo request button', () => {
      render(<Landing />);
      
      expect(screen.getByText(/Request Demo/i)).toBeInTheDocument();
    });

    it('renders product tour button', () => {
      render(<Landing />);
      
      expect(screen.getByText(/Watch Product Tour/i)).toBeInTheDocument();
    });
  });

  describe('Theme Toggle', () => {
    it('toggles dark mode when clicked', () => {
      render(<Landing />);
      
      const themeToggle = screen.getByRole('button', { name: /theme/i });
      
      // Initial state should be light mode
      expect(document.documentElement.classList.contains('dark')).toBe(false);
      
      // Click to toggle
      fireEvent.click(themeToggle);
      
      // Should now be dark mode
      expect(document.documentElement.classList.contains('dark')).toBe(true);
    });

    it('toggles back to light mode when clicked again', () => {
      render(<Landing />);
      
      const themeToggle = screen.getByRole('button', { name: /theme/i });
      
      // Toggle to dark
      fireEvent.click(themeToggle);
      expect(document.documentElement.classList.contains('dark')).toBe(true);
      
      // Toggle back to light
      fireEvent.click(themeToggle);
      expect(document.documentElement.classList.contains('dark')).toBe(false);
    });
  });

  describe('Data Fetching', () => {
    it('fetches landing data on mount', async () => {
      const mockData = {
        successStories: [{ title: 'Story 1' }],
        features: [{ title: 'Feature 1' }],
        trustedBy: [{ name: 'Brand 1' }],
      };

      global.fetch = vi.fn().mockResolvedValueOnce({
        json: () => Promise.resolve(mockData),
      }).mockResolvedValueOnce({
        json: () => Promise.resolve({}),
      });

      render(<Landing />);

      await waitFor(() => {
        expect(global.fetch).toHaveBeenCalledWith('/api/v1/data');
      });
    });

    it('fetches UI config on mount', async () => {
      const mockConfig = {
        landing_metadata: { title: 'Test' },
      };

      global.fetch = vi.fn()
        .mockResolvedValueOnce({ json: () => Promise.resolve({}) })
        .mockResolvedValueOnce({ json: () => Promise.resolve(mockConfig) });

      render(<Landing />);

      await waitFor(() => {
        expect(global.fetch).toHaveBeenCalledWith('/api/v1/ui/global_config');
      });
    });

    it('handles fetch errors gracefully', async () => {
      global.fetch = vi.fn()
        .mockRejectedValueOnce(new Error('Network error'))
        .mockRejectedValueOnce(new Error('Network error'));

      const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {});

      render(<Landing />);

      await waitFor(() => {
        expect(consoleSpy).toHaveBeenCalled();
      });

      consoleSpy.mockRestore();
    });
  });

  describe('Responsive Behavior', () => {
    it('hides navigation on mobile', () => {
      // Mock mobile viewport
      Object.defineProperty(window, 'innerWidth', {
        writable: true,
        configurable: true,
        value: 600,
      });

      render(<Landing />);

      // Navigation should be hidden on mobile
      const nav = screen.queryByRole('navigation');
      expect(nav?.classList.contains('hidden')).toBe(true);
    });

    it('shows navigation on desktop', () => {
      // Mock desktop viewport
      Object.defineProperty(window, 'innerWidth', {
        writable: true,
        configurable: true,
        value: 1200,
      });

      render(<Landing />);

      // Navigation should be visible on desktop
      const nav = screen.queryByRole('navigation');
      expect(nav?.classList.contains('hidden')).toBe(false);
    });
  });

  describe('Accessibility', () => {
    it('has proper heading hierarchy', () => {
      render(<Landing />);
      
      expect(screen.getByRole('heading', { level: 1 })).toBeInTheDocument();
    });

    it('has accessible buttons', () => {
      render(<Landing />);
      
      const buttons = screen.getAllByRole('button');
      expect(buttons.length).toBeGreaterThan(0);
    });

    it('has proper link text', () => {
      render(<Landing />);
      
      const links = screen.getAllByRole('link');
      expect(links.length).toBeGreaterThan(0);
    });
  });
});
