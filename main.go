package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
)

type AWSResult struct {
	ARN           string   `json:"ARN"`
	Name          string   `json:"Name"`
	VersionId     string   `json:"VersionId"`
	SecretString  string   `json:"SecretString"`
	CreatedDate   string   `json:"CreatedDate"`
	VersionStages []string `json:"VersionStages"`
}

func main() {
	secretName := os.Args[1]
	fmt.Println("Secret name: ", secretName)
	cmd := exec.Command("aws", "secretsmanager", "get-secret-value", "--secret-id", secretName)
	var outb, errb bytes.Buffer
	cmd.Stdout = &outb
	cmd.Stderr = &errb
  err := cmd.Run()
  if err != nil {
    panic(err)
  }
	var result AWSResult
	json.Unmarshal(outb.Bytes(), &result)
	var s interface{}
	json.Unmarshal([]byte(result.SecretString), &s)
	pp, _ := json.MarshalIndent(s, "", "  ")
	os.WriteFile(".aws-secret-tmp", pp, 0644)
  var newSecret string
  for {
    cmd = exec.Command("vim", ".aws-secret-tmp")
    cmd.Stdin = os.Stdin
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr
    err = cmd.Run()
    if err != nil {
      panic(err)
    }
    content, err := ioutil.ReadFile(".aws-secret-tmp")
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
    os.Remove(".aws-secret-tmp")
    break
  }

	fmt.Println("New Secret value: ")
	fmt.Println(newSecret)
	fmt.Printf("Press enter to update secret %s", secretName)
	fmt.Scanln()
	fmt.Println("Updating secret...")
	cmd = exec.Command("aws", "secretsmanager", "update-secret", "--secret-id", secretName, "--secret-string", newSecret)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
}
