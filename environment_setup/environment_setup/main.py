"""Recursively searches a root path and validates environment variable declarations.

Different actions may be specified with flags;
however, if no actions are specified, then validation alone is performed.
"""

import argparse
import enum
from pathlib import Path
import sys
import typing as t

from .environment import raise_if_bad_environment_variable_declarations
from .hosts import host_declaration_from_line
from . import parsing
from .usage_error import UsageError, UsageErrorList
from .variable_declaration import environment_variable_declaration_from_line

def main():
    try:
        cli()
    except UsageError as e:
        print(f"Usage Error: {e.problem}", file=sys.stderr)
        sys.exit(1)
    except UsageErrorList as e:
        print(e.header + "\n" + "\n".join(map(lambda x: f"{'':>4}{x}", map(lambda y: y.problem, e.errors))))
        sys.exit(1)

def cli():
    parser = argparse.ArgumentParser(
        description=__doc__
    )

    parser.add_argument("root", default=".", help="the root path, which will be searched recursively for environment variable and host files.")
    parser.add_argument("--list-environment-variables", "-le", action="store_true", help="if set, lists all variables for which there are declarations")
    parser.add_argument("--list-hosts", "-lh", action="store_true", help="if set, lists all hosts for which there are declarations")
    parser.add_argument("--validate-environment-variables", "-ve", action="store_true", help="if set, validates that all environment variables are well defined; leads an an undefined or improperly defined environment variable to cause the progra to exit with a non-zero exit status.")
    parser.add_argument("--env-files-name", default="env_vars.txt", help="the name of the files that contain environment variable declarations.")
    parser.add_argument("--host-files-name", default="hosts.txt", help="the name of the files that contain environment variable declarations.")

    args = parser.parse_args()
    root = Path(args.root)
    env_file_name = str(args.env_files_name)
    host_files_name = str(args.host_files_name)
    actions_to_perform = actions if (actions := get_actions_to_perform(
        list_declarations=args.list_environment_variables,
        list_hosts = args.list_hosts,
        validate_environment=args.validate_environment_variables)) else {Action.ValidateDeclarations}

    environment_files = list(filter(lambda f: not f.is_dir(), root.glob(f"**/{env_file_name}")))
    host_files = list(filter(lambda f: not f.is_dir(), root.glob(f"**/{host_files_name}")))


    try:
        var_declarations = parsing.parse_from_files(environment_variable_declaration_from_line, environment_files)
    except UsageError as e:
        raise UsageError("parsing environment variable declarations", e.problem)
    
    try:
        host_declarations = parsing.parse_from_files(host_declaration_from_line, host_files)
    except UsageError as e:
        raise UsageError("parsing environment variable declarations", e.problem)
    
    if Action.ListVars in actions_to_perform:
        print(make_table(attributes_to_columns(
            {
                "name": "VAR NAME",
                "var_type": "TYPE",
                "description": "DESCRIPTION",
            },
            list(map(lambda x: x.data, var_declarations)),
            field_serializers={
                "var_type": lambda x: x.__name__
            })))
        
    if Action.ListHosts in actions_to_perform:
        print(make_table(attributes_to_columns(
            {
                "name": "HOST NAME",
                "description": "DESCRIPTION",
            },
            list(map(lambda x: x.data, host_declarations)))))
        
    
    if Action.ValidateDeclarations in actions_to_perform:
        raise_if_bad_environment_variable_declarations(var_declarations)
        print("All environment variables are defined and passed type validation.")
    


class Action(enum.Enum):
    ListVars = 1
    ValidateDeclarations = 2
    ListHosts = 3

def get_actions_to_perform(*, list_declarations: bool, validate_environment: bool, list_hosts: bool) -> set[Action]:
    def not_none(v: Action | None) -> t.TypeGuard[Action]:
        return v is not None
    return set[Action](filter(not_none, [
        Action.ListVars if list_declarations else None,
        Action.ValidateDeclarations if validate_environment else None,
        Action.ListHosts if list_hosts else None
    ]))


def attributes_to_columns[T](field_name_to_display_name: dict[str, str], data: t.Sequence[T], *, field_serializers: dict[str, t.Callable[[t.Any], str]] = {}) -> dict[str, list[str]]:
    
    return {
        display_name: [field_serializers.get(field_name, lambda x: x)(getattr(datum, field_name)) for datum in data] for field_name, display_name in field_name_to_display_name.items()
    }
def make_table(columns: dict[str, list[str]], column_spacing: int = 4) -> str:

    assert len(set([len(rows) for rows in columns.values()])) == 1

    num_rows = len(next(v for v in columns.values()))

    if len(columns) == 0:
        return ""
    column_lengths = {
        column: max(map(len, [column] + rows)) for column, rows in columns.items()
        }
    column_length_with_spaces = {
        column: length + column_spacing for column, length in column_lengths.items()
    }

    table = (
        make_spaced_row([(column, length) for column, length in column_length_with_spaces.items()])
        + "\n"
        + "\n".join(make_spaced_row(
            [(columns[column][row_index], column_length_with_spaces[column]) for column in columns]
        ) for row_index in range(num_rows)))
    return table

            

def make_spaced_row(row: t.Sequence[tuple[str, int]]):
    
    str_row = ""
    for column, length_with_spacing in row:
        str_row += f"{column:<{length_with_spacing}}"
    return str_row




if __name__ == "__main__":

    try:
        main()
    except UsageError as e:
        print(f"Usage Error: {e.problem}", file=sys.stderr)
        sys.exit(1)
