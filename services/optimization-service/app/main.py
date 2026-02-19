from fastapi import FastAPI, Depends, Header
from typing import Annotated

from .middleware import ProblemDetailsMiddleware
from .errors import NotFoundError, RateLimitError, UnprocessableError

app = FastAPI(title="Route Optimization Service")

# Register Middleware (Exception Handling)
app.add_middleware(ProblemDetailsMiddleware)

@app.get("/health")
async def health_check():
    return {"status": "ok"}

@app.get("/api/v1/routes/{route_id}")
async def get_route(route_id: str):
    if route_id == "not-found":
        raise NotFoundError(resource="Route", resource_id=route_id)
    if route_id == "rate-limit":
        raise RateLimitError(retry_after=60)
    if route_id == "error":
        raise RuntimeError("Unexpected failure")
    
    return {"id": route_id, "status": "optimized"}

@app.post("/api/v1/optimize")
async def optimize_route(
    locations: list[str],
    x_request_id: Annotated[str | None, Header()] = None
):
    if len(locations) < 2:
        raise UnprocessableError("At least two locations are required for optimization.")
    
    return {"optimized": True, "stops": len(locations)}
