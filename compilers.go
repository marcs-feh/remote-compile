package main

// A LanguageBuilder creates the command to build/run a particular `path`
type LanguageBuilder interface {
	Name() string
	Ext() string // File extension, without the .
	Build(path string, args ...string) []string
	Run(path string, args ...string) []string
	Test(path string, args ...string) []string
}

type OdinBuilder struct {}

var DefaultOdinBuilder OdinBuilder

func (b OdinBuilder) Name() string {
	return "Odin"
}

func (b OdinBuilder) Ext() string {
	return "odin"
}

func (b OdinBuilder) Build(path string, args ...string) []string {
	cmd := []string{
		"odin", "build", path, "-file",
	}
	cmd = append(cmd, args...)
	return cmd
}

func (b OdinBuilder) Run(path string, args ...string) []string {
	cmd := []string{
		path,
	}
	cmd = append(cmd, args...)
	return cmd
}

func (b OdinBuilder) Test(path string, args ...string) []string {
	cmd := []string{
		"odin", "test", path, "-file",
	}
	cmd = append(cmd, args...)
	return cmd
}

