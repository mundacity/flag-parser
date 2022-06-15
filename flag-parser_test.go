package flagParser

import (
	"os"
	"strconv"
	"testing"
	"time"
)

type parsing_test_case struct {
	args           []string
	expected       []string
	name           string
	systemFlags    func() []FlagInfo
	err            error
	dateTimeFormat string
}

func _getNoSpaceBodyTestCases() []parsing_test_case {
	return []parsing_test_case{{
		args:        []string{"-b", "this is a body"},
		expected:    []string{"-b", "this is a body"},
		name:        "standard body in quotes",
		systemFlags: _getTodoAddTestCases,
		err:         nil,
	}, {
		args:        []string{"-b", "this", "is", "a", "body"},
		expected:    []string{"-b", "this is a body"},
		name:        "no other tags",
		systemFlags: _getTodoAddTestCases,
		err:         nil,
	}, {
		args:        []string{"-b", "this", "is", "a", "body", "-", "with", "a", "dash"},
		expected:    []string{"-b", "this is a body - with a dash"},
		name:        "dash char in body",
		systemFlags: _getTodoAddTestCases,
		err:         nil,
	}, {
		args:        []string{"-b", "this", "is", "a", "body", "-", "with", "a", "dash", "and", "a", "t tag", "-t", "1stTag"},
		expected:    []string{"-b", "this is a body - with a dash and a t tag", "-t", "1stTag"},
		name:        "dash char in body and subsequent tag",
		systemFlags: _getTodoAddTestCases,
		err:         nil,
	}, {
		args:        []string{"-t", "1stTag", "-b", "this", "is", "a", "body", "-", "with", "a", "dash", "and", "a", "t", "tag", "before", "the body tag"},
		expected:    []string{"-t", "1stTag", "-b", "this is a body - with a dash and a t tag before the body tag"},
		name:        "tag before body and dash char in body and subsequent tag",
		systemFlags: _getTodoAddTestCases,
		err:         nil,
	}, {
		args:        []string{"-t", "1st tag with a space", "-b", "this", "is", "a", "body", "-", "with", "a", "dash", "and", "a", "t", "tag", "before", "the body tag"},
		expected:    []string{"-t", "1st tag with a space", "-b", "this is a body - with a dash and a t tag before the body tag"},
		name:        "tag with spaces before body and dash char in body and subsequent tag",
		systemFlags: _getTodoAddTestCases,
		err:         &ExceedMaxLengthError{},
	}, {
		args:        []string{"-t", "first", "tag", "with", "a", "few", "spaces", "-b", "this", "is", "a", "body", "with", "a", "dash", "-", "or", "two", "-", "and", "with", "odd", "flag", "ordering", "and", "a", "tag", "after", "the", "body", "-m", "p", "-c", "1", "-p", "88"},
		expected:    []string{"-t", "first tag with a few spaces", "-b", "this is a body with a dash - or two - and with odd flag ordering and a tag after the body", "-m", "p", "-c", "1", "-p", "88"},
		name:        "wrong order spaces in body and tag dashes and tags after body",
		systemFlags: _getTodoAddTestCases,
		err:         &ExceedMaxLengthError{},
	}, {
		args:        []string{"-t", "tag", "w", "spc", "-c", "9", "-b", "body", "with", "spaces", "-m", "p"},
		expected:    []string{"-t", "tag w spc", "-c", "9", "-b", "body with spaces", "-m", "p"},
		name:        "space body and tag with other tags",
		systemFlags: _getTodoAddTestCases,
		err:         nil,
	}, {
		args:        []string{"this", "is", "a", "body", "with no body flag"},
		expected:    []string{"-b", "this is a body with no body flag"},
		name:        "body no quotes and no body flag",
		systemFlags: _getTodoAddTestCases,
		err:         nil,
	}, {
		args:        []string{"this", "is", "a", "body", "with no body flag", "-m", "p"},
		expected:    []string{"-m", "p", "-b", "this is a body with no body flag"},
		name:        "body no quotes no body flag but subsequent flags",
		systemFlags: _getTodoAddTestCases,
		err:         nil,
	}, {
		args:        []string{"-c", "7", "body", "with no body flag", "and spaces", "-m", "p"},
		expected:    []string{"-c", "7", "-m", "p", "-b", "body with no body flag and spaces"},
		name:        "body no quotes no body flag but other flags before and after body",
		systemFlags: _getTodoAddTestCases,
		err:         nil,
	}, {
		args:        []string{"-c", "7", "-t", "tagNoSpace", "body", "with no body flag", "and spaces", "-m", "p"},
		expected:    []string{"-c", "7", "-t", "tagNoSpace", "-m", "p", "-b", "body with no body flag and spaces"},
		name:        "correct parse for spaceless 10-digit tag",
		systemFlags: _getTodoAddTestCases,
		err:         &ExceedMaxLengthError{},
	}, {
		args:        []string{"-c", "7", "-t", "t1 & t2", "-b", "body", "with no body flag", "and spaces", "-m", "p"},
		expected:    []string{"-c", "7", "-t", "t1 & t2", "-b", "body with no body flag and spaces", "-m", "p"},
		name:        "correct parse for spaceful tag below max len",
		systemFlags: _getTodoAddTestCases,
		err:         nil,
	}, {
		args:        []string{"-c", "7", "-t", "tag", "with", "spaces", "body", "with no body flag", "and spaces", "-m", "p"},
		expected:    []string{"-c", "7", "-t", "tag with s", "-m", "p", "-b", "paces body with no body flag and spaces"},
		name:        "Malformed: tag of first 10 chars and implicity body with remainder",
		systemFlags: _getTodoAddTestCases,
		err:         &ExceedMaxLengthError{},
	}, {
		args:        []string{"-c", "7", "-t", "tag", "spaces", "-b", "body", "with body flag", "and spaces", "-m", "p"},
		expected:    []string{"-c", "7", "-t", "tag spaces", "-b", "body with body flag and spaces", "-m", "p"},
		name:        "tag with space before body with flag and no quotes and 1 flag before tag and 1 flag after body",
		systemFlags: _getTodoAddTestCases,
		err:         nil,
	}, {
		args:        []string{"-m", "commit", "message", "with", "spaces"},
		expected:    []string{"-m", "commit message with spaces"},
		name:        "commit with spaces",
		systemFlags: _getCommitTestCases,
		err:         nil,
	}, {
		args:        []string{"-m", "commit", "message", "with", "spaces", "-f", "unknown flag"},
		expected:    []string{},
		name:        "commit with spaces and unknown flag",
		systemFlags: _getCommitTestCases,
		err:         &UserArgsContainsUnknownFlag{},
	}, {
		args:        []string{"-t", "testing", "-m", "p", "new", "item", "with", "spaces"},
		expected:    []string{"-t", "testing", "-m", "p", "-b", "new item with spaces"},
		name:        "tag below max and implicit body",
		systemFlags: _getTodoAddTestCases,
		err:         nil,
	},
	}
}

func _getVariableMaxTagLengthTestCases() []parsing_test_case {
	return []parsing_test_case{{
		args:        []string{"-t", "tag1|tag2|tag3"},
		expected:    []string{"-t", "tag1|tag2|tag3"},
		name:        "multiple tag input",
		systemFlags: _getTodoAddTestCases,
		err:         nil,
	}, {
		args:        []string{"-t", "tag1;tag2;tag3"},
		expected:    []string{"-t", "tag1;tag2;tag3"},
		name:        "multiple tag input2",
		systemFlags: _getTodoAddTestCases,
		err:         nil,
	}}
}

func _getDateParsingTestCases() []parsing_test_case {
	return []parsing_test_case{{
		args:        []string{"-d", "-2m:-1m"},
		expected:    []string{"-d", "2022-01-14:2022-02-14"},
		name:        "date range both months",
		systemFlags: _getCanonicalFlagsForGodoGettingTests,
		err:         nil,
	}, {
		args:        []string{"-d", "-2m:-7d"},
		expected:    []string{"-d", "2022-01-14:2022-03-07"},
		name:        "date range month operand day operand",
		systemFlags: _getCanonicalFlagsForGodoGettingTests,
		err:         nil,
	}, {
		args:        []string{"-d", "-2y:-8d"},
		expected:    []string{"-d", "2020-03-14:2022-03-06"},
		name:        "date range year operand day operand",
		systemFlags: _getCanonicalFlagsForGodoGettingTests,
		err:         nil,
	}, {
		args:        []string{"-d", "-2y:"},
		expected:    []string{},
		name:        "date range bad input missing operand",
		systemFlags: _getCanonicalFlagsForGodoGettingTests,
		err:         &MalformedDateRangeError{},
	}, {
		args:        []string{"-d", "-1y:1"},
		expected:    []string{},
		name:        "date range bad input missing time indicator",
		systemFlags: _getCanonicalFlagsForGodoGettingTests,
		err:         &MalformedDateRangeError{},
	}, {
		args:        []string{"-d", "-1y:d"},
		expected:    []string{},
		name:        "date range bad input missing time integer",
		systemFlags: _getCanonicalFlagsForGodoGettingTests,
		err:         &UnknownDateInputError{},
	}, {
		args:        []string{"-b", "this is a spaceful body", "-d", "1y1m2d"},
		expected:    []string{"-b", "this is a spaceful body", "-d", "2023-04-16"},
		name:        "add spaceless time",
		systemFlags: _getTodoAddTestCases,
		err:         nil,
	}, {
		args:        []string{"this is a spaceful body", "-d", "1y", "1m", "2d"},
		expected:    []string{"-d", "2023-04-16", "-b", "this is a spaceful body"},
		name:        "add spaced time",
		systemFlags: _getTodoAddTestCases,
		err:         nil,
	}, {
		args:        []string{"this is a spaceful body", "-d", "-1y", "1m", "2d"},
		expected:    []string{"-d", "2021-04-16", "-b", "this is a spaceful body"},
		name:        "spaced negative time",
		systemFlags: _getTodoAddTestCases,
		err:         nil,
	}, {
		args:        []string{"this is a spaceful body", "-d", "-1y"},
		expected:    []string{"-d", "2021-03-14", "-b", "this is a spaceful body"},
		name:        "negative year only",
		systemFlags: _getTodoAddTestCases,
		err:         nil,
	}, {
		args:        []string{"this is a spaceful body", "-d", "-1m"},
		expected:    []string{"-d", "2022-02-14", "-b", "this is a spaceful body"},
		name:        "negative month only",
		systemFlags: _getTodoAddTestCases,
		err:         nil,
	}, {
		args:        []string{"this is a spaceful body", "-d", "-4d"},
		expected:    []string{"-d", "2022-03-10", "-b", "this is a spaceful body"},
		name:        "negative day only",
		systemFlags: _getTodoAddTestCases,
		err:         nil,
	}, {
		args:        []string{"this is a spaceful body", "-d", "-4d-1m-3y"},
		expected:    []string{"-d", "2019-02-10", "-b", "this is a spaceful body"},
		name:        "all negative time vals no space",
		systemFlags: _getTodoAddTestCases,
		err:         nil,
	}, {
		args:           []string{"this is a spaceful body", "-d", "-4d", "-1m", "-3y"},
		expected:       []string{"-d", "2019-02-10", "-b", "this is a spaceful body"},
		name:           "all negative time vals with spaces",
		systemFlags:    _getTodoAddTestCases,
		err:            nil,
		dateTimeFormat: "",
	}, {
		args:           []string{"this is a spaceful body", "-d", "-4 d", "-1", "m", "-", "3", "y"},
		expected:       []string{"-d", "2019-02-10", "-b", "this is a spaceful body"},
		name:           "all negative time vals with weird spacing",
		systemFlags:    _getTodoAddTestCases,
		err:            nil,
		dateTimeFormat: "",
	}, {
		args:           []string{"this", "is", "a spaceful", "body", "-4 d", "-1", "m", "-", "3", "y"},
		expected:       []string{"-b", "this is a spaceful body -4 d -1 m - 3 y"},
		name:           "negative time vals no date or body flag",
		systemFlags:    _getTodoAddTestCases,
		err:            nil,
		dateTimeFormat: "",
	}, {
		args:           []string{"-b", "valid", "spaceful", "body with invalid date input", "-4 d", "-1", "m", "-", "3", "y"},
		expected:       []string{"-b", "valid spaceful body with invalid date input -4 d -1 m - 3 y"},
		name:           "negative time vals no date flag",
		systemFlags:    _getTodoAddTestCases,
		err:            nil,
		dateTimeFormat: "",
	}}
}

func _getTestCasesForGoDooGetting() []parsing_test_case {
	return []parsing_test_case{{
		args:        []string{"-b", "testing", "a", "body", "with", "spaces"},
		expected:    []string{"-b", "testing a body with spaces"},
		name:        "spaceful body search",
		systemFlags: _getCanonicalFlagsForGodoGettingTests,
		err:         nil,
	}, {
		args:        []string{"-t", "testing"},
		expected:    []string{"-t", "testing"},
		name:        "basic tag search",
		systemFlags: _getCanonicalFlagsForGodoGettingTests,
		err:         nil,
	}, {
		args:        []string{"testing", "tagless", "body", "search", "with spaces"},
		expected:    []string{"-b", "testing tagless body search with spaces"},
		name:        "tagless body spaceful search",
		systemFlags: _getCanonicalFlagsForGodoGettingTests,
		err:         nil,
	}, {
		args:        []string{"testing", "spaceful", "tagless", "body", "search", "with another tag", "-c", "3"},
		expected:    []string{"-c", "3", "-b", "testing spaceful tagless body search with another tag"},
		name:        "tagless body spaceful search plus another tag",
		systemFlags: _getCanonicalFlagsForGodoGettingTests,
		err:         nil,
	}, {
		args:        []string{"-d", "2d"},
		expected:    []string{"-d", "2022-03-16"},
		name:        "search by deadline 2 days hence",
		systemFlags: _getCanonicalFlagsForGodoGettingTests,
		err:         nil,
	}, {
		args:        []string{"-d", "-2d"},
		expected:    []string{"-d", "2022-03-12"},
		name:        "search by deadline 2 days before",
		systemFlags: _getCanonicalFlagsForGodoGettingTests,
		err:         nil,
	}, {
		args:        []string{"-d", "2022-04-17"},
		expected:    []string{"-d", "2022-04-17"},
		name:        "search by literal date string",
		systemFlags: _getCanonicalFlagsForGodoGettingTests,
		err:         nil,
	}, {
		args:           []string{"-a"},
		expected:       []string{"-a"},
		name:           "get all",
		systemFlags:    _getCanonicalFlagsForGodoGettingTests,
		err:            nil,
		dateTimeFormat: "",
	}}
}

// probably overkill - stems from fact I tried to reassemble in original order but couldn't make it work
// some history in a few commits if I need it
func _getTestCasesForGoDooEditing() []parsing_test_case {
	return []parsing_test_case{{
		args:        []string{"-i", "38", "--append", "-B", "adding", "to", "body", "with", "edit command"},
		expected:    []string{"-i", "38", "-B", "adding to body with edit command", "--append"},
		name:        "spaceful body edit by id",
		systemFlags: _getCanonicalFlagsForGodooEditing,
		err:         nil,
	}, {
		args:        []string{"--append", "-b", "body", "key", "phrase", "-B", "adding", "to", "body", "with", "edit command"},
		expected:    []string{"-b", "body key phrase", "-B", "adding to body with edit command", "--append"},
		name:        "spaceful body search & edit",
		systemFlags: _getCanonicalFlagsForGodooEditing,
		err:         nil,
	}, {
		args:        []string{"--append", "body", "key", "phrase", "-B", "adding", "to", "body", "with", "edit command"},
		expected:    []string{"-b", "body key phrase", "-B", "adding to body with edit command", "--append"},
		name:        "spaceful body search & edit but no body search tag",
		systemFlags: _getCanonicalFlagsForGodooEditing,
		err:         nil,
	}, {
		args:        []string{"-t", "workNotes", "-B", "work", "note", "edited", "by", "tag", "and", "parent id", "--append", "-c", "43", "-C", "88"},
		expected:    []string{"-t", "workNotes", "-B", "work note edited by tag and parent id", "-c", "43", "-C", "88", "--append"},
		name:        "long but fairly simple",
		systemFlags: _getCanonicalFlagsForGodooEditing,
		err:         nil,
	}, {
		args:        []string{"-B", "work", "note", "edited", "by", "tag", "and", "parent id", "--append", "body", "key", "search", "phrase", "-c", "43", "-C", "88"},
		expected:    []string{"-B", "work note edited by tag and parent id", "-b", "body key search phrase", "-c", "43", "-C", "88", "--append"},
		name:        "long but fairly simple - missing body tag",
		systemFlags: _getCanonicalFlagsForGodooEditing,
		err:         nil,
	}, {
		args:        []string{"-B", "work", "note", "edited", "by", "tag", "and", "parent id", "--append", "body", "key", "search", "phrase", "--replace", "-c", "43", "-C", "88"},
		expected:    []string{"-B", "work note edited by tag and parent id", "-b", "body key search phrase", "-c", "43", "-C", "88", "--append", "--replace"},
		name:        "2 standalone flags plus missing body tag",
		systemFlags: _getCanonicalFlagsForGodooEditing,
		err:         nil,
	}, {
		args:        []string{"-B", "work", "note", "edited", "by", "tag", "and", "parent id", "--append", "body", "key", "search", "phrase", "--replace", "-c", "43", "--nonsensea", "-C", "88"},
		expected:    []string{"-B", "work note edited by tag and parent id", "-b", "body key search phrase", "-c", "43", "-C", "88", "--append", "--replace", "--nonsensea"},
		name:        "3 standalone flags plus missing body tag",
		systemFlags: _getCanonicalFlagsForGodooEditing,
		err:         nil,
	}, {
		args:        []string{"--nonsenseb", "-B", "work", "note", "edited", "by", "tag", "and", "parent id", "--append", "body", "key", "search", "phrase", "--replace", "-c", "43", "--nonsensea", "-C", "88"},
		expected:    []string{"-B", "work note edited by tag and parent id", "-b", "body key search phrase", "-c", "43", "-C", "88", "--nonsenseb", "--append", "--replace", "--nonsensea"},
		name:        "4 standalone flags plus missing body tag",
		systemFlags: _getCanonicalFlagsForGodooEditing,
		err:         nil,
	}, {
		args:        []string{"--nonsenseb", "--append", "-b", "body", "key", "search", "phrase", "--replace", "-c", "43", "--nonsensea", "-C", "88"},
		expected:    []string{"-b", "body key search phrase", "-c", "43", "-C", "88", "--nonsenseb", "--append", "--replace", "--nonsensea"},
		name:        "4 standalone flags plus missing body tag",
		systemFlags: _getCanonicalFlagsForGodooEditing,
		err:         nil,
	}}
}

func getBadInputTesting() []parsing_test_case {
	return []parsing_test_case{{
		args:        []string{"-F", "-d"},
		expected:    []string{},
		name:        "missing arg",
		systemFlags: _getCanonicalFlagsForGodooEditing,
		err:         &MissingArgumentError{},
	}, {
		args:        []string{"-a", "-d", "0d", "-t"},
		expected:    []string{""},
		name:        "missing arg2",
		systemFlags: _getCanonicalFlagsForGodoGettingTests,
		err:         &MissingArgumentError{},
	}, {
		args:        []string{"-a", "-f"},
		expected:    []string{"-a", "-f"},
		name:        "standalones only",
		systemFlags: _getCanonicalFlagsForGodoGettingTests,
		err:         nil,
	}, {
		args:        []string{"-3", "garbage", "-u", "input"},
		expected:    []string{},
		name:        "garbage input",
		systemFlags: _getCanonicalFlagsForGodoGettingTests,
		err:         &UserArgsContainsUnknownFlag{},
	}, {
		args:        []string{"-t", "1d2m3y"},
		expected:    []string{"-t", "1d2m3y"},
		name:        "wrong arg data type",
		systemFlags: _getCanonicalFlagsForGodoGettingTests,
		err:         nil,
	}, {
		args:        []string{"-t", "33", "-u", "input"},
		expected:    []string{},
		name:        "wrong input data type and garbage flag/arg",
		systemFlags: _getCanonicalFlagsForGodoGettingTests,
		err:         &UserArgsContainsUnknownFlag{},
	}}
}
func TestVariableTagLengthsWithMultipleTags(t *testing.T) {
	os.Setenv("MAX_LENGTH", "2000")
	os.Setenv("MAX_TAG_LENGTH", "2000")
	os.Setenv("MAX_INT_DIGITS", "4")

	tcs := _getVariableMaxTagLengthTestCases()
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			_runParseTest(t, tc)
		})
	}
}

func TestParseGoDooAdding(t *testing.T) {

	os.Setenv("MAX_LENGTH", "2000")
	os.Setenv("MAX_TAG_LENGTH", "10")
	os.Setenv("MAX_INT_DIGITS", "4")

	tcs := _getNoSpaceBodyTestCases()
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			_runParseTest(t, tc)
		})
	}
}

func TestParseGoDooGetting(t *testing.T) {
	os.Setenv("MAX_LENGTH", "2000")
	os.Setenv("MAX_TAG_LENGTH", "10")
	os.Setenv("MAX_INT_DIGITS", "4")
	os.Setenv("DATETIME_FORMAT", "2006-01-02")

	tcs := _getTestCasesForGoDooGetting()
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			_runParseTest(t, tc)
		})
	}
}

func TestParseGoDooEditing(t *testing.T) {
	os.Setenv("MAX_LENGTH", "2000")
	os.Setenv("MAX_INT_DIGITS", "4")
	os.Setenv("DATETIME_FORMAT", "2006-01-02")
	os.Setenv("TAG_DELIMITER", "*")

	tcs := _getTestCasesForGoDooEditing()
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			_runParseTest(t, tc)
		})
	}
}

func _runParseTest(t *testing.T, tc parsing_test_case) {

	fp := NewFlagParser(tc.systemFlags(), tc.args, WithNowAs(returnNowString(), "2006-01-02"))
	got, err := fp.ParseUserInput()

	if err != nil && err == tc.err {
		t.Logf(">>>>PASSED: operation threw correct error. \nExp\t'%v', \nGot\t'%v'", tc.err, err)
		return
	} else if err != nil && err != tc.err {
		t.Errorf(">>>>FAILED: operation threw incorrect error. \nExp\t'%v', \nGot\t'%v'", tc.err, err)
	}

	if len(tc.expected) != len(got) {
		t.Errorf(">>>>FAILED: slices not same length. \nInp\t'%v' \nExp\t'%v' \nGot\t'%v'", tc.args, tc.expected, got)
		return
	}

	if _slicesAreTheSame(tc.expected, got) {
		t.Logf(">>>>PASSED: slices are equal in value. \nInp\t'%v' \nExp\t'%v', \nGot\t'%v'", tc.args, tc.expected, got)
	} else {
		t.Errorf(">>>>FAILED: slices are not equal in value. \nInp\t'%v' \nExp\t'%v', \nGot\t'%v'", tc.args, tc.expected, got)
	}
}

func _slicesAreTheSame(s1 []string, s2 []string) bool {
	for i, s := range s1 {
		if s != s2[i] {
			return false
		}
	}
	return true
}

func _getTodoAddTestCases() []FlagInfo {
	var ret []FlagInfo

	lenMax, _ := strconv.Atoi(os.Getenv("MAX_LENGTH"))
	maxTagLen, _ := strconv.Atoi(os.Getenv("MAX_TAG_LENGTH"))
	maxIntDigits, _ := strconv.Atoi(os.Getenv("MAX_INT_DIGITS"))

	f2 := FlagInfo{FlagName: "-b", FlagType: Str, MaxLen: lenMax}
	f3 := FlagInfo{FlagName: "-m", FlagType: Str, MaxLen: 1}
	f4 := FlagInfo{FlagName: "-t", FlagType: Str, MaxLen: maxTagLen}
	f5 := FlagInfo{FlagName: "-c", FlagType: Integer, MaxLen: maxIntDigits}
	f6 := FlagInfo{FlagName: "-p", FlagType: Integer, MaxLen: maxIntDigits}
	f7 := FlagInfo{FlagName: "-d", FlagType: DateTime, MaxLen: 20}

	ret = append(ret, f2, f3, f4, f5, f6, f7)
	return ret
}

func _getCommitTestCases() []FlagInfo {
	var ret []FlagInfo

	f2 := FlagInfo{FlagName: "-m", FlagType: Str, MaxLen: 2000}
	ret = append(ret, f2)
	return ret
}

func _getCanonicalFlagsForGodoGettingTests() []FlagInfo {
	var ret []FlagInfo
	lenMax, _ := strconv.Atoi(os.Getenv("MAX_LENGTH"))
	maxIntDigits, _ := strconv.Atoi(os.Getenv("MAX_INT_DIGITS"))
	maxTagLen, _ := strconv.Atoi(os.Getenv("MAX_TAG_LENGTH"))

	f8 := FlagInfo{FlagName: "-b", FlagType: Str, MaxLen: lenMax}
	f2 := FlagInfo{FlagName: "-i", FlagType: Integer, MaxLen: maxIntDigits}
	f3 := FlagInfo{FlagName: "-n", FlagType: Boolean}
	f4 := FlagInfo{FlagName: "-d", FlagType: DateTime, MaxLen: 20}
	f5 := FlagInfo{FlagName: "-t", FlagType: Str, MaxLen: maxTagLen}
	f6 := FlagInfo{FlagName: "-c", FlagType: Integer, MaxLen: maxIntDigits}
	f7 := FlagInfo{FlagName: "-p", FlagType: Integer, MaxLen: maxIntDigits}
	f9 := FlagInfo{FlagName: "-e", FlagType: DateTime, MaxLen: 20}
	f10 := FlagInfo{FlagName: "-a", FlagType: Boolean, Standalone: true}
	f11 := FlagInfo{FlagName: "-f", FlagType: Boolean, Standalone: true}

	ret = append(ret, f8, f2, f3, f4, f5, f6, f7, f9, f10, f11)

	return ret
}

func _getCanonicalFlagsForGodooEditing() []FlagInfo {
	var ret []FlagInfo

	lenMax := 2000
	maxIntDigits := 4

	f1 := FlagInfo{FlagName: "-b", FlagType: Str, MaxLen: lenMax}
	f2 := FlagInfo{FlagName: "-i", FlagType: Integer, MaxLen: maxIntDigits}
	f3 := FlagInfo{FlagName: "-d", FlagType: DateTime, MaxLen: 20}
	f4 := FlagInfo{FlagName: "-t", FlagType: Str, MaxLen: lenMax}
	f5 := FlagInfo{FlagName: "-c", FlagType: Integer, MaxLen: maxIntDigits}
	f6 := FlagInfo{FlagName: "-e", FlagType: DateTime, MaxLen: 20}

	f7 := FlagInfo{FlagName: "--append", FlagType: Boolean, Standalone: true}
	f8 := FlagInfo{FlagName: "--replace", FlagType: Boolean, Standalone: true}

	f9 := FlagInfo{FlagName: "-B", FlagType: Str, MaxLen: lenMax}
	f10 := FlagInfo{FlagName: "-T", FlagType: Str, MaxLen: lenMax}
	f11 := FlagInfo{FlagName: "-C", FlagType: Integer, MaxLen: maxIntDigits}
	f12 := FlagInfo{FlagName: "-D", FlagType: Str, MaxLen: 20}
	f13 := FlagInfo{FlagName: "-F", FlagType: Boolean, Standalone: true}

	f14 := FlagInfo{FlagName: "--nonsensea", FlagType: Boolean, Standalone: true}
	f15 := FlagInfo{FlagName: "--nonsenseb", FlagType: Boolean, Standalone: true}

	ret = append(ret, f1, f2, f3, f4, f5, f6, f7, f8, f9, f10, f11, f12, f13, f14, f15)
	return ret
}

func TestDateParsing(t *testing.T) {
	os.Setenv("MAX_LENGTH", "2000")
	os.Setenv("MAX_TAG_LENGTH", "10")
	os.Setenv("MAX_INT_DIGITS", "4")
	os.Setenv("DATETIME_FORMAT", "2006-01-02")

	tcs := _getDateParsingTestCases()
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			_runDateParseTest(t, tc)
		})
	}
}

func _runDateParseTest(t *testing.T, tc parsing_test_case) {
	fp := NewFlagParser(tc.systemFlags(), tc.args, WithNowAs(returnNowString(), "2006-01-02"))
	got, err := fp.ParseUserInput()

	if err != nil {
		if err == tc.err {
			t.Logf(">>>>PASSED: operation threw error. \nExp\t'%v', \nGot\t'%v'", tc.err, err)
			return
		} else {
			t.Errorf(">>>>FAILED: operation threw wrong error. \nExp\t'%v', \nGot\t'%v'", tc.err, err)
		}
	}

	if len(tc.expected) != len(got) {
		t.Errorf(">>>>FAILED: slices not same length. \nInp\t'%v' \nExp\t'%v' \nGot\t'%v'", tc.args, tc.expected, got)
		return
	}

	if _slicesAreTheSame(tc.expected, got) {
		t.Logf(">>>>PASSED: slices are equal in value. \nInp\t'%v' \nExp\t'%v', \nGot\t'%v'", tc.args, tc.expected, got)
	} else {
		t.Errorf(">>>>FAILED: slices are not equal in value. \nInp\t'%v' \nExp\t'%v', \nGot\t'%v'", tc.args, tc.expected, got)
	}
}

func returnNowString() string {
	n := time.Date(2022, 03, 14, 0, 0, 0, 0, time.UTC)
	return StringFromDate(n)
}

func TestBadInput(t *testing.T) {
	os.Setenv("MAX_LENGTH", "2000")
	os.Setenv("MAX_TAG_LENGTH", "10")
	os.Setenv("MAX_INT_DIGITS", "4")
	os.Setenv("DATETIME_FORMAT", "2006-01-02")

	tcs := getBadInputTesting()
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			_runParseTest(t, tc)
		})
	}
}
