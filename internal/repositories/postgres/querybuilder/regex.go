package querybuilder

import (
	"fmt"
	"regexp"
)

// intrinsic patterns
const (
	numericPattern            = "[0-9]"
	lowercaseCharacterPattern = "[a-z]"
	uppercaseCharacterPattern = "[A-Z]"
	underscorePattern         = "[_]"
)

const (
	characterPattern                  = (lowercaseCharacterPattern + uppercaseCharacterPattern)
	characterWithUnderscorePattern    = (underscorePattern + characterPattern)
	alphanumericPattern               = (characterPattern + numericPattern)
	alphanumericWithUnderscorePattern = (underscorePattern + alphanumericPattern)
	numericWithUnderscorePattern      = (underscorePattern + numericPattern)
)

var (
	tableNameRegex = regexp.MustCompile(fmt.Sprintf("%s+%s*", characterPattern, alphanumericWithUnderscorePattern))
)

// did not factor in <x>.<column-name>.
// make it work first.
var (
	columnNameRegex          = regexp.MustCompile(fmt.Sprintf("^%s+%s*$", characterPattern, alphanumericWithUnderscorePattern))
	aliasNameRegex           = regexp.MustCompile(fmt.Sprintf("^%s+%s*$", characterPattern, alphanumericWithUnderscorePattern))
	columnNameWithAliasRegex = regexp.MustCompile(fmt.Sprintf("^%s AS %s$", columnNameRegex, aliasNameRegex))
)
