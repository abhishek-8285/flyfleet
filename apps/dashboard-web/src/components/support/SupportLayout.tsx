import React from 'react';
import SupportSidebar from './SupportSidebar';

interface SupportLayoutProps {
  children: React.ReactNode;
}

const SupportLayout: React.FC<SupportLayoutProps> = ({ children }) => {
  return (
    <div className="flex h-screen w-full bg-[#0b1219] text-white">
      <SupportSidebar />
      <div className="flex-1 flex flex-col min-w-0 bg-[#0b1219] overflow-hidden">
        {children}
      </div>
    </div>
  );
};

export default SupportLayout;
