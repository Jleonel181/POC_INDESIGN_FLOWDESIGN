import json
from ..domain.entities import LayoutDocument, SpreadPage, Frame
from ..domain.layout_fetcher_port import LayoutFetcherPort
from .layout_api_adapter import LayoutApiAdapter


class LayoutJsonAdapter(LayoutFetcherPort):
    """Reads layout from a local JSON file instead of the HTTP API."""

    def __init__(self, json_path: str):
        self._json_path = json_path

    def fetch(self, edition_id: int) -> LayoutDocument:
        with open(self._json_path, "r", encoding="utf-8") as f:
            data = json.load(f)
        return LayoutApiAdapter("")._map(data)
