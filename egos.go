// Copyright (c) 2015 eel3
//
// This software is provided 'as-is', without any express or implied
// warranty. In no event will the authors be held liable for any damages
// arising from the use of this software.
//
// Permission is granted to anyone to use this software for any purpose,
// including commercial applications, and to alter it and redistribute it
// freely, subject to the following restrictions:
//
//     1. The origin of this software must not be misrepresented; you must not
//     claim that you wrote the original software. If you use this software
//     in a product, an acknowledgment in the product documentation would be
//     appreciated but is not required.
//
//     2. Altered source versions must be plainly marked as such, and must not be
//     misrepresented as being the original software.
//
//     3. This notice may not be removed or altered from any source
//     distribution.

package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/format"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	exit_success = iota // EXIT_SUCCESS
	exit_failure        // EXIT_FAILURE
)

// Script template for no option.
const template = `// generated by egos(1)
package main

%s

func main() {
%s
}
`

// Script template for n/p option.
const template_n = `// generated by egos(1)
package main

import (
	"bufio"
	"fmt"
	"os"
)

%s

func fn(line string) {
%s%s
}

func doIt(in *os.File) {
	scanner := bufio.NewScanner(in)
	for scanner.Scan() {
		fn(scanner.Text())
	}
}

func main() {
	argc := len(os.Args)
	if argc <= 1 {
		doIt(os.Stdin)
	} else {
		for i := 1; i < argc; i++ {
			infile := os.Args[i]
			if infile == "-" {
				doIt(os.Stdin)
			} else {
				in, err := os.Open(infile)
				if err != nil {
					fmt.Fprintln(os.Stderr, err)
				} else {
					doIt(in)
					in.Close()
				}
			}
		}
	}
}
`

// Generate import statements.
func generateImport(i_opt string, pred func(string) bool) string {
	tmp := strings.Split(strings.Trim(i_opt, " ;"), ";")
	var pkgs []string
	for _, v := range tmp {
		v = strings.Trim(v, " \t")
		if pred(v) {
			pkgs = append(pkgs, v)
		}
	}
	if len(pkgs) == 0 {
		return ""
	}
	return "import (\n" + strings.Join(pkgs, "\n") + "\n)"
}

// Generate golang script.
func generateScript(orig_src, i_opt string, n_opt, p_opt bool) string {
	src := strings.Trim(orig_src, "\t\n\v\f\r ")
	if n_opt || p_opt {
		is := generateImport(i_opt, func(s string) bool {
			switch s {
			case ``, `"bufio"`, "`bufio`", `"fmt"`, "`fmt`", `"os"`, "`os`":
				return false
			default:
				return true
			}
		})
		var appendage string
		if n_opt {
			appendage = ""
		} else {
			appendage = "\nfmt.Println(line)"
		}
		return fmt.Sprintf(template_n, is, src, appendage)
	} else {
		is := generateImport(i_opt, func(s string) bool { return s != "" })
		return fmt.Sprintf(template, is, src)
	}
}

// Format golang script.
func formatScript(src string) ([]byte, error) {
	return format.Source([]byte(src))
}

// Run golang script.
func runScript(script []byte, args []string) int {
	dir, err := ioutil.TempDir("", "egos-")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return exit_failure
	}
	defer os.RemoveAll(dir)
	// TODO handle signal and remove dir.

	scriptFile := filepath.Join(dir, "script.go")
	execFile := filepath.Join(dir, "script.exe")

	err = ioutil.WriteFile(scriptFile, script, os.ModePerm)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return exit_failure
	}

	var stderr bytes.Buffer
	cmd := exec.Command("go", "build", "-o", execFile, scriptFile)
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		// TODO process build error messages.
		s, _ := stderr.ReadString(0)
		fmt.Fprintln(os.Stderr, s)
		return exit_failure
	}

	cmd = exec.Command(execFile, args...)
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	err = cmd.Run()
	if err != nil {
		return exit_failure
	}

	return exit_success
}

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s [-d] <-i packages> [-n|-p] 'script' [file ...]\n", filepath.Base(os.Args[0]))
		flag.PrintDefaults()
	}
	d_opt := flag.Bool("d", false, "print the compiled script but do not run it")
	i_opt := flag.String("i", "", "specify import packages")
	n_opt := flag.Bool("n", false, "assume 'read line' loop around your script")
	p_opt := flag.Bool("p", false, "assume loop like -n but print line also like sed")
	flag.Parse()

	if flag.NArg() <= 0 {
		flag.Usage()
		os.Exit(exit_failure)
	}

	if *n_opt && *p_opt {
		flag.Usage()
		os.Exit(exit_failure)
	}

	script, err := formatScript(generateScript(flag.Arg(0), *i_opt, *n_opt, *p_opt))
	if err != nil {
		fmt.Fprintln(os.Stderr, "syntax error")
		os.Exit(exit_failure)
	}

	retcode := exit_success
	if *d_opt {
		fmt.Printf("%s", script)
	} else {
		retcode = runScript(script, flag.Args()[1:])
	}

	os.Exit(retcode)
}
