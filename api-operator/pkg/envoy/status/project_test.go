package status

import (
	"reflect"
	"testing"
)

func TestUpdatedProjects(t *testing.T) {
	var tests = []struct {
		name      string
		st, newSt *ProjectsStatus
		want      map[string]bool
	}{
		// Same ingresses
		{
			name:  "No hosts added or deleted",
			st:    &ProjectsStatus{"foo/ing1": map[string]string{"a_com": "_", "b_com": "_"}},
			newSt: &ProjectsStatus{"foo/ing1": map[string]string{"a_com": "_", "b_com": "_"}},
			want:  map[string]bool{"a_com": true, "b_com": true},
		},
		{
			name:  "Add and delete hosts",
			st:    &ProjectsStatus{"foo/ing1": map[string]string{"a_com": "_", "b_com": "_"}},
			newSt: &ProjectsStatus{"foo/ing1": map[string]string{"a_com": "_", "c_com": "_"}},
			want:  map[string]bool{"a_com": true, "b_com": true, "c_com": true},
		},
		// Different ingresses
		{
			name:  "Add new ingress",
			st:    &ProjectsStatus{"foo/ing1": map[string]string{"a_com": "_", "b_com": "_"}},
			newSt: &ProjectsStatus{"foo/ing2": map[string]string{"b_com": "_", "c_com": "_"}},
			want:  map[string]bool{"b_com": true, "c_com": true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := tt.st.UpdatedProjects(tt.newSt)
			if !reflect.DeepEqual(p, tt.want) {
				t.Errorf("%v.UpdatedProjects(%v) = %v; want %v", tt.st, tt.newSt, p, tt.want)
			}
		})
	}
}
