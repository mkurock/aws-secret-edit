package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"

	"github.com/TylerBrock/colorjson"
	"github.com/fatih/color"
)

type AWSResult struct {
	ARN           string   `json:"ARN"`
	Name          string   `json:"Name"`
	VersionId     string   `json:"VersionId"`
	SecretString  string   `json:"SecretString"`
	CreatedDate   string   `json:"CreatedDate"`
	VersionStages []string `json:"VersionStages"`
}

var TMP_FILE_NAME = ".aws-secret-tmp.json"

func main() {
	secretName := os.Args[1]
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vim"
	}
	fmt.Println("Secret name: ", secretName)
	cmd := exec.Command("aws", "secretsmanager", "get-secret-value", "--secret-id", secretName)
	var outb, errb bytes.Buffer
	cmd.Stdout = &outb
	cmd.Stderr = &errb
	err := cmd.Run()
	if err != nil {
		fmt.Printf("%s", errb.String())
		os.Exit(1)
	}
	var result AWSResult
	json.Unmarshal(outb.Bytes(), &result)
	var s interface{}
	json.Unmarshal([]byte(result.SecretString), &s)
	pp, _ := json.MarshalIndent(s, "", "  ")
	os.WriteFile(TMP_FILE_NAME, pp, 0644)
	var newSecret string
	for {
		cmd = exec.Command(editor, TMP_FILE_NAME)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err = cmd.Run()
		if err != nil {
			fmt.Println("Error opening editor")
			panic(err)
		}
		content, err := ioutil.ReadFile(TMP_FILE_NAME)
		if err != nil {
			fmt.Println("Cannot read tmp file")
		}
		var testJson interface{}
		err = json.Unmarshal(content, &testJson)
		if err != nil {
			fmt.Println("Invalid JSON:\n", err)
			fmt.Println("----------------------------------------")
			fmt.Println(string(content))
			fmt.Println("----------------------------------------")
			fmt.Println("Press enter to try again")
			fmt.Scanln()
			continue
		}

		newSecret = string(content)
		os.Remove(TMP_FILE_NAME)
		break
	}

	f := colorjson.NewFormatter()
	f.Indent = 2
  f.KeyColor = color.New(color.FgBlue)
	var newSecretObj interface{}
	json.Unmarshal([]byte(newSecret), &newSecretObj)
	newSecretPrettyStr, _ := f.Marshal(newSecretObj)
	fmt.Println("New Secret value: ")
	fmt.Println("----------------------------------------")
	fmt.Println(string(newSecretPrettyStr))
	fmt.Println("----------------------------------------")
	fmt.Printf("Press enter to update secret %s", secretName)
	fmt.Scanln()
	fmt.Println("Updating secret...")
	cmd = exec.Command("aws", "secretsmanager", "update-secret", "--secret-id", secretName, "--secret-string", newSecret)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
}
