package flagParser

type UserArgsContainsUnknownFlag struct{}

func (u *UserArgsContainsUnknownFlag) Error() string {
	return "unknown flag in user-provided args"
}

type ExceedMaxLengthError struct{}

func (e *ExceedMaxLengthError) Error() string {
	return "maximum argument length exceeded"
}

type UnknownDateInputError struct{}

func (u *UnknownDateInputError) Error() string {
	return "unknown elements in date argument"
}

type FlagMapperInitialisationError struct{}

func (f *FlagMapperInitialisationError) Error() string {
	return "flag mapper initialisation failed"
}

type MissingArgumentError struct{}

func (m *MissingArgumentError) Error() string {
	return "flag missing argument"
}

type MalformedDateRangeError struct{}

func (m *MalformedDateRangeError) Error() string {
	return "malformed date range provided"
}

type DateRangeNotAllowedError struct{}

func (d *DateRangeNotAllowedError) Error() string {
	return "date range not allowed with flag provided"
}
