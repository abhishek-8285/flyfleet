#!/bin/bash

echo "Deploying remaining services..."

# Build and start trip service
echo "Building trip-service..."
cd /home/abhishek/Desktop/flyfleet && docker build -t flyfleet_trip-service ./services/microservices/trip-service
echo "Starting trip-service..."
docker run -d --name fleet-trip-service -p 8084:8080 --network fleet-network \
  -e DATABASE_URL=postgres://fleet_user:fleet_password@172.17.0.1:5433/fleet_db?sslmode=disable \
  flyfleet_trip-service

# Build and start tracking service
echo "Building tracking-service..."
cd /home/abhishek/Desktop/flyfleet && docker build -t flyfleet_tracking-service ./services/microservices/tracking-service
echo "Starting tracking-service..."
docker run -d --name fleet-tracking-service -p 8085:8080 --network fleet-network \
  -e DATABASE_URL=postgres://fleet_user:fleet_password@172.17.0.1:5433/fleet_db?sslmode=disable \
  -e REDIS_URL=redis://172.17.0.1:6380 \
  flyfleet_tracking-service

# Build and start route optimization service
echo "Building route-optimization-service..."
cd /home/abhishek/Desktop/flyfleet && docker build -t flyfleet_route-optimization-service ./services/route-optimization-service
echo "Starting route-optimization-service..."
docker run -d --name fleet-route-service -p 8086:8080 --network fleet-network \
  -e DATABASE_URL=postgres://fleet_user:fleet_password@172.17.0.1:5433/fleet_db?sslmode=disable \
  flyfleet_route-optimization-service

# Build and start maintenance service
echo "Building maintenance-service..."
cd /home/abhishek/Desktop/flyfleet && docker build -t flyfleet_maintenance-service ./services/microservices/maintenance-service
echo "Starting maintenance-service..."
docker run -d --name fleet-maintenance-service -p 8087:8080 --network fleet-network \
  -e DATABASE_URL=postgres://fleet_user:fleet_password@172.17.0.1:5433/fleet_db?sslmode=disable \
  flyfleet_maintenance-service

# Build and start geofencing service
echo "Building geofencing-service..."
cd /home/abhishek/Desktop/flyfleet && docker build -t flyfleet_geofencing-service ./services/geofencing-service
echo "Starting geofencing-service..."
docker run -d --name fleet-geofencing-service -p 8088:8080 --network fleet-network \
  -e DATABASE_URL=postgres://fleet_user:fleet_password@172.17.0.1:5433/fleet_db?sslmode=disable \
  flyfleet_geofencing-service

# Build and start alert service
echo "Building alert-service..."
cd /home/abhishek/Desktop/flyfleet && docker build -t flyfleet_alert-service ./services/microservices/alert-service
echo "Starting alert-service..."
docker run -d --name fleet-alert-service -p 8089:8080 --network fleet-network \
  -e DATABASE_URL=postgres://fleet_user:fleet_password@172.17.0.1:5433/fleet_db?sslmode=disable \
  -e NATS_URL=nats://172.17.0.1:4222 \
  flyfleet_alert-service

# Build and start video service
echo "Building video-service..."
cd /home/abhishek/Desktop/flyfleet && docker build -t flyfleet_video-service ./services/video-service
echo "Starting video-service..."
docker run -d --name fleet-video-service -p 8090:8080 --network fleet-network \
  -e DATABASE_URL=postgres://fleet_user:fleet_password@172.17.0.1:5433/fleet_db?sslmode=disable \
  flyfleet_video-service

# Build and start india integration service
echo "Building india-integration-service..."
cd /home/abhishek/Desktop/flyfleet && docker build -t flyfleet_india-integration-service ./services/microservices/india-integration-service
echo "Starting india-integration-service..."
docker run -d --name fleet-india-service -p 8091:8080 --network fleet-network \
  -e DATABASE_URL=postgres://fleet_user:fleet_password@172.17.0.1:5433/fleet_db?sslmode=disable \
  flyfleet_india-integration-service

# Build and start analytics service
echo "Building analytics-service..."
cd /home/abhishek/Desktop/flyfleet && docker build -t flyfleet_analytics-service ./services/microservices/analytics-service
echo "Starting analytics-service..."
docker run -d --name fleet-analytics-service -p 8092:8080 --network fleet-network \
  -e DATABASE_URL=postgres://fleet_user:fleet_password@172.17.0.1:5433/fleet_db?sslmode=disable \
  flyfleet_analytics-service

echo "All services deployed!"
echo "Checking service status..."
docker ps
