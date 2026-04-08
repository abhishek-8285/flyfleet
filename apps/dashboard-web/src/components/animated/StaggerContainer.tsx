import React, { Children, cloneElement, isValidElement, ReactElement, useEffect, useState } from 'react';
import { motion, AnimatePresence, Variants } from 'framer-motion';

interface StaggerContainerProps {
  children: React.ReactNode;
  staggerDelay?: number;
  maxDelay?: number;
  className?: string;
  initial?: 'hidden' | 'visible';
}

/**
 * StaggerContainer wraps a list of children and applies staggered entrance animations.
 * Uses Framer Motion's staggerChildren pattern.
 *
 * Usage:
 * <StaggerContainer staggerDelay={0.1}>
 *   {items.map(item => <Card key={item.id} />)}
 * </StaggerContainer>
 */
export const StaggerContainer: React.FC<StaggerContainerProps> = ({
  children,
  staggerDelay = 0.1,
  maxDelay = 0.5, // Cap total stagger at 500ms
  className = '',
  initial = 'hidden'
}) => {
  const [shouldReduceMotion, setShouldReduceMotion] = React.useState(false);

  useEffect(() => {
    const mediaQuery = window.matchMedia('(prefers-reduced-motion: reduce)');
    setShouldReduceMotion(mediaQuery.matches);

    const handler = (e: MediaQueryListEvent) => {
      setShouldReduceMotion(e.matches);
    };

    mediaQuery.addEventListener('change', handler);
    return () => mediaQuery.removeEventListener('change', handler);
  }, []);

  const containerVariants: Variants = {
    hidden: {},
    visible: {
      transition: {
        staggerChildren: shouldReduceMotion ? 0 : staggerDelay,
        delayChildren: 0
      }
    }
  };

  const childVariants: Variants = {
    hidden: {
      opacity: 0,
      y: 20,
      scale: 0.98
    },
    visible: {
      opacity: 1,
      y: 0,
      scale: 1,
      transition: {
        duration: shouldReduceMotion ? 0 : 0.4,
        ease: [0.16, 1, 0.3, 1] // cubic-bezier equivalent
      }
    }
  };

  const childrenArray = Children.toArray(children);
  const cappedStagger = Math.min(staggerDelay, maxDelay / childrenArray.length);

  return (
    <motion.div
      variants={containerVariants}
      initial={shouldReduceMotion ? 'visible' : initial}
      animate="visible"
      className={className}
    >
      <AnimatePresence mode="popLayout">
        {childrenArray.map((child, index) =>
          isValidElement(child) ? (
            <motion.div
              key={child.key || index}
              variants={childVariants}
              style={{
                transitionDelay: shouldReduceMotion ? undefined : `calc(${cappedStagger * index}s * 0.5)`
              }}
            >
              {child}
            </motion.div>
          ) : (
            child
          )
        )}
      </AnimatePresence>
    </motion.div>
  );
};

export default StaggerContainer;
