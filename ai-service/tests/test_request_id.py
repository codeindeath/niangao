import logging

from fastapi import FastAPI
from fastapi.testclient import TestClient

from app.middleware.request_id import REQUEST_ID_HEADER, RequestIDMiddleware


def test_request_id_middleware_echoes_and_logs_request_id(caplog):
    app = FastAPI()
    app.add_middleware(RequestIDMiddleware)

    @app.get("/health")
    async def health():
        return {"status": "ok"}

    with caplog.at_level(logging.INFO, logger="app.middleware.request_id"):
        response = TestClient(app).get("/health", headers={REQUEST_ID_HEADER: "client-request-1"})

    assert response.status_code == 200
    assert response.headers[REQUEST_ID_HEADER] == "client-request-1"
    assert "request_id=client-request-1" in caplog.text
