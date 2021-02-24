// Copyright (c)  WSO2 Inc. (http://www.wso2.org) All Rights Reserved.
//
// WSO2 Inc. licenses this file to you under the Apache License,
// Version 2.0 (the "License"); you may not use this file except
// in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package utils

import (
	"testing"
)

func TestExtractArchive(t *testing.T) {
	var tmpPath string
	var err error

	// Zip file is available
	tmpPath, err = ExtractArchive("../../test/utils/PizzaShackAPI_1.0.0.zip")

	if err != nil  {
		t.Error("extracting the zip should not return an error")
	}

	if tmpPath == "" {
		t.Error("extracting the zip should not return an empty path")
	}

	// Zip file path is invalid
	tmpPath, err = ExtractArchive("../../invalid_path/PizzaShackAPI_1.0.0.zip")

	if err == nil  {
		t.Error("extracting the zip should return an error for invalid file path use case")
	}

	if tmpPath != "" {
		t.Error("extracting the zip should return an empty path for invalid file path use case")
	}

	// Zip file is invalid
	tmpPath, err = ExtractArchive("../../test/utils/PizzaShackAPI_1.0.0_Invalid.zip")

	if err != nil  {
		t.Error("extracting the zip should return an error for invalid zip file")
	}

	if tmpPath == "" {
		t.Error("extracting the zip should not return an empty path for invalid zip file")
	}
}

func TestCreateZipFileFromProject(t *testing.T) {
	var tmpPath string
	var err error
	var cleanup func()
	skipCleanup := true

	// File is available for zipping
	tmpPath, err, cleanup = CreateZipFileFromProject("../../test/utils/Pizza", skipCleanup)

	if err != nil  {
		t.Error("zipping the file should not return an error for a valid file content")
	}

	if tmpPath == "" {
		t.Error("zipping the zip should not return an empty path for a valid file content")
	}

	if cleanup == nil  {
		t.Error("zipping the zip should return a func for a valid file content")
	}

	// File is not available for zipping
	tmpPath, err, cleanup = CreateZipFileFromProject("../../invalid_path/PizzaShackAPI_1.0.0", skipCleanup)

	if err != nil  {
		t.Error("zipping the file should not return an error for a invalid file content")
	}

	if tmpPath == "" {
		t.Error("zipping the zip should not return an empty path for a invalid file content")
	}

	if cleanup != nil  {
		t.Error("zipping the zip should not return a nil func for a invalid file content")
	}

	// File is available for zipping with other files
	tmpPath, err, cleanup = CreateZipFileFromProject("../../test/utils/Pizza/PizzaShackAPI-1.0.0", skipCleanup)

	if err != nil  {
		t.Error("zipping the file should not return an error for a valid file content")
	}

	if tmpPath == "" {
		t.Error("zipping the zip should not return an empty path for a valid file content")
	}

	if cleanup == nil  {
		t.Error("zipping the zip should return a func for a valid file content")
	}
}

func TestCreateZipFileFromProjectWithCleanup(t *testing.T) {
	var tmpPath string
	var err error
	var cleanup func()
	skipCleanup := false

	// File is available for zipping
	tmpPath, err, cleanup = CreateZipFileFromProject("../../test/utils/Pizza", skipCleanup)

	if err != nil  {
		t.Error("zipping the file should not return an error for a valid file content")
	}

	if tmpPath == "" {
		t.Error("zipping the zip should not return an empty path for a valid file content")
	}

	if cleanup == nil  {
		t.Error("zipping the zip should return a func for a valid file content")
	}

}
