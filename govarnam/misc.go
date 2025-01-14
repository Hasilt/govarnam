package govarnam

/**
 * govarnam - An Indian language transliteration library
 * Copyright Subin Siby <mail at subinsb (.) com>, 2021
 * Licensed under AGPL-3.0-only. See LICENSE.txt
 */

import (
	"fmt"
	"io/fs"
	"path/filepath"
)

// GetAllSchemePaths get available IDs' location as a string array
func GetAllSchemePaths() ([]string, error) {
	vstsDir, err := FindVSTDir()

	if err != nil {
		return nil, err
	}

	var schemeIDs []string

	filepath.WalkDir(vstsDir, func(s string, d fs.DirEntry, e error) error {
		if e != nil {
			return e
		}
		if filepath.Ext(d.Name()) == ".vst" {
			schemeIDs = append(schemeIDs, s)
		}
		return nil
	})

	return schemeIDs, nil
}

// GetAllSchemeDetails get information of all schemes available
func GetAllSchemeDetails() ([]SchemeDetails, error) {
	schemePaths, err := GetAllSchemePaths()

	if err != nil {
		return nil, err
	}

	var schemeDetails []SchemeDetails

	for _, vstPath := range schemePaths {
		fmt.Println(vstPath)

		varnam := Varnam{}
		varnam.InitVST(vstPath)

		schemeDetails = append(schemeDetails, varnam.SchemeDetails)

		varnam.Close()
	}

	return schemeDetails, nil
}
