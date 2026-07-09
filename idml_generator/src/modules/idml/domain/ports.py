from abc import ABC, abstractmethod
from ..domain.entities import LayoutDocument


class IdmlGeneratorPort(ABC):
    @abstractmethod
    def generate(self, document: LayoutDocument, output_path: str) -> str:
        """Generate an IDML file and return the output path."""
