package ftp

type config struct {
	logger Logger

	randomTempPath       bool
	randomTempPathPrefix string
	tempPath             string
}
