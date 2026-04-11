import { useEffect, useRef, useState } from 'react';

interface WSMessage {
    subject: string;
    data: any;
}

export const useWebSocket = (url: string) => {
    const [messages, setMessages] = useState<WSMessage[]>([]);
    const [lastMessage, setLastMessage] = useState<WSMessage | null>(null);
    const [isConnected, setIsConnected] = useState(false);
    const ws = useRef<WebSocket | null>(null);

    useEffect(() => {
        let isMounted = true;

        const connect = () => {
            if (!isMounted) return;
            
            // Don't connect if there's no token in the URL
            const urlObj = new URL(url.startsWith('ws') ? url.replace('ws', 'http') : url);
            const token = urlObj.searchParams.get('token');
            if (!token) {
                console.warn('useWebSocket: No token provided in URL, skipping connection');
                return;
            }

            const socket = new WebSocket(url);

            socket.onopen = () => {
                setIsConnected(true);
                console.log('WebSocket Connected');
            };

            socket.onmessage = (event) => {
                try {
                    const message: WSMessage = JSON.parse(event.data);
                    setLastMessage(message);
                    setMessages((prev) => [...prev.slice(-49), message]); // Keep last 50
                } catch (e) {
                    console.error('Failed to parse WS message', e);
                }
            };

            socket.onclose = () => {
                setIsConnected(false);
                if (isMounted) {
                    console.log('WebSocket Disconnected. Retrying in 5s...');
                    setTimeout(connect, 5000);
                } else {
                    console.log('WebSocket Disconnected. Component unmounted.');
                }
            };

            socket.onerror = (error) => {
                console.error('WebSocket Error', error);
                socket.close();
            };

            ws.current = socket;
        };

        connect();

        return () => {
            isMounted = false;
            ws.current?.close();
        };
    }, [url]);

    const sendMessage = (msg: any) => {
        if (ws.current?.readyState === WebSocket.OPEN) {
            ws.current.send(JSON.stringify(msg));
        }
    };

    return { lastMessage, messages, isConnected, sendMessage };
};
