package database

import "testing"

func TestDatabaseFromYaml(t *testing.T) {

	tests := []struct {
		yaml           string
		userIdToGroups map[string][]string
		isError        bool
	}{
		{
			yaml: `
 cs100:
  instructor:
   - bob@ucr.edu
   - fred@ucr.edu


  student:
   - charlie@ucr.edu

   - fred@ucr.edu

  assistant:
   ta:
    - tom@ucr.edu


   ula:
     - dick@ucr.edu`,
			userIdToGroups: map[string][]string{
				"bob@ucr.edu":     {"cs100.instructor"},
				"fred@ucr.edu":    {"cs100.instructor", "cs100.student"},
				"charlie@ucr.edu": {"cs100.student"},
				"tom@ucr.edu":     {"cs100.assistant.ta"},
				"dick@ucr.edu":    {"cs100.assistant.ula"},
			},
		},
	}

	for _, test := range tests {
		database, err := GetGroupDBFromYaml(test.yaml)

		if (err != nil) != test.isError {
			if test.isError {
				t.Fatalf("got err=nil, want err!=nil")
			} else {
				t.Fatalf("got err=%s, want err=nil", err.Error())
			}
		}

		expectedUsers := make([]string, 0, len(test.userIdToGroups))
		for k := range test.userIdToGroups {
			expectedUsers = append(expectedUsers, k)
		}

		gotUsers := database.GetAllUsersIds()

		if diff := listSetMinus(expectedUsers, gotUsers); len(diff) != 0 {
			t.Fatalf("missing users %v", diff)
		}
		if diff := listSetMinus(gotUsers, expectedUsers); len(diff) != 0 {
			t.Fatalf("spurious users %v", diff)
		}

		for userId, expectedGroups := range test.userIdToGroups {
			gotGroups, err := database.GetGroups(userId)
			if (err != nil) != test.isError {
				if test.isError {
					t.Fatalf("got err=nil, want err!=nil")
				} else {
					t.Fatalf("got err=%s, want err=nil", err.Error())
				}
			}

			if diff := listSetMinus(expectedGroups, gotGroups); len(diff) != 0 {
				t.Errorf("missing %v as groups for %s", diff, userId)
			}

			if diff := listSetMinus(gotGroups, expectedGroups); len(diff) != 0 {
				t.Errorf("incorrect assigned %v as groups for %s", diff, userId)
			}

		}

	}

}

func listSetMinus[T comparable](l1 []T, l2 []T) []T {
	var l3 []T

outer:
	for _, e1 := range l1 {
		for _, e2 := range l2 {
			if e1 == e2 {
				continue outer
			}
		}
		l3 = append(l3, e1)
	}

	return l3
}
