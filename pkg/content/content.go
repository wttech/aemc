package content

import (
	"github.com/spf13/cast"
	"github.com/wttech/aemc/pkg/base"
)

type Opts struct {
	baseOpts *base.Opts

	FilesDotContent      []string
	FilesDeleted         []PathRule
	FilesFlattened       []string
	PropertiesSkipped    []PathRule
	MixinTypesSkipped    []PathRule
	NamespacesSkipped    bool
	ParentsBackupEnabled bool
	ParentsBackupSuffix  string
}

type PathRule struct {
	Patterns      []string
	ExcludedPaths []string
	IncludedPaths []string
}

func NewOpts(baseOpts *base.Opts) *Opts {
	cv := baseOpts.Config().Values()

	return &Opts{
		baseOpts: baseOpts,

		FilesDotContent:      cv.GetStringSlice("content.files_dot_content"),
		FilesDeleted:         determinePathRules(cv.Get("content.files_deleted")),
		FilesFlattened:       cv.GetStringSlice("content.files_flattened"),
		PropertiesSkipped:    determinePathRules(cv.Get("content.properties_skipped")),
		MixinTypesSkipped:    determinePathRules(cv.Get("content.mixin_types_skipped")),
		NamespacesSkipped:    cv.GetBool("content.namespaces_skipped"),
		ParentsBackupEnabled: cv.GetBool("content.parents_backup_enabled"),
		ParentsBackupSuffix:  cv.GetString("content.parents_backup_suffix"),
	}
}

func determinePathRules(values interface{}) []PathRule {
	var result []PathRule
	for _, value := range cast.ToSlice(values) {
		result = append(result, PathRule{
			Patterns:      determineStringSlice(value, "patterns"),
			ExcludedPaths: determineStringSlice(value, "excluded_paths"),
			IncludedPaths: determineStringSlice(value, "included_paths"),
		})
	}
	return result
}

func determineStringSlice(values interface{}, key string) []string {
	var result []string
	value := cast.ToStringMap(values)[key]
	if value != nil {
		result = cast.ToStringSlice(value)
	}
	return result
}