# **BUILD**

## **Table of contents** 
### 1. [**About**](#about)
### 2. [**Environment**](#environment)
### 3. [**Development**](#development)
>#### 3.1. [**Style Guide**](#style-guide)
>#### 3.2. [**Tests**](#tests)
>#### 3.3. [**Security**](#security)

## **About**

The **BUILD.md** is a file to check the environment and build specifications of **horusec-engine** project.


## **Environment**

- [**Golang**](https://go.dev/dl/):  ^1.17.X
- [**GNU Make**](https://www.gnu.org/software/make/): ^4.2.X

## **Development**

Horusec-Engine is the standalone SAST engine used by [Horusec-CLI](https://github.com/ZupIT/horusec).

You can use it as a package by running the following command in your Golang project:

```bash
go get github.com/ZupIT/horusec-engine
```

### **Style Guide**

The project uses the [**golangci-lint**](https://golangci-lint.run) tool as a Go linter aggregator.
You can perform the lint check via the `make` command available in each microservice:

```bash
make lint
```

To perform the indentation and removal of unused code automatically, just run the following command:

```bash
make fmt
```

The project also has a pattern of dependency imports, and the command below organizes your code in the pattern defined by the Horusec team, this command must be run in each microservice:

```bash
make fix-imports
```

All project files must have the [**license header**](./copyright.txt). You can check if all files are in agreement by running the following command in project root:

```bash
make license
```

If it is necessary to add the license in any file, run the command below to insert it in all files that do not have it:

```bash
make license-fix
```

### **Tests**

The unit tests were written with the [**standard package**](https://pkg.go.dev/testing) and some mock and assert snippets, we used the [**testify**](https://github.com/stretchr/testify). You can run the tests using the command below:

```bash
make test
```

To check test coverage, run command below:

```bash
make coverage
```

### **Security**

We use the latest version of [**Horusec-CLI**](https://github.com/ZupIT/horusec) to maintain the security of our source code. Through the command below, you can perform this verification in the project:

```bash
make security
```
