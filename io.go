package xlisp

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/spy16/sabre"
)

// Println is an alias for fmt.Println which ignores the return values.
func Println(args ...interface{}) error {
	_, err := fmt.Println(args...)
	return err
}

// Printf is an alias for fmt.Printf which ignores the return values.
func Printf(format string, args ...interface{}) error {
	_, err := fmt.Printf(format, args...)
	return err
}

// Reads from stdin and returns string
func Read(prompt string) (string, error) {

	reader := bufio.NewReader(os.Stdin)
	fmt.Print(prompt)
	text, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	return text[:len(text)-1], nil
}

func Random(max int) int {
	rand.Seed(time.Now().UnixNano())
	result := rand.Intn(max)
	return result
}

func Shuffle(seq sabre.Seq) sabre.Seq {
	rand.Seed(time.Now().UnixNano())
	list := Realize(seq)
	values := list.Values
	rand.Shuffle(len(list.Values), func(i, j int) {
		values[i], values[j] = values[j], values[i]
	})
	return list
}

func ReadFile(name string) (string, error) {

	content, err := ioutil.ReadFile(name)
	if err != nil {
		return "", err
	}

	return string(content), nil
}

func createShellOutput(out, err string, exit int) *sabre.HashMap {
	return &sabre.HashMap{
		Data: map[sabre.Value]sabre.Value{
			sabre.Keyword("exit"): sabre.Int64(exit),
			sabre.Keyword("out"):  sabre.String(out),
			sabre.Keyword("err"):  sabre.String(err),
		},
	}
}

func Shell(command string) (*sabre.HashMap, error) {

	cmd := exec.Command("bash", "-c", command)
	var cmdout, cmderr bytes.Buffer

	cmd.Stdout = &cmdout
	cmd.Stderr = &cmderr

	err := cmd.Run()
	if exitErr, ok := err.(*exec.ExitError); ok {
		errMsg := strings.TrimSpace(cmderr.String())
		return createShellOutput("", errMsg, exitErr.ExitCode()), nil
	} else if err != nil {
		return &sabre.HashMap{}, err
	}

	output := strings.TrimSpace(cmdout.String())

	return createShellOutput(output, "", 0), nil
}

func splitString(str, sep sabre.String) *sabre.List {
	result := strings.Split(string(str), string(sep))
	values := make([]sabre.Value, 0, len(result))
	for _, v := range result {
		values = append(values, sabre.String(v))
	}
	return &sabre.List{Values: values}
}
