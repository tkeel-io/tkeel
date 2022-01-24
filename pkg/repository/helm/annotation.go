package helm

import (
	"strings"
)

const (
	tKeelPluginEnableKey     = "tkeel.io/enable"
	tKeelPluginDeploymentKey = "tkeel.io/deployment-name"
	tKeelPluginPortKey       = "tkeel.io/plugin-port"
	tKeelPluginTypeTag       = "tkeel.io/tag"
	tKeelPluginVersion       = "tkeel.io/version"

	trueString = "true"
)

func getBoolAnnotationOrDefault(annotations map[string]string, key string, defaultValue bool) bool {
	enabled, ok := annotations[key]
	if !ok {
		return defaultValue
	}
	s := strings.ToLower(enabled)
	// trueString is used to silence a lint error.
	return (s == "y") || (s == "yes") || (s == trueString) || (s == "on") || (s == "1")
}

func getStringAnnotation(annotations map[string]string, key string) string {
	return annotations[key]
}

func getTagAnnotations(annotations map[string]string) string {
	s := strings.ToLower(getStringAnnotation(annotations, tKeelPluginTypeTag))
	if s == "manager" {
		return s
	}
	return "user"
}
