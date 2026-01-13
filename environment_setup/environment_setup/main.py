"""Recursively searches a root path and collects environment variable declarations."""

import argparse
from dataclasses import dataclass
from pathlib import Path
import sys
import typing as t


def main():
    try:
        cli()
    except UsageError as e:
        print(f"Usage Error: {e.problem}", file=sys.stderr)
        sys.exit(1)

def cli():
    parser = argparse.ArgumentParser(
        description=__doc__
    )

    parser.add_argument("root", default=".", help="the root path, which will be searched recursively for environment variable and host files.")
    parser.add_argument("--env-file-name", default="env_vars.txt", help="the name of the files that contain environment variable definitions")

    args = parser.parse_args()
    root = Path(args.root)
    env_file_name = str(args.env_file_name)

    environment_files = list(filter(lambda f: not f.is_dir(), root.glob(f"**/{env_file_name}")))


    try:
        var_declarations: list[EnvironmentVariableDeclaration] = get_environment_variable_declarations_from_files(environment_files)
    except UsageError as e:
        raise UsageError("parsing environment variable declarations", e.problem)
    
    for var in var_declarations:
        print(var)
    
            

        

            
            
        

        

@dataclass
class EnvironmentVariableDeclaration:
    name: str
    var_type: t.Type[int] | t.Type[str]
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



def get_type(name: str) -> t.Type[int] | t.Type[str]:
    match name.lower():
        case "str" | "string":
            return str
        case "int" | "integer":
            return int
        case invalid_type_name:
            raise UsageError(f"{invalid_type_name} is not a valid type")
        
def get_environment_variable_declarations_from_file(file: Path) -> list[EnvironmentVariableDeclaration]:
    var_declarations: list[EnvironmentVariableDeclaration] = []
    for i, line in filter(lambda x: x[1].strip(), enumerate(file.read_text().splitlines())):
        try:
            var_declaration = EnvironmentVariableDeclaration.from_line(line)
        except UsageError as e:
            raise UsageError(f"on line {i}", e.problem)
        var_declarations.append(var_declaration)
    return var_declarations

def get_environment_variable_declarations_from_files(files: t.Iterable[Path]) -> list[EnvironmentVariableDeclaration]:
    var_declarations: list[EnvironmentVariableDeclaration] = []
    for file in files:
        try: 
            var_declarations_in_file = get_environment_variable_declarations_from_file(file)
        except UsageError as e:
            raise UsageError(f"in {file}", e.problem) 
        var_declarations.extend(var_declarations_in_file)
    return var_declarations 


class UsageError(RuntimeError):
    def __init__(self, problem: str, subproblem: t.Optional[str] = None):
        super().__init__()
        self.problem = f"{problem}"
        if subproblem:
            self.problem += f": {subproblem}"    

if __name__ == "__main__":

    try:
        main()
    except UsageError as e:
        print(f"Usage Error: {e.problem}", file=sys.stderr)
        sys.exit(1)
