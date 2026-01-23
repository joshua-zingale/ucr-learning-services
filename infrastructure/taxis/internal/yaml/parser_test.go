package yaml

import (
	"fmt"
	"reflect"
	"testing"
)

func TestParseYaml(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		want      YamlValue
		wantErr   bool
		errorLine int
		errorChar int
	}{
		{
			name:    "root list",
			input:   "- tom \n- dick\n   \n- harry\n    ",
			want:    YamlList{"tom", "dick", "harry"},
			wantErr: false,
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
					t.Errorf("expected error at L%d:C%d, got L%d:C%d (Error: %s)",
						tt.errorLine, tt.errorChar, line, char, err.Error())
				}
				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseYaml() = %v, want %v", got, tt.want)
			}
		})
	}
}
