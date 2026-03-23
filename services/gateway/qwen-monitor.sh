#!/bin/bash
# Qwen Message Monitor - Fast polling

MESSAGE_FILE="$HOME/.gateway_message.txt"
RESPONSE_FILE="$HOME/.gateway_response.txt"

echo "🤖 Qwen Message Monitor (Fast)"
echo "=============================="
echo "Watching: $MESSAGE_FILE"
echo "Responding to: $RESPONSE_FILE"
echo ""
echo "💬 Messages will appear below"
echo "📝 Type response + Enter to send"
echo ""
echo "Press Ctrl+C to stop"
echo ""

# Initialize
> "$MESSAGE_FILE"
> "$RESPONSE_FILE"

while true; do
    if [ -s "$MESSAGE_FILE" ]; then
        current_message=$(cat "$MESSAGE_FILE" 2>/dev/null)
        if [ -n "$current_message" ]; then
            echo ""
            echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
            echo "💬 MESSAGE:"
            echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
            echo "$current_message"
            echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
            echo ""
            echo "🤖 Your response:"
            read -r response
            if [ -n "$response" ]; then
                echo "$response" > "$RESPONSE_FILE"
                echo "✅ Sent!"
            fi
            > "$MESSAGE_FILE"
        fi
    fi
    sleep 0.5
done
