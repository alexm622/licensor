package utils

const LicensorYAML = "licensor.yml"

const UniqueIdentifier = "License generated by licensor(https://github.com/Marvin9/licensor)."

const (
	PROJECT  = "-project"
	EXT      = "-ext"
	LICENSE  = "-license"
	IGNORE   = "-ignore"
	TEMPLATE = "-template"
)

var Commands = []string{
	PROJECT, EXT, LICENSE, IGNORE, TEMPLATE,
}

var SupportedFileExtensions = []string{
	"go", "c", "cpp", "js", "css",
}
