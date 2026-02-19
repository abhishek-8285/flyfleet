from app.services.forecasting_service import ForecastingService

class PricingEngine:
    def __init__(self):
        self.forecasting_service = ForecastingService()

    def suggest_multiplier(self, route_id: str, date: str, base_capacity: int = 50) -> float:
        # 1. Get Forecast
        forecast = self.forecasting_service.predict_demand(route_id, date)
        
        # 2. Calculate Utilization Ratio
        predicted_demand = forecast.predicted_demand
        utilization = predicted_demand / base_capacity
        
        # 3. Apply basic rules
        if utilization > 0.9:
            return 1.5 # 50% surge
        elif utilization > 0.7:
            return 1.2 # 20% surge
        elif utilization < 0.3:
            return 0.9 # 10% discount
            
        return 1.0
