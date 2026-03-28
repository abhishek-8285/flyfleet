// Isolated refresh client using fetch to avoid bundling full axios
// This file is dynamically imported only during token refresh

export async function refreshToken(API_URL: string): Promise<{ token: string; user?: any } | null> {
  try {
    const response = await fetch(`${API_URL}/auth/refresh`, {
      method: 'POST',
      credentials: 'include', // Send refresh cookie
      headers: {
        'Content-Type': 'application/json',
      },
    });

    if (!response.ok) {
      throw new Error(`Refresh failed: ${response.status}`);
    }

    const data = await response.json();
    return {
      token: data.token || data.access_token,
      user: data.user,
    };
  } catch (error) {
    console.error('Token refresh failed:', error);
    return null;
  }
}

export async function logout(API_URL: string): Promise<void> {
  try {
    await fetch(`${API_URL}/auth/logout`, {
      method: 'POST',
      credentials: 'include',
      headers: {
        'Content-Type': 'application/json',
      },
    });
  } catch (error) {
    // Ignore logout errors
  }
}
