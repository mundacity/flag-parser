package flagParser

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
)

type flag_origin int

const (
	user flag_origin = iota
	system
)

type FlagParser struct {
	canonicalFlags  []FlagInfo
	userPassedFlags [][]string
	system_intKey   map[int]FlagInfo
	system_strKey   map[string]flag_info_key
	user_intKey     map[int]string
	user_strKey     map[string]int
	implicitFlag    string
	HasUnknownFlags bool
	DateTimeLayout  string
	NowMoment       time.Time
}

type FlagDataType string

const (
	Str      FlagDataType = "string"
	Integer  FlagDataType = "int"
	Boolean  FlagDataType = "bool"
	DateTime FlagDataType = "dateTime"
)

type FlagInfo struct {
	FlagName   string
	FlagType   FlagDataType
	MaxLen     int
	Standalone bool
}

type flag_info_key struct {
	index      int
	flgType    FlagDataType
	maxLen     int
	standalone bool
}

type NowMomentFunc func(*FlagParser)

func WithNowAs(nowStr, dateTimeFormat string) NowMomentFunc {
	return func(fp *FlagParser) {
		fp.DateTimeLayout = dateTimeFormat
		fp.NowMoment, _ = time.Parse(fp.DateTimeLayout, nowStr)
	}
}

// Sets up a new FlagParser. allFlags[0] assumed to be implicit flag
func NewFlagParser(allFlags []FlagInfo, userFlags []string, nowFunc NowMomentFunc) *FlagParser {
	fp := FlagParser{canonicalFlags: allFlags}
	fp.userPassedFlags = append(fp.userPassedFlags, userFlags)
	fp.implicitFlag = allFlags[0].FlagName

	fp.system_intKey = make(map[int]FlagInfo)
	fp.system_strKey = make(map[string]flag_info_key)

	for i, fi := range allFlags {
		fp.system_intKey[i] = fi
		fik := flag_info_key{index: i, flgType: fi.FlagType, maxLen: fi.MaxLen, standalone: fi.Standalone}
		fp.system_strKey[fi.FlagName] = fik
	}

	err := fp.setupUserMaps(userFlags)
	if err != nil {
		fp.HasUnknownFlags = true // don't return error from constructor
	}

	nowFunc(&fp)
	return &fp
}

func (fp *FlagParser) CheckInitialisation() error {
	if fp.system_intKey == nil || fp.system_strKey == nil {
		return &FlagMapperInitialisationError{}
	}
	return nil
}

// Populate user input (flag/arg) maps. Separate method supports
// use of implicit flags
func (fp *FlagParser) setupUserMaps(args []string) error {
	fp.user_intKey, fp.user_strKey = make(map[int]string), make(map[string]int)
	for i, s := range args {

		if len(s) > 1 && strings.HasPrefix(s, "-") {
			_, inCanonicalList := fp.GetIndexFromFlagValue(system, s)
			if !inCanonicalList {
				// allows for negative number input (for shorthand date parsing)
				asRunes := []rune(s)
				_, err := strconv.Atoi(string(asRunes[1])) // 1 after asRunes[0] (which = "-")
				if err != nil {
					return &UserArgsContainsUnknownFlag{}
				}
			}
		}

		fp.user_intKey[i], fp.user_strKey[s] = s, i
	}
	return nil
}

// Get flag from a given index in canonical or
// user-provided flag list. Defaults to canonical list
func (fp FlagParser) GetFlagValueFromIndex(fType flag_origin, idx int) (string, bool) {

	switch fType {
	case user:
		v, exists := fp.user_intKey[idx]
		return v, exists
	default:
		x, e := fp.system_intKey[idx]
		if e {
			return x.FlagName, e
		}
	}

	return "", false
}

// Get index from a given flag in canonical
// or user-provided flag list. Defaults to canonical list
func (fp FlagParser) GetIndexFromFlagValue(fType flag_origin, flag string) (int, bool) {

	switch fType {
	case user:
		v, exists := fp.user_strKey[flag]
		return v, exists
	default:
		x, e := fp.system_strKey[flag]
		if e {
			return x.index, e
		}
	}

	return -1, false
}

// Get flag details from flag name. Canonical only
func (fp FlagParser) GetFlagInfoFromName(name string) (flag_info_key, bool) {
	v, e := fp.system_strKey[name]
	return v, e
}

// Get location of canonical flags in user-passed flags
func (fp *FlagParser) GetFlagLocations(iteration int) []int {
	ret := []int{}

	for i, s := range fp.userPassedFlags[iteration] {
		_, exists := fp.GetIndexFromFlagValue(system, s) // prevents use of non-canonical flags
		if exists {
			ret = append(ret, i)
		}
	}

	return ret
}

// Compares user-provided flag input with canonical list.
// Can handle both use & non-use of quotation marks.
//
// Also handles implicit flags - or flags that can be assumed even if not provided.
func (fp *FlagParser) ParseUserInput() ([]string, error) {
	var newArgs []string
	if fp.HasUnknownFlags {
		return newArgs, &UserArgsContainsUnknownFlag{}
	}

	if len(fp.userPassedFlags[0]) == 1 {
		return fp.userPassedFlags[0], nil
	}
	newArgs, err := fp.parse()
	if err != nil {
		return nil, err
	}
	fp._updateUserMaps(newArgs)

	return newArgs, nil
}

func (fp *FlagParser) parse() ([]string, error) {
	ret := fp.handleSpaces()

	latest := fp._updateUserMaps(ret)
	ufLocations := fp.GetFlagLocations(latest)

	ret = fp.handleNumericalInput(ret, ufLocations)

	var removed bool
	var standalones map[int]string
	if removed, ret, standalones = fp.removeStandaloneFlags(ret, ufLocations); removed {
		latest = fp._updateUserMaps(ret)
		ufLocations = fp.GetFlagLocations(latest)
	}

	err := checkForMissingArgs(ret, ufLocations)
	if err != nil {
		return ret, err
	}
	ret, insuff := fp.handleInsufficientFlags(ret, ufLocations)
	if insuff {
		latest = fp._updateUserMaps(ret)
		ufLocations = fp.GetFlagLocations(latest)
	}

	ret, err = fp.handleArgumentLengthAndRemainders(ret, ufLocations)
	if err != nil {
		return ret, err
	}
	ret, err = fp.handleDates(ret, ufLocations)
	if err != nil {
		return ret, err
	}

	if removed {
		ret = fp.reassemble(ret, standalones)
	}

	return ret, nil
}

// Handles spaces in user-passed arguments relative to flags. Moves
// flag-less arguments to the end of the input.
//
// Multiple-string input that falls between two canonical flags
// is condensed to a single string and allocated to the preceding flag.
// If between positon zero and a canonical flag, its assumed to be the
// arg for an implicit flag.
func (fp *FlagParser) handleSpaces() []string {
	var ret, suffix []string
	latest := len(fp.userPassedFlags) - 1
	usrArgs := fp.userPassedFlags[latest] // latest version; first time = actual user input

	flagLocations := fp.GetFlagLocations(latest)
	if len(flagLocations) == 0 {
		ret = append(ret, StringFromSlice(usrArgs)) // no flags, return input
		return ret
	}

	for i, flgLoc := range flagLocations {

		if i == 0 && flgLoc != 0 {
			// most likely an arg with an implicit flag
			suffix = append(suffix, StringFromSlice(usrArgs[i:flgLoc]))
		}

		start := flgLoc + 1

		if i+1 < len(flagLocations) { // more than one flag left
			end := flagLocations[i+1]
			arg := StringFromSlice(usrArgs[start:end])
			if len(arg) > 0 {
				ret = append(ret, usrArgs[flgLoc], arg)
			} else {
				ret = append(ret, usrArgs[flgLoc]) // standalone flags
			}
		}
		if i+1 == len(flagLocations) { // one more flag
			arg := StringFromSlice(usrArgs[start:])
			if len(arg) > 0 {
				ret = append(ret, usrArgs[flgLoc], arg)
			} else {
				ret = append(ret, usrArgs[flgLoc]) // standalone flags
			}
		}
	}
	if len(suffix) > 0 {
		ret = append(ret, suffix...)
	}
	return ret
}

// Check for numbers at start of args. If FlagInfo.FlagType is integer,
// numerical data taken as arg & remainder appended to end of input
func (fp *FlagParser) handleNumericalInput(input []string, locs []int) []string {
	for _, v := range locs {

		fi, _ := fp.GetFlagInfoFromName(input[v])
		if fi.flgType != Integer {
			continue
		}

		hasNums, nums := fp.argHasNumericalPrefix(input[v+1])
		if !hasNums {
			continue
		}

		arg := nums
		runes := []rune(input[v+1])
		remainder := strings.Trim(string(runes[len(nums):]), " ")
		if len(remainder) > 0 {
			input = append(input, remainder)
			input[v+1] = arg
		}
	}
	return input
}

// Determines whether input string begins with any numbers.
// If so, return true plus numerical input
func (fp *FlagParser) argHasNumericalPrefix(input string) (bool, string) {
	nums := ""
	for _, v := range input {

		_, err := strconv.Atoi(string(v))
		if err == nil {
			nums += string(v)
		} else {
			break
		}
	}
	if len(nums) > 0 {
		return true, nums
	}
	return false, ""
}

// Removes standalone flags from input. Makes it easier to
// determine if implicit flags are missing
func (fp *FlagParser) removeStandaloneFlags(input []string, locs []int) (removed bool, output []string, standaloneLocs map[int]string) {
	standaloneLocs = make(map[int]string)

	for _, v := range locs {
		fi, ok := fp.GetFlagInfoFromName(input[v])
		if ok {
			if fi.standalone {
				standaloneLocs[v] = input[v]
			}
		}
	}

	for i, v := range input {
		if _, ok := standaloneLocs[i]; !ok {
			output = append(output, v)
		}
	}

	if len(standaloneLocs) > 0 {
		removed = true
	}

	return removed, output, standaloneLocs
}

func checkForMissingArgs(input []string, locs []int) error {
	flgCount := len(locs)
	argCount := len(input) - flgCount

	if argCount < flgCount { // standalones removed --> bad input
		return &MissingArgumentError{}
	}
	return nil
}

// Compares flag & arg count. If insufficient flags, adds implicit flag, else returns input.
func (fp *FlagParser) handleInsufficientFlags(input []string, locs []int) ([]string, bool) {
	var ret []string
	recurs := false
	flgCount := len(locs)
	argCount := len(input) - flgCount

	if argCount == flgCount { // optimal (with standalones removed)
		return input, false
	}

	for i, v := range input {

		if recurs { // recursive call only needed once
			break
		}
		if i%2 != 0 { // only check flags in even positions (as well as pos 0)
			continue
		}
		fi, exists := fp.GetFlagInfoFromName(v)
		if exists && (!fi.standalone) {
			continue // valid flag
		}

		// else add implicit flag
		prior := make([]string, len(input[0:i]))
		copy(prior, input[0:i])

		ret = append(ret, prior...)
		ret = append(ret, fp.implicitFlag)
		ret = append(ret, input[i:]...)

		latest := fp._updateUserMaps(ret)
		locs2 := fp.GetFlagLocations(latest)

		_, insuff := fp.handleInsufficientFlags(ret, locs2)
		recurs = true

		if !insuff {
			break
		}
	}

	if len(ret) > 0 {
		return ret, true
	}
	return input, false
}

// Trims input to user-determined limits. Attempts to match any extra input
// to implicit flag. If not possible, returns error.
func (fp *FlagParser) handleArgumentLengthAndRemainders(input []string, ufLocations []int) ([]string, error) {
	var lenChecked, suffix []string

	for _, v := range ufLocations {

		arg, remainder := fp.checkAgainstMaxLength(input, v)
		if len(remainder) == 0 {
			if len(arg) > 0 {
				lenChecked = append(lenChecked, input[v], arg)
			} else if len(arg) == 0 {
				lenChecked = append(lenChecked, input[v]) // flags with no arg
			}
		} else {
			lenChecked = append(lenChecked, input[v], arg)
			suffix = append(suffix, remainder)
		}
	}
	if len(suffix) > 0 {
		req := fp.implicitFlagRequired(ufLocations, input)
		if req {
			lenChecked = append(lenChecked, fp.implicitFlag, StringFromSlice(suffix))
			return lenChecked, nil
		}
		return nil, &ExceedMaxLengthError{}
	}
	return lenChecked, nil
}

// Trims arg input to user-determined max length per flag.
// Returns trimmed input plus remainder
func (fp *FlagParser) checkAgainstMaxLength(input []string, flagLocation int) (arg, remainder string) {

	fi, _ := fp.GetFlagInfoFromName(input[flagLocation])

	if fi.standalone {
		return "", ""
	}

	argRunes := []rune(input[flagLocation+1])

	if len(argRunes) <= fi.maxLen {
		arg = string(argRunes)
	} else {
		arg = string(argRunes[:fi.maxLen])
		remainder = strings.Trim(string(argRunes[fi.maxLen:]), " ")
	}

	return arg, remainder
}

// Checks input to see whether implicit flag is missing.
func (fp *FlagParser) implicitFlagRequired(ufLocations []int, input []string) bool {

	_, exists := fp.GetIndexFromFlagValue(user, fp.implicitFlag) // user passed implicit flag
	if exists {
		return false
	}

	required := true
	for _, v := range ufLocations {
		if input[v] == fp.implicitFlag {
			required = false
		}
	}

	return required
}

// Checks args of DateTime flags for literal date strings
// and date relative shorthand ('3d 9m 4y')
func (fp *FlagParser) handleDates(input []string, ufLocations []int) ([]string, error) {
	var yInt, mInt, dInt int
	for _, v := range ufLocations {

		flgInf, _ := fp.GetFlagInfoFromName(input[v])
		if flgInf.flgType != DateTime {
			continue
		}

		noSpaces := strings.ToLower(strings.ReplaceAll(input[v+1], " ", ""))
		mp, literalDateStr, err := getDateMap(noSpaces)
		if err != nil {
			return nil, err
		}
		if literalDateStr {
			return input, nil
		}

		val, ok := mp["y"]
		if ok {
			yInt = val
		}
		val, ok = mp["m"]
		if ok {
			mInt = val
		}
		val, ok = mp["d"]
		if ok {
			dInt = val
		}

		now := fp.NowMoment.Local()
		newNow := now.AddDate(yInt, mInt, dInt)
		yy, mm, dd := newNow.Date()

		zPrefM, zPrefD := "", ""
		if int(mm) < 10 {
			zPrefM = "0"
		}
		if dd < 10 {
			zPrefD = "0"
		}

		input[v+1] = fmt.Sprintf("%v-%v%v-%v%v", yy, zPrefM, int(mm), zPrefD, dd)
	}
	return input, nil
}

// Checks for existence/location of date identifiers ('y', 'm', 'd')
// in '3d1m5y' format. Populates date identifier map with relevant values.
func getDateMap(inputStr string) (mp map[string]int, literalDateStr bool, e error) {
	letterLocs := []int{}
	mp = getEmptyDateMap()
	e = nil
	literalDateStr = true

	for i, v := range []rune(inputStr) {
		if string(v) == "y" || string(v) == "m" || string(v) == "d" {
			letterLocs = append(letterLocs, i)
			literalDateStr = false
		}
	}

	// pupulate date map
	start := 0
	for _, v := range letterLocs {

		dateIdfr := string(rune(inputStr[v]))

		if _, exists := mp[dateIdfr]; exists {
			intPrefix, e := strconv.Atoi(inputStr[start:v]) // number (n) that comes before dateIdfr; e.g. if input = '3m', dateIdfr = 'm' & n = '3'
			if e != nil {
				return nil, literalDateStr, &UnknownDateInputError{}
			}
			mp[dateIdfr] = intPrefix
		}

		start = v + 1
	}
	return mp, literalDateStr, e
}

func getEmptyDateMap() map[string]int {
	mp := make(map[string]int)
	y, m, d := "y", "m", "d"
	mp[y] = 0
	mp[m] = 0
	mp[d] = 0

	return mp
}

// Appends standalone flags to end of input. Since they're
// standalone, placement isn't significant but sequential order is maintained.
func (fp *FlagParser) reassemble(input []string, standalones map[int]string) []string {
	var locs []int

	// can't guarantee order in maps
	// no real difference; just for predictability/testing
	for i := range standalones {
		locs = append(locs, i)
	}
	sort.Ints(locs)

	for _, v := range locs {
		input = append(input, standalones[v])
	}
	return input
}

func (fp *FlagParser) _updateUserMaps(addition []string) int {
	fp.userPassedFlags = append(fp.userPassedFlags, addition)
	fp.setupUserMaps(addition)
	return len(fp.userPassedFlags) - 1
}

func StringFromSlice(sl []string) string {
	bodyStr := ""
	for _, s := range sl {
		bodyStr += s + " "
	}
	return strings.Trim(bodyStr, " ")
}

func StringFromDate(d time.Time) string {
	m := int(d.Month())
	var mStr, dStr string
	if m < 10 {
		mStr = "0" + fmt.Sprint(m)
	} else {
		mStr = fmt.Sprint(m)
	}
	if d.Day() < 10 {
		dStr = "0" + fmt.Sprint(d.Day())
	} else {
		dStr = fmt.Sprint(d.Day())
	}
	final := fmt.Sprintf("%v-%v-%v", d.Year(), mStr, dStr)

	return final
}
