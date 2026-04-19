import React, { useEffect, useState } from "react";
import billingService from "../../../services/billingService";
import { Check, Zap, Shield, Crown, RefreshCw } from "lucide-react";
import { useUIStrings } from "../../../hooks/useUIStrings";

const SubscriptionPlans: React.FC = () => {
  const [plans, setPlans] = useState<any[]>([]);
  const [subscription, setSubscription] = useState<any>(null);
  const [loading, setLoading] = useState(true);
  const { uiStrings } = useUIStrings("admin_billing");

  useEffect(() => {
    fetchData();
  }, []);

  const fetchData = async () => {
    try {
      const [plansData, subData] = await Promise.all([
        billingService.getPlans(),
        billingService.getSubscription()
      ]);
      setPlans(plansData);
      setSubscription(subData);
    } catch (error) {
      console.error("Failed to fetch subscription data", error);
    } finally {
      setLoading(false);
    }
  };

  if (loading) return (
    <div className="p-12 flex justify-center">
      <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
    </div>
  );

  return (
    <div className="space-y-8">
      <div className="flex flex-col items-center text-center max-w-2xl mx-auto space-y-4">
        <h2 className="text-3xl font-bold text-white">
          {uiStrings?.plans_title || "Choose Your Flight Path"}
        </h2>
        <p className="text-text-muted">
          Scale your fleet operations with precision. Switch between monthly and annual billing to optimize costs.
        </p>
        
        {subscription && (
          <div className="inline-flex items-center gap-2 bg-primary/10 border border-primary/20 px-4 py-1.5 rounded-full mt-4">
            <span className="text-[10px] font-black text-primary uppercase tracking-widest">Active Plan</span>
            <span className="text-white text-sm font-bold">{subscription.plan_name || 'Pro Tier'}</span>
          </div>
        )}
      </div>

      <div className="grid grid-cols-1 md:grid-cols-3 gap-8">
        {plans.length > 0 ? plans.map((plan) => (
          <div 
            key={plan.id} 
            className={`bg-surface-dark border ${plan.is_popular ? 'border-primary shadow-2xl shadow-primary/10' : 'border-border-color'} rounded-3xl p-8 flex flex-col relative transition-all hover:scale-[1.02]`}
          >
            {plan.is_popular && (
              <div className="absolute -top-4 left-1/2 -translate-x-1/2 bg-primary text-white text-[10px] font-black uppercase px-4 py-1 rounded-full tracking-widest">
                Most Popular
              </div>
            )}

            <div className="mb-8">
              <div className={`w-12 h-12 rounded-2xl flex items-center justify-center mb-6 ${plan.is_popular ? 'bg-primary text-white' : 'bg-white/5 text-primary'}`}>
                {plan.tier === 'basic' ? <Shield className="w-6 h-6" /> : plan.tier === 'pro' ? <Zap className="w-6 h-6" /> : <Crown className="w-6 h-6" />}
              </div>
              <h3 className="text-xl font-bold text-white mb-2">{plan.name}</h3>
              <p className="text-sm text-text-muted">{plan.description}</p>
            </div>

            <div className="mb-8">
              <div className="flex items-baseline gap-1">
                <span className="text-4xl font-bold text-white">₹{plan.base_price?.toLocaleString()}</span>
                <span className="text-text-muted text-sm">/mo</span>
              </div>
              <p className="text-[10px] text-text-muted font-bold uppercase mt-2">Plus usage fees</p>
            </div>

            <div className="space-y-4 mb-10 flex-grow">
              {plan.features?.map((feature: string, idx: number) => (
                <div key={idx} className="flex items-center gap-3 text-sm">
                  <div className="bg-primary/20 text-primary p-0.5 rounded-full">
                    <Check className="w-3 h-3" />
                  </div>
                  <span className="text-gray-300">{feature}</span>
                </div>
              ))}
            </div>

            <button 
              className={`w-full py-3 rounded-xl font-bold transition-all ${
                subscription?.plan_id === plan.id 
                ? 'bg-white/5 text-white/50 cursor-default' 
                : plan.is_popular 
                  ? 'bg-primary text-white hover:bg-primary/90 shadow-lg shadow-primary/20' 
                  : 'bg-white/10 text-white hover:bg-white/20'
              }`}
            >
              {subscription?.plan_id === plan.id ? 'Current Plan' : uiStrings?.btn_choose_plan || 'Upgrade Now'}
            </button>
          </div>
        )) : (
          // Fallback static plans for demonstration
          ['Basic', 'Pro', 'Enterprise'].map((name, i) => (
            <div key={i} className="bg-surface-dark border border-border-color rounded-3xl p-8 flex flex-col opacity-50 grayscale pointer-events-none">
              <h3 className="text-xl font-bold text-white mb-2 font-mono uppercase tracking-widest">{name}</h3>
              <p className="text-4xl font-bold text-white mt-4">₹--</p>
              <div className="mt-8 space-y-4">
                <div className="h-4 bg-white/5 rounded-lg w-3/4 animate-pulse"></div>
                <div className="h-4 bg-white/5 rounded-lg w-1/2 animate-pulse"></div>
              </div>
            </div>
          ))
        )}
      </div>

      <div className="bg-surface-darker/50 border border-border-color rounded-3xl p-8 flex flex-col md:flex-row items-center justify-between gap-6">
        <div className="flex items-center gap-6">
          <div className="p-4 bg-primary/10 rounded-2xl text-primary">
            <RefreshCw className="w-8 h-8" />
          </div>
          <div>
            <h4 className="text-lg font-bold text-white">Annual Savings</h4>
            <p className="text-sm text-text-muted max-w-md">
              Switch to annual billing and save up to 20% on all plans. Includes priority support and unlimited historical data.
            </p>
          </div>
        </div>
        <button className="bg-white/10 text-white px-8 py-3 rounded-xl font-bold hover:bg-white/20 transition-all whitespace-nowrap">
          Switch to Annual
        </button>
      </div>
    </div>
  );
};

export default SubscriptionPlans;
