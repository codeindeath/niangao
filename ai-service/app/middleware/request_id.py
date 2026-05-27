import logging
import time
import uuid
from typing import Optional

from starlette.middleware.base import BaseHTTPMiddleware

REQUEST_ID_HEADER = "X-Request-ID"

logger = logging.getLogger(__name__)


def sanitize_request_id(value: Optional[str]) -> str:
    request_id = (value or "").strip()
    if not request_id or len(request_id) > 128:
        return ""
    if any(ord(ch) < 0x20 or ord(ch) == 0x7F for ch in request_id):
        return ""
    return request_id


class RequestIDMiddleware(BaseHTTPMiddleware):
    async def dispatch(self, request, call_next):
        request_id = sanitize_request_id(request.headers.get(REQUEST_ID_HEADER))
        if not request_id:
            request_id = f"req-{uuid.uuid4()}"

        request.state.request_id = request_id
        started = time.perf_counter()
        response = await call_next(request)
        response.headers[REQUEST_ID_HEADER] = request_id

        latency_ms = (time.perf_counter() - started) * 1000
        logger.info(
            "ai_request request_id=%s method=%s path=%s status=%s latency_ms=%.3f",
            request_id,
            request.method,
            request.url.path,
            response.status_code,
            latency_ms,
        )
        return response
