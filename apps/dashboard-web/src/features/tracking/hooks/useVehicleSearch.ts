// hooks/useVehicleSearch.ts

import { useState, useEffect, useCallback } from 'react';
import { useDispatch } from 'react-redux';
import { setSearchQuery } from '../store/trackingSlice';

export function useVehicleSearch() {
    const dispatch = useDispatch();
    const [inputValue, setInputValue] = useState('');

    const debouncedSearch = useCallback((value: string) => {
        const timer = setTimeout(() => {
            dispatch(setSearchQuery(value));
        }, 300);
        return () => clearTimeout(timer);
    }, [dispatch]);

    useEffect(() => {
        return debouncedSearch(inputValue);
    }, [inputValue, debouncedSearch]);

    const clear = () => setInputValue('');

    return {
        inputValue,
        setInputValue,
        clear
    };
}
