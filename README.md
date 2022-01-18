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

This repository contains the standalone SAST engine used by [Horusec](https://github.com/ZupIT/horusec). 
By now we only have a pattern matching rule implementation, but a semantic analysis is already is being planned.

This is an internal repository of the [Horusec CLI](https://github.com/ZupIT/horusec), so we don't guarantee
compatibility between versions.

### **What is a SAST tool?**

A Static Application Security Testing tool is an automated scanner for security issues in your source code. 
The main goal is to identify, as soon as possible in your development lifecycle, any possible threat to your
infrastructure and your user's data. SAST tools don't actually find vulnerabilities because the tool never executes the
program being analyzed, therefore, you still have to keep testing your applications with more traditional pen testing
and any other tests that you can execute.

## **Usage**

To use this implementation will be needed to create a new engine instance informing the goroutines pool size and the
slice of the extensions that should be analyzed. After the analysis is finished, a slice of findings will be returned.

#### **1. Goroutines Pool**

The pool size informed during instantiation will directly affect memory usage and analysis time. The larger the pool,
the shorter the analysis time, but the greater the amount of memory required.

#### **2. Rule**

Contains all the data needed to identify and report a vulnerability. All rules are defined by a generic interface with
a `Run` function. The idea is that we have several specific implementations of rules, like the one we currently have in
the text package, but each one with it own specific strategy.

#### **3. Finding**

It contains all the possible vulnerabilities found after the analysis, it also has the necessary data to identify and
treat the vulnerability.

### **Example**

```go
    eng := engine.NewEngine(10, ".java")

    rules := []engine.Rule{
        &text.Rule{
            Metadata: engine.Metadata{
                ID:          "HORUSEC-EXAMPLE-1",
                Name:        "Hello World",
                Description: "This is a example of the engine usage",
                Severity:    "HIGH",
                Confidence:  "HIGH",
            },
            Type: text.OrMatch,
            Expressions: []*regexp.Regexp{
                regexp.MustCompile(`System\.out\.println\("Hello World"\);`),
             },
        },
        ...
    }

    findings, err := eng.Run(context.Background(), "path-to-analyze", rules...)
    if err != nil {
        return err
    }

    for _, finding := range findings {
        // do something
    }
```

## **Documentation**

For more information about Horusec, please check out the [**documentation**](https://horusec.io/docs/).

## **Contributing**

If you want to contribute to this repository, access our 
[**Contributing Guide**](https://github.com/ZupIT/horusec-engine/blob/main/CONTRIBUTING.md).

### **Developer Certificate of Origin - DCO**

This is a security layer for the project and for the developers. It is mandatory.

Follow one of these two methods to add DCO to your commits:

**1. Command line**
Follow the steps:
**Step 1:** Configure your local git environment adding the same name and e-mail configured at your GitHub account. 
It helps to sign commits manually during reviews and suggestions.

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

**Step 1:** When the commit changes box opens, manually type or paste your signature in the comment box, see the 
example:

```
Signed-off-by: Name < e-mail address >
```

For this method, your name and e-mail must be the same registered on your GitHub account.

## **License**

[**Apache License 2.0**](https://github.com/ZupIT/horusec-engine/blob/main/LICENSE).

## **Community**

Do you have any question about Horusec? Let's chat in our [**forum**](https://forum.zup.com.br/).

This project exists thanks to all the contributors. You rock! ❤️🚀
