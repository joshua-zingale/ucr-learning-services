from dataclasses import dataclass
import typing as t
from pathlib import Path

from .usage_error import UsageError, UsageErrorList


@dataclass
class DataInFile[T]:
    file: Path
    line: int
    data: T


def parse_from_file[T](line_parser: t.Callable[[str], T], file: Path) -> list[DataInFile[T]]:
    var_declarations: list[DataInFile[T]] = []
    for i, line in filter(lambda x: x[1].strip(), enumerate(file.read_text().splitlines())):
        try:
            var_declaration = line_parser(line)
        except UsageError as e:
            raise UsageError(f"on line {i+1}", e.problem)
        var_declarations.append(
            DataInFile(
                file=file,
                line=i,
                data=var_declaration)
            )
    return var_declarations

def parse_from_files[T](line_parser: t.Callable[[str], T], files: t.Iterable[Path]) -> list[DataInFile[T]]:
    var_declarations: list[DataInFile[T]] = []
    errors: list[UsageError] = []
    for file in files:
        try: 
            var_declarations_in_file = parse_from_file(line_parser, file)
        except UsageError as e:
            errors.append(UsageError(f"in {file}", e.problem) )
            continue
        var_declarations.extend(var_declarations_in_file)
    
    if errors:
        raise UsageErrorList("Error(s) while parsing", errors)
    return var_declarations 