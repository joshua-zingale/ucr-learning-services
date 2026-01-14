import os
import typing as t

from .variable_declaration import DataInFile, EnvironmentVariableDeclaration
from .usage_error import UsageError, UsageErrorList


class Validator[T](t.Protocol):
    @property
    def __name__(self) -> str:
        """The name of the type being validated."""
        ...

    def __call__(self, value: str, /) -> T:
        """Raises ValueError if the string value cannot be validated."""
        ...


def raise_if_bad_environment_variable_declarations(var_declarations: list[DataInFile[EnvironmentVariableDeclaration]]):
    errors: list[UsageError] = []
    for var_declaration in var_declarations:
        try:
            validate_environment_variable(var_declaration.data.name, var_declaration.data.var_type)
        except UsageError as e:
            errors.append(e)
    if errors:
        raise UsageErrorList("Invalid environment variable declaration(s):", errors)

def validate_environment_variable[T](name: str, type_to_validate: Validator[T]) -> T:
    if not environment_variable_is_defined(name):
        raise UsageError(f"'{name}' is not defined as an environment variable")
    
    try:
        return validate(os.environ[name], type_to_validate)
    except UsageError as e:
        raise UsageError(f"'{name}' has an invalid value", e.problem)


def validate[T](string: str, type_to_validate: Validator[T]) -> T:
    try:
        return type_to_validate(string)
    except ValueError as e:
        raise UsageError(f"'{string}' is not a valid {type_to_validate.__name__}: {e}")
    

def environment_variable_is_defined(name: str) -> bool:
    return name in os.environ