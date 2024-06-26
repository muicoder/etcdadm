/*
Copyright 2020 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package hosts

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRemovesDuplicateGuardedBlocks(t *testing.T) {
	in := `
foo 1.2.3.4

# Begin host entries managed by etcd-manager[etcd] - do not edit
# End host entries managed by etcd-manager[etcd]
# Begin host entries managed by etcd-manager[etcd] - do not edit
# End host entries managed by etcd-manager[etcd]
# Begin host entries managed by etcd-manager[etcd] - do not edit
# End host entries managed by etcd-manager[etcd]
# Begin host entries managed by etcd-manager[etcd] - do not edit
# End host entries managed by etcd-manager[etcd]
# Begin host entries managed by etcd-manager[etcd] - do not edit
# End host entries managed by etcd-manager[etcd]
# Begin host entries managed by etcd-manager[etcd] - do not edit
# End host entries managed by etcd-manager[etcd]
# Begin host entries managed by etcd-manager[etcd] - do not edit
# End host entries managed by etcd-manager[etcd]
# Begin host entries managed by etcd-manager[etcd] - do not edit
# End host entries managed by etcd-manager[etcd]
`

	expected := `
foo 1.2.3.4

# Begin host entries managed by etcd-manager[etcd] - do not edit
a\t10.0.1.1 10.0.1.2
b\t10.0.2.1
c\t
# End host entries managed by etcd-manager[etcd]
`

	runTest(t, in, expected)
}

func TestRecoversFromBadNesting(t *testing.T) {
	in := `
foo 10.2.3.4

# End host entries managed by etcd-manager[etcd]
# End host entries managed by etcd-manager[etcd]
# Begin host entries managed by etcd-manager[etcd] - do not edit
# Begin host entries managed by etcd-manager[etcd] - do not edit
# End host entries managed by etcd-manager[etcd]
# Begin host entries managed by etcd-manager[etcd] - do not edit
# End host entries managed by etcd-manager[etcd]
# Begin host entries managed by etcd-manager[etcd] - do not edit
old 10.4.5.6
# End host entries managed by etcd-manager[etcd]
# Begin host entries managed by etcd-manager[etcd] - do not edit
# End host entries managed by etcd-manager[etcd]
# Begin host entries managed by etcd-manager[etcd] - do not edit
# End host entries managed by etcd-manager[etcd]

bar 10.1.2.3
`

	expected := `
foo 10.2.3.4

bar 10.1.2.3

# Begin host entries managed by etcd-manager[etcd] - do not edit
a\t10.0.1.1 10.0.1.2
b\t10.0.2.1
c\t
# End host entries managed by etcd-manager[etcd]
`

	runTest(t, in, expected)
}

func runTest(t *testing.T, in string, expected string) {
	expected = strings.Replace(expected, "\\t", "\t", -1)

	dir := t.TempDir()

	p := filepath.Join(dir, "hosts")
	key := "etcd-manager[etcd]"
	addrToHosts := map[string][]string{
		"a": {"10.0.1.2", "10.0.1.1"},
		"b": {"10.0.2.1"},
		"c": {},
	}

	if err := os.WriteFile(p, []byte(in), 0755); err != nil {
		t.Fatalf("error writing hosts file: %v", err)
	}

	for i := 0; i < 100; i++ {
		if err := updateHostsFileWithRecords(p, key, addrToHosts); err != nil {
			t.Fatalf("error updating hosts file: %v", err)
		}

		b, err := os.ReadFile(p)
		if err != nil {
			t.Fatalf("error reading output file: %v", err)
		}

		actual := string(b)
		if actual != expected {
			t.Errorf("unexpected output.\n  expected=%q\n    actual=%q", expected, actual)
		}
	}
}
