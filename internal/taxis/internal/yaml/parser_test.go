package yaml

import (
	"fmt"
	"reflect"
	"testing"
)

func TestParseYaml_MapOfStringLists(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		want      map[string][]string
		wantErr   bool
		errorLine int
		errorChar int
	}{
		{
			name: "Single key single item list",
			input: `
items:
 - one
`,
			want: map[string][]string{
				"items": {"one"},
			},
		},
		{
			name: "Single key multi item list",
			input: `
items:
 - one
 - two
 - three
`,
			want: map[string][]string{
				"items": {"one", "two", "three"},
			},
		},
		{
			name: "Multiple keys with lists",
			input: `
fruits:
 - apple
 - banana
vegetables:
 - carrot
 - onion
`,
			want: map[string][]string{
				"fruits":     {"apple", "banana"},
				"vegetables": {"carrot", "onion"},
			},
		},
		{
			name: "Invalid indentation",
			input: `
items:
  - one
   - b
`,
			wantErr:   true,
			errorLine: 4,
			errorChar: 4,
		},
		{
			name: "Missing dash in list",
			input: `
items:
 one
`,
			wantErr:   true,
			errorLine: 3,
			errorChar: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseYaml(tt.input)

			if (err != nil) != tt.wantErr {
				t.Fatalf("ParseYaml() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr {
				var line, char int
				_, scanErr := fmt.Sscanf(err.Error(), "line %d: char %d:", &line, &char)
				if scanErr != nil {
					t.Fatalf("failed to parse error coordinates from %q: %v", err.Error(), scanErr)
				}

				if line != tt.errorLine || char != tt.errorChar {
					t.Errorf(
						"expected error at L%d:C%d, got L%d:C%d (Error: %s)",
						tt.errorLine, tt.errorChar, line, char, err.Error(),
					)
				}
				return
			}

			gotMap, ok := got.(map[string]any)
			if !ok {
				t.Fatalf("ParseYaml() returned %T, want map[string][]string", got)
			}

			normalized := make(map[string][]string)
			for k, v := range gotMap {
				list, ok := v.([]string)
				if !ok {
					t.Fatalf("value for key %q is %T, want []string", k, v)
				}
				normalized[k] = list
			}

			if !reflect.DeepEqual(normalized, tt.want) {
				t.Errorf("ParseYaml() = %#v, want %#v", normalized, tt.want)
			}
		})
	}
}
