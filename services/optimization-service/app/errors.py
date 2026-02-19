from typing import Optional, List, Any

class AppError(Exception):
    """Base application error."""
    def __init__(
        self,
        message: str,
        code: str = "INTERNAL_ERROR",
        status_code: int = 500,
        details: Optional[List[Any]] = None
    ):
        self.message = message
        self.code = code
        self.status_code = status_code
        self.details = details
        super().__init__(self.message)

class ValidationError(AppError):
    def __init__(self, message: str = "Validation failed", details: Optional[List[Any]] = None):
        super().__init__(message, code="VALIDATION_ERROR", status_code=400, details=details)

class AuthenticationError(AppError):
    def __init__(self, message: str = "Authentication required"):
        super().__init__(message, code="UNAUTHORIZED", status_code=401)

class AuthorizationError(AppError):
    def __init__(self, message: str = "Permission denied"):
        super().__init__(message, code="FORBIDDEN", status_code=403)

class NotFoundError(AppError):
    def __init__(self, resource: str, resource_id: str):
        super().__init__(f"{resource} with ID {resource_id} not found", code="NOT_FOUND", status_code=404)

class ConflictError(AppError):
    def __init__(self, message: str):
        super().__init__(message, code="CONFLICT", status_code=409)

class UnprocessableError(AppError):
    def __init__(self, message: str):
        super().__init__(message, code="UNPROCESSABLE", status_code=422)

class RateLimitError(AppError):
    def __init__(self, retry_after: int):
        self.retry_after = retry_after
        super().__init__(f"Rate limit exceeded. Retry after {retry_after} seconds.", code="RATE_LIMIT", status_code=429)

class InternalError(AppError):
    def __init__(self, message: str = "An internal error occurred"):
        super().__init__(message, code="INTERNAL_ERROR", status_code=500)
