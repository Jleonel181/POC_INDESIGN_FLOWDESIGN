from abc import ABC, abstractmethod
from ..domain.entities import LayoutDocument


class LayoutFetcherPort(ABC):
    @abstractmethod
    def fetch(self, edition_id: int) -> LayoutDocument:
        """Fetch layout data for a given edition and return a LayoutDocument."""
