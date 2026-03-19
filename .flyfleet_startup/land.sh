#!/bin/bash
echo -e "\033]0;Landing Page\007"
echo "Waiting for Backend to initialize... (8s)"
sleep 8
cd apps/landing-page
npm install >/dev/null 2>&1
PORT=3001 npm run dev
exec bash
