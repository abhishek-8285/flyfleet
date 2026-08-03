import type { Metadata } from "next";
import { Inter, Outfit } from "next/font/google";
import "./globals.css";

const inter = Inter({
  variable: "--font-inter",
  subsets: ["latin"],
});

const outfit = Outfit({
  variable: "--font-outfit",
  subsets: ["latin"],
});

export const metadata: Metadata = {
  title: {
    default: "FlyFleet | Commercial Fleet Management Platform",
    template: "%s | FlyFleet",
  },
  description:
    "Secure, scalable, and compliant fleet management for logistics leaders. Reduce costs and mitigate risk with AI-driven route optimization and real-time tracking.",
  keywords: [
    "fleet management",
    "telematics",
    "gps tracking",
    "route optimization",
    "logistics technology",
    "commercial fleet",
  ],
  authors: [{ name: "FlyFleet Command" }],
  creator: "FlyFleet",
  publisher: "FlyFleet Inc.",
  formatDetection: {
    email: false,
    address: false,
    telephone: false,
  },
  openGraph: {
    title: "FlyFleet | The Trust Standard for Commercial Fleets",
    description:
      "Next-generation fleet management with 99.9% uptime SLA and ISO 27001 security.",
    url: "https://flyfleet.io",
    siteName: "FlyFleet",
    images: [
      {
        url: "https://flyfleet.io/og-image.png",
        width: 1200,
        height: 630,
        alt: "FlyFleet Command Dashboard Preview",
      },
    ],
    locale: "en_US",
    type: "website",
  },
  twitter: {
    card: "summary_large_image",
    title: "FlyFleet Command",
    description: "Transform your logistics with AI-driven fleet optimization.",
    creator: "@flyfleet",
    images: ["https://flyfleet.io/twitter-image.png"],
  },
  robots: {
    index: true,
    follow: true,
    googleBot: {
      index: true,
      follow: true,
      "max-video-preview": -1,
      "max-image-preview": "large",
      "max-snippet": -1,
    },
  },
  icons: {
    icon: "/favicon.ico",
    shortcut: "/shortcut-icon.png",
    apple: "/apple-touch-icon.png",
  },
  verification: {
    google: "C79XYq2b1ZDDpm2lDWImcI6VUPHh31JrXxYA2Ryv8mo",
  },
  manifest: "/site.webmanifest",
};

import CookieConsent from "@/components/landing/CookieConsent";

import { ThemeProvider } from "@/context/ThemeContext";

import { CSPostHogProvider } from "@/components/providers/PostHogProvider";

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en" suppressHydrationWarning={true}>
      <head>
        <link
          rel="stylesheet"
          href="https://fonts.googleapis.com/css2?family=Material+Symbols+Outlined:opsz,wght,FILL,GRAD@20..48,100..700,0..1,-50..200"
        />
        <script src="https://cdn.jsdelivr.net/npm/gsap@3.12.5/dist/gsap.min.js"></script>
        <script src="https://cdn.jsdelivr.net/npm/gsap@3.12.5/dist/ScrollTrigger.min.js"></script>
        <script src="https://cdn.jsdelivr.net/gh/studio-freight/lenis@1.0.29/bundled/lenis.min.js"></script>
        <script dangerouslySetInnerHTML={{
          __html: `
            try {
              gsap.registerPlugin(ScrollTrigger);

              // Initialize Lenis smooth scrolling
              const lenis = new Lenis({
                duration: 1.2,
                easing: (t) => Math.min(1, 1.001 - Math.pow(2, -10 * t)),
                direction: 'vertical',
                gestureOrientation: 'vertical',
                smooth: true,
                mouseMultiplier: 1,
                smoothTouch: false,
                touchMultiplier: 2,
                infinite: false,
              });

              function raf(time) {
                lenis.raf(time);
                requestAnimationFrame(raf);
              }

              requestAnimationFrame(raf);

              // Integrate Lenis with ScrollTrigger
              lenis.on('scroll', ScrollTrigger.update);

              gsap.ticker.add((time) => {
                lenis.raf(time * 1000);
              });

              gsap.ticker.lagSmoothing(0);
            } catch (e) {
              console.warn('Smooth scroll setup failed:', e);
            }
          `
        }}></script>
      </head>
      <body className={`${inter.variable} ${outfit.variable} antialiased`}>
        <CSPostHogProvider>
          <ThemeProvider>
            {children}
            <CookieConsent />
          </ThemeProvider>
        </CSPostHogProvider>
      </body>
    </html>
  );
}
