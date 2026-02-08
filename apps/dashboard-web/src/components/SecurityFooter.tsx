import React from 'react';

const SecurityFooter: React.FC = () => (
    <footer className="bg-surface-darker border-t border-border-color px-6 py-2 flex items-center justify-between text-[10px] text-text-muted z-20 mt-auto flex-shrink-0">
        <span>© 2026 Enterprise Fleet Manager. Security Module V1.2</span>
        <div className="flex items-center gap-4">
            <span>Encryption: <span className="text-success">AES-256</span></span>
            <span>System Status: <span className="text-success">Operational</span></span>
            <span>Server Time: <span className="font-mono text-white">{new Date().toLocaleTimeString()} UTC</span></span>
        </div>
    </footer>
);

export default SecurityFooter;
