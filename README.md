<p align="center" margin="20 0"><a href="https://horusec.io/">
    <img src="https://github.com/ZupIT/horusec-devkit/blob/main/assets/horusec_logo.png?raw=true" 
            alt="logo_header" width="65%" style="max-width:100%;"/></a></p>

<p align="center">
    <a href="https://github.com/ZupIT/horusec-engine/pulse" alt="activity">
        <img src="https://img.shields.io/github/commit-activity/m/ZupIT/horusec-engine?label=activity"/></a>
    <a href="https://github.com/ZupIT/horusec-engine/graphs/contributors" alt="contributors">
        <img src="https://img.shields.io/github/contributors/ZupIT/horusec-engine?label=contributors"/></a>
    <a href="https://github.com/ZupIT/horusec-engine/actions/workflows/lint.yml" alt="lint">
        <img src="https://img.shields.io/github/workflow/status/ZupIT/horusec-engine/Lint?label=lint"/></a>
    <a href="https://github.com/ZupIT/horusec-engine/actions/workflows/test.yml" alt="test">
        <img src="https://img.shields.io/github/workflow/status/ZupIT/horusec-engine/Test?label=test"/></a>
    <a href="https://github.com/ZupIT/horusec-engine/actions/workflows/security.yml" alt="security">
        <img src="https://img.shields.io/github/workflow/status/ZupIT/horusec-engine/Security?label=security"/></a>
    <a href="https://github.com/ZupIT/horusec-engine/actions/workflows/coverage.yml" alt="coverage">
        <img src="https://img.shields.io/github/workflow/status/ZupIT/horusec-engine/Coverage?label=coverage"/></a>
    <a href="https://opensource.org/licenses/Apache-2.0" alt="license">
        <img src="https://img.shields.io/badge/license-Apache%202-blue"/></a>

# **Horusec Engine**

This repository contains the standalone SAST engine used by Horusec.

Horusec-Engine provides baseline functionality and the basic building blocks for you to build your own SAST tool.

## **What is a SAST tool?**
A Static Application Security Testing tool is an automated scanner for security issues in your source code or binary artifact. The main goal is to identify, as soon as possible in your development lifecycle, any possible threat to your infrastructure and your user's data.
SAST tools don't actually find vulnerabilities because the tool never executes the program being analyzed, therefore, you still have to keep testing your applications with more traditional pen testing and any other tests that you can execute.


### **There are many SAST tools out there, why write my own?**
The main benefit you can get for writing your own is the amount of knowledge about your application you can add to your tool. In the beginning, the off-the-shelf tool will have more rules and more corner cases covered than yours, but with the right amount of dedication to improve and expand the techniques your tool uses, you can easily overcome the regular market tools.

### **Why does this engine help me?**
The only built-in technique our engine uses is the syntax pattern matching technique, a powerful yet simple technique to uncover the most common mistakes you can leave in your codebase. But, the extensibility of the engine is the main advantage it presents. All the design around our solution was focused on the extensibility and interoperability of techniques in one single analysis.

To achieve that, Horusec-Engine uses three simple and expressive components that can be easily extended to suit your needs and it allows you to expand the functionality of the engine with new techniques, while still having a common ground for all of them. 
See it below: 


### **1. Unit**
A unit is a piece of your code that makes sense to be analyzed as one. So every Unit is a lexical scope, for example, a C++ namespace or a Java Class. The engine will treat all the files and code inside a unit as one thing, and it will only be able to cross-reference anything inside a single unit.
We are working on a complex lexical analysis between units and even a more deeper one inside units, [**if you want to help us, check out our GitHub issues**](https://github.com/ZupIT/horusec-engine/issues).

### **2. Rule**
The engine won't help you here because you have to provide your own rules. The FOSS version of Horusec's tool has a lot of rules you can use, but this interface is here to encourage you to learn more about how security issues manifest in your favorite language syntax, and therefore how to identify them with your tool.

### **3. Finding**
The finding is a key part of your tool. You can actually extract useful insights from the source code analysis.
The struct is focused on simplicity, but we are working to implement it following the[**SARIF**](https://github.com/oasis-tcs/sarif-spec) specification, so you can have complete control of where you import your data.


## **Examples**

A simple analysis of an inmemory string:

```go
	var exampleGoFile = `package version

import (
	"github.com/ZupIT/horusec/development-kit/pkg/utils/logger"
	"github.com/spf13/cobra"
)

type IVersion interface {
	CreateCobraCmd() *cobra.Command
}

type Version struct {
}

func NewVersionCommand() IVersion {
	return &Version{}
}

func (v *Version) CreateCobraCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "version",
		Short:   "Actual version installed of the horusec",
		Example: "horusec version",
		RunE: func(cmd *cobra.Command, args []string) error {
			logger.LogPrint(cmd.Short + " is: ")
			return nil
		},
	}
}
`

	var textUnit TextUnit = TextUnit{}
	goTextFile, err := NewTextFile("example/cmd/version.go", []byte(exampleGoFile))

	if err != nil {
		t.Error(err)
	}

	textUnit.Files = append(textUnit.Files, goTextFile)

	var regularMatchRule TextRule = TextRule{}
	regularMatchRule.Type = Regular
	regularMatchRule.Expressions = append(regularMatchRule.Expressions, regexp.MustCompile(`cmd\.Short`))

	rules := []engine.Rule{regularMatchRule}
	program := []engine.Unit{textUnit}

	findings := engine.Run(program, rules)

	for _, finding := range findings {
		t.Log(finding.SourceLocation)
	}
```
## **Documentation**

For more information about Horusec, please check out the [**documentation**](https://horusec.io/docs/).


## **Contributing**

If you want to contribute to this repository, access our [**Contributing Guide**](https://github.com/ZupIT/charlescd/blob/main/CONTRIBUTING.md). 
And if you want to know more about Horusec, check out some of our other projects:


- [**Admin**](https://github.com/ZupIT/horusec-admin)
- [**Charts**](https://github.com/ZupIT/charlescd/tree/main/circle-matcher)
- [**Devkit**](https://github.com/ZupIT/horusec-devkit)
- [**Jenkins**](https://github.com/ZupIT/horusec-jenkins-sharedlib)
- [**Operator**](https://github.com/ZupIT/horusec-operator)
- [**Platform**](https://github.com/ZupIT/horusec-platform)
- [**VSCode plugin**](https://github.com/ZupIT/horusec-vscode-plugin)
- [**Kotlin**](https://github.com/ZupIT/horusec-tree-sitter-kotlin)
- [**Vulnerabilities**](https://github.com/ZupIT/horusec-examples-vulnerabilities)

## **Community**
Feel free to reach out to us at:

- [**GitHub Issues**](https://github.com/ZupIT/horusec-devkit/issues)
- [**Zup Open Source Forum**](https://forum.zup.com.br)


This project exists thanks to all the contributors. You rock! ❤️🚀