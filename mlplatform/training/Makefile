.PHONY: install train train-dry deploy validate

install:
	pip install -r requirements.txt

# Full pipeline: extract → train → export → deploy
train:
	python train_dispatch_ranker.py --config config.yaml

# Train and evaluate only, don't touch the live model
train-dry:
	python train_dispatch_ranker.py --config config.yaml --dry-run

# Train on last 30 days only (faster iteration)
train-30d:
	python train_dispatch_ranker.py --config config.yaml --dry-run --days 30

# Validate an existing ONNX model loads correctly in Go
validate:
	@echo "Loading model into mlplatform..."
	curl -s -X POST http://localhost:8080/v1/models/load \
		-H 'Content-Type: application/json' \
		-d '{ \
			"name": "dispatch_ranker", \
			"version": "test", \
			"path": "'$$(ls -t models/*.onnx | head -1)'", \
			"format": "onnx", \
			"features": ["distance_to_rider_km","acceptance_rate_7d","rides_completed_today","minutes_since_last_ride","driver_rating","is_heading_toward_rider","surge_multiplier","time_of_day_hour","day_of_week","zone_demand_density","zone_supply_density","supply_demand_ratio","effective_distance"], \
			"activate": false \
	}' | python -m json.tool
	@echo ""
	@echo "Listing models..."
	curl -s http://localhost:8080/v1/models | python -m json.tool