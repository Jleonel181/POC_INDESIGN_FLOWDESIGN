from ..domain.ports import IdmlGeneratorPort
from ..domain.layout_fetcher_port import LayoutFetcherPort


class GenerateIdmlUseCase:
    def __init__(self, fetcher: LayoutFetcherPort, generator: IdmlGeneratorPort):
        self._fetcher = fetcher
        self._generator = generator

    def execute(self, edition_id: int, output_path: str) -> str:
        document = self._fetcher.fetch(edition_id)
        return self._generator.generate(document, output_path)
