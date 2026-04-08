import React, { useEffect } from 'react';
import { motion, AnimatePresence, LayoutGroup, Variants } from 'framer-motion';
import { useLocation } from 'react-router-dom';

/**
 * PageTransition wrapper for route changes.
 * Respects prefers-reduced-motion and provides smooth slide/fade transitions.
 *
 * Usage in App.tsx:
 * <AnimatePresence mode="wait">
 *   <Routes location={location}>
 *     <Route path="/dashboard" element={
 *       <PageTransition>
 *         <Dashboard />
 *       </PageTransition>
 *     } />
 *   </Routes>
 * </AnimatePresence>
 */

const pageVariants: Variants = {
  initial: {
    opacity: 0,
    x: 20,
    y: 0,
    scale: 0.998
  },
  animate: {
    opacity: 1,
    x: 0,
    y: 0,
    scale: 1,
    transition: {
      duration: 0.3,
      ease: [0.16, 1, 0.3, 1]
    }
  },
  exit: {
    opacity: 0,
    x: -20,
    y: 0,
    scale: 0.998,
    transition: {
      duration: 0.25,
      ease: [0.16, 1, 0.3, 1]
    }
  }
};

interface PageTransitionProps {
  children: React.ReactNode;
  variant?: 'slide' | 'fade' | 'scale';
}

export const PageTransition: React.FC<PageTransitionProps> = ({
  children,
  variant = 'slide'
}) => {
  const [shouldAnimate, setShouldAnimate] = React.useState(true);

  useEffect(() => {
    // Check reduced motion preference
    const mediaQuery = window.matchMedia('(prefers-reduced-motion: reduce)');
    if (mediaQuery.matches) {
      setShouldAnimate(false);
    }

    const handleChange = (e: MediaQueryListEvent) => {
      setShouldAnimate(!e.matches);
    };

    mediaQuery.addEventListener('change', handleChange);
    return () => mediaQuery.removeEventListener('change', handleChange);
  }, []);

  if (!shouldAnimate) {
    return <>{children}</>;
  }

  const getVariants = () => {
    switch (variant) {
      case 'fade':
        return {
          initial: { opacity: 0 },
          animate: { opacity: 1, transition: { duration: 0.2 } },
          exit: { opacity: 0, transition: { duration: 0.15 } }
        };
      case 'scale':
        return {
          initial: { opacity: 0, scale: 0.98 },
          animate: { opacity: 1, scale: 1, transition: { duration: 0.3 } },
          exit: { opacity: 0, scale: 0.98, transition: { duration: 0.2 } }
        };
      case 'slide':
      default:
        return pageVariants;
    }
  };

  return (
    <motion.div
      variants={getVariants()}
      initial="initial"
      animate="animate"
      exit="exit"
    >
      {children}
    </motion.div>
  );
};

/**
 * RouteTransition component that wraps Routes with AnimatePresence
 * Handles page transitions automatically based on location changes
 *
 * Usage:
 * <RouteTransition>
 *   <Routes location={location}>
 *     <Route path="/dashboard" element={<Dashboard />} />
 *   </Routes>
 * </RouteTransition>
 */
interface RouteTransitionProps {
  children: React.ReactNode;
  className?: string;
}

export const RouteTransition: React.FC<RouteTransitionProps> = ({
  children,
  className = ''
}) => {
  const location = useLocation();

  useEffect(() => {
    // Scroll to top on route change
    window.scrollTo(0, 0);
  }, [location.pathname]);

  return (
    <AnimatePresence mode="wait" initial={false}>
      <motion.div
        key={location.pathname}
        className={className}
        variants={pageVariants}
        initial="initial"
        animate="animate"
        exit="exit"
      >
        {children}
      </motion.div>
    </AnimatePresence>
  );
};

export default PageTransition;
