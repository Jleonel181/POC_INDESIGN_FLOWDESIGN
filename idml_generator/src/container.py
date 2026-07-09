from .modules.idml.application.generate_idml_use_case import GenerateIdmlUseCase
from .modules.idml.infrastructure.idml_generator_adapter import IdmlGeneratorAdapter
from .modules.idml.infrastructure.layout_api_adapter import LayoutApiAdapter
from .modules.idml.infrastructure.layout_json_adapter import LayoutJsonAdapter


def build_use_case_from_api(backend_url: str) -> GenerateIdmlUseCase:
    return GenerateIdmlUseCase(
        fetcher=LayoutApiAdapter(backend_url),
        generator=IdmlGeneratorAdapter(),
    )


def build_use_case_from_json(json_path: str) -> GenerateIdmlUseCase:
    return GenerateIdmlUseCase(
        fetcher=LayoutJsonAdapter(json_path),
        generator=IdmlGeneratorAdapter(),
    )
