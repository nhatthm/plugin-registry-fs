# File system installer for Plugin Registry

[![GitHub Releases](https://img.shields.io/github/v/release/nhatthm/plugin-registry-fs)](https://github.com/nhatthm/plugin-registry-fs/releases/latest)
[![Build Status](https://github.com/nhatthm/plugin-registry-fs/actions/workflows/test.yaml/badge.svg)](https://github.com/nhatthm/plugin-registry-fs/actions/workflows/test.yaml)
[![codecov](https://codecov.io/gh/nhatthm/plugin-registry-fs/branch/master/graph/badge.svg?token=eTdAgDE2vR)](https://codecov.io/gh/nhatthm/plugin-registry-fs)
[![Go Report Card](https://goreportcard.com/badge/github.com/nhatthm/plugin-registry-fs)](https://goreportcard.com/report/github.com/nhatthm/plugin-registry-fs)
[![GoDevDoc](https://img.shields.io/badge/dev-doc-00ADD8?logo=go)](https://pkg.go.dev/github.com/nhatthm/plugin-registry-fs)
[![Donate](https://img.shields.io/badge/Donate-PayPal-green.svg)](https://www.paypal.com/donate/?hosted_button_id=PJZSGJN57TDJY)

A file system installer for [plugin-registry](https://github.com/nhatthm/plugin-registry)

## Prerequisites

- `Go >= 1.15`

## Install

```bash
go get github.com/nhatthm/plugin-registry-fs
```

## Usage

Import the library while bootstrapping the application (see the [examples](#examples))

This installer supports installing:
- A binary file
- A folder
- An archive (`.tar.gz`, `.gz.` or `zip`)

The source must be in this format:

```
./my-project/
├── .plugin.registry.yaml
└── my-plugin/
    └── (plugin files) 
```

For example, if source is an archive, it should be:

```
./my-project/
├── .plugin.registry.yaml
└── my-plugin-1.0.0-darwin-amd64.tar.gz 
```

## Examples

```go
package mypackage

import (
	"context"

	registry "github.com/nhatthm/plugin-registry"
	_ "github.com/nhatthm/plugin-registry-fs" // Add file system installer.
)

var defaultRegistry = mustCreateRegistry()

func mustCreateRegistry() registry.Registry {
	r, err := createRegistry()
	if err != nil {
		panic(err)
	}

	return r
}

func createRegistry() (registry.Registry, error) {
	return registry.NewRegistry("~/plugins")
}

func installPlugin(source string) error {
	return defaultRegistry.Install(context.Background(), source)
}

```

## Donation

If this project help you reduce time to develop, you can give me a cup of coffee :)

### Paypal donation

[![paypal](https://www.paypalobjects.com/en_US/i/btn/btn_donateCC_LG.gif)](https://www.paypal.com/donate/?hosted_button_id=PJZSGJN57TDJY)

&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;or scan this

<img src="https://user-images.githubusercontent.com/1154587/113494222-ad8cb200-94e6-11eb-9ef3-eb883ada222a.png" width="147px" />
