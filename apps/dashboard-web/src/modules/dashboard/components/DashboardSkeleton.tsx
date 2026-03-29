import React from "react";
import { motion } from "framer-motion";

const Skeleton: React.FC<{ className?: string }> = ({ className }) => (
  <motion.div
    className={`bg-gray-800 rounded-xl ${className}`}
    initial={{ opacity: 0.6 }}
    animate={{ opacity: 1 }}
    transition={{ duration: 1.5, repeat: Infinity, repeatType: "reverse" }}
  />
);

// Enhanced shimmer skeleton with gradient animation
const ShimmerSkeleton: React.FC<{ className?: string }> = ({ className }) => {
  return (
    <motion.div
      className={`relative overflow-hidden bg-gray-800 rounded-xl ${className}`}
    >
      <motion.div
        className="absolute inset-0 bg-gradient-to-r from-transparent via-gray-700/30 to-transparent"
        initial={{ x: "-100%" }}
        animate={{ x: "100%" }}
        transition={{
          duration: 1.5,
          repeat: Infinity,
          ease: "linear",
          repeatDelay: 0.5,
        }}
      />
    </motion.div>
  );
};

export const DashboardSkeleton: React.FC = () => {
  return (
    <motion.div
      className="p-6 space-y-6 bg-black min-h-screen"
      initial={{ opacity: 0 }}
      animate={{ opacity: 1 }}
      transition={{ duration: 0.3 }}
    >
      <motion.div
        className="flex justify-between items-center mb-10"
        initial={{ opacity: 0, y: -20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ delay: 0.1, duration: 0.4 }}
      >
        <div className="space-y-3">
          <Skeleton className="h-8 w-64" />
          <Skeleton className="h-4 w-96" />
        </div>
      </motion.div>

      <motion.div
        className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4"
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ delay: 0.2, duration: 0.4 }}
      >
        {[1, 2, 3, 4].map((i) => (
          <ShimmerSkeleton key={i} className="h-32" />
        ))}
      </motion.div>

      <motion.div
        className="grid grid-cols-1 lg:grid-cols-3 gap-6"
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ delay: 0.3, duration: 0.4 }}
      >
        <div className="lg:col-span-2 space-y-6">
          <ShimmerSkeleton className="h-[500px]" />
          <ShimmerSkeleton className="h-[300px]" />
        </div>
        <div className="space-y-6">
          <ShimmerSkeleton className="h-[200px]" />
          <ShimmerSkeleton className="h-[400px]" />
        </div>
      </motion.div>
    </motion.div>
  );
};
