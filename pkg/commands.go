package pkg

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"

	"github.com/TylerBrock/colorjson"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/fatih/color"
)

var TMP_FILE_NAME = ".aws-secret-tmp.json"

type updateListCmd struct{ list []awsResult }
type editorClosed struct {
	err         error
	beforeValue string
}
type secretUpdated struct{ err error }
type editorResult struct {
	error   bool
	changed bool
	msg     string
	value   string
}

func openEditor(secretName string, loadSecret bool) tea.Cmd {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vim"
	}
	beforeValue := ""
	if loadSecret {
		cmd := exec.Command("aws", "secretsmanager", "get-secret-value", "--secret-id", secretName)
		var outb, errb bytes.Buffer
		cmd.Stdout = &outb
		cmd.Stderr = &errb
		err := cmd.Run()
		if err != nil {
      os.WriteFile(TMP_FILE_NAME, []byte(""), 0644);
      fmt.Printf("%s", errb.String());
		} else {
      var result awsResult
      err = json.Unmarshal(outb.Bytes(), &result)
      if err != nil {
        os.WriteFile(TMP_FILE_NAME, outb.Bytes(), 0644);
        beforeValue = outb.String();
      } else {
        var s interface{}
        json.Unmarshal([]byte(result.SecretString), &s)
        pp, _ := json.MarshalIndent(s, "", "  ")
        os.WriteFile(TMP_FILE_NAME, pp, 0644)
        beforeValue = string(pp)
      }
    }
	}
	// var newSecret string
	cmd := exec.Command(editor, TMP_FILE_NAME)
	return tea.ExecProcess(cmd, func(err error) tea.Msg {
		return editorClosed{err: err, beforeValue: beforeValue}
	})
}

func updateSecretCmd(secretName string, value string) tea.Cmd {
	return func() tea.Msg {
		cmd := exec.Command("aws", "secretsmanager", "update-secret", "--secret-id", secretName, "--secret-string", value)
		err := cmd.Run()
		if err != nil {
			return secretUpdated{err: err}
		}
		return secretUpdated{err: nil}
	}
}

func checkSecretValid(m model) tea.Cmd {
	return func() tea.Msg {
		content, err := ioutil.ReadFile(TMP_FILE_NAME)
		if err != nil {
			return tea.Quit
		}
		var testJson interface{}
		err = json.Unmarshal(content, &testJson)
		if err != nil {
			msg := fmt.Sprintf(
				`Invalid JSON: 
--------------------
%s------------------------------
Press enter to edit again`, content)
			return editorResult{error: true, msg: msg}
		}
		newSecret, _ := ioutil.ReadFile(TMP_FILE_NAME)
    os.Remove(TMP_FILE_NAME)
		if string(newSecret) == m.beforeValue {
			return editorResult{error: false, changed: false, msg: "No changes made, exiting"}
		}
		f := colorjson.NewFormatter()
		f.Indent = 2
		f.KeyColor = color.New(color.FgBlue)
		var newSecretObj interface{}
		json.Unmarshal([]byte(newSecret), &newSecretObj)
		newSecretPrettyStr, _ := f.Marshal(newSecretObj)
		strContent := string(newSecretPrettyStr)
		msg := fmt.Sprintf(
			`Update secret '%s' with new value
------------------------------
%s
------------------------------
Press enter to update, or q to quit`, m.selectedSecret, strContent)
		return editorResult{error: false, msg: msg, changed: true, value: string(newSecret)}
	}
}

func getSecrets() tea.Msg {
	cmd := exec.Command("aws", "secretsmanager", "list-secrets")
	var outb, errb bytes.Buffer
	cmd.Stdout = &outb
	cmd.Stderr = &errb
	err := cmd.Run()
	if err != nil {
		fmt.Printf("%s", errb.String())
		os.Exit(1)
	}
	var secretList AWSListResult
	json.Unmarshal(outb.Bytes(), &secretList)
	return updateListCmd{list: secretList.SecretList}
}
