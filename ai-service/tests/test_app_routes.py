from fastapi import FastAPI

from app.core.config import Settings
from app.main import app, register_api_routers


def route_paths(fastapi_app: FastAPI) -> set[str]:
    return {getattr(route, "path", "") for route in fastapi_app.routes}


def test_ai_service_registers_only_gateway_routes_by_default():
    paths = route_paths(app)

    assert "/api/v1/ai-gateway/call" in paths
    assert "/api/v1/chat/send" not in paths
    assert "/api/v1/chat/generate-interpretation" not in paths
    assert "/api/v1/review" not in paths
    assert "/api/v1/translate" not in paths
    assert "/api/v1/normalize" not in paths


def test_legacy_ai_routes_require_explicit_enablement():
    assert Settings().enable_legacy_ai_routes is False

    test_app = FastAPI()
    register_api_routers(test_app, enable_legacy_ai_routes=True)
    paths = route_paths(test_app)

    assert "/api/v1/ai-gateway/call" in paths
    assert "/api/v1/chat/send" in paths
    assert "/api/v1/chat/generate-interpretation" in paths
    assert "/api/v1/review" in paths
    assert "/api/v1/translate" in paths
    assert "/api/v1/normalize" in paths
