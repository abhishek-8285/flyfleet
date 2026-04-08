import React from 'react';
import { motion, HTMLMotionProps } from 'framer-motion';

/**
 * AnimatedCard - A card component with hover lift effect and entrance animation.
 * Extends all motion.div props for flexibility.
 *
 * Usage:
 * <AnimatedCard whileHover={{ y: -4 }} className="bg-surface-dark...">
 *   <h3>Title</h3>
 *   <p>Content</p>
 * </AnimatedCard>
 */
export const AnimatedCard: React.FC<HTMLMotionProps<"div">> = ({
  children,
  className = '',
  initial = { opacity: 0, y: 20 },
  animate = { opacity: 1, y: 0 },
  transition = { duration: 0.4, ease: [0.16, 1, 0.3, 1] },
  whileHover = {
    y: -4,
    boxShadow: '0 20px 40px rgba(0, 0, 0, 0.15), 0 0 0 1px rgba(255, 255, 255, 0.05)'
  },
  ...props
}) => {
  return (
    <motion.div
      className={className}
      initial={initial}
      animate={animate}
      transition={transition}
      whileHover={whileHover}
      {...props}
    >
      {children}
    </motion.div>
  );
};

/**
 * StatCard variant - specifically for dashboard KPI cards
 * Has a floating effect on hover with border glow
 */
export const StatCard: React.FC<HTMLMotionProps<"div">> = ({
  children,
  className = '',
  ...props
}) => {
  return (
    <AnimatedCard
      className={`bg-surface-dark rounded-xl p-5 border border-border-color shadow-sm relative overflow-visible group ${className}`}
      whileHover={{
        y: -6,
        boxShadow: '0 25px 50px -12px rgba(19, 127, 236, 0.15)',
        borderColor: 'rgba(19, 127, 236, 0.3)'
      }}
      transition={{
        type: 'spring',
        stiffness: 300,
        damping: 20
      }}
      {...props}
    >
      {children}
    </AnimatedCard>
  );
};

export default AnimatedCard;
