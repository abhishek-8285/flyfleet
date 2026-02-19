from fastapi import APIRouter
from pydantic import BaseModel
from datetime import date
from app.services.pricing_engine import PricingEngine

router = APIRouter()

class PricingRequest(BaseModel):
    route_id: str
    date: date
    base_capacity: int = 50

class PricingResponse(BaseModel):
    route_id: str
    date: date
    suggested_multiplier: float

@router.post("/suggest", response_model=PricingResponse)
def suggest_pricing(request: PricingRequest):
    engine = PricingEngine()
    multiplier = engine.suggest_multiplier(request.route_id, request.date, request.base_capacity)
    
    return PricingResponse(
        route_id=request.route_id,
        date=request.date,
        suggested_multiplier=multiplier
    )
