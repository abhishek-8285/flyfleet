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
        <div className="relative">
            <input
                type="text"
                placeholder="Search vehicles by registration..."
                value={inputValue}
                onChange={(e) => setInputValue(e.target.value)}
                className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
            />
            {inputValue && (
                <button
                    onClick={() => setInputValue('')}
                    className="absolute right-2 top-1/2 -translate-y-1/2 text-gray-400 hover:text-gray-600"
                >
                    ✕
                </button>
            )}
        </div>
    );
};
