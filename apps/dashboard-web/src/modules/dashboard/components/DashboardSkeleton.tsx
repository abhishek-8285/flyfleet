import React from "react";

const Skeleton: React.FC<{ className?: string }> = ({ className }) => (
  <div className={`bg-gray-800 animate-pulse rounded-xl ${className}`} />
);

export const DashboardSkeleton: React.FC = () => {
  return (
    <div className="p-6 space-y-6 bg-black min-h-screen">
      <div className="flex justify-between items-center mb-10">
        <div className="space-y-3">
          <Skeleton className="h-8 w-64" />
          <Skeleton className="h-4 w-96" />
        </div>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        {[1, 2, 3, 4].map((i) => (
          <Skeleton key={i} className="h-32" />
        ))}
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        <div className="lg:col-span-2 space-y-6">
          <Skeleton className="h-[500px]" />
          <Skeleton className="h-[300px]" />
        </div>
        <div className="space-y-6">
          <Skeleton className="h-[200px]" />
          <Skeleton className="h-[400px]" />
        </div>
      </div>
    </div>
  );
};
