import { useEffect, useRef } from "react";
import { motion, useMotionValue, useSpring } from "framer-motion";
import React from "react";

/**
 * MagneticButton - A button with a subtle magnetic hover effect.
 * The button slightly follows the mouse cursor when nearby, creating a premium interactive feel.
 * Respects prefers-reduced-motion via CSS fallback.
 *
 * Note: This is a button element only. For navigation, pass an onClick handler that uses navigate().
 */
const MagneticButton: React.FC<{
  onClick?: () => void;
  children: React.ReactNode;
  variant?: "primary" | "secondary";
  className?: string;
  disabled?: boolean;
  type?: "button" | "submit" | "reset";
}> = ({
  onClick,
  children,
  variant = "primary",
  className = "",
  disabled = false,
  type = "button",
}) => {
  const buttonRef = useRef<HTMLButtonElement>(null);
  const x = useMotionValue(0);
  const y = useMotionValue(0);
  const springConfig = { damping: 15, stiffness: 150 };
  const xSpring = useSpring(x, springConfig);
  const ySpring = useSpring(y, springConfig);

  const handleMouseMove = (e: React.MouseEvent<HTMLButtonElement>) => {
    if (disabled) return;

    const rect = buttonRef.current?.getBoundingClientRect();
    if (!rect) return;

    const centerX = rect.left + rect.width / 2;
    const centerY = rect.top + rect.height / 2;
    const mouseX = e.clientX - centerX;
    const mouseY = e.clientY - centerY;
    // Magnetic pull factor - subtle movement
    const pullFactor = 0.2;
    x.set(mouseX * pullFactor);
    y.set(mouseY * pullFactor);
  };

  const handleMouseLeave = () => {
    x.set(0);
    y.set(0);
  };

  // Respect reduced motion preference
  const [shouldReduceMotion, setShouldReduceMotion] = React.useState(false);

  useEffect(() => {
    const mediaQuery = window.matchMedia("(prefers-reduced-motion: reduce)");
    setShouldReduceMotion(mediaQuery.matches);

    const handler = (e: MediaQueryListEvent) => {
      setShouldReduceMotion(e.matches);
    };

    mediaQuery.addEventListener("change", handler);
    return () => mediaQuery.removeEventListener("change", handler);
  }, []);

  const baseStyles =
    "inline-flex items-center justify-center gap-2 px-6 py-3 rounded-xl font-semibold transition-all focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-offset-gray-900 disabled:opacity-50 disabled:cursor-not-allowed";

  const variantStyles = {
    primary:
      "bg-primary-600 hover:bg-primary-700 text-white shadow-lg shadow-primary-500/30 focus:ring-primary-500",
    secondary:
      "bg-gray-800 hover:bg-gray-700 text-gray-200 border border-gray-700 focus:ring-gray-500",
  };

  // If reduced motion, disable magnetic effect
  const motionStyle = shouldReduceMotion ? {} : { x: xSpring, y: ySpring };

  return (
    <motion.button
      ref={buttonRef}
      type={type}
      onClick={onClick}
      style={motionStyle}
      onMouseMove={handleMouseMove}
      onMouseLeave={handleMouseLeave}
      whileTap={{ scale: 0.97 }}
      className={`${baseStyles} ${variantStyles[variant]} ${className}`}
      aria-disabled={disabled}
    >
      {children}
    </motion.button>
  );
};

export default MagneticButton;
