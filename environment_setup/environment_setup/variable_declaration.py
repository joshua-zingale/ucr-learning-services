from dataclasses import dataclass
from pathlib import Path


import typing as t

from .usage_error import UsageError, UsageErrorList


@dataclass
class DataInFile[T]:
    file: Path
    line: int
    data: T


class AbsolutePath(Path):
    def __init__(self, path: str) -> None: # type: ignore
        super().__init__(path)
        if not self.is_absolute():
            import os
            raise ValueError(f"path is not absolute, i.e. it must begin with '{os.sep}'")


@dataclass
class EnvironmentVariableDeclaration:
    name: str
    var_type: t.Type[int] | t.Type[str] | t.Type[Path] | t.Type[AbsolutePath]
    description: str

    @staticmethod
    def from_line(line: str) -> "EnvironmentVariableDeclaration":
        try:
            name, name_of_type, description = map(str.strip, line.split("|"))
        except ValueError as e:
            raise UsageError(f"each line must have three values, separated by |, like VAR_NAME|type|A description")

        try:
            var_type = get_type(name_of_type)
        except UsageError as e:
            raise UsageError(f"defining {name}", e.problem)
        
        return EnvironmentVariableDeclaration(
                name=name,
                var_type=var_type,
                description=description
            )



def get_type(name: str) -> t.Type[int] | t.Type[str] | t.Type[Path] | t.Type[AbsolutePath]:
    match name.lower():
        case "str" | "string":
            return str
        case "int" | "integer":
            return int
        case "path":
            return Path
        case "absolutepath" | "absolute path":
            return AbsolutePath
        case invalid_type_name:
            raise UsageError(f"{invalid_type_name} is not a valid type")
        
def get_environment_variable_declarations_from_file(file: Path) -> list[DataInFile[EnvironmentVariableDeclaration]]:
    var_declarations: list[DataInFile[EnvironmentVariableDeclaration]] = []
    for i, line in filter(lambda x: x[1].strip(), enumerate(file.read_text().splitlines())):
        try:
            var_declaration = EnvironmentVariableDeclaration.from_line(line)
        except UsageError as e:
            raise UsageError(f"on line {i}", e.problem)
        var_declarations.append(
            DataInFile(
                file=file,
                line=i,
                data=var_declaration)
            )
    return var_declarations

def get_environment_variable_declarations_from_files(files: t.Iterable[Path]) -> list[DataInFile[EnvironmentVariableDeclaration]]:
    var_declarations: list[DataInFile[EnvironmentVariableDeclaration]] = []
    errors: list[UsageError] = []
    for file in files:
        try: 
            var_declarations_in_file = get_environment_variable_declarations_from_file(file)
        except UsageError as e:
            errors.append(UsageError(f"in {file}", e.problem) )
            continue
        var_declarations.extend(var_declarations_in_file)
    
    if errors:
        raise UsageErrorList("Error(s) parsing enviromnent variable declarations", errors)
    return var_declarations 