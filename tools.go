//go:build tools
// +build tools

// Package tools tracks tool dependencies that are required by the project
// but not directly imported by application code. This file ensures these
// dependencies are tracked in go.mod.
//
// See: https://github.com/golang/go/wiki/Modules#how-can-i-track-tool-dependencies-for-a-module
package tools

import (
	// Azure Blob Storage SDK - used for file storage in production
	// Implementation will be added in sc-308
	_ "github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
)
