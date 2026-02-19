import time
import uuid
import logging
from typing import Union, Callable

from fastapi import Request, Response
from fastapi.responses import JSONResponse
from fastapi.exceptions import RequestValidationError
from starlette.middleware.base import BaseHTTPMiddleware
from starlette.exceptions import HTTPException as StarletteHTTPException

from .schemas import ProblemDetails, ErrorDetail
from .errors import AppError, RateLimitError

logger = logging.getLogger(__name__)

class ProblemDetailsMiddleware(BaseHTTPMiddleware):
    async def dispatch(self, request: Request, call_next: Callable) -> Response:
        request_id = request.headers.get("X-Request-ID", str(uuid.uuid4()))
        
        try:
            response = await call_next(request)
            return response
        except Exception as exc:
            return await self.handle_exception(request, exc, request_id)

    async def handle_exception(self, request: Request, exc: Exception, request_id: str) -> JSONResponse:
        status_code = 500
        title = "Internal Server Error"
        type_uri = "https://api.fleetmanager.com/errors/internal-error"
        detail = "An unexpected error occurred."
        errors = None
        headers = {}

        if isinstance(exc, AppError):
            status_code = exc.status_code
            title = exc.__class__.__name__
            type_uri = f"https://api.fleetmanager.com/errors/{exc.code.lower().replace('_', '-')}"
            detail = exc.message
            if isinstance(exc, RateLimitError):
                headers["Retry-After"] = str(exc.retry_after)
            if exc.details:
                # Map specific details if needed
                pass

        elif isinstance(exc, RequestValidationError):
            status_code = 422
            title = "Validation Error"
            type_uri = "https://api.fleetmanager.com/errors/validation-error"
            detail = "The request parameters failed validation."
            errors = [
                ErrorDetail(
                    field=".".join(map(str, err["loc"])),
                    code=err["type"],
                    message=err["msg"]
                ) for err in exc.errors()
            ]

        elif isinstance(exc, StarletteHTTPException):
            status_code = exc.status_code
            title = http_status_text(status_code)
            type_uri = f"https://api.fleetmanager.com/errors/{status_code}"
            detail = str(exc.detail)

        else:
            logger.error(f"Unhandled exception: {exc}", exc_info=True)

        problem = ProblemDetails(
            type=type_uri,
            title=title,
            status=status_code,
            detail=detail,
            instance=request.url.path,
            request_id=request_id,
            errors=errors
        )

        return JSONResponse(
            status_code=status_code,
            content=problem.model_dump(exclude_none=True),
            headers=headers
        )

def http_status_text(code: int) -> str:
    """Helper to get text for status code."""
    # Simplified mapping
    mapping = {
        400: "Bad Request",
        401: "Unauthorized",
        403: "Forbidden",
        404: "Not Found",
        405: "Method Not Allowed",
        429: "Too Many Requests",
        500: "Internal Server Error",
    }
    return mapping.get(code, "Error")
