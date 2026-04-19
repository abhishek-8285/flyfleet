import React, { useState, useEffect, useRef } from 'react';
import { Search, User, Mail, Phone, MapPin } from 'lucide-react';
import { customerV2Service, Customer } from '../../services/customerV2Service';

interface CustomerSearchProps {
  onSelect: (customer: Customer) => void;
  defaultValue?: string;
  className?: string;
}

const CustomerSearch: React.FC<CustomerSearchProps> = ({ onSelect, defaultValue = '', className = '' }) => {
  const [query, setQuery] = useState(defaultValue);
  const [suggestions, setSuggestions] = useState<Customer[]>([]);
  const [isOpen, setIsOpen] = useState(false);
  const [loading, setLoading] = useState(false);
  const containerRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (containerRef.current && !containerRef.current.contains(event.target as Node)) {
        setIsOpen(false);
      }
    };
    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, []);

  const searchCustomers = async (val: string) => {
    setQuery(val);
    if (val.length < 2) {
      setSuggestions([]);
      setIsOpen(false);
      return;
    }

    setLoading(true);
    try {
      const { customers } = await customerV2Service.listCustomers(10, 0);
      // Client-side filtering for better UX on small lists, though API should handle it ideally
      const filtered = customers.filter(c => 
        c.name.toLowerCase().includes(val.toLowerCase()) || 
        c.email?.toLowerCase().includes(val.toLowerCase()) ||
        c.phone?.includes(val)
      );
      setSuggestions(filtered);
      setIsOpen(filtered.length > 0);
    } catch (err) {
      console.error('Failed to fetch customers', err);
    } finally {
      setLoading(false);
    }
  };

  const handleSelect = (customer: Customer) => {
    setQuery(customer.name);
    setSuggestions([]);
    setIsOpen(false);
    onSelect(customer);
  };

  return (
    <div className={`relative ${className}`} ref={containerRef}>
      <div className="relative">
        <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none text-gray-400">
          <Search className="w-4 h-4" />
        </div>
        <input
          type="text"
          value={query}
          onChange={(e) => searchCustomers(e.target.value)}
          onFocus={() => query.length >= 2 && suggestions.length > 0 && setIsOpen(true)}
          placeholder="Search registered customers..."
          className="w-full bg-white/5 border border-white/10 rounded-lg pl-10 pr-4 py-2 text-white focus:outline-none focus:border-primary transition-all text-sm"
        />
        {loading && (
          <div className="absolute inset-y-0 right-0 pr-3 flex items-center pointer-events-none">
            <div className="animate-spin h-3 w-3 border-b-2 border-primary rounded-full"></div>
          </div>
        )}
      </div>

      {isOpen && (
        <div className="absolute z-50 w-full mt-2 bg-gray-900 border border-white/10 rounded-xl shadow-2xl overflow-hidden backdrop-blur-xl max-h-60 overflow-y-auto">
          {suggestions.map((customer) => (
            <div
              key={customer.id}
              onClick={() => handleSelect(customer)}
              className="p-3 hover:bg-white/5 cursor-pointer border-b border-white/5 last:border-0 transition-colors group"
            >
              <div className="flex items-center justify-between mb-1">
                <span className="font-medium text-white group-hover:text-primary transition-colors">{customer.name}</span>
                {customer.gstin && (
                  <span className="text-[10px] bg-white/5 text-gray-500 px-1.5 py-0.5 rounded uppercase font-mono">{customer.gstin}</span>
                )}
              </div>
              <div className="flex flex-wrap gap-x-4 gap-y-1">
                {customer.email && (
                  <div className="flex items-center gap-1.5 text-[11px] text-gray-400">
                    <Mail className="w-3 h-3" /> {customer.email}
                  </div>
                )}
                {customer.phone && (
                  <div className="flex items-center gap-1.5 text-[11px] text-gray-400">
                    <Phone className="w-3 h-3" /> {customer.phone}
                  </div>
                )}
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
};

export default CustomerSearch;
