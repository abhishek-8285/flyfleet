#!/bin/bash
echo -e "\033]0;Dashboard Web\007"
echo "Waiting for Backend to initialize... (8s)"
sleep 8
cd apps/dashboard-web
npm install >/dev/null 2>&1
export REACT_APP_API_URL="http://localhost:8084/api/v1"
PORT=3000 BROWSER=none npm start
exec bash
