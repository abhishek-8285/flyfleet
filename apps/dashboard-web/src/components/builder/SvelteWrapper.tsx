import React, { useEffect, useRef } from 'react';
import { mount, unmount } from 'svelte';
// @ts-ignore
import InvoiceBuilder from './InvoiceBuilderCompiled.js';
import './InvoiceBuilderCompiled.css';

interface SvelteWrapperProps {
  customerName?: string;
  initialAmount?: number;
}

const SvelteWrapper: React.FC<SvelteWrapperProps> = ({ customerName, initialAmount }) => {
  const containerRef = useRef<HTMLDivElement>(null);
  const svelteAppRef = useRef<any>(null);

  useEffect(() => {
    if (containerRef.current && !svelteAppRef.current) {
      // Mount the Svelte component
      try {
        const Component = (InvoiceBuilder as any).default || InvoiceBuilder;
        svelteAppRef.current = mount(Component, {
          target: containerRef.current,
          props: {
            customerName: customerName || "Dynamic Enterprise",
            amount: initialAmount || 75000
          }
        });
      } catch (e) {
        console.error("Failed to mount Svelte", e);
      }
    }

    return () => {
      // Clean up Svelte when React component unmounts
      if (svelteAppRef.current) {
        try {
          unmount(svelteAppRef.current);
        } catch(e) {}
        svelteAppRef.current = null;
      }
    };
  }, [customerName, initialAmount]);

  return (
    <div className="svelte-host-container">
      <div className="flex justify-between items-center mb-4">
        <h2 className="text-xl font-bold">Interactive Template Designer</h2>
        <span className="bg-orange-100 text-orange-700 px-3 py-1 rounded-full text-xs font-bold">
          Svelte Powered
        </span>
      </div>
      <div ref={containerRef}></div>
    </div>
  );
};

export default SvelteWrapper;
