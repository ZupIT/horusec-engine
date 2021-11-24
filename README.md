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

## **Table of contents**
### 1. [**About**](#about)
### 2. [**Usage**](#usage)
>#### 2.1. [**Why does this engine help me?**](#why-does-this-engine-help-me)
>#### 2.2. [**Examples**](#examples)
### 3. [**Documentation**](#documentation)
### 4. [**Contributing**](#contributing)
### 5. [**License**](#license)
### 6. [**Community**](#community)

## **About** 

This repository contains the standalone SAST engine used by Horusec.

Horusec-Engine provides baseline functionality and the basic building blocks for you to build your own SAST tool.

### **What is a SAST tool?**
A Static Application Security Testing tool is an automated scanner for security issues in your source code or binary artifact. The main goal is to identify, as soon as possible in your development lifecycle, any possible threat to your infrastructure and your user's data.
SAST tools don't actually find vulnerabilities because the tool never executes the program being analyzed, therefore, you still have to keep testing your applications with more traditional pen testing and any other tests that you can execute.


### **There are many SAST tools out there, why write my own?**
The main benefit you can get for writing your own is the amount of knowledge about your application you can add to your tool. In the beginning, the off-the-shelf tool will have more rules and more corner cases covered than yours, but with the right amount of dedication to improve and expand the techniques your tool uses, you can easily overcome the regular market tools.

## **Usage**

### **Why does this engine help me?**
The only built-in technique our engine uses is the syntax pattern matching technique, a powerful yet simple technique to uncover the most common mistakes you can leave in your codebase. But, the extensibility of the engine is the main advantage it presents. All the design around our solution was focused on the extensibility and interoperability of techniques in one single analysis.

To achieve that, Horusec-Engine uses three components that allow you to expand the functionality of the engine with new techniques and still with common ground for all of them. They can be extended to suit your needs. Check them out below: 


#### **1. Unit**
A unit is a piece of your code that makes sense to be analyzed as one. So every Unit is a lexical scope, for example, a C++ namespace or a Java Class. The engine will treat all the files and code inside a unit as one thing, and it will only be able to cross-reference anything inside a single unit.
We are working on a complex lexical analysis between units and even a more deeper one inside units, [**if you want to help us, check out our GitHub issues**](https://github.com/ZupIT/horusec-engine/issues).

#### **2. Rule**
The engine won't help you here because you have to provide your own rules. The FOSS version of Horusec's tool has a lot of rules you can use, but this interface is here to encourage you to learn more about how security issues manifest in your favorite language syntax, and therefore how to identify them with your tool.

#### **3. Finding**
The finding is a key part of your tool. You can extract useful insights from the source code analysis.
The structure is focused on simplicity, but we are working to implement it following the [**SARIF**](https://github.com/oasis-tcs/sarif-spec) specification, so you can have complete control of where you import your data.


### **Examples**

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

If you want to contribute to this repository, access our [**Contributing Guide**](https://github.com/ZupIT/horusec-engine/blob/main/CONTRIBUTING.md). 


### **Developer Certificate of Origin - DCO**

 This is a security layer for the project and for the developers. It is mandatory.
 
 Follow one of these two methods to add DCO to your commits:
 
**1. Command line**
 Follow the steps: 
 **Step 1:** Configure your local git environment adding the same name and e-mail configured at your GitHub account. It helps to sign commits manually during reviews and suggestions.

 ```
git config --global user.name “Name”
git config --global user.email “email@domain.com.br”
```
**Step 2:** Add the Signed-off-by line with the `'-s'` flag in the git commit command:

```
$ git commit -s -m "This is my commit message"
```

**2. GitHub website**
You can also manually sign your commits during GitHub reviews and suggestions, follow the steps below: 

**Step 1:** When the commit changes box opens, manually type or paste your signature in the comment box, see the example:

```
Signed-off-by: Name < e-mail address >
```

For this method, your name and e-mail must be the same registered on your GitHub account.

## **License**
[**Apache License 2.0**](https://github.com/ZupIT/horusec-engine/blob/main/LICENSE).

## **Community**
Do you have any question about Horusec? Let's chat in our [**forum**](https://forum.zup.com.br/).


This project exists thanks to all the contributors. You rock! ❤️🚀
