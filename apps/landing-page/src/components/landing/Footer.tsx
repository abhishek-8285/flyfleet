'use client';


const fadeInUp = {
  initial: { opacity: 0, y: 30 },
  whileInView: { opacity: 1, y: 0 },
  viewport: { once: true },
  transition: { duration: 0.6 }
};

const Footer = ({ companyName }: { companyName: string }) => {
  return (
    <footer className="border-t border-slate-200 dark:border-border-dark bg-white dark:bg-background-dark pt-20 pb-10">
      <div className="px-4 md:px-10 max-w-[1280px] mx-auto">
        <div className="grid grid-cols-2 md:grid-cols-4 lg:grid-cols-6 gap-12 mb-16">
          <div className="col-span-2 lg:col-span-2">
            <div className="flex items-center gap-3 text-slate-900 dark:text-white mb-6">
              <div className="size-8 flex items-center justify-center text-primary bg-primary/10 rounded-lg">
                <span className="material-symbols-outlined !text-2xl">local_shipping</span>
              </div>
              <h2 className="text-xl font-bold font-display tracking-tight">{companyName}</h2>
            </div>
            <p className="text-slate-500 dark:text-text-secondary text-base leading-relaxed mb-8 max-w-sm">
              The world's most advanced AI-driven fleet management platform. Optimization made simple for global logistics leaders.
            </p>
            <div className="flex gap-4">
              {['facebook', 'twitter', 'linkedin', 'youtube'].map((social) => (
                <button key={social} className="size-10 rounded-full border border-slate-200 dark:border-border-dark flex items-center justify-center text-slate-400 hover:text-primary hover:border-primary transition-all">
                  <span className="material-symbols-outlined !text-xl">
                    {social === 'linkedin' ? 'work' : social === 'twitter' ? 'share' : social === 'facebook' ? 'public' : 'play_circle'}
                  </span>
                </button>
              ))}
            </div>
          </div>
          
          <div>
            <h4 className="font-bold text-slate-900 dark:text-white mb-6 font-display">Platform</h4>
            <ul className="space-y-4 text-sm text-slate-500 dark:text-text-secondary">
              <li><a href="#" className="hover:text-primary transition-colors">Real-time Tracking</a></li>
              <li><a href="#" className="hover:text-primary transition-colors">Route Optimization</a></li>
              <li><a href="#" className="hover:text-primary transition-colors">Fuel Monitoring</a></li>
              <li><a href="#" className="hover:text-primary transition-colors">Driver Safety</a></li>
            </ul>
          </div>
          
          <div>
            <h4 className="font-bold text-slate-900 dark:text-white mb-6 font-display">Company</h4>
            <ul className="space-y-4 text-sm text-slate-500 dark:text-text-secondary">
              <li><a href="#" className="hover:text-primary transition-colors">About Us</a></li>
              <li><a href="#" className="hover:text-primary transition-colors">Success Stories</a></li>
              <li><a href="#" className="hover:text-primary transition-colors">Careers</a></li>
              <li><a href="#" className="hover:text-primary transition-colors">Contact</a></li>
            </ul>
          </div>

          <div>
            <h4 className="font-bold text-slate-900 dark:text-white mb-6 font-display">Resources</h4>
            <ul className="space-y-4 text-sm text-slate-500 dark:text-text-secondary">
              <li><a href="#" className="hover:text-primary transition-colors">Documentation</a></li>
              <li><a href="#" className="hover:text-primary transition-colors">Help Center</a></li>
              <li><a href="#" className="hover:text-primary transition-colors">API Reference</a></li>
              <li><a href="#" className="hover:text-primary transition-colors">Blog</a></li>
            </ul>
          </div>

          <div>
            <h4 className="font-bold text-slate-900 dark:text-white mb-6 font-display">Legal</h4>
            <ul className="space-y-4 text-sm text-slate-500 dark:text-text-secondary">
              <li><a href="#" className="hover:text-primary transition-colors">Privacy Policy</a></li>
              <li><a href="#" className="hover:text-primary transition-colors">Terms of Service</a></li>
              <li><a href="#" className="hover:text-primary transition-colors">Security</a></li>
              <li><a href="#" className="hover:text-primary transition-colors">Trust Center</a></li>
            </ul>
          </div>
        </div>
        <div className="border-t border-slate-100 dark:border-border-dark pt-10 flex flex-col md:flex-row justify-between items-center gap-6">
          <div className="flex items-center gap-6">
            <p className="text-slate-400 text-sm font-medium">© 2026 {companyName} Inc. All rights reserved.</p>
            <div className="h-4 w-px bg-slate-200 dark:bg-border-dark hidden md:block"></div>
            <div className="flex gap-4">
              <span className="text-xs text-slate-400">ISO 27001 Certified</span>
              <span className="text-xs text-slate-400">GDPR Compliant</span>
            </div>
          </div>
          <div className="flex gap-8">
            <a href="#" className="text-slate-400 hover:text-primary text-xs font-bold uppercase tracking-wider transition-colors">Security</a>
            <a href="#" className="text-slate-400 hover:text-primary text-xs font-bold uppercase tracking-wider transition-colors">Status</a>
          </div>
        </div>
      </div>
    </footer>
  );
};

export default Footer;
