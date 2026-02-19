from fastapi import APIRouter
from pydantic import BaseModel
from typing import List, Optional
from datetime import date

router = APIRouter()

class DemandPredictionRequest(BaseModel):
    route_id: str
    date: date
    
class DemandPredictionResponse(BaseModel):
    route_id: str
    date: date
    predicted_demand: int
    confidence_score: float

@router.post("/predict", response_model=DemandPredictionResponse)
def predict_demand(request: DemandPredictionRequest):
    from app.services.forecasting_service import ForecastingService
    service = ForecastingService()
    return service.predict_demand(request.route_id, request.date)
