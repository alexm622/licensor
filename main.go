/* License generated by licensor(https://github.com/Marvin9/licensor).

Copyright 2020 mayursinh

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const uniqueIdentifier = "License generated by licensor(https://github.com/Marvin9/licensor)."

type cmdModel struct {
	projectPath string
	// files       []string
	extensions  []string
	license     string
	ignore      []string
	licenseText []byte
	template    map[string]string
}

func logError(msg interface{}) {
	fmt.Printf("\nError: %v\n", msg)
	os.Exit(1)
}

func invalidFlagError(flag string) {
	logError(fmt.Sprintf("invalid use of %v flag.", flag))
}

var commands = []string{
	"-project", "-ext", "-license", "-ignore", "-template",
}

func exists(key string, strs []string) bool {
	for _, str := range strs {
		if str == key {
			return true
		}
	}
	return false
}

func isKeywordCommand(str string) bool {
	return exists(str, commands)
}

var validExtensions = []string{
	"go", "c", "cpp", "js", "html", "css", "rb",
}

func isValidExtension(ext string) bool {
	return exists(ext, validExtensions)
}

// not implementing now
// licensor -files ./main.go ./fixtures/foo.go -license "FREE BSD CLOSE" -ignore *_test.go

// licensor -project ./ -ext go -license "FREE BSD CLOSE" -template '{'owner':'foo', 'year': '2020'}' -ignore *_test.go
func main() {
	mainArgs := os.Args[1:]
	i := 0
	mainArgsLen := len(mainArgs)
	var model cmdModel
	for i < mainArgsLen {
		arg := mainArgs[i]

		switch arg {
		case "-project":
			i++
			if i >= mainArgsLen || isKeywordCommand(mainArgs[i]) {
				invalidFlagError("-project")
			}
			model.projectPath = mainArgs[i]
			i++
		// case "-files":
		case "-ext":
			i++
			for i < mainArgsLen && !isKeywordCommand(mainArgs[i]) {
				model.extensions = append(model.extensions, mainArgs[i])
				i++
			}

		case "-license":
			i++
			if i >= mainArgsLen || isKeywordCommand(mainArgs[i]) {
				invalidFlagError("-license")
			}

			model.license = mainArgs[i]
			i++
		case "-ignore":
			i++
			for i < mainArgsLen && !isKeywordCommand(mainArgs[i]) {
				model.ignore = append(model.ignore, mainArgs[i])
				i++
			}
		case "-template":
			i++
			if i >= mainArgsLen || isKeywordCommand(mainArgs[i]) {
				invalidFlagError("-template")
			}
			json.Unmarshal([]byte(mainArgs[i]), &model.template)
			i++
		default:
			i++
		}
	}

	if len(model.extensions) == 0 {
		logError("You must provide atleast one extension to -ext flag.")
	}
	// STEP 1 ABOVE: MAKE MODEL FROM COMMAND

	// STEP 2: VALIDATION
	/*
		project path must exist & it must be directory
		extensions must be valid and implemented
		license must be valid
	*/

	// project path must exist
	projectDir, err := os.Stat(model.projectPath)
	if err != nil {
		logError(err)
	}

	// project path must be directory
	if !projectDir.IsDir() {
		logError(fmt.Sprintf("%v is not directory.", model.projectPath))
	}

	// extensions must be valid and implemented
	for _, ext := range model.extensions {
		if !isValidExtension(ext) {
			logError(fmt.Sprintf("We do not support %v extension right now.", ext))
		}
	}
	// COMMAND MODEL VALIDATED

	// STEP 3: LOAD LICENSE IN BUFFER
	var licenseText []byte
	_, errPath := os.Stat(model.license)

	licenseFileError := fmt.Sprintf("%v is neither valid path nor valid url", model.license)

	if errPath != nil {
		res, errURL := http.Get(model.license)
		if errURL != nil {
			logError(licenseFileError)
		}
		rd, errURL := ioutil.ReadAll(res.Body)
		if errURL != nil {
			logError(errURL)
		}
		res.Body.Close()
		licenseText = rd
	} else {
		licenseFile, err := os.Open(model.license)
		if err != nil {
			logError(err)
		}
		licenseTextLc, errFile := ioutil.ReadAll(licenseFile)
		if errFile != nil {
			logError(licenseFileError)
		}
		licenseText = licenseTextLc
	}

	templateReg := regexp.MustCompile(`{{[[:alpha:]]+}}`)
	templateMatches := templateReg.FindAll(licenseText, -1)
	for _, match := range templateMatches {
		matchStr := string(match)
		variableName := matchStr[2 : len(matchStr)-2]
		variableValue, ok := model.template[variableName]
		if !ok {
			logError(fmt.Sprintf("%v is not defined in template.", variableName))
		}
		licenseText = bytes.Replace(licenseText, match, []byte(variableValue), 1)
	}

	splitted := bytes.Split(licenseText, []byte("\n"))
	licenseText = []byte("")
	for _, line := range splitted {
		licenseText = append(licenseText, append([]byte("\n "), line...)...)
	}

	// -template "name:foo, organization:bar"
	model.licenseText = licenseText

	model.start()
}

func getExtension(file string) string {
	ext := filepath.Ext(file)
	return strings.TrimPrefix(ext, ".")
}

func (model cmdModel) start() {
	model.iterateDirectory(model.projectPath)
}

func (model cmdModel) iterateDirectory(path string) {
	files, _ := ioutil.ReadDir(path)

	for _, file := range files {
		filename := file.Name()
		fullpath := path + "/" + filename
		if file.IsDir() {
			// IGNORE SPECIFIC DIRECTORY
			model.iterateDirectory(fullpath)
			continue
		}

		// FILE SPOTTED
		// USE FILE ONLY IF IT HAS EXTENSION GIVEN IN COMMAND
		ext := getExtension(filename)
		if exists(ext, model.extensions) {
			// GET FILE CONTENT
			fileContent, err := ioutil.ReadFile(fullpath)
			if err != nil {
				logError(err)
			}

			// FOR golang only: TODO
			commentPrefix := "/* "
			commentPostfix := "*/"

			uniqueHeader := append([]byte(commentPrefix), []byte(uniqueIdentifier)...)
			// CHECK IF LICENSE EXIST
			exist := bytes.Index(fileContent, uniqueHeader)
			if exist != -1 {
				// CHECK IF CURRENT LICENSE IS NOT EQUAL TO PREVIOUS ONE
				endOfComment := bytes.Index(fileContent, []byte(commentPostfix))
				oldLicenseText := bytes.TrimPrefix(fileContent[exist:endOfComment], uniqueHeader)

				null := []byte("")
				reg := regexp.MustCompile(`\s+|\n`)
				t1 := reg.ReplaceAll(oldLicenseText, null)
				t2 := reg.ReplaceAll(model.licenseText, null)
				if bytes.Equal(t1, t2) {
					// IF SAME THEN SKIP
					continue
				} else {
					// REMOVE EXISTING LICENSE
					lastIdx := endOfComment + len(commentPostfix) - 1 + 2
					fileContent = append(fileContent[0:exist], fileContent[lastIdx+1:len(fileContent)]...)
				}
			}

			f, err := os.OpenFile(fullpath, os.O_WRONLY, os.ModePerm)
			if err != nil {
				logError(err)
			}
			defer f.Close()

			// COMMENT OUT LICENSE TEXT
			// ---------------------- template --------------------------
			// commentPrefix uniqueIdentifier
			// license text
			// commentPostfix
			//
			//
			// actual code
			// -----------------------------------------------------------
			f.WriteString(fmt.Sprintf("%v%v\n%v\n%v\n\n", commentPrefix, uniqueIdentifier, string(model.licenseText), commentPostfix))
			f.Write(fileContent)
			fmt.Printf("\nFile updated: %v\n", fullpath)
		}
	}
}

func printModel(m cmdModel) {
	fmt.Printf("\n%v: %v\n", "Project path", m.projectPath)
	fmt.Printf("\n%v: %v\n", "Extension", m.extensions)
	fmt.Printf("\n%v: %v\n", "License", m.license)
	fmt.Printf("\n%v: %v\n", "Ignore", m.ignore)
	fmt.Printf("\n%v:%v\n", "Template", m.template)
}
