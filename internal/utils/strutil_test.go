package utils

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReplaceInvalidFileSystemCharacters(t *testing.T) {
	cases := map[string]string{
		"":                          "",
		"   ":                       "   ",
		"Foobar RP":                 "Foobar RP",
		"C:\\temp\\foo.bar":         "Ctempfoo.bar",
		"/tmp/foo.bar":              "tmpfoo.bar",
		"Hello 世界":                  "Hello 世界",
		"ls -la > files.txt":        "ls -la  files.txt",
		"Je prends un ☕ à Montréal": "Je prends un ☕ à Montréal",

		"mkdir -p ~/'tmp_$(date +%Y)' && cd '$_' || exit 1":                            "mkdir -p ~'tmp_$(date +%Y)' && cd '$_'  exit 1",
		": > ./log_?.txt; cat < /etc/[oO]s-* | grep -E 'ID|NAME' > ./log_1.txt 2>&1 &": "  .log_.txt; cat  etc[oO]s-  grep -E 'IDNAME'  .log_1.txt 2&1 &",
		"[ `wc -l < ./log_1.txt` -gt 0 ] && echo 'Success!' || echo 'Failed!'":         "[ `wc -l  .log_1.txt` -gt 0 ] && echo 'Success!'  echo 'Failed!'",
	}
	for in, expected := range cases {
		actual := ReplaceInvalidFileSystemCharacters(in, "")
		require.Equal(t, expected, actual)
	}
}

func TestSanitizeString(t *testing.T) {
	cases := map[string]string{
		"":                          "",
		"   ":                       "   ",
		"Foobar RP":                 "Foobar RP",
		"/tmp/foo.bar":              "/tmp/foo.bar",
		"Hello 世界":                  "Hello ",
		"Je vais à l'école":         "Je vais à l'école",
		"Je prends un ☕ à Montréal": "Je prends un  à Montréal",
	}
	for in, expected := range cases {
		actual := SanitizeString(in, "")
		require.Equal(t, expected, actual)
	}
}
