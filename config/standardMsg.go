package config

var (
	InternalServerError     int = 500
	Success                 int = 200
	ResolveInfoError        int = 300
	DuplicatedUsername      int = 400
	DuplicatedCourse int=401
	CaptchaError            int = 402
	LanguageNotSupported    int = 602
	BuildFailed             int = 603
	FormatError             int = 604
	NoTokenFound            int = 700
	TokenExpired            int = 701
	TokenInvaild            int = 702
	WrongUsernameOrPassword int = 703
	PermissionDenied int=800
)

const (
	TypeTeacher int = 1
	TypeStudent int = 2
)