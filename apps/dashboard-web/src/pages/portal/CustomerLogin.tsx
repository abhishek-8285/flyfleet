import axios from 'axios';
import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';

const CustomerLogin: React.FC = () => {
    const [apiKey, setApiKey] = useState('');
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState('');
    const navigate = useNavigate();

    const handleLogin = async (e: React.FormEvent) => {
        e.preventDefault();
        setLoading(true);
        setError('');

        try {
            // For now, we are using API Key based login as per backend implementation
            // Ideally, a proper portal would have email/password, but we'll stick to the verified backend auth
            // which uses API Key for now or we can assume there's a way to get token via API Key
            // The backend endpoint is POST /api/v1/customer/auth/token with body { api_key: "..." }

            const response = await axios.post('/api/v1/customer/auth/token', {
                api_key: apiKey
            });

            const token = response.data.token;
            if (token) {
                localStorage.setItem('customer_token', token);
                navigate('/portal/dashboard');
            } else {
                setError('Invalid credentials received');
            }
        } catch (err: any) {
            console.error('Login failed', err);
            setError(err.response?.data?.error || 'Authentication failed. Please check your API Key.');
        } finally {
            setLoading(false);
        }
    };

    return (
        <div className="min-h-screen bg-background-dark flex items-center justify-center p-4 relative overflow-hidden">
            {/* Background Elements */}
            <div className="absolute top-0 left-0 w-full h-full overflow-hidden z-0">
                <div className="absolute top-[-10%] left-[-10%] w-[40%] h-[40%] rounded-full bg-primary/20 blur-[120px]" />
                <div className="absolute bottom-[-10%] right-[-10%] w-[40%] h-[40%] rounded-full bg-purple-500/10 blur-[120px]" />
            </div>

            <div className="max-w-md w-full bg-surface-darker/60 backdrop-blur-xl border border-border-color rounded-2xl p-8 shadow-2xl z-10 relative">
                <div className="text-center mb-8">
                    <div className="w-12 h-12 bg-primary/20 rounded-xl flex items-center justify-center mx-auto mb-4 text-primary">
                        <span className="material-symbols-outlined text-[28px]">local_shipping</span>
                    </div>
                    <h1 className="text-2xl font-bold text-white mb-2">Customer Portal</h1>
                    <p className="text-text-muted text-sm">Access your shipments, invoices, and bookings</p>
                </div>

                {error && (
                    <div className="bg-red-500/10 border border-red-500/20 text-red-400 p-3 rounded-lg text-sm mb-6 flex items-center gap-2">
                        <span className="material-symbols-outlined text-[18px]">error</span>
                        {error}
                    </div>
                )}

                <form onSubmit={handleLogin} className="space-y-5">
                    <div>
                        <label className="block text-xs font-medium text-text-muted uppercase tracking-wider mb-1.5">
                            API Access Key
                        </label>
                        <div className="relative">
                            <span className="material-symbols-outlined absolute left-3 top-2.5 text-text-muted text-[20px]">key</span>
                            <input
                                type="password"
                                value={apiKey}
                                onChange={(e) => setApiKey(e.target.value)}
                                className="w-full bg-surface-dark border border-border-color rounded-lg py-2.5 pl-10 pr-4 text-white focus:outline-none focus:border-primary focus:ring-1 focus:ring-primary transition-all placeholder:text-text-muted/30"
                                placeholder="Enter your API Key"
                                required
                            />
                        </div>
                        <p className="text-[10px] text-text-muted mt-1.5">
                            Please use the API Key provided by your account manager.
                        </p>
                    </div>

                    <button
                        type="submit"
                        disabled={loading}
                        className="w-full bg-primary hover:bg-primary-dark text-white font-medium py-2.5 rounded-lg transition-all duration-200 shadow-[0_0_20px_rgba(19,127,236,0.2)] hover:shadow-[0_0_25px_rgba(19,127,236,0.4)] disabled:opacity-70 disabled:cursor-not-allowed flex items-center justify-center gap-2"
                    >
                        {loading ? (
                            <>
                                <span className="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin"></span>
                                <span>Authenticating...</span>
                            </>
                        ) : (
                            <>
                                <span>Sign In</span>
                                <span className="material-symbols-outlined text-[18px]">arrow_forward</span>
                            </>
                        )}
                    </button>
                </form>

                <div className="mt-8 pt-6 border-t border-border-color text-center">
                    <p className="text-xs text-text-muted">
                        Need access?{' '}
                        <button
                            type="button"
                            onClick={() => navigate('/support')}
                            className="text-primary hover:underline"
                        >
                            Contact Support
                        </button>
                    </p>
                </div>
            </div>
        </div>
    );
};

export default CustomerLogin;
