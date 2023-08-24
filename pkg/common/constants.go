package common

const (
	AppId          = "aem-compose"
	AppName        = "AEM Compose"
	MainDir        = "aem"
	HomeDirName    = "home"
	HomeDir        = MainDir + "/" + HomeDirName
	VarDirName     = "var"
	VarDir         = HomeDir + "/" + VarDirName
	ConfigDirName  = "etc"
	ConfigDir      = HomeDir + "/" + ConfigDirName
	LogDirName     = "log"
	LogDir         = VarDir + "/" + LogDirName
	LogFile        = LogDir + "/aem.log"
	CacheDirName   = "cache"
	CacheDir       = VarDir + "/" + CacheDirName
	ToolDirName    = "opt"
	ToolDir        = HomeDir + "/" + ToolDirName
	LibDirName     = "lib"
	LibDir         = HomeDir + "/" + LibDirName
	TmpDirName     = "tmp"
	TmpDir         = HomeDir + "/" + TmpDirName
	DefaultDirName = "default"
	DefaultDir     = MainDir + "/" + DefaultDirName
)

const (
	STDIn           = "STDIN"
	STDOut          = "STDOUT"
	OutputValueAll  = "ALL"
	OutputValueNone = "NONE"
)
