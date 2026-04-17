// components/SearchBar.tsx

import React, { useState, useEffect } from 'react';
import { useDispatch } from 'react-redux';
import { setSearchQuery } from '../store/trackingSlice';

export const SearchBar: React.FC = () => {
    const dispatch = useDispatch();
    const [inputValue, setInputValue] = useState('');

    useEffect(() => {
        const timer = setTimeout(() => {
            dispatch(setSearchQuery(inputValue));
        }, 300); // 300ms debounce

        return () => clearTimeout(timer);
    }, [inputValue, dispatch]);

    return (
        <div className="relative group">
            <span className="absolute left-6 top-1/2 -translate-y-1/2 material-symbols-outlined text-on-surface-variant text-xl opacity-50 group-focus-within:opacity-100 group-focus-within:text-primary transition-all">
                search
            </span>
            <input
                type="text"
                placeholder="Query Registry / Manifest / Registration..."
                value={inputValue}
                onChange={(e) => setInputValue(e.target.value)}
                className="w-full pl-16 pr-12 py-5 bg-surface-dim border border-outline rounded-3xl text-sm font-bold text-white placeholder:text-on-surface-variant/40 focus:outline-none focus:border-primary/50 focus:ring-4 focus:ring-primary/10 transition-all shadow-inner uppercase tracking-tighter"
            />
            {inputValue && (
                <button
                    onClick={() => setInputValue('')}
                    className="absolute right-6 top-1/2 -translate-y-1/2 text-on-surface-variant hover:text-white transition-colors h-8 w-8 flex items-center justify-center rounded-xl bg-surface-container border border-outline"
                >
                    <span className="material-symbols-outlined text-base">close</span>
                </button>
            )}
        </div>
    );
};
