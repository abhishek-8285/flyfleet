import { useEffect, useState } from 'react';

// Types
interface SearchResult {
  id: string;
  type: 'vehicle' | 'driver' | 'trip' | 'alert';
  title: string;
  subtitle: string;
  description: string;
  status: string;
  metadata: Record<string, any>;
  updated_at: string;
}

interface SearchFacet {
  label: string;
  value: string;
  count: number;
}

const SearchPage = () => {
  const [searchTerm, setSearchTerm] = useState("");
  const [results, setResults] = useState<SearchResult[]>([]);
  const [loading, setLoading] = useState(false);
  const [facets, setFacets] = useState<Record<string, SearchFacet[]>>({});

  // Fetch results based on query
  useEffect(() => {
    const fetchResults = async () => {
      setLoading(true);
      try {
        const queryParams = new URLSearchParams({ q: searchTerm });
        const response = await fetch(`/api/v1/search?${queryParams}`);
        const data = await response.json();
        if (data.data) {
          setResults(data.data);
        }
      } catch (error) {
        console.error("Failed to fetch search results:", error);
      } finally {
        setLoading(false);
      }
    };

    // Debounce search
    const timer = setTimeout(() => {
        fetchResults();
    }, 300);

    return () => clearTimeout(timer);
  }, [searchTerm]);

  // Fetch facets on mount
  useEffect(() => {
    const fetchFacets = async () => {
      try {
        const response = await fetch('/api/v1/search/facets');
        const data = await response.json();
        setFacets(data);
      } catch (error) {
        console.error("Failed to fetch facets:", error);
      }
    };
    fetchFacets();
  }, []);

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'active': return 'text-success bg-success/10 border-success/20';
      case 'warning': return 'text-warning bg-warning/10 border-warning/20';
      case 'critical': return 'text-danger bg-danger/10 border-danger/20';
      default: return 'text-text-muted bg-surface-darker border-border-color';
    }
  };

  const getTypeIcon = (type: string) => {
    switch (type) {
      case 'vehicle': return 'directions_car';
      case 'driver': return 'person';
      case 'trip': return 'route';
      case 'alert': return 'warning';
      default: return 'search';
    }
  };

  const getTypeColor = (type: string) => {
      switch (type) {
        case 'vehicle': return 'text-primary bg-primary/10 border-primary/20';
        case 'driver': return 'text-indigo-500 bg-indigo-500/10 border-indigo-500/20';
        case 'trip': return 'text-emerald-500 bg-emerald-500/10 border-emerald-500/20';
        case 'alert': return 'text-danger bg-danger/10 border-danger/20';
        default: return 'text-text-muted bg-surface-darker border-border-color';
      }
  }

  return (
    <div className="flex h-screen w-full">
      {/* Sidebar is now part of layout, but kept here if standalone page structure is needed */}
      
      <main className="flex-1 flex flex-col min-w-0 bg-background-dark overflow-hidden">
        <header className="h-20 flex items-center justify-between px-6 border-b border-border-color bg-surface-darker flex-shrink-0 z-20 gap-8">
          <div className="flex flex-col">
            <div className="flex items-center text-[10px] text-text-muted font-mono uppercase tracking-wider mb-1">
              <span>Platform</span>
              <span className="mx-2">/</span>
              <span className="text-primary">Global Search</span>
            </div>
            <h2 className="text-white text-lg font-bold">Unified Search</h2>
          </div>
          <div className="flex-1 max-w-3xl">
            <div className="relative group">
              <div className="absolute inset-y-0 left-0 pl-4 flex items-center pointer-events-none">
                <span className="material-symbols-outlined text-text-muted group-focus-within:text-primary transition-colors">search</span>
              </div>
              <input 
                autoFocus
                className="block w-full pl-12 pr-12 py-3 bg-surface-dark border border-border-color rounded-lg text-sm text-white placeholder-text-muted focus:ring-2 focus:ring-primary focus:border-primary outline-none transition-all shadow-lg" 
                placeholder="Search for vehicles, drivers, trips, or alerts..." 
                type="text"
                value={searchTerm}
                onChange={(e) => setSearchTerm(e.target.value)}
              />
              <div className="absolute inset-y-0 right-0 pr-4 flex items-center pointer-events-none">
                <span className="search-shortcut">/</span>
              </div>
            </div>
          </div>
          <div className="flex items-center gap-4">
            <button className="p-2 text-text-muted hover:text-white hover:bg-surface-dark rounded-lg transition-colors relative">
              <span className="material-symbols-outlined">notifications</span>
              <span className="absolute top-1.5 right-1.5 w-2 h-2 bg-danger rounded-full border border-surface-darker"></span>
            </button>
            <button className="p-2 text-text-muted hover:text-white hover:bg-surface-dark rounded-lg transition-colors">
              <span className="material-symbols-outlined">settings</span>
            </button>
          </div>
        </header>
        
        <div className="flex-1 flex overflow-hidden">
          <aside className="w-80 bg-surface-dark border-r border-border-color flex-shrink-0 overflow-y-auto custom-scrollbar p-6">
            <div className="flex items-center justify-between mb-6">
              <h3 className="text-sm font-semibold text-white flex items-center gap-2">
                <span className="material-symbols-outlined text-primary text-[18px]">filter_alt</span>
                Faceted Filters
              </h3>
              <button className="text-[10px] text-primary hover:text-primary-dark uppercase font-bold tracking-wider">Reset All</button>
            </div>
            
            {/* Dynamic Facets */}
            {facets.entity_types && (
            <div className="space-y-6">
              <div className="space-y-3">
                <label className="text-xs font-semibold text-text-muted uppercase tracking-wider">Entity Type</label>
                <div className="space-y-2">
                    {facets.entity_types.map((facet) => (
                        <label key={facet.value} className="flex items-center gap-3 p-2 rounded hover:bg-surface-darker cursor-pointer transition-colors group">
                            <input defaultChecked className="w-4 h-4 rounded border-border-color bg-surface-darker text-primary focus:ring-offset-0 focus:ring-1 focus:ring-primary" type="checkbox"/>
                            <div className="flex items-center gap-2 text-sm text-white group-hover:text-primary transition-colors">
                            <span className="material-symbols-outlined text-[18px] text-text-muted">
                                {getTypeIcon(facet.value)}
                            </span>
                            {facet.label}
                            </div>
                            <span className="ml-auto text-[10px] bg-surface-darker px-1.5 py-0.5 rounded text-text-muted border border-border-color">{facet.count}</span>
                        </label>
                    ))}
                </div>
              </div>
              <div className="h-px bg-border-color"></div>
            </div>
            )}
            
            {/* Status Filter (Static for now) */}
            <div className="space-y-3 mt-6">
                <label className="text-xs font-semibold text-text-muted uppercase tracking-wider">Status</label>
                <div className="flex flex-wrap gap-2">
                    <button className="px-3 py-1 rounded-full text-[10px] font-medium bg-success/10 text-success border border-success/20 hover:bg-success/20 transition-colors">Active</button>
                    <button className="px-3 py-1 rounded-full text-[10px] font-medium bg-warning/10 text-warning border border-warning/20 hover:bg-warning/20 transition-colors">Warning</button>
                    <button className="px-3 py-1 rounded-full text-[10px] font-medium bg-danger/10 text-danger border border-danger/20 hover:bg-danger/20 transition-colors">Critical</button>
                    <button className="px-3 py-1 rounded-full text-[10px] font-medium bg-surface-darker text-text-muted border border-border-color hover:text-white hover:border-white transition-colors">Inactive</button>
                </div>
            </div>

          </aside>
          
          <div className="flex-1 overflow-y-auto custom-scrollbar p-6 bg-background-dark">
            <div className="flex flex-col gap-4 mb-6">
              <div className="flex justify-between items-end border-b border-border-color pb-2">
                <div>
                  <h3 className="text-white font-bold text-lg">Search Results</h3>
                  <p className="text-xs text-text-muted">Found <span className="text-white font-medium">{results.length}</span> matches for query</p>
                </div>
                <div className="flex items-center gap-2">
                  <span className="text-xs text-text-muted">Sort by:</span>
                  <select className="bg-surface-dark border border-border-color rounded px-2 py-1 text-xs text-white focus:ring-1 focus:ring-primary outline-none">
                    <option>Relevance</option>
                    <option>Date (Newest)</option>
                    <option>Date (Oldest)</option>
                  </select>
                </div>
              </div>
            </div>
            
            <div className="space-y-4 max-w-5xl">
                {loading ? (
                    <div className="text-text-muted text-center py-10">Searching...</div>
                ) : (
                    results.map((result) => (
                        <div key={result.id} className="bg-surface-dark rounded-xl border border-border-color p-4 hover:border-primary/50 transition-colors group cursor-pointer relative overflow-hidden">
                            <div className={`absolute top-0 left-0 w-1 h-full ${result.type === 'alert' ? 'bg-danger' : result.type === 'trip' ? 'bg-emerald-500' : 'bg-primary'}`}></div>
                            <div className="flex items-start gap-4">
                                <div className="w-10 h-10 rounded-lg bg-surface-darker flex items-center justify-center border border-border-color flex-shrink-0 group-hover:border-primary/30 transition-colors">
                                    <span className={`material-symbols-outlined ${result.type === 'alert' ? 'text-danger' : 'text-primary'}`}>{getTypeIcon(result.type)}</span>
                                </div>
                                <div className="flex-1 min-w-0">
                                    <div className="flex items-center gap-2 mb-1">
                                        <span className={`px-1.5 py-0.5 rounded text-[10px] font-bold uppercase tracking-wider border ${getTypeColor(result.type)}`}>{result.type}</span>
                                        <span className="text-xs text-text-muted font-mono">ID: {result.id}</span>
                                        <span className="w-1 h-1 rounded-full bg-text-muted"></span>
                                        <span className={`text-xs flex items-center gap-1 capitalize ${result.status === 'critical' ? 'text-danger' : 'text-success'}`}>
                                            <span className={`w-1.5 h-1.5 rounded-full ${result.status === 'critical' ? 'bg-danger' : 'bg-success'}`}></span>
                                            {result.status}
                                        </span>
                                    </div>
                                    <h4 className="text-white font-semibold text-base mb-1 group-hover:text-primary transition-colors">{result.title}</h4>
                                    <p className="text-sm text-text-muted line-clamp-2">
                                        {result.description}
                                    </p>
                                    <div className="mt-3 flex items-center gap-4 text-xs text-text-muted">
                                        {result.metadata && Object.entries(result.metadata).map(([key, value]) => (
                                            <span key={key} className="flex items-center gap-1 capitalize">
                                                <span className="material-symbols-outlined text-[14px]">info</span> {key}: {value}
                                            </span>
                                        ))}
                                    </div>
                                </div>
                                <div className="flex flex-col items-end gap-2 self-center">
                                    <button className="px-3 py-1.5 bg-surface-darker hover:bg-primary hover:text-white border border-border-color text-text-muted text-xs font-medium rounded transition-colors flex items-center gap-2">
                                        Open Detail
                                        <span className="material-symbols-outlined text-[14px]">arrow_forward</span>
                                    </button>
                                </div>
                            </div>
                        </div>
                    ))
                )}
            </div>
            
            <div className="mt-8 flex justify-center pb-8">
              <button className="px-4 py-2 bg-surface-darker border border-border-color text-sm text-text-muted hover:text-white hover:border-primary rounded-lg transition-colors">
                Load More Results
              </button>
            </div>
          </div>
        </div>
        
        <footer className="bg-surface-darker border-t border-border-color px-6 py-2 flex items-center justify-between text-[10px] text-text-muted z-20 flex-shrink-0">
          <span>© 2023 Enterprise Fleet Manager. Global Search Module V2.0</span>
          <div className="flex items-center gap-4">
            <span>Index Status: <span className="text-success">Synced</span></span>
            <span>Server Time: <span className="font-mono text-white">14:35 UTC</span></span>
          </div>
        </footer>
      </main>
    </div>
  );
};

export default SearchPage;
