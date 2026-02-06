import React, { useState } from 'react';
import { Screen, User } from './types';
import LoginScreen from './components/LoginScreen';
import MFAScreen from './components/MFAScreen';
import OrgSelectionScreen from './components/OrgSelectionScreen';

const App: React.FC = () => {
  const [currentScreen, setCurrentScreen] = useState<Screen>(Screen.LOGIN);
  const [user, setUser] = useState<User | null>(null);

  // Mock User Data
  const mockUser: User = {
    name: "Alex Morgan",
    email: "alex.morgan@enterprise.fleet",
    avatarUrl: "https://picsum.photos/200", 
  };

  const handleLoginSuccess = (email: string) => {
    // In a real app, we'd validate creds here.
    setUser({ ...mockUser, email });
    setCurrentScreen(Screen.MFA);
  };

  const handleMFASuccess = () => {
    setCurrentScreen(Screen.ORG_SELECTION);
  };

  const handleSignOut = () => {
    setUser(null);
    setCurrentScreen(Screen.LOGIN);
  };

  return (
    <div className="min-h-screen w-full flex flex-col">
      {currentScreen === Screen.LOGIN && (
        <LoginScreen onLoginSuccess={handleLoginSuccess} />
      )}
      {currentScreen === Screen.MFA && (
        <MFAScreen 
          onSuccess={handleMFASuccess} 
          onCancel={handleSignOut}
        />
      )}
      {currentScreen === Screen.ORG_SELECTION && user && (
        <OrgSelectionScreen user={user} onSignOut={handleSignOut} />
      )}
    </div>
  );
};

export default App;